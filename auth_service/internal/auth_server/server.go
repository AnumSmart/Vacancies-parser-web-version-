// описание сервера авторизации
package authserver

import (
	"auth_service/internal/auth_server/dto"
	"auth_service/internal/auth_server/handlers"
	"context"
	"log"
	"net/http"
	"shared/config"
	"shared/middleware"

	"github.com/gin-gonic/gin"
)

// структура сервера авторизации
type AuthServer struct {
	httpServer *http.Server
	router     *gin.Engine
	config     *config.ServerConfig
	Handler    handlers.AuthHandlerInterface
}

// Конструктор для сервера
func NewAuthServer(ctx context.Context, config *config.ServerConfig, handler handlers.AuthHandlerInterface) (*AuthServer, error) {
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

	router.Use(middleware.CORSMiddleware()) // используем для всех маршруторв работу с CORS

	return &AuthServer{
		router:  router,
		config:  config,
		Handler: handler,
	}, nil
}

// Метод для маршрутизации сервера
func (a *AuthServer) SetUpRoutes() {
	a.router.GET("/hello", a.Handler.EchoAuthServer) // тестовый ендпоинт
	a.router.POST("/register", middleware.ValidateAuthMiddleware(&dto.RegisterRequest{}), a.Handler.RegisterHandler)
	a.router.POST("/login", middleware.ValidateAuthMiddleware(&dto.LoginRequest{}), a.Handler.LoginHandler)
	a.router.GET("/refresh", a.Handler.ProcessRefreshTokenHandler)

	// Эндпоинт для валидации access токена (используется nginx)
	a.router.POST("/api/v1/validate", a.Handler.ValidateTokenHandler)

	// зазищённые эндпоинты
	a.router.GET("/logout", a.Handler.LogoutHandler)
}

// Метод для запуска сервера
func (a *AuthServer) Run() error {
	a.SetUpRoutes()

	a.httpServer = &http.Server{
		Handler: a.router,
	}
	// Используем обычный порт для HTTP
	a.httpServer.Addr = a.config.Addr()
	log.Printf("Starting HTTP server on %s", a.config.Addr())
	return a.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (a *AuthServer) Shutdown(ctx context.Context) error {

	// Останавливаем HTTP сервер
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
