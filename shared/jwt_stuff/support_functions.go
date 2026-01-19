package jwt_stuff

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// вспомогательная функция для создании структуры информации для JWT
func NewClaims(TokenExp time.Duration, email, tokenType, issuer string) jwt.Claims {
	return &CustomClaims{
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			ID:        uuid.New().String(),
		},
	}
}

// вспомогательная фукнция парсинга токена с клэймами
func ParseTokenWithClaims(c *gin.Context, tokenString string, key string) (*jwt.Token, error) {
	// Проверяем не отменен ли контекст
	if err := c.Err(); err != nil {
		return nil, err
	}

	//создаём новый парсер, который учитываем метод шифрования и подтверждение срока действия
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}), // проверять токлько наличие метода шифрования HS256
		jwt.WithExpirationRequired(),            // проверка наличия срока действия токена
	)

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

	// создаём новый парсер без валидации клэймов
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithoutClaimsValidation(),
	)

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
