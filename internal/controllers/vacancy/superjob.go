package vacancy

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
)

func GetVacanciesFromSuperjob(query string, limit int) ([]models.Vacancy, error) {
	url := createSuperjobRequestUrl(query, limit)
	headers, err := GetMapSuperjobHeaders()
	if err != nil {
		return nil, err
	}
	resp, status := DecondeJsonResponse(url, headers, &Superjob{}, "GET")
	if status != 200 {
		logger.Log.Warning("Пытаемся обновить токен SuperJob")
		err = UpdateSuperjobToken()
		if err != nil {
			logger.Log.Error("Не смогли обновить токен! " + err.Error())
		}
		logger.Log.Info("Пытаемся снова выполнить запрос с новым токеном")
		return GetVacanciesFromSuperjob(query, limit)
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
func createSuperjobRequestUrl(query string, limit int) string {
	params := url.Values{}
	params.Add("count", strconv.Itoa(limit))
	params.Add("keywords[0][srws]", "1")          // Поиск по названию
	params.Add("keywords[0][skwc]", "particular") // Поиск точной фразы
	params.Add("keywords[0][keys]", query)        // Запрос
	return "https://api.superjob.ru/2.0/vacancies/?" + params.Encode()
}

func createSuperjobUpdateTokenUrl() (string, error) {
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

func UpdateSuperjobToken() error {
	refreshUrl, err := createSuperjobUpdateTokenUrl()
	if err != nil {
		return err
	}
	resp, status := DecondeJsonResponse(refreshUrl, nil, &SuperjobToken{}, "POST")
	if status != 200 {
		return errors.New("cant connect to superjob")
	}
	data := resp.(*SuperjobToken)
	setNewSuperjobTokenToConfig(data.AccessToken)
	return nil
}
