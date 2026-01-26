package config

import (
	"time"
)

// структура для конфига сервера
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           string        `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	// дополнительные поля, если нужно чтобы эта реализация сервера могла использовать https
	/*
		EnableTLS      bool          `yaml:"enable_tls"`    // флаг, который говорит о том, что нужно использовать HTTPS
		TLSCertFile    string        `yaml:"tls_cert_file"` // путь к файлу сертиыфикации (пасспорт сервера)
		TLSKeyFile     string        `yaml:"tls_key_file"`  // путь к приватному ключу (сертификация)
		TLSPort        string        `yaml:"tls_port"`      // стандартный HTTPS порт
	*/
}

// функция для создания конфига сервера по - дефолту
func UseDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:           "localhost",
		Port:           "8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,

		/*
			// TLS значения по умолчанию
			EnableTLS:   false,
			TLSCertFile: "certs/localhost.pem",
			TLSKeyFile:  "certs/localhost-key.pem",
			TLSPort:     "8443",
		*/
	}
}

// метод конфига сервера для формирования адреса
func (c *ServerConfig) Addr() string {
	return c.Host + ":" + c.Port
}

/*
// метод для формирования TLS адреса (если используется HTTPS)
func (c *ServerConfig) TLSAddr() string {
	return c.Host + ":" + c.TLSPort
}

// метод для проверки валидности TLS конфигурации
func (c *ServerConfig) ValidateTLS() error {
	if !c.EnableTLS {
		return nil // TLS не включен, ничего проверять не нужно
	}

	if c.TLSCertFile == "" {
		return &ConfigError{Field: "TLSCertFile", Msg: "TLS certificate file is required when TLS is enabled"}
	}

	if c.TLSKeyFile == "" {
		return &ConfigError{Field: "TLSKeyFile", Msg: "TLS key file is required when TLS is enabled"}
	}

	return nil
}
*/

// Вспомогательная структура для ошибок конфигурации
type ConfigError struct {
	Field string
	Msg   string
}

// метод вспомогательной функции для формирования ошибок
func (e *ConfigError) Error() string {
	return "config error in field '" + e.Field + "': " + e.Msg
}
