package vacancy

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"
)

// FIXME: Panics
func GetVacanciesFromSuperjob(query string, limit int) ([]models.Vacancy, error) {
	url := createSuperjobRequestUrl(query, limit)
	headers, err := GetMapSuperjobHeaders()
	if err != nil {
		return nil, err
	}
	resp, status := DecondeJsonResponse(url, headers, &Superjob{}, "GET")
	if status != 200 {
		err = UpdateSuperjobToken()
		if err != nil {
			panic("не обновили токен: " + err.Error())
		}
		fmt.Println("Пробуем снова")
		return GetVacanciesFromSuperjob(query, limit)
		// panic(url + "STATUS " + string(status))
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
