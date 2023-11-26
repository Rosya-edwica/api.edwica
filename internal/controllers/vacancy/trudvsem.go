package vacancy

import (
	"errors"
	"net/url"
	"regexp"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
)

func GetVacanciesFromTrudvsem(query string, limit int) ([]models.Vacancy, error) {
	link := "http://opendata.trudvsem.ru/api/v1/vacancies/?text=" + url.PathEscape(query) // trudvsem не принимает слова с пробелом
	resp, _ := DecondeJsonResponse(link, nil, &Trudvsem{}, "GET")
	if resp == nil {
		return nil, errors.New("не удалось получить данные trudvsem для запроса: " + query)
	}
	data := resp.(*Trudvsem)

	var vacancies []models.Vacancy
	for i, item := range data.Results.Vacancies {
		if i >= limit && limit > 0 {
			break
		}
		vacancies = append(vacancies, models.Vacancy{
			Id:         item.Vacancy.Id,
			Name:       item.Vacancy.Name,
			Url:        item.Vacancy.Url,
			Platform:   "trudvsem",
			Currency:   tools.FilterCurrency(item.Vacancy.Currency),
			City:       getTrudvsemCity(item.Vacancy.Addressses.Address[0].Location),
			Skills:     []string{},
			SalaryFrom: item.Vacancy.SalaryFrom,
			SalaryTo:   item.Vacancy.SalaryTo,
		})
	}

	return vacancies, nil
}

func getTrudvsemCity(text string) string {
	var (
		reCity    = regexp.MustCompile(`г. .*?`)
		reSubCity = regexp.MustCompile(`г. `)
	)
	city := reCity.FindString(text)
	city = reSubCity.ReplaceAllString(city, "")
	return city
}
