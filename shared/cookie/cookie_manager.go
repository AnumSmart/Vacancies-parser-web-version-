package cookie

import (
	"fmt"
	"net/http"
	"shared/config"

	"github.com/gin-gonic/gin"
)

// интерфейс для использовании во внешних модулях
type CookieManagerInterface interface {
	SetCookie(c *gin.Context, opts CookieOptions) error
	GetCookie(c *gin.Context, name string) (string, error)
	DeleteCookie(c *gin.Context, name, path string)
}

// Manager - только базовая установка кук
type Manager struct {
	config config.CookieManagerConfig
}

// конструктор для мэнеджера куки
func NewManager(config config.CookieManagerConfig) *Manager {
	return &Manager{config: config}
}

// структура опций для работы с куки
type CookieOptions struct {
	Name     string // имя куки
	Value    string // значение
	MaxAge   int    // в секундах
	Path     string // путь
	HttpOnly *bool  // nil = использовать дефолт (true)
}

// SetCookie - самый общий метод. Установка куки, согласно переданным параетрам
func (m *Manager) SetCookie(c *gin.Context, opts CookieOptions) error {
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
	secure := m.config.Secure
	sameSite := m.parseSameSite()
	domain := m.getDomain()

	// Путь по умолчанию если не указан
	path := opts.Path
	if path == "" {
		path = m.config.DefaultPath
	}

	// HttpOnly по умолчанию true для безопасности
	httpOnly := true
	if opts.HttpOnly != nil {
		httpOnly = *opts.HttpOnly
	}

	c.SetSameSite(sameSite)

	c.SetCookie(
		cookieName,
		opts.Value,
		opts.MaxAge,
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
	secure := m.config.Secure
	sameSite := m.parseSameSite()
	domain := m.getDomain()

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
func (m *Manager) getDomain() string {
	if m.config.ProjectMode == "production" && m.config.Domain != "" {
		return m.config.Domain
	}
	return "" // для localhost/development
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
