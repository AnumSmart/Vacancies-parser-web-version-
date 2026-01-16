package search_server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"search_service/internal/search_server/handlers"
	"shared/config"
	"shared/toolkit"

	"github.com/gin-gonic/gin"
)

// структура сервера поиска вакансий
type VacancySearchServer struct {
	httpServer *http.Server
	router     *gin.Engine
	config     *config.ServerConfig
	Handler    *handlers.SearchHandler
}

// Конструктор для сервера
func NewSearchServer(ctx context.Context, config *config.ServerConfig, handler *handlers.SearchHandler) (*VacancySearchServer, error) {
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

	router.Use(toolkit.CORSMiddleware()) // используем для всех маршруторв работу с CORS

	return &VacancySearchServer{
		router:  router,
		config:  config,
		Handler: handler,
	}, nil
}

// Метод для маршрутизации сервера
func (s *VacancySearchServer) SetUpRoutes() {
	s.router.GET("/hello", s.Handler.EchoSearchServer)                  // тестовый ендпоинт
	s.router.POST("/multisearch", s.Handler.ProcessMultisearchRequest)  // эндпоинт поиска всех доступных вакансий из всех доступных источников (согласно строке поиска)
	s.router.POST("/quickoverview", s.Handler.ProcessQuickRequest)      // эндпоинт получения краткой инфы по конкретной найденной вакансии
	s.router.POST("/vac_details", s.Handler.ProcessDetailedVacancyInfo) // эндпоинт получения подробной инфы по конкретной вакансии (отдельный запрос на внешний сервис)
}

// Метод для запуска сервера
func (s *VacancySearchServer) Run() error {
	s.SetUpRoutes()

	s.httpServer = &http.Server{
		Addr:    s.config.Addr(),
		Handler: s.router,
	}

	// если установлен флаг о том, что нужно использовать HTTPS, то запускаем сервер, который работает с HTTPS
	if s.config.EnableTLS {
		// Создаем TLS конфигурацию
		tlsConfig, err := s.config.CreateTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}

		s.httpServer.TLSConfig = tlsConfig

		// Запускаем HTTPS сервер
		log.Printf("Starting HTTPS server on %s", s.config.TLSAddr())
		return s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
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
