package vacancy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"
	"github.com/gin-gonic/gin"
)

// TODO: Обработка ошибок

func GetVacancies(c *gin.Context) {
	var (
		response []models.QueryVacancies
		wg       sync.WaitGroup
	)
	platform := c.Query("platform")
	limit, _ := strconv.Atoi(c.Query("count"))
	queryList := tools.UniqueSlice(c.QueryArray("text"))
	wg.Add(len(queryList))
	for _, s := range queryList {
		go func(s string) {
			defer wg.Done()
			data, _ := GetQueryVacancies(s, platform, limit)
			response = append(response, data)
		}(s)
	}
	wg.Wait()
	c.JSON(200, response)
}

func GetQueryVacancies(query string, platform string, limit int) (models.QueryVacancies, error) {
	var QueryVacancies []models.Vacancy
	switch platform {
	case "trudvsem":
		vacancies, _ := GetVacanciesFromTrudvsem(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	case "superjob":
		vacancies, _ := GetVacanciesFromSuperjob(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	case "geekjob":
		vacancies, _ := GetVacanciesFromGeekjob(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	default:
		vacancies, _ := GetVacanciesFromAllPlatforms(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	}
	return models.QueryVacancies{Query: query, VacancyList: QueryVacancies}, nil
}

func GetVacanciesFromAllPlatforms(query string, limit int) ([]models.Vacancy, error) {
	var (
		QueryVacancies []models.Vacancy
		wg             sync.WaitGroup
	)
	const platformsCount = 3
	wg.Add(platformsCount)
	go func(query string) {
		defer wg.Done()
		vacancies, _ := GetVacanciesFromTrudvsem(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)

	}(query)
	go func(query string) {
		defer wg.Done()
		vacancies, _ := GetVacanciesFromSuperjob(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	}(query)
	go func(query string) {
		defer wg.Done()
		vacancies, _ := GetVacanciesFromGeekjob(query, limit)
		QueryVacancies = append(QueryVacancies, vacancies...)
	}(query)
	wg.Wait()
	return QueryVacancies, nil
}

func DecondeJsonResponse(url string, headers map[string]string, dataStruct interface{}, method string) (data interface{}, statusCode int) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(dataStruct)
	if err != nil {
		fmt.Println(err, resp.StatusCode)
		panic(err)
	}
	return dataStruct, resp.StatusCode
}
