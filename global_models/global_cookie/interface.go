package global_cookie

import (
	globalmodels "global_models"

	"github.com/gin-gonic/gin"
)

// интерфейс для использовании во внешних модулях
type CookieManagerInterface interface {
	SetCookie(c *gin.Context, opts globalmodels.CookieOptions) error
	GetCookie(c *gin.Context, name string) (string, error)
	DeleteCookie(c *gin.Context, name, path string)
}
