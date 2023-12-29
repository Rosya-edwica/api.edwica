package models

type Vacancy struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	City       string   `json:"city"`
	Platform   string   `json:"platform"`
	Url        string   `json:"url"`
	SalaryFrom int      `json:"salary_from"`
	SalaryTo   int      `json:"salary_to"`
	Currency   string   `json:"currency"`
	Company    string   `json:"company"`
	Skills     []string `json:"skills"`
}

type QueryVacancies struct {
	Query       string    `json:"skill"`
	VacancyList []Vacancy `json:"vacancies"`
}

type Salary struct {
	From     int
	To       int
	Currency string
}
