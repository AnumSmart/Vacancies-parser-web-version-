package dto

// SearchRequest - DTO для входящего запроса
type SearchRequest struct {
	Query    string `json:"query" binding:"required,min=2,max=100"`
	Location string `json:"location"`
	PerPage  int    `json:"per_page" binding:"min=1,max=100"`
	Page     int    `json:"page" binding:"min=0"`
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

// SearchResponse - основной ответ API
type SearchVacanciesResponse struct {
	Success    bool              `json:"success"`
	Data       []VacancyResponse `json:"data"`
	ParserName string            `json:"parser_name"`
	TookMs     int64             `json:"took_ms"` // Время выполнения
}
