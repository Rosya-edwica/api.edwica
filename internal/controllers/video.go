package controllers

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/video"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"

	"github.com/gin-gonic/gin"
)

func GetVideos(c *gin.Context) {
	var (
		response          []models.QueryVideos
		notFoundedQueries []string
		selectErrors      []error
		wg                sync.WaitGroup
	)
	db := database.GetDB()
	r := video.NewRepository(db)
	queryList := c.QueryArray("text")
	limit, _ := strconv.Atoi(c.Query("count"))
	uniqueQueryList := tools.UniqueSlice(queryList)
	wg.Add(len(uniqueQueryList))

	for _, query := range uniqueQueryList {
		go func(query string) {
			defer wg.Done()
			videos, err := r.GetByQuery(query, limit)
			if err != nil {
				selectErrors = append(selectErrors, err)
				return
			}
			if len(videos) > 0 {
				response = append(response, models.QueryVideos{
					Query:     query,
					VideoList: videos,
				})
			} else {
				notFoundedQueries = append(notFoundedQueries, query)
			}

		}(query)
	}
	wg.Wait()
	// Метод для поиска в ютубе
	fmt.Println("Not founded: ", notFoundedQueries)
	if response == nil {
		c.JSON(207, "Not found")
	} else {
		c.JSON(200, response)
	}
}

// func GetVideosFromYoutube()
