package vacancy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Rosya-edwica/api.edwica/internal/models"
	"github.com/Rosya-edwica/api.edwica/tools"
	"github.com/go-faster/errors"
	"github.com/gocolly/colly"
)

func GetVacanciesFromGeekjob(query string, limit int) ([]models.Vacancy, error) {
	url := "https://geekjob.ru/json/find/vacancy?qs=" + query
	resp, _ := DecondeJsonResponse(url, nil, &GeekJob{}, "GET")
	data := resp.(*GeekJob)
	var vacancies []models.Vacancy
	var wg sync.WaitGroup
	var countVacancies int
	const defaultLimit = 3

	if limit > 0 && len(data.Items) >= limit {
		countVacancies = limit
	} else if limit == 0 && len(data.Items) >= defaultLimit {
		countVacancies = defaultLimit
	} else {
		countVacancies = len(data.Items)
	}
	fmt.Println(countVacancies)
	wg.Add(countVacancies)

	for _, item := range data.Items[:countVacancies] {
		go func(item GeekJobItem) {
			defer wg.Done()
			vacancy, _ := ParseGeekjobVacancyById(item.Id)
			vacancies = append(vacancies, vacancy)
		}(item)
	}
	wg.Wait()
	return vacancies, nil
}

func ParseGeekjobVacancyById(id string) (models.Vacancy, error) {
	var vacancy models.Vacancy
	url := "https://geekjob.ru/vacancy/" + id
	html, err := GetHTMLBody(url)
	if err != nil {
		return models.Vacancy{}, errors.Wrap(err, "html content geekjob vacancy by url: "+url)
	}
	vacancy.Name = getTitle(html)
	if vacancy.Name == "" {
		return models.Vacancy{}, nil
	}
	vacancy.Url = url
	vacancy.Id = id
	vacancy.Platform = "geekjob"

	vacancy.City = html.ChildText("div.location")
	salary := getSalary(html)
	vacancy.SalaryFrom = salary.From
	vacancy.SalaryTo = salary.To
	vacancy.Currency = tools.FilterCurrency(salary.Currency)
	vacancy.Skills = getSkills(html)

	return vacancy, nil
}

func GetHTMLBody(url string) (*colly.HTMLElement, error) {
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
