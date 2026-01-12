package models

import "time"

// Структура джобы для очереди в мэнеджере парсеров для поиска спика вакансий
type SearchVacanciesJob struct {
	ID         string
	Params     SearchParams
	ResultChan chan *JobSearchVacanciesResult
	CreatedAt  time.Time
}

// Структура результата по выполнении работы поиска списка вакансий
type JobSearchVacanciesResult struct {
	Results []SearchVacanciesResult
	Error   error
}

func (s *SearchVacanciesJob) GetID() string {
	return s.ID
}
