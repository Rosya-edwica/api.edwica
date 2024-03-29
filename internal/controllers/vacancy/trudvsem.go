package vacancy

import (
	"fmt"
	"github.com/go-faster/errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
)

const trudvsemUrl = "http://opendata.trudvsem.ru/api/v1/vacancies"

// CollectVacanciesFromTrudvsem парсинг сайта занятости trudvsem.ru (Работа России)
func CollectVacanciesFromTrudvsem(query models.VacancyQuery) ([]models.Vacancy, error) {
	// Пытаемся получить JSON в структуре Trudvsem
	var requestUrl string
	if query.RegionCode != 0 {
		requestUrl = fmt.Sprintf("%s/region/%d/?text=%s", trudvsemUrl, query.RegionCode, url.PathEscape(query.Query))
	} else {
		requestUrl = fmt.Sprintf("%s?text=%s", trudvsemUrl, url.PathEscape(query.Query))
	}
	response, err := DecodeJsonResponse(requestUrl, nil, &Trudvsem{}, "GET")
	if response == nil || err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("На trudvsem.ru ничего не нашлось по запросу: '%s'", query))
	}
	data := response.(*Trudvsem)

	// Распарсим структуру Trudvsem в список вакансий
	var vacancies []models.Vacancy
	for i, item := range data.Results.Vacancies {
		if i >= query.Limit && query.Limit > 0 {
			break
		}
		// Пытаемся распарсить город из полного адреса
		var cityAddress string
		if len(item.Vacancy.Addressses.Address) > 0 {
			cityAddress = item.Vacancy.Addressses.Address[0].Location
		}
		vacancies = append(vacancies, models.Vacancy{
			Platform:   "trudvsem",
			Id:         item.Vacancy.Id,
			Name:       item.Vacancy.Name,
			Url:        item.Vacancy.Url,
			SalaryFrom: item.Vacancy.SalaryFrom,
			SalaryTo:   item.Vacancy.SalaryTo,
			Currency:   tools.FilterCurrency(item.Vacancy.Currency),
			City:       parseTrudvsemCity(cityAddress),
		})
	}
	return vacancies, nil
}

// parseTrudvsemCity с помощью регулярок пытаемся вытащить город из полного адреса
func parseTrudvsemCity(text string) string {
	var (
		reCity    = regexp.MustCompile(`г. .*?|г .*?,|,.*?район`)
		reSubCity = regexp.MustCompile(`г. |, |г |,`)
	)
	city := reCity.FindString(strings.ToLower(text))
	city = reSubCity.ReplaceAllString(city, "")
	if city == "" {
		return text
	}
	return strings.Title(city)
}
