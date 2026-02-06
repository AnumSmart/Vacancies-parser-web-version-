// описание хэндлеров для сервера авторизации
package handlers

import (
	"auth_service/internal/auth_server/dto"
	"auth_service/internal/auth_server/service"
	"auth_service/internal/domain"
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
	LoginHandler(c *gin.Context)
	ProcessRefreshTokenHandler(c *gin.Context)
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
	userID, err := a.service.Register(c.Request.Context(), user.Email, user.Password)
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

	// формируем объект для ответа
	response := dto.RegisterResponse{
		Message: "User registered successfully",
		UserID:  userID,
		Email:   user.Email,
	}

	// в ответе пользователю отдаём сообщение и ID пользователя
	c.JSON(http.StatusCreated, response)
}

// метод слоя Handlers для обработки входящего POST запроса, валидация запроса, проверка пользователя в базе, в ответе: пара JWT токенов
func (a *AuthHandler) LoginHandler(c *gin.Context) {
	//проверяем, есть ли в контексте валидированные данные
	validatedData, exists := c.Get("validatedData")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation data not found"})
		return
	}

	// Приведение типа с проверкой
	user, ok := validatedData.(*dto.LoginRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid request type",
		})
		return
	}

	//пробуем залогировать пользователя
	err := a.service.Login(c.Request.Context(), user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем access и refresh токены
	accessToken, refreshToken, err := a.service.GetTokens(c.Request.Context(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка при получении токенов токена",
			"error":   err.Error(),
		})
		return
	}

	// пробуем добавить refresh токен в базу
	err = a.service.AddHashRefreshTokenToDb(c.Request.Context(), user.Email, refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка записи refreshToken в БД",
			"error":   err.Error(),
		})
		return
	}

	// структура jwt токенов
	tokenPair := domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// формируем ответ для пользователя
	responce := dto.LoginResponse{
		Tokens:    tokenPair,
		TokenType: "Bearer",
	}

	c.JSON(http.StatusOK, responce)
}

// Хэндлер генерации нового access токена, при предоставлении валидного refresh токена
func (a *AuthHandler) ProcessRefreshTokenHandler(c *gin.Context) {
	//Проверка того, что JSON из запроса мапится в нужную структуру refresh токена
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}
	// 3. Вызов сервиса
	tokens, err := a.service.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		// Обработка ошибок: токен невалиден, отозван и т.д.
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
