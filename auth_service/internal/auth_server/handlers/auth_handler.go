// описание хэндлеров для сервера авторизации
package handlers

import (
	"auth_service/internal/auth_server/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// структура хэндлера сервера авторизации
type AuthHandler struct {
	service service.AuthServiceInterface
}

func (a *AuthHandler) EchoAuthServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello from auth server!"})
}
