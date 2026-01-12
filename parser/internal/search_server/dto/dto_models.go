package dto

import "errors"

// SearchRequest - DTO для входящего запроса
type SearchRequest struct {
	Query    string `json:"query" binding:"required,min=2,max=100"`
	Location string `json:"location"`
	PerPage  int    `json:"per_page" binding:"min=1,max=100"`
	Page     int    `json:"page" binding:"min=0"`
}

type SearchVacancyRequest struct {
	VacancyID string `json:"vacancy_id"`
	Source    string `json:"source"`
}

// VacancyResponse - DTO для ответа клиенту
type VacancyResponse struct {
	ID         string `json:"id"`
	Job        string `json:"job"`
	Company    string `json:"company"`
	Salary     string `json:"salary"` // "от 150 000 ₽" или "не указана"
	Currency   string `json:"currency"`
	Location   string `json:"location"`
	Experience string `json:"experience"` // "Нет опыта", "1-3 года"
	Schedule   string `json:"schedule"`   // "Полный день", "Удаленная работа"
	Source     struct {
		Name string `json:"name"` // "hh.ru", "SuperJob"
		Icon string `json:"icon"` // URL иконки
	} `json:"source"`
	URL         string `json:"url"`
	Description string `json:"description"`
	PublishedAt string `json:"published_at"` // "2 дня назад"
}

// SearchVacanciesResponse - DTO для ответа с группировкой по источникам
type SearchVacanciesResponse struct {
	Results map[string]SourceVacancies `json:"results"`
	Total   int                        `json:"total"`
}

// DTO для ответа - расширенная информация по конкретной вакансии
type VacancyDetailsResponce struct {
	ID          string `json:"id"`
	Job         string `json:"job"`
	Company     string `json:"company"`
	Salary      string `json:"salary"` // "от 150 000 ₽" или "не указана"
	Description string `json:"description"`
	URL         string `json:"url"`
}

// SourceVacancies - вакансии одного источника
type SourceVacancies struct {
	Name      string            `json:"name"`
	Icon      string            `json:"icon"`
	Vacancies []VacancyResponse `json:"vacancies"`
	Count     int               `json:"count"`
	HasError  bool              `json:"has_error"`
	Error     string            `json:"error,omitempty"`
	Duration  string            `json:"duration,omitempty"` // "1.2s"
}

// метод валидации и нормализации данных из запроса поиска вакансий
func (r *SearchRequest) ValidateAndNormalize() error {
	if r.Query == "" {
		return errors.New("search text cannot be empty")
	}

	// Нормализация
	if r.PerPage == 0 {
		r.PerPage = 50
	} else if r.PerPage > 100 {
		r.PerPage = 100 // также можно добавить ограничение
	}

	if r.Page == 0 {
		r.Page = 1
	}

	// Дополнительная валидация
	if r.PerPage < 1 {
		return errors.New("per_page must be positive")
	}

	if r.Page < 1 {
		return errors.New("page must be positive")
	}

	return nil
}

// метод валидации и нормализации данных из запроса для поиска вакансии, среди уже найденных по ID
func (r *SearchVacancyRequest) ValidateAndNormalize() error {
	if r.VacancyID == "" {
		return errors.New("Vacancy ID can not be empty!")
	}
	if r.Source == "" {
		return errors.New("Vacancy Source can not be empty!")
	}
	return nil
}
