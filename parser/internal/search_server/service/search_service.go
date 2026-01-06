// поисковый сервис.
package service

// интерфейс поискового сервиса
type SearchServiceInterface interface {
}

// структура поискового сервиса
type SearchService struct{}

// конструктор поискового сервиса
func NewSearchService() *SearchService {
	return &SearchService{}
}
