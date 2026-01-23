package middleware

import (
	"errors"
	"log"
	"net/http"
	"shared/jwt_service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(config *jwt_service.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		authHeader := c.GetHeader("Authorization")

		// проверяем наличие токена в заготовке, если его нет, выдаём ошибку и не пускаем запрос дальше
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Проверяем формат "Bearer <token>"
		tokenString, err := CheckBearerFormat(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		//Парсим токен
		token, err := jwt_service.ParseTokenWithClaims(c, tokenString, config.SecretAccKey)
		if err != nil {
			log.Println("Invalid token")
			return
		}

		// Проверяем claims
		if claims, ok := token.Claims.(*jwt_service.CustomClaims); ok && token.Valid {
			// Добавляем данные пользователя в контекст
			c.Set("user_email", claims.Email)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}

func CheckBearerFormat(authHeader string) (string, error) {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}
	return "", errors.New("Invalid authorization header format")
}
