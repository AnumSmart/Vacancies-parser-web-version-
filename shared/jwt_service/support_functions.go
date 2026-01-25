package jwt_service

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// создаём новый парсер, который учитываем метод шифрования и подтверждение срока действия
var parser = jwt.NewParser(
	jwt.WithValidMethods([]string{"HS256"}), // проверять токлько наличие метода шифрования HS256
	jwt.WithExpirationRequired(),            // проверка наличия срока действия токена
)

// LoadJWTConfig - специальная функция для JWT (без дефолтов!)
func LoadJWTConfig(configPath string) (*JWTConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("JWT config path is required")
	}

	// Проверяем существование файла
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("JWT config file not found: %w", err)
	}

	// Читаем файл
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT config: %w", err)
	}

	// Парсим
	var config JWTConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JWT config: %w", err)
	}

	// ВАЛИДАЦИЯ (самое важное!)
	if err := validateJWTConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid JWT config: %w", err)
	}
	return &config, nil
}

// validateJWTConfig - строгая валидация
func validateJWTConfig(cfg *JWTConfig) error {
	// 1. Ключи не должны быть пустыми
	if cfg.SecretAccKey == "" {
		return fmt.Errorf("access_secret is required")
	}
	if cfg.SecretRefKey == "" {
		return fmt.Errorf("refresh_secret is required")
	}

	// 2. Минимальная длина ключей (рекомендация: 32+ символа)
	if len(cfg.SecretAccKey) < 32 {
		return fmt.Errorf("access_secret too short (min 32 chars)")
	}
	if len(cfg.SecretRefKey) < 32 {
		return fmt.Errorf("refresh_secret too short (min 32 chars)")
	}

	// 3. Ключи должны быть разными (опционально, но рекомендуется)
	if cfg.SecretAccKey == cfg.SecretRefKey {
		return fmt.Errorf("access_secret and refresh_secret must be different")
	}

	// 4. Валидация времени жизни
	if cfg.AccessTokenExp <= 0 {
		return fmt.Errorf("access_token_expiry must be positive")
	}
	if cfg.RefreshTokenExp <= 0 {
		return fmt.Errorf("refresh_token_expiry must be positive")
	}

	// 5. Refresh должен жить дольше Access
	if cfg.RefreshTokenExp <= cfg.AccessTokenExp {
		return fmt.Errorf("refresh_token_expiry must be longer than access_token_expiry")
	}

	// 6. Максимальные значения (опционально)
	if cfg.AccessTokenExp > 24*time.Hour {
		return fmt.Errorf("access_token_expiry too long (max 24h)")
	}
	if cfg.RefreshTokenExp > 90*24*time.Hour { // 90 дней
		return fmt.Errorf("refresh_token_expiry too long (max 90 days)")
	}

	return nil
}

// вспомогательная функция для создании структуры информации для JWT
func NewClaims(TokenExp time.Duration, email, tokenType, issuer string) CustomClaims {
	newClaim := CustomClaims{
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			ID:        uuid.New().String(),
		},
	}
	return newClaim
}

// вспомогательная фукнция парсинга токена с клэймами
func ParseTokenWithClaims(c *gin.Context, tokenString string, key string) (*jwt.Token, error) {
	// Проверяем не отменен ли контекст
	if err := c.Err(); err != nil {
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
			return []byte(key), nil
		})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid token",
			"details": err.Error(),
		})
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
