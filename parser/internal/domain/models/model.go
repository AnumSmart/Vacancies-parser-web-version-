package models

import "time"

// общая структура поиска
type SearchParams struct {
	Text     string
	Location string
	PerPage  int
	Page     int
}

// Стуктура общей вакансии для всех ответов
type Vacancy struct {
	ID          string
	Job         string
	Company     string
	Salary      *string
	Currency    string
	Location    string
	Experience  string
	Schedule    string
	URL         string
	Source      string // "hh", "superjob", ...
	Description string
	PublishedAt time.Time
}

// Структура для определния результатов поиска списка вакансий по всем доступным парсерам
type SearchVacanciesResult struct {
	Vacancies  []Vacancy
	ParserName string
	SearchHash string
	Error      error
	Duration   time.Duration
}

// Employer представляет информацию о работодателе
type Employer struct {
	ID   string
	Name string
}

// Area представляет информацию о местоположении
type Area struct {
	ID   string
	Name string
}

// Salary представляет информацию о зарплате
type Salary struct {
	From     int
	To       int
	Currency string
	Gross    bool
}

// Структура для определния результатов поиска деталей конкретной вакансии
type SearchVacancyDetailesResult struct {
	Employer    Employer
	Location    Area
	Salary      Salary
	Description string
	Name        string
	ID          string
	Url         string
}
