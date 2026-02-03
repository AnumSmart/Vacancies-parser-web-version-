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
	SecretAccKey    string        `yaml:"access_secret"`        //секретный ключ для access токена
	SecretRefKey    string        `yaml:"refresh_secret"`       //секретный ключ для refresh токена
	TokenPepper     string        `yaml:"token_pepper"`         // секретный ключ для хэширования токенов
	AccessTokenExp  time.Duration `yaml:"access_token_expiry"`  // время жизни для access токена (обычно около 15 мин)
	RefreshTokenExp time.Duration `yaml:"refresh_token_expiry"` // время жизни для refresh токена (обычно дни ...)
}

// CustomClaims для JWT
type CustomClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"type"` // "access" или "refresh"
	jwt.RegisteredClaims
}
