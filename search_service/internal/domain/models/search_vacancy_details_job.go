package models

type SearchVacancyDetailesJob struct {
	ID         string
	VacancyID  string
	ParserName string
	ResultChan chan *JobSearchVacancyDetailesResult
}

type JobSearchVacancyDetailesResult struct {
	Result SearchVacancyDetailesResult
	Error  error
}

func (s *SearchVacancyDetailesJob) GetID() string {
	return s.ID
}
