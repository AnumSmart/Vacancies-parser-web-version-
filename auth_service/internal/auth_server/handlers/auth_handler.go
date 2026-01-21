// описание хэндлеров для сервера авторизации
package handlers

import (
	"auth_service/internal/auth_server/service"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// описание интерфейса слоя хэндлеров
type AuthHandlerInterface interface {
	EchoAuthServer(c *gin.Context)
	ShutDown(ctx context.Context)
}

// структура хэндлера сервера авторизации
type AuthHandler struct {
	service service.AuthServiceInterface
}

// конструктор для слоя хэндлеров
func NewAuthHandler(service service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// метод проверки работоспособности слоя хэндлеров
func (a *AuthHandler) EchoAuthServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from auth server!"})
}

// метод хэндлера для остановки сервиса поиска
func (a *AuthHandler) ShutDown(ctx context.Context) {
	a.service.StopServices(ctx)
}
