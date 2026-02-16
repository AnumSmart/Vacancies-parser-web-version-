// описание моделей сервиса авторизации
package dto

import (
	"time"
)

// структура запроса для логина пользователя
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// структура ответа на запрос login  от сервиса авторизации
type LoginResponse struct {
	AccessToken string `json:"access_token"` // access и refresh
	TokenType   string `json:"token_type"`   // Bearer
	UserID      int64  `json:"user_id,omitempty"`
	ExpiresIn   int64  `json:"expires_in"`
}

// структура запроса для регистрации пользователя
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,containsany=!@#$"`
}

// структура ответа при успешной регистрации
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
}

// Структура для входящего запроса
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Структура для логаута
type LogOutParams struct {
	UserID    string        // ID пользователя
	TokenID   string        // ID токена (access)
	TokenType string        // Тип токена
	Token     string        // переданный токен
	TTL       time.Duration // время жизни токена (оставшееся)
}
