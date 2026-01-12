package model

import "parser/pkg"

// HHVacancy представляет структуру вакансии с HH.ru
type HHVacancy struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Salary      Salary   `json:"salary"`
	Employer    Employer `json:"employer"`
	Area        Area     `json:"area"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
}

// Salary представляет информацию о зарплате
type Salary struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	Currency string `json:"currency"`
	Gross    bool   `json:"gross"`
}

// Employer представляет информацию о работодателе
type Employer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Area представляет информацию о местоположении
type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchResponse представляет ответ от API HH.ru
type SearchResponse struct {
	Items []HHVacancy `json:"items"`
	Found int         `json:"found"`
	Pages int         `json:"pages"`
}

// предоставляет ответ API HH.ru по запросу с ID
type SearchDetails struct {
	Employer    Employer `json:"employer"`
	Area        Area     `json:"area"`
	Salary      Salary   `json:"salary"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	ID          string   `json:"id"`
	Url         string   `json:"alternate_url"`
}

// GetSalaryString возвращает форматированную строку зарплаты
func (v HHVacancy) GetSalaryString() string {
	if v.Salary.From == 0 && v.Salary.To == 0 {
		return "не указана"
	}

	if v.Salary.From > 0 && v.Salary.To > 0 {
		return pkg.FormatSalary(v.Salary.From, v.Salary.To, v.Salary.Currency)
	} else if v.Salary.From > 0 {
		return pkg.FormatSalary(v.Salary.From, 0, v.Salary.Currency)
	} else {
		return pkg.FormatSalary(0, v.Salary.To, v.Salary.Currency)
	}
}
