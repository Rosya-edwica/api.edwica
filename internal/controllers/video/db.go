// Здесь мы проверяем, запрашивали ли раньше видосы по таким запросам. Если да, то они должны храниться у нас в базе истории запросов, если нет, значит мы должны вернуть список ненайденных видео

package video

import (
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/database/video"
	"github.com/Rosya-edwica/api.edwica/internal/models"
)

func GetVideosFromDB(queryList []string, limit int, r *video.Repository) (response []models.QueryVideos, notFounded []string, errors []error) {
	wg := sync.WaitGroup{}
	wg.Add(len(queryList))
	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()
			videos, err := r.GetByQuery(query, limit)
			if err != nil {
				errors = append(errors, err)
				return
			}
			if len(videos) > 0 {
				response = append(response, models.QueryVideos{
					Query:     query,
					VideoList: videos,
				})
				VideoCache[query] = videos
			} else {
				notFounded = append(notFounded, query)
			}

		}(query)
	}
	wg.Wait()
	return
}
