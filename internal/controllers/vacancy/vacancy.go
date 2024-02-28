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

	limit, _ := strconv.Atoi(c.Query("count"))
	regionCode, _ := strconv.Atoi(c.Query("region"))
	params := models.VacancyParams{
		Texts:      tools.UniqueSlice(c.QueryArray("text")),
		City:       "",
		Platform:   c.Query("platform"),
		RegionCode: regionCode,
		Limit:      limit,
	}

	errc := make(chan error, len(params.Texts)) // Канал с ошибками
	wg.Add(len(params.Texts))                   // Создаем горутины, количество которых будет равно количеству параметров text
	for _, query := range params.Texts {
		vacancyQuery := models.VacancyQuery{
			Query:      query,
			City:       params.City,
			Platform:   params.Platform,
			RegionCode: params.RegionCode,
			Limit:      params.Limit,
		}
		go func(query string) {
			defer wg.Done()
			data, err := GetQueryVacancies(vacancyQuery)
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
func GetQueryVacancies(query models.VacancyQuery) (result models.QueryVacancies, err error) {
	var vacancies []models.Vacancy
	switch query.Platform {
	case "trudvsem":
		vacancies, err = CollectVacanciesFromTrudvsem(query)
	case "superjob":
		vacancies, err = CollectVacanciesFromSuperjob(query)
	case "geekjob":
		vacancies, err = CollectVacanciesFromGeekjob(query)
	default:
		vacancies, err = CollectVacancies(query)
	}
	if len(vacancies) == 0 {
		return models.QueryVacancies{Query: query.Query, VacancyList: []models.Vacancy{}}, err
	}
	return models.QueryVacancies{Query: query.Query, VacancyList: vacancies}, err
}

// CollectVacancies Собирает вакансии со всех источников в один список
func CollectVacancies(query models.VacancyQuery) ([]models.Vacancy, error) {
	var (
		vacancies []models.Vacancy
		wg        sync.WaitGroup
	)
	wg.Add(platformsCount)
	errc := make(chan error, platformsCount)
	go func(query models.VacancyQuery) {
		defer wg.Done()
		v, err := CollectVacanciesFromTrudvsem(query)
		if err != nil {
			errc <- err
		} else {
			vacancies = append(vacancies, v...)
		}
	}(query)
	go func(query models.VacancyQuery) {
		defer wg.Done()
		v, err := CollectVacanciesFromSuperjob(query)
		if err != nil {
			errc <- err
		} else {
			vacancies = append(vacancies, v...)
		}
	}(query)
	go func(query models.VacancyQuery) {
		defer wg.Done()
		v, err := CollectVacanciesFromGeekjob(query)
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
