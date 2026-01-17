// описание моделей сервиса авторизации
package dto

// структура запроса для логина пользователя
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// структура ответа от сервиса авторизации
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"` // "Bearer"
}
