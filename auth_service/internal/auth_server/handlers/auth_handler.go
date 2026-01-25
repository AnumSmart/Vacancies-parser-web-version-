// описание хэндлеров для сервера авторизации
package handlers

import (
	"auth_service/internal/auth_server/dto"
	"auth_service/internal/auth_server/service"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// описание интерфейса слоя хэндлеров
type AuthHandlerInterface interface {
	EchoAuthServer(c *gin.Context)
	ShutDown(ctx context.Context)
	RegisterHandler(c *gin.Context)
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

// метод слоя Handlers для обработки входящего POST запроса, валидации запроса и регистрации нового пользователя
func (a *AuthHandler) RegisterHandler(c *gin.Context) {
	validatedData, exists := c.Get("validatedData")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, ok := validatedData.(*dto.RegisterRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
		return
	}

	// вызываем метод сервиса для регистрации нового пользователя
	userID, err := a.service.Register(c, user.Email, user.Password)
	if err != nil {
		// Обработка разных типов ошибок
		if errors.Is(err, ErrUserExists) {
			code, apiErr := ToAPIError(err)
			c.JSON(http.StatusConflict, gin.H{"status code": code, "error": apiErr})
			return
		}
		code, apiErr := ToAPIError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status code": code, "error": apiErr})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": userID,
	})
}
