package config

import "time"

type CookieManagerConfig struct {
	Domain        string        `yaml:"domain" env:"COOKIE_DOMAIN"`                                  // Domain для production (пустая строка для localhost)
	ProjectMode   string        `yaml:"project_mode" env:"PROJECT_MODE"`                             // Режим работы: production, staging, development
	Secure        bool          `yaml:"secure" env:"COOKIE_SECURE"`                                  // Secure flag (true в production)
	SameSite      string        `yaml:"same_site" env:"COOKIE_SAME_SITE" default:"lax"`              // SameSite режим: lax, strict, none
	DefaultPath   string        `yaml:"default_path" env:"COOKIE_DEFAULT_PATH" default:"/"`          // Путь по умолчанию для кук
	RefreshMaxAge time.Duration `yaml:"refresh_max_age" env:"COOKIE_REFRESH_MAX_AGE" default:"168h"` // 7 дней
	Prefix        string        `yaml:"prefix" env:"COOKIE_PREFIX"`                                  // Префикс для имен кук (опционально)
}

// DefaultConfig возвращает конфиг по умолчанию
func DefaultCookieConfig() *CookieManagerConfig {
	return &CookieManagerConfig{
		SameSite:      "lax",
		DefaultPath:   "/",
		Secure:        false,
		RefreshMaxAge: 7 * 24 * time.Hour,
		// Domain и ProjectMode пустые - должны быть явно заданы
	}
}
