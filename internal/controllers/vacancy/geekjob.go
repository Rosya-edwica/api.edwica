package vacancy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/pkg/tools"
	"github.com/go-faster/errors"
	"github.com/gocolly/colly"
)

const geekjobAPIUrl = "https://geekjob.ru/json/find/vacancy?qs="
const geekjobUrl = "https://geekjob.ru/vacancy/"

// CollectVacanciesFromGeekjob парсинг сайта занятости geekjob.ru (тут только IT вакансии)
func CollectVacanciesFromGeekjob(query models.VacancyQuery) ([]models.Vacancy, error) {
	// Пытаемся получить JSON в структуре GeekJob
	response, err := DecodeJsonResponse(geekjobAPIUrl+query.Query, nil, &GeekJob{}, "GET")
	if response == nil || err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("На geekjob.ru ничего не нашлось по запросу: '%s'", query))
	}
	data := response.(*GeekJob)

	var (
		countVacancies int
		vacancies      []models.Vacancy
		wg             sync.WaitGroup
	)

	// Пытаемся определить сколько вакансий нам нужно распарсить
	if query.Limit > 0 && len(data.Items) >= query.Limit {
		countVacancies = query.Limit
	} else if query.Limit == 0 && len(data.Items) >= defaultLimit {
		countVacancies = defaultLimit
	} else {
		countVacancies = len(data.Items)
	}

	// Запускаем горутины, которые по отдельности будут парсить вакансии с geekjob
	wg.Add(countVacancies)
	errc := make(chan error)
	for _, item := range data.Items[:countVacancies] {
		go func(item GeekJobItem) {
			defer wg.Done()
			vacancy, err := parseGeekjobSiteByVacancyId(item.Id)
			if err != nil {
				errc <- err
			} else {
				vacancies = append(vacancies, vacancy)
			}
		}(item)
	}
	wg.Wait()
	close(errc)
	// Если что-то смогли спарсить, значит отправляем все, что есть
	if len(vacancies) > 0 {
		return vacancies, nil
	}
	// Если ничего не смогли спарсить, значит где-то была ошибка, которую можно показать пользователю
	for err := range errc {
		return nil, errors.Wrap(err, "Ошибка при парсинге вакансий с geekjob")
	}
	// Ошибок не было и вакансий тоже
	return nil, nil
}

// Парсим HTML сайта по id вакансии
func parseGeekjobSiteByVacancyId(id string) (vacancy models.Vacancy, err error) {
	vacancy.Id = id
	vacancy.Url = geekjobUrl + vacancy.Id
	// Пытаемся получить HTML для парсинга с помощью colly
	html, err := getBodyHTML(vacancy.Url)
	if err != nil {
		return models.Vacancy{}, errors.Wrap(err, fmt.Sprintf("Не удалось получить HTML страницы: %s", vacancy.Url))
	}
	vacancy.Name = getTitle(html)
	if vacancy.Name == "" {
		// Проверяем смогли ли мы получить текст с HTML
		return models.Vacancy{}, errors.Wrap(err, fmt.Sprintf("Ошибка при чтении HTML страницы: %s", vacancy.Url))
	}

	vacancy.Platform = "geekjob"
	salary := getSalary(html)
	vacancy.SalaryFrom = salary.From
	vacancy.SalaryTo = salary.To
	vacancy.Currency = tools.FilterCurrency(salary.Currency)
	vacancy.Skills = getSkills(html)
	vacancy.City = html.ChildText("div.location")
	return vacancy, nil
}

func getBodyHTML(url string) (*colly.HTMLElement, error) {
	c := colly.NewCollector()

	var body *colly.HTMLElement
	c.OnHTML("body", func(h *colly.HTMLElement) {
		body = h
	})
	err := c.Visit(url)
	if err != nil {
		return nil, errors.Wrap(err, "colly visit:"+url)
	}
	return body, nil
}

func getSkills(html *colly.HTMLElement) []string {
	tags := strings.Split(html.ChildText("div.tags"), "•")[1:]
	return removeAreasFromSkills(tags)
}
func getTitle(html *colly.HTMLElement) (title string) {
	title = html.ChildText("h1")
	return
}

func getSalary(html *colly.HTMLElement) (salary models.Salary) {
	var salaryText string
	reDigits := regexp.MustCompile(`\d+`)
	reCurrency := regexp.MustCompile(`$|€|₽`)
	html.ForEach("span.salary", func(i int, h *colly.HTMLElement) {
		if i == 0 {
			salaryText = strings.ReplaceAll(h.Text, " ", "")
		}
	})
	currency := reCurrency.FindString(salaryText)
	if strings.Contains(salaryText, "от") && strings.Contains(salaryText, "до") {
		digits := reDigits.FindAllString(salaryText, 2)
		from, _ := strconv.Atoi(digits[0])
		to, _ := strconv.Atoi(digits[1])
		salary = models.Salary{
			From:     from,
			To:       to,
			Currency: currency,
		}

	} else if strings.Contains(salaryText, "от") {
		digit := reDigits.FindString(salaryText)
		from, _ := strconv.Atoi(digit)
		salary = models.Salary{
			From:     from,
			Currency: currency,
		}
	} else {
		digit := reDigits.FindString(salaryText)
		to, _ := strconv.Atoi(digit)
		salary = models.Salary{
			To:       to,
			Currency: currency,
		}
	}
	return salary
}

func removeAreasFromSkills(skills []string) (updated []string) {
	for _, item := range skills {
		if !checkSkillIsArea(item) {
			updated = append(updated, item)
		}
	}
	return
}

// checkSkillIsArea Проверяем, что не забираем профобласть вместе с навыками
func checkSkillIsArea(skill string) bool {
	areas := []string{
		"Торговля и общепит",
		"СМИ, Медиа и индустрия развлечений",
		"Образование",
		"Заказная разработка",
		"Производство",
		"Промышленность",
		"Логистика и транспорт",
		"Медицина и фармацевтика",
		"Телекоммуникации",
		"Строительство и недвижимость",
		"Банковская и страховая сфера",
		"Наука",
		"Сельское хозяйство",
		"Консалтинг, профессиональные услуги",
		"Культура и искусство",
		"Государственные проекты",
	}
	for _, area := range areas {
		if strings.TrimSpace(area) == strings.TrimSpace(skill) {
			return true
		}
	}
	return false
}
