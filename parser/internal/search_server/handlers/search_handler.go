// создаём поисковый хэндлер
package handlers

import (
	"net/http"
	"parser/internal/search_server/converters"
	"parser/internal/search_server/dto"
	"parser/internal/search_server/service"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service service.SearchServiceInterface // интерфейс сервисного слоя для поиска
}

// конструктор для создания поискового хэндлера
func NewSearchHandler(service service.SearchServiceInterface) *SearchHandler {
	return &SearchHandler{
		service: service,
	}
}

// метод для теста запуска сервера
func (s *SearchHandler) EchoSearchServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from search server!"})
}

// метод обработки запроса на поиск всех вакансий (согласно условиям поиска) во всех доступных источниках
func (s *SearchHandler) ProcessMultisearchRequest(c *gin.Context) {
	// Парсинг DTO запроса
	var req dto.SearchRequest

	// парсим данные запроса из JSON в необходимую структуру
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// проводим валидацию и нормализацию входных данных
	if err := req.ValidateAndNormalize(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертация DTO -> Domain
	params := converters.SearchRequestDTOToParamsDomain(req)

	// запускаем комплексный метод поиска (идём в сервисный слой)
	searchVacanciesResults, err := s.service.SearchVacancies(c, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что получили хоть какаеи данные после поиска
	if len(searchVacanciesResults) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to find vacancies"})
		return
	}

	// Конвертация Domain -> DTO
	result := converters.SearchVacanciesResultDomainToDTO(searchVacanciesResults)

	// отдаём результат клиенту
	c.JSON(http.StatusOK, result)
}

// метод обработки запроса одной вакансии из списка найденных по ID
func (s *SearchHandler) ProcessQuickRequest(c *gin.Context) {
	// Парсинг DTO запроса
	var req dto.SearchVacancyRequest

	// парсим данные запроса из JSON в необходимую структуру
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Message": "invalid request", "error": err.Error()})
		return
	}

	// проводим валидацию и нормализацию входных данных
	if err := req.ValidateAndNormalize(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// вызываем поиск вакансии по ID
	resultDoamin, err := s.service.GetBriefVacancyDetails(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертация Domain -> DTO
	resultDTO := converters.ConvertVacancyToDTO(resultDoamin)

	// отдаём результат клиенту
	c.JSON(http.StatusOK, resultDTO)
}

// метод получения детальной информации по вакансии (отдельный запрос на внешний источник)
func (s SearchHandler) ProcessDetailedVacancyInfo(c *gin.Context) {
	// Парсинг DTO запроса
	var req dto.SearchVacancyRequest

	// парсим данные запроса из JSON в необходимую структуру
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Message": "invalid request", "error": err.Error()})
		return
	}

	// проводим валидацию и нормализацию входных данных
	if err := req.ValidateAndNormalize(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// вызываем поиск расширенной информации вакансии по ID
	resultDomain, err := s.service.GetVacancyDetails(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертация Domain -> DTO
	resultDTO := converters.ConvertVacancyResultInfoDomainToDTO(resultDomain)

	// отдаём результат клиенту
	c.JSON(http.StatusOK, resultDTO)
}
