package toolkit

import "github.com/gin-gonic/gin"

// middleware для CORS политики
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Список разрешенных доменов
		allowedOrigins := []string{
			"http://localhost:8080",
		}

		origin := c.Request.Header.Get("Origin")

		// Если Origin не указан (например, запрос из curl или postman)
		if origin == "" {
			// Разрешаем любые источники (или задайте конкретные)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Проверяем по списку разрешенных
			isAllowed := false
			for _, domain := range allowedOrigins {
				if domain == origin {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				c.AbortWithStatusJSON(403, gin.H{
					"error":  "Origin not allowed",
					"origin": origin,
				})
				return
			}
		}

		// Разрешенные методы
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")

		// Разрешенные заголовки
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		// Заголовки, которые можно читать клиенту
		c.Writer.Header().Set("Access-Control-Expose-Headers",
			"Content-Length, Content-Type, Authorization")

		// Разрешаем отправку кук/авторизации
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Кеширование предзапроса (в секундах)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
