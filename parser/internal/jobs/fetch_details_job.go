package jobs

// FetchDetailsJob - джоба для получения деталей вакансии
type FetchDetailsJob struct {
	BaseJob
	Source    string
	VacancyID string
}
