package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
)

// создаём экзмепляр валидатора (чтобы он создавался в памяти только при загрузке модуля)
var validate = validator.New()

// ValidateMiddleware создает middleware для валидации
func ValidateAuthMiddleware(model interface{}) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Создаем новый экземпляр структуры для валидации
		request := reflect.New(reflect.TypeOf(model).Elem()).Interface()

		// Парсим БЕЗ встроенной валидации Gin
		// Используем ShouldBindBodyWith с binding.JSON
		if err := c.ShouldBindBodyWith(request, binding.JSON); err != nil {
			// Только ошибки парсинга JSON (не валидации!)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON format",
				"code":  "INVALID_JSON",
			})
			c.Abort()
			return
		}

		// Валидируем структуру
		if err := validate.Struct(request); err != nil {
			errors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				errors[err.Field()] = err.Tag()
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": errors,
			})

			c.Abort()
			return
		}

		// Сохраняем валидированные данные в контекст для использования в обработчике
		c.Set("validatedData", request)
		c.Next()
	}
}
