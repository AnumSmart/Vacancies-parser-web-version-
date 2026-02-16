// описание хэндлеров для сервера авторизации
package handlers

import (
	"auth_service/internal/auth_server/dto"
	"auth_service/internal/auth_server/service"
	"context"
	"errors"
	"fmt"
	globalmodels "global_models"
	"global_models/global_cookie"
	"net/http"
	"shared/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// описание интерфейса слоя хэндлеров
type AuthHandlerInterface interface {
	EchoAuthServer(c *gin.Context) // ЭХО для тестирования!
	ShutDown(ctx context.Context)
	RegisterHandler(c *gin.Context)            // Хэндлер для регистрации нового пользователя (публичный)
	LoginHandler(c *gin.Context)               // Хэндлер для логина зарегестрированного пользователя (публичный)
	ProcessRefreshTokenHandler(c *gin.Context) // Хэндлер для обновления пары JWT токенов (публичный)
	ValidateTokenHandler(c *gin.Context)       // Хэндлер для проверки JWT токена от API Gateway (nginx)
	LogoutHandler(c *gin.Context)              // Хэндлер для логаута (защищён проверкой токена от nginx)
}

// структура хэндлера сервера авторизации
type AuthHandler struct {
	service       service.AuthServiceInterface
	cookieManager global_cookie.CookieManagerInterface
}

// конструктор для слоя хэндлеров
func NewAuthHandler(service service.AuthServiceInterface, cookieManager global_cookie.CookieManagerInterface) *AuthHandler {
	return &AuthHandler{
		service:       service,
		cookieManager: cookieManager,
	}
}

// метод проверки работоспособности слоя хэндлеров
func (a *AuthHandler) EchoAuthServer(c *gin.Context) {
	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hello from auth server!"})
}

// метод хэндлера для остановки сервиса поиска
func (a *AuthHandler) ShutDown(ctx context.Context) {
	a.service.StopServices(ctx)
}

// метод слоя Handlers для обработки входящего POST запроса, валидации запроса и регистрации нового пользователя
func (a *AuthHandler) RegisterHandler(c *gin.Context) {
	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

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
	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

	//проверяем, есть ли в контексте валидированные данные
	validatedData, exists := c.Get("validatedData")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation data not found"})
		return
	}

	// Приведение типа с проверкой
	loginInfo, ok := validatedData.(*dto.LoginRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid request type",
		})
		return
	}

	//пробуем залогировать пользователя
	userID, err := a.service.Login(c.Request.Context(), loginInfo.Email, loginInfo.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем access и refresh токены, если метод Login отработал без ошибок
	tokenPair, err := a.service.GetTokens(c.Request.Context(), loginInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка при получении токенов токена",
			"error":   err.Error(),
		})
		return
	}

	// создаём хэш рэфрэш токена и пробуем добавить в базу
	err = a.service.AddHashRefreshTokenToDb(c.Request.Context(), loginInfo.Email, tokenPair.RefreshToken, tokenPair.RefreshJTI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка записи refreshToken в БД",
			"error":   err.Error(),
		})
		return
	}

	// refresh token возвращаем через cookieManager в куке
	// создаём опцции для куки, согласно глобальной модели

	opts := globalmodels.CookieOptions{
		Name:  "refresh_token",        // имя куки
		Value: tokenPair.RefreshToken, // значение куки (тут это refresh токен)
		Path:  "/api/auth/refresh",    // путь
	}

	// устанавливаем рефрэш токен в куки (используе куки-менеджер)
	err = a.cookieManager.SetCookie(c, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка установки Cookie (refresh_token)",
			"error":   err.Error(),
		})
		return
	}

	// формируем ответ для пользователя (в ответе возвращаем только access токен)
	responce := dto.LoginResponse{
		AccessToken: tokenPair.AccessToken,
		TokenType:   "Bearer",
		UserID:      userID, // возвращаем пользователю ID
	}

	c.JSON(http.StatusOK, responce)
}

// Хэндлер генерации нового access токена, при предоставлении валидного refresh токена
func (a *AuthHandler) ProcessRefreshTokenHandler(c *gin.Context) {
	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

	// БРАУЗЕР АВТОМАТИЧЕСКИ прислал refresh_token в куке
	// Нам не нужно читать его из тела запроса или заголовков!

	refreshToken, err := a.cookieManager.GetCookie(c, "refresh_token")
	if err != nil {
		// Нет refresh токена в куке - просим перелогиниться
		c.JSON(http.StatusUnauthorized, gin.H{
			"Message": "no refresh token",
			"Code":    "REFRESH_TOKEN_MISSING",
			"Error":   err.Error(),
		})
		return
	}

	// Вызов сервиса
	tokens, err := a.service.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		// Токен невалидный - удаляем куку
		a.cookieManager.DeleteCookie(c, "refresh_token", "/api/auth/refresh")
		// Обработка ошибок: токен невалиден, отозван и т.д.
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// refresh token возвращаем через cookieManager в куке
	// создаём опцции для куки, согласно глобальной модели

	opts := globalmodels.CookieOptions{
		Name:  "refresh_token",     // имя куки
		Value: tokens.RefreshToken, // значение куки (тут это refresh токен)
		Path:  "/api/auth/refresh", // путь
	}

	// устанавливаем рефрэш токен в куки (используе куки-менеджер)
	err = a.cookieManager.SetCookie(c, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Ошибка установки Cookie (refresh_token)",
			"error":   err.Error(),
		})
		return
	}

	// возвращаем новый access токен в ответе пользователю
	c.JSON(http.StatusOK, gin.H{
		"Access_token": tokens.AccessToken,
		"TokenType":    "Bearer",
	})
}

// Хэндлер для валидации access токена от API Gateway (это будет внутренний эндпоинт nginx)
func (a *AuthHandler) ValidateTokenHandler(c *gin.Context) {
	// Проверяем Content-Type (nginx отправляет JSON)
	if c.ContentType() != "application/json" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Content-Type must be application/json",
			"code":  "INVALID_CONTENT_TYPE",
		})
		return
	}

	// Проверяем не отменён ли контекст
	if c.Request.Context().Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

	// Извлекаем токен из тела запроса (nginx отправляет JSON), будет access токен
	var request struct {
		Token string `json:"token"`
	}

	// ВАЖНО: nginx отправляет JSON с полем "token"
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "INVALID_JSON",
		})
		return
	}

	// Проверяем наличие токена
	if request.Token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Token is required",
			"code":  "MISSING_TOKEN",
		})
		return
	}

	// Проверяем формат "Bearer <token>"
	tokenString, err := middleware.CheckBearerFormat(request.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// валидируем токен в сервисном слое
	claims, valid, err := a.service.ValidateToken(c.Request.Context(), tokenString)

	// обрабатываем ошибку
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// проверяем, валидный ли токен
	if !valid {
		c.Status(401)
		return // ← Выходим, не идем дальше!
	}

	// проверяем, что claim не nil
	if claims == nil {
		// Это не должно происходить при valid=true
		fmt.Printf("BUG: claims is nil but token is valid")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Проверяем обязательные поля в claims
	if claims.UserID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token claims",
			"code":  "INVALID_CLAIMS",
		})
		return
	}

	// если все успешно, возвращаем необходимые заголовки и статус
	c.Header("X-User-ID", claims.UserID) // обязательныое поле отведа для nginx
	c.Header("X-User-Roles", "user")     // обязательныое поле отведа для nginx (пока используем "заглушку" - user)
	c.Header("X-Token-ID", claims.ID)
	c.Header("X-Token-Type", claims.TokenType)
	c.Header("X-Token-Exp", claims.ExpiresAt.String())
	c.Status(200)
}

// Хэндлер для логаута
func (a *AuthHandler) LogoutHandler(c *gin.Context) {
	ctx := c.Request.Context()
	// Проверяем не отменён ли контекст
	if ctx.Err() != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
		return
	}

	// Токен уже валидирован API Gateway
	// Извлекаем данные из заголовков, установленных nginx
	userID := c.GetHeader("X-User-ID")
	tokenID := c.GetHeader("X-Token-ID")
	tokenType := c.GetHeader("X-Token-Type")
	expStr := c.GetHeader("X-Token-Exp")

	// Если API Gateway не передал claims, возвращаем ошибку
	if userID == "" || tokenID == "" || expStr == "" || tokenType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Missing authentication headers",
		})
		return
	}

	// Проверяем, что это access токен (не refresh)
	if tokenType != "access" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_token_type",
			"message": "Only access tokens can be used for logout",
		})
		return
	}

	// Парсим expiration time
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_expiration",
			"message": "Invalid token expiration format",
		})
		return
	}

	// получаем оставшееся время жизни токена (в данном случае access)
	expiresAt := time.Unix(expUnix, 0)

	ttl := time.Until(expiresAt)

	// формируем структуру параметров для сервисного слоя
	logOutParams := dto.LogOutParams{
		UserID:    userID,
		TokenID:   tokenID,
		TokenType: tokenType,
		TTL:       ttl,
	}

	// продуем провести logout
	err = a.service.LogOut(ctx, &logOutParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "LogOut error",
			"message": err.Error(),
		})
		return
	}

	// ----------------------------------------------------------------------------------------------------------------
	// доделать работу с cookie через cookieManager (ивалидация cookie)------------------------------------------------
	// ----------------------------------------------------------------------------------------------------------------

	//  успешный ответ пользователю
	c.JSON(http.StatusOK, gin.H{
		"message": "LogOut finished successfully",
	})

}
