// поисковый сервис.
package service

import (
	"context"
	"fmt"
	"search_service/internal/domain/models"
	"search_service/internal/parsers_manager"
	"search_service/internal/search_server/dto"
)

// интерфейс поискового сервиса
type SearchServiceInterface interface {
	SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error)
	GetBriefVacancyDetails(getVacReq dto.SearchVacancyRequest) (models.Vacancy, error)
	GetVacancyDetails(ctx context.Context, getVacReq dto.SearchVacancyRequest) (models.SearchVacancyDetailesResult, error)
	StopServices(ctx context.Context)
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

// метод сервисного слоя для поиска всех доступных вакансий, согласно запроса
func (s *SearchService) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {
	// запускаем комплексный метод поиска
	results, err := s.searchManager.SearchVacancies(ctx, params)
	if err != nil {
		return []models.SearchVacanciesResult{}, err // позвращаем пустой результат поиска (не nil слайс), так как дальше будет конвертация в DTO, а потом в JSON
	}

	return results, nil
}

// метод сервисного слоя для получения сжатой информации по конкретной вакансии из списка уже найденных по ID и сервису
func (s *SearchService) GetBriefVacancyDetails(getVacReq dto.SearchVacancyRequest) (models.Vacancy, error) {
	// создаём составной индекс, в котором будет ID вакансии и сервис, в котором этот ID нужно будет искать
	// этот составной индекс - будет ключем для кэша №2
	compositeID := fmt.Sprintf("%s_%s", getVacReq.Source, getVacReq.VacancyID)

	// создаём переменную для искомой вакансии
	var targetVacancy models.Vacancy

	// пытаемся найти в кэше №2 данные по заданному ключу (составному индексу)
	searchResIndex, ok := s.searchManager.VacancyIndex.GetItem(compositeID)
	if !ok {
		return models.Vacancy{}, fmt.Errorf("No Vacancy with ID:%s was found in cache", getVacReq.VacancyID)
	}

	// проводим type assertion, проверяем нужный тип (так как нам функция GetItem возвращает интерфейс)
	searchResIndexChecked, ok := searchResIndex.(models.VacancyIndex)
	if !ok {
		fmt.Println("Type assertion after GetVacancyDetails ---> failed!")
		return models.Vacancy{}, fmt.Errorf("Type assertion after GetVacancyDetails ---> failed!\n")
	}

	// теперь из полученного из кэша индексов индекса мы можем найти нужный хэш запроса,
	// чтобы потом по этому хэшу из кэша поиска найти нужную вакансию по ID

	// пытаемся найти в кэше данные по заданному хэш ключу
	searchRes, ok := s.searchManager.SearchCache.GetItem(searchResIndexChecked.SearchHash)
	if ok {
		// если можно получить данные из кэша, то получаем интерфейс.
		// проводим type assertion, проверяем нужный тип
		searchResChecked, ok := searchRes.([]models.SearchVacanciesResult)
		if !ok {
			return models.Vacancy{}, fmt.Errorf("Type assertion after multi-search ---> failed!\n")
		}

		for _, neededElementRes := range searchResChecked {
			if neededElementRes.ParserName == getVacReq.Source {
				for _, vacancyRes := range neededElementRes.Vacancies {
					if vacancyRes.ID == getVacReq.VacancyID {
						targetVacancy.ID = vacancyRes.ID
						targetVacancy.Job = vacancyRes.Job
						targetVacancy.Salary = vacancyRes.Salary
						targetVacancy.Company = vacancyRes.Company
						targetVacancy.Location = vacancyRes.Location
						targetVacancy.URL = vacancyRes.URL
					}
				}
			}
		}
	} else {
		s.searchManager.VacancyIndex.DeleteItem(compositeID)
		return models.Vacancy{}, fmt.Errorf("Данные устарели, сделайте повторный запрос поиска всех доступных вакансий\n")
	}

	return targetVacancy, nil
}

// метод сервисного слоя для получения полной информации по конкретной вакансии по ID и названию сервиса (делается отдельный запрос на внешний ресурс)
func (s *SearchService) GetVacancyDetails(ctx context.Context, getVacReq dto.SearchVacancyRequest) (models.SearchVacancyDetailesResult, error) {
	vacInfo, err := s.searchManager.ExecuteSearchVacancyDetailes(ctx, getVacReq.VacancyID, getVacReq.Source)
	if err != nil {
		return models.SearchVacancyDetailesResult{}, err
	}
	return vacInfo, nil
}

// метод для остановки всех воркеров
func (s *SearchService) StopServices(ctx context.Context) {
	s.searchManager.Shutdown()
}
