package vacancy

import (
	"regexp"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"
)

func GetVacanciesFromTrudvsem(query string, limit int) ([]models.Vacancy, error) {
	url := "http://opendata.trudvsem.ru/api/v1/vacancies/?text=" + query
	resp, status := DecondeJsonResponse(url, nil, &Trudvsem{}, "GET")
	if status != 200 {
		panic(url + "STATUS " + string(status))
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
