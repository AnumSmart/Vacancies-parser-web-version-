package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TrustedAuthMiddleware - Gin middleware который доверяет Nginx
func TrustedAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Проверяем обязательный заголовок X-User-ID от Nginx
		userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
		if userID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "User identification missing",
			})
			c.Abort()
			return
		}

		// 2. Проверяем что аутентификация была пройдена
		authValidated := strings.TrimSpace(c.GetHeader("X-Auth-Validated"))
		if authValidated != "true" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_server_error",
				"message": "Authentication not validated",
			})
			c.Abort()
			return
		}

		// 3. Сохраняем userID в контексте Gin
		c.Set("userID", userID)

		// 4. Передаем управление следующему обработчику
		c.Next()
	}
}
