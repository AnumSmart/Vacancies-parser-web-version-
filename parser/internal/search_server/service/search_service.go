// поисковый сервис.
package service

import (
	"context"
	"parser/internal/domain/models"
	"parser/internal/parsers_manager"
)

// интерфейс поискового сервиса
type SearchServiceInterface interface {
	SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error)
}

// структура поискового сервиса
type SearchService struct {
	searchManager *parsers_manager.ParsersManager
}

// конструктор поискового сервиса
func NewSearchService(searchManager *parsers_manager.ParsersManager) *SearchService {
	return &SearchService{
		searchManager: searchManager,
	}
}

func (s *SearchService) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {
	// запускаем комплексный метод поиска
	results, err := s.searchManager.SearchVacancies(ctx, params)
	if err != nil {
		return []models.SearchVacanciesResult{}, err // позвращаем пустой результат поиска (не nil слайс), так как дальше будет конвертация в DTO, а потом в JSON
	}

	return results, nil
}
