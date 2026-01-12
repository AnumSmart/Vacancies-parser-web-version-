package search_server

import (
	"context"
	"log"
	"net/http"
	"parser/configs"
	"parser/internal/search_server/handlers"

	"github.com/gin-gonic/gin"
)

type VacancySearchServer struct {
	httpServer *http.Server
	router     *gin.Engine
	config     *configs.ServerConfig
	handler    *handlers.SearchHandler
}

// Конструктор для сервера
func NewServer(ctx context.Context, config *configs.ServerConfig, handler *handlers.SearchHandler) (*VacancySearchServer, error) {
	// создаём экземпляр роутера
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	// Добавляем middleware для проброса контекста
	router.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "request_id", c.GetHeader("X-Request-ID"))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	router.Use(CORSMiddleware()) // используем для всех маршруторв работу с CORS

	return &VacancySearchServer{
		router:  router,
		config:  config,
		handler: handler,
	}, nil
}

// Метод для маршрутизации сервера
func (s *VacancySearchServer) SetUpRoutes() {
	s.router.GET("/hello", s.handler.EchoSearchServer)                  // тестовый ендпоинт
	s.router.POST("/multisearch", s.handler.ProcessMultisearchRequest)  // эндпоинт поиска всех доступных вакансий из всех доступных источников (согласно строке поиска)
	s.router.POST("/quickoverview", s.handler.ProcessQuickRequest)      // эндпоинт получения краткой инфы по конкретной найденной вакансии
	s.router.POST("/vac_details", s.handler.ProcessDetailedVacancyInfo) // эндпоинт получения подробной инфы по конкретной вакансии (отдельный запрос на внешний сервис)
}

// Метод для запуска сервера
func (s *VacancySearchServer) Run() error {
	s.SetUpRoutes()

	s.httpServer = &http.Server{
		Addr:    s.config.Addr(),
		Handler: s.router,
	}
	log.Println("Server is running on port 8080")
	return s.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (s *VacancySearchServer) Shutdown(ctx context.Context) error {

	// Останавливаем HTTP сервер
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}

// middleware для CORS политики
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Список разрешенных доменов
		allowedOrigins := []string{
			"http://localhost:8080",
		}

		origin := c.Request.Header.Get("Origin")

		// Если Origin не указан (например, запрос из curl или postman)
		if origin == "" {
			// Разрешаем любые источники (или задайте конкретные)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Проверяем по списку разрешенных
			isAllowed := false
			for _, domain := range allowedOrigins {
				if domain == origin {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				c.AbortWithStatusJSON(403, gin.H{
					"error":  "Origin not allowed",
					"origin": origin,
				})
				return
			}
		}

		// Разрешенные методы
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")

		// Разрешенные заголовки
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		// Заголовки, которые можно читать клиенту
		c.Writer.Header().Set("Access-Control-Expose-Headers",
			"Content-Length, Content-Type, Authorization")

		// Разрешаем отправку кук/авторизации
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Кеширование предзапроса (в секундах)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
