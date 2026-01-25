package jwt_service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService - рабочий сервис с методами
type JWTService struct {
	config *JWTConfig // Конфиг внутри сервиса
}

// Конфигурация JWTConfig
type JWTConfig struct {
	SecretAccKey    string        //секретный ключ для access токена
	SecretRefKey    string        //секретный ключ для refresh токена
	AccessTokenExp  time.Duration // время жизни для access токена (обычно около 15 мин)
	RefreshTokenExp time.Duration // время жизни для refresh токена (обычно дни ...)
}

// CustomClaims для JWT
type CustomClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"type"` // "access" или "refresh"
	jwt.RegisteredClaims
}
