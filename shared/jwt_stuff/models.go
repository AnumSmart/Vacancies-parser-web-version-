package jwt_stuff

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Конфигурация JWT
type JWT struct {
	SecretAccKey    string        //секретный ключ для access токена
	SecretRefKey    string        //секретный ключ для refresh токена
	AccessTokenExp  time.Duration // время жизни для access токена (обычно около 15 мин)
	RefreshTokenExp time.Duration // время жизни для refresh токена (обычно дни ...)
}

// Claims для JWT
type CustomClaims struct {
	Email     string `json:"email"` // адрес электромнной почты юзера
	TokenType string `json:"type"`  // "access" или "refresh"
	jwt.RegisteredClaims
}
