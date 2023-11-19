package video

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/database/video"
	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"
	"github.com/go-faster/errors"

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
	for _, i := range notFoundedQueries {
		GetVideosFromYoutube(i)
		break
	}
	if response == nil {
		c.JSON(207, "Not found")
	} else {
		c.JSON(200, response)
	}
}

func GetVideosFromYoutube(query string) {
	url := "https://www.youtube.com/results?search_query=" + query
	html, err := getHTML(url)
	if err != nil {
		panic(err)
	}
	readHTMLContent(html)

}

func getHTML(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "page connect:"+url)
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", errors.Wrap(err, "reading HTML content:"+url)
	}
	return string(content), nil
}

func readHTMLContent(content string) string {
	file, err := os.Create("name.txt")
	if err != nil {
		panic(err)
	}
	file.Write([]byte(content))
	file.Close()
	var (
		reScripts    = regexp.MustCompile(`<script.*?</script>`)
		reSubScripts = regexp.MustCompile(`<script.*? =|;</script>`)
	)

	scripts := reScripts.FindAllString(content, -1)
	data := reSubScripts.ReplaceAllString(scripts[5], "")
	fmt.Println("JSON", data)

	return ""
}
