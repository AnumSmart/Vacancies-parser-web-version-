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

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Конвертация DTO -> Domain
	params := converters.SearchRequestToParams(req)

	c.JSON(http.StatusOK, params)

}
