package jobs

import "search_service/internal/domain/models"

// SearchJob - джоба для поиска вакансий
type SearchJob struct {
	BaseJob
	Params models.SearchParams
}
