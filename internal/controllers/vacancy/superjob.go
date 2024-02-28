package vacancy

import (
	"fmt"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/go-faster/errors"
	"net/url"
	"strconv"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
)

// CollectVacanciesFromSuperjob главная функция, которая вызывает вспомогательную функцию для парсинга вакансий superjob
func CollectVacanciesFromSuperjob(query models.VacancyQuery) ([]models.Vacancy, error) {
	return collectVacanciesFromSuperjobWithRetries(query, 3)

}

// collectVacanciesFromSuperjobWithRetries вспомогательная функция, которая будет несколько раз пробовать обновить
// токен superjob.
func collectVacanciesFromSuperjobWithRetries(query models.VacancyQuery, retries int) ([]models.Vacancy, error) {
	for i := 0; i < retries; i++ {
		vacancies, err := collectVacanciesFromSuperjob(query)
		if err != nil {
			err = updateSuperjobToken()
			if err != nil {
				return nil, errors.Wrap(err, "Не смогли подключиться к superjob. Переподключение не помогло.")
			}
		} else {
			return vacancies, nil
		}
	}
	return nil, nil
}

// collectVacanciesFromSuperjob логика парсинга superjob
func collectVacanciesFromSuperjob(query models.VacancyQuery) ([]models.Vacancy, error) {
	requestUrl := buildSuperjobRequestUrl(query)
	headers, err := GetMapSuperjobHeaders()
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка при создании запроса к superjob")
	}
	resp, err := DecodeJsonResponse(requestUrl, headers, &Superjob{}, "GET")
	if resp == nil {
		return nil, errors.Wrap(err, fmt.Sprintf("На superjob.ru ничего не нашлось по запросу: '%s'", query))
	}
	if err != nil {
		return nil, errors.Wrap(err, "Проблемы с superjob (обновление токена не помогает)")
	}
	data := resp.(*Superjob)
	var vacancies []models.Vacancy
	for _, item := range data.Items {
		vacancies = append(vacancies, models.Vacancy{
			Platform:   "superjob",
			Id:         strconv.Itoa(item.Id),
			Name:       item.Name,
			City:       item.City.Name,
			Url:        item.Url,
			Skills:     []string{},
			Currency:   tools.FilterCurrency(item.Currency),
			SalaryFrom: item.SalaryFrom,
			SalaryTo:   item.SalaryTo,
			Company:    item.Company,
		})
	}
	return vacancies, nil

}

// buildSuperjobRequestUrl собираем ссылку вместе с необходимыми параметрами для успешного GET-запроса к API Superjob
func buildSuperjobRequestUrl(query models.VacancyQuery) string {
	params := url.Values{}
	params.Add("count", strconv.Itoa(query.Limit))
	params.Add("keywords[0][srws]", "1")          // Поиск по названию
	params.Add("keywords[0][skwc]", "particular") // Поиск точной фразы
	params.Add("keywords[0][keys]", query.Query)  // Запрос
	if query.City != "0" {
		params.Add("town", query.City)
	}
	return "https://api.superjob.ru/2.0/vacancies/?" + params.Encode()
}

// buildSuperjobUpdateTokenUrl собираем ссылку вместе с необходимыми параметрами для успешного обновления токена API Superjob
func buildSuperjobUpdateTokenUrl() (string, error) {
	headers, err := GetSuperjobHeaders()
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("refresh_token", headers.Token)
	params.Add("client_id", headers.ClientId)
	params.Add("client_secret", headers.SecretId)
	return "https://api.superjob.ru/2.0/oauth2/refresh_token?" + params.Encode(), nil

}

// updateSuperjobToken логика обновления токена
func updateSuperjobToken() error {
	logger.Log.Info("Пробуем обновить токен superjob...")
	refreshUrl, err := buildSuperjobUpdateTokenUrl()
	if err != nil {
		return errors.Wrap(err, "Не удалось создать ссылку для обновления токена")
	}
	resp, err := DecodeJsonResponse(refreshUrl, nil, &SuperjobToken{}, "POST")
	if err != nil {
		return errors.Wrap(err, "Не смогли распарсить токен в структуру SuperjobToken")
	}
	data := resp.(*SuperjobToken)
	err = setNewSuperjobTokenToConfig(data.AccessToken)
	if err != nil {
		return errors.Wrap(err, "Ошибка при обновлении токена в конфиг-файле")
	}
	return nil
}
