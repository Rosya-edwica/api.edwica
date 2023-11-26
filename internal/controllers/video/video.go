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

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/video"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
	"github.com/gin-gonic/gin"
)

const DefaultLimit = 3

func GetVideos(c *gin.Context) {
	db := database.GetDB()
	r := video.NewRepository(db)
	queryList := tools.UniqueSlice(c.QueryArray("text"))
	limit, err := strconv.Atoi(c.Query("count"))
	if err != nil || limit == 0 {
		limit = DefaultLimit
	}
	response, notFounded, _ := GetVideosFromDB(queryList, limit, r)

	if len(notFounded) > 0 {
		newVideos, _ := GetUndiscoveredVideos(notFounded, limit)
		for _, v := range newVideos {
			done, err := r.SaveVideos(v)
			logger.Log.Info(fmt.Sprintf("controllers.video.db: videos %s saving=%v err:%s", v.Query, done, err))
		}
		response = append(response, newVideos...)
	}
	if response == nil {
		c.JSON(207, "Not found")
	} else {
		c.JSON(200, response)
	}
}
