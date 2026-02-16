package cookie

import (
	"fmt"
	globalmodels "global_models"
	"global_models/global_cookie"
	"net/http"
	"shared/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// Проверяем, что Manager реализует интерфейс
var _ global_cookie.CookieManagerInterface = (*Manager)(nil)

// Manager - только базовая установка кук
type Manager struct {
	config *config.CookieManagerConfig
}

// конструктор для мэнеджера куки
func NewManager(config *config.CookieManagerConfig) global_cookie.CookieManagerInterface {
	return &Manager{config: config}
}

// SetCookie - самый общий метод. Установка куки, согласно переданным параметрам
func (m *Manager) SetCookie(c *gin.Context, opts globalmodels.CookieOptions) error {
	// проверяем наличие имени
	if opts.Name == "" {
		return fmt.Errorf("cookie name must not be empty")
	}

	// Добавляем префикс если задан (если несколько приложений использует этот менеджер, чтобы куки не конфликтовали)
	cookieName := opts.Name
	if m.config.Prefix != "" {
		cookieName = fmt.Sprintf("%s_%s", m.config.Prefix, opts.Name)
	}

	// Определяем параметры безопасности
	// Динамический Secure
	isSecure := c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil
	secure := m.config.Secure || isSecure

	sameSite := m.parseSameSite()
	domain := m.getDomain(cookieName)

	// Путь по умолчанию если не указан
	path := opts.Path
	if path == "" || !strings.HasPrefix(path, "/") {
		path = m.config.DefaultPath
	}

	// HttpOnly по умолчанию true для безопасности
	httpOnly := true
	if opts.HttpOnly != nil {
		httpOnly = *opts.HttpOnly
	}

	c.SetSameSite(sameSite)

	// устанавливаем заголоки для ответа
	c.SetCookie(
		cookieName,
		opts.Value,
		int(m.config.RefreshMaxAge),
		path,
		domain,
		secure,
		httpOnly,
	)

	return nil
}

// GetCookie - получить куку по имени
func (m *Manager) GetCookie(c *gin.Context, name string) (string, error) {
	// Добавляем префикс если задан
	cookieName := name
	if m.config.Prefix != "" {
		cookieName = fmt.Sprintf("%s_%s", m.config.Prefix, name)
	}

	// Получаем куку из контекста Gin
	value, err := c.Cookie(cookieName)
	if err != nil {
		// Возвращаем разные ошибки в зависимости от причины
		if err == http.ErrNoCookie {
			return "", fmt.Errorf("cookie %s not found: %w", cookieName, err)
		}
		return "", fmt.Errorf("failed to get cookie %s: %w", cookieName, err)
	}

	return value, nil
}

// DeleteCookie - очистить куку по имени и path
func (m *Manager) DeleteCookie(c *gin.Context, name, path string) {
	// Добавляем префикс если задан
	cookieName := name
	if m.config.Prefix != "" {
		cookieName = fmt.Sprintf("%s_%s", m.config.Prefix, name)
	}

	// Определяем путь для удаления
	deletePath := path
	if deletePath == "" {
		deletePath = m.config.DefaultPath
	}

	// Определяем параметры безопасности
	// Динамический Secure
	isSecure := c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil
	secure := m.config.Secure || isSecure

	sameSite := m.parseSameSite()
	domain := m.getDomain(cookieName)

	// Устанавливаем SameSite
	c.SetSameSite(sameSite)

	// Устанавливаем куку с истекшим сроком действия
	c.SetCookie(
		cookieName,
		"", // пустое значение
		-1, // отрицательный MaxAge = удалить куку
		deletePath,
		domain,
		secure,
		true, // HttpOnly всегда true для удаления
	)
}

// Вспомогательные методы
func (m *Manager) getDomain(cookieName string) string {
	// Domain - безопасное значение
	if m.config.ProjectMode == "production" && m.config.Domain != "" {
		// Проверяем, что не пытаемся установить __Host- куку с domain
		if !strings.HasPrefix(cookieName, "__Host-") {
			return m.config.Domain
		}
	}
	return ""
}

func (m *Manager) parseSameSite() http.SameSite {
	switch m.config.SameSite {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}
