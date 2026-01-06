package jobs

import "parser/internal/domain/models"

// SearchJob - джоба для поиска вакансий
type SearchJob struct {
	BaseJob
	Params models.SearchParams
}
