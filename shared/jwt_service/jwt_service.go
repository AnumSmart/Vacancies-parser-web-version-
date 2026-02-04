package jwt_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManagerInterface interface {
	GenerateTokens(email string) (string, string, error)
	GetJTWConfig() *JWTConfig
	CalculateTokenTTL(claims *CustomClaims) time.Duration
}

// NewJWTService создаёт рабочий сервис с конфигом
func NewJWTService(config *JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

// создаём новый парсер, который учитываем метод шифрования и подтверждение срока действия
var parser = jwt.NewParser(
	jwt.WithValidMethods([]string{"HS256"}), // проверять токлько наличие метода шифрования HS256
	jwt.WithExpirationRequired(),            // проверка наличия срока действия токена
)

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

func (j *JWTService) GetJTWConfig() *JWTConfig {
	return j.config
}

// метод, который определяет оставшеесф время жизни токена
func (j *JWTService) CalculateTokenTTL(claims *CustomClaims) time.Duration {
	if claims == nil || claims.ExpiresAt == nil {
		// Если ExpiresAt не установлен, используем стандартное значение
		// Например, 24 часа для refresh токена
		return 24 * time.Hour
	}

	expiryTime := claims.ExpiresAt.Time
	now := time.Now()

	remainingTime := expiryTime.Sub(now)

	// Гарантируем минимальный TTL (специально хардкод, при необходимости можно вынести в JWTConfig)
	minTTL := 5 * time.Minute
	maxTTL := 30 * 24 * time.Hour // Максимум 30 дней

	// Минимальное и максимальное время можно сделать настраиваемым
	if remainingTime < minTTL {
		return minTTL
	}
	if remainingTime > maxTTL {
		return maxTTL
	}

	return remainingTime
}

// вспомогательная фукнция парсинга токена с клэймами
// передаём контекст, строку рефрэш токена и секрет для рэфрэш токена
func ParseTokenWithClaims(ctx context.Context, tokenString string, refreshSecret string) (*jwt.Token, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// пытаемся получить токен
	token, err := parser.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(refreshSecret), nil
		})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// parseTokenWithoutVerification парсит JWT токен без проверки подписи,
// но с проверкой базовой структуры и обязательных полей
func ParseTokenWithoutVerification(tokenString string) (*CustomClaims, error) {
	// Базовые проверки токена
	if tokenString == "" {
		return nil, errors.New("empty token string")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format: expected 3 parts")
	}

	// Парсим токен без верификации подписи
	token, _, err := parser.ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Приводим claims к нашему типу
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims structure")
	}

	// Проверяем обязательные поля
	if claims.ID == "" {
		return nil, errors.New("token missing jti claim")
	}
	if claims.ExpiresAt == nil {
		return nil, errors.New("token missing exp claim")
	}
	if claims.Email == "" {
		return nil, errors.New("token missing email claim")
	}
	if claims.TokenType != "access" && claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type, expected 'access' or 'refresh'")
	}

	return claims, nil
}
