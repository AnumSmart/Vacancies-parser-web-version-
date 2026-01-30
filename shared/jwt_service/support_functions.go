package jwt_service

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
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
