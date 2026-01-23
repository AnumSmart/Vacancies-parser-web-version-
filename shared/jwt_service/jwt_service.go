package jwt_service

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager interface {
	GenerateTokens(email string) (string, string, error)
}

// NewJWTService создаёт рабочий сервис с конфигом
func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// метод структуры JWT для генерации токенов (access и refresh)
func (j *JWTService) GenerateTokens(email string) (string, string, error) {
	// Access токен
	accessClaims := NewClaims(j.config.AccessTokenExp, email, "access", "my_app")
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.config.SecretAccKey))
	if err != nil {
		return "", "", err
	}

	// Refresh токен
	refreshClaims := NewClaims(j.config.RefreshTokenExp, email, "refresh", "my_app")
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.config.SecretRefKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
