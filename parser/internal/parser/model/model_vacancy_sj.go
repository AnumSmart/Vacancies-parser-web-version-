package model

import "parser/pkg"

// Структуры для SuperJob API
type SuperJobResponse struct {
	Items []SJVacancy `json:"objects"`
	Total int         `json:"total"`
}

type SJVacancy struct {
	ID              int    `json:"id"`
	Profession      string `json:"profession"`
	FirmName        string `json:"firm_name"`
	PaymentFrom     int    `json:"payment_from"`
	PaymentTo       int    `json:"payment_to"`
	Currency        string `json:"currency"`
	Town            Town   `json:"town"`
	Link            string `json:"link"`
	VacancyRichText string `json:"vacancyRichText"`
}

type Town struct {
	Title string `json:"title"`
}

func (v SJVacancy) GetSalaryString() string {
	if v.PaymentFrom == 0 && v.PaymentTo == 0 {
		return "не указана"
	}

	if v.PaymentFrom > 0 && v.PaymentTo > 0 {
		return pkg.FormatSalary(v.PaymentFrom, v.PaymentTo, v.Currency)
	} else if v.PaymentFrom > 0 {
		return pkg.FormatSalary(v.PaymentFrom, 0, v.Currency)
	} else {
		return pkg.FormatSalary(0, v.PaymentTo, v.Currency)
	}
}
