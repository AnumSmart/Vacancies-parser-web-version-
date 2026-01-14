package models

import "time"

// модель составного индекса для получения деталей найденных вакансий
type VacancyIndex struct {
	SearchHash string
	ParserName string
	Index      int
	CreatedAt  time.Time
}
