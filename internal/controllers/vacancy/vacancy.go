package vacancy

import (
	"encoding/json"
	"fmt"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-faster/errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
	"github.com/gin-gonic/gin"
)

const (
	defaultLimit   = 3
	platformsCount = 3
)

// GetVacancies Обработчик /vacancies
func GetVacancies(c *gin.Context) {
	var (
		response []models.QueryVacancies // Ответ обработчика
		wg       sync.WaitGroup
	)
	platform := c.Query("platform")                      // С какого сайта спарсить вакансии?
	limit, _ := strconv.Atoi(c.Query("count"))           // Сколько максимум вакансий отдать?
	queryList := tools.UniqueSlice(c.QueryArray("text")) // Что искать на сайтах занятости?

	errc := make(chan error, len(queryList)) // Канал с ошибками
	wg.Add(len(queryList))                   // Создаем горутины, количество которых будет равно количеству параметров text
	for _, query := range queryList {
		go func(query string) {
			defer wg.Done()
			data, err := GetQueryVacancies(query, platform, limit)
			if err != nil {
				fmt.Println("main err", err)
				errc <- errors.Wrap(err, fmt.Sprintf("Для '%s' произошла ошибка", query))
			} else {
				response = append(response, data)
			}
		}(query)
	}
	wg.Wait()
	close(errc)

	// Если что-то смогли получить, показываем пользователю положительный результат
	if len(response) != 0 {
		c.JSON(200, response)
		return
	}
	// Если нечего показать, значит нужно проверить происходили ли какие-то ошибки
	for err := range errc {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}

// GetQueryVacancies "Маршрутизатор", который определяет с какого сайта нужно парсить данные
func GetQueryVacancies(query string, platform string, limit int) (result models.QueryVacancies, err error) {
	var vacancies []models.Vacancy
	switch platform {
	case "trudvsem":
		vacancies, err = CollectVacanciesFromTrudvsem(query, limit)
	case "superjob":
		vacancies, err = CollectVacanciesFromSuperjob(query, limit)
	case "geekjob":
		vacancies, err = CollectVacanciesFromGeekjob(query, limit)
	default:
		vacancies, err = CollectVacancies(query, limit)
	}
	if len(vacancies) == 0 {
		return models.QueryVacancies{Query: query, VacancyList: []models.Vacancy{}}, err
	}
	return models.QueryVacancies{Query: query, VacancyList: vacancies}, err
}

// CollectVacancies Собирает вакансии со всех источников в один список
func CollectVacancies(query string, limit int) ([]models.Vacancy, error) {
	var (
		vacancies []models.Vacancy
		wg        sync.WaitGroup
	)
	wg.Add(platformsCount)
	errc := make(chan error, platformsCount)
	go func(query string) {
		defer wg.Done()
		v, err := CollectVacanciesFromTrudvsem(query, limit)
		if err != nil {
			errc <- err
		} else {
			vacancies = append(vacancies, v...)
		}
	}(query)
	go func(query string) {
		defer wg.Done()
		v, err := CollectVacanciesFromSuperjob(query, limit)
		if err != nil {
			errc <- err
		} else {
			vacancies = append(vacancies, v...)
		}
	}(query)
	go func(query string) {
		defer wg.Done()
		v, err := CollectVacanciesFromGeekjob(query, limit)
		if err != nil {
			errc <- err
		} else {
			vacancies = append(vacancies, v...)
		}
	}(query)
	wg.Wait()
	close(errc)
	if len(vacancies) > 0 {
		return vacancies, nil
	} else {
		for err := range errc {
			return nil, errors.Wrap(err, "Не удалось получить вакансии")
		}
	}
	return nil, nil
}

// DecodeJsonResponse старается распарсить любой JSON в структуру dataStruct.
// dataStruct - &структуры, в которую нужно распарсить JSON.
// Возвращает заполненную структуру и ошибку
func DecodeJsonResponse(url string, headers map[string]string, dataStruct interface{}, method string) (data interface{}, err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка при создании запроса: "+url)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := client.Do(request)
	if err != nil || response.StatusCode != 200 {
		return nil, errors.Wrap(err, "Ошибка при отправке запроса: "+url)
	}

	err = json.NewDecoder(response.Body).Decode(dataStruct)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("decodeJsonResponse by url:%s err:%s", url, err))
		return nil, errors.Wrap(err, "Ошибка при наполнении структуры")
	}
	err = response.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "Не удалось закрыть тело ответа")
	}
	return dataStruct, nil
}
