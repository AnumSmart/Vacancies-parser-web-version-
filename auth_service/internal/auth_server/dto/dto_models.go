// описание моделей сервиса авторизации
package dto

// структура запроса для логина пользователя
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// структура ответа на запрос login  от сервиса авторизации
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
}

// структура запроса для регистрации пользователя
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,containsany=!@#$"`
}

// структура ответа при успешной регистрации
type RegisterResponse struct {
	Message     string `json:"message"`
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}
