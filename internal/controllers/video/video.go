/*
Поиск видео:
1. Получаем список запросов, по которым нужны видосы
2. Ищем для этих запросов видосы в истории запросов БД
3. Если запрос в истории не существует и видосов по нему нет, то подбираем видосы напрямую из ютуб
4. Сохраняем видосы для нового запроса в БД, чтобы в следующий раз не обращаться к Ютубу и не парсить его ответ

Минусы:
1. Историю запросов нужно переодически обновлять, т.к они могут ссылаться на удаленные или неактуальные видео
*/

package video

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/video"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
	"github.com/gin-gonic/gin"
)

const DefaultLimit = 3

var VideoCache = map[string][]models.Video{}
var VideoCacheMutex = sync.RWMutex{}

func GetVideos(c *gin.Context) {
	var response []models.QueryVideos
	db := database.GetDB()
	r := video.NewRepository(db)
	queryList, limit := valideVideoParams(c)
	cacheResponse, notFounded := checkNewQueriesInCache(tools.UniqueSlice(queryList))
	response, notFounded, _ = GetVideosFromDB(notFounded, limit, r)

	if len(notFounded) > 0 {
		newVideos, _ := GetUndiscoveredVideosByAPI(notFounded, limit)
		for _, v := range newVideos {
			done, err := r.SaveVideos(v)
			logger.Log.Info(fmt.Sprintf("controllers.video.db: videos %s saving=%v err:%s", v.Query, done, err))
		}
		response = append(response, newVideos...)
	}

	response = append(response, cacheResponse...)
	if response == nil {
		c.JSON(207, "Not found")
	} else {
		c.JSON(200, response)
	}
}

func valideVideoParams(c *gin.Context) (queryList []string, limit int) {
	queryList = c.QueryArray("text")
	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	return
}

func checkNewQueriesInCache(items []string) (cacheResponse []models.QueryVideos, notFoundedInCache []string) {
	for _, query := range items {
		VideoCacheMutex.RLock()
		if val, ok := VideoCache[query]; ok {
			cacheResponse = append(cacheResponse, models.QueryVideos{Query: query, VideoList: val})
		} else {
			notFoundedInCache = append(notFoundedInCache, query)
		}
		VideoCacheMutex.RUnlock()
	}
	return cacheResponse, notFoundedInCache
}
