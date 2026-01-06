package configs

import "time"

// структура для конфига сервера
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           string        `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

// функция для создания конфига сервера по - дефолту
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:           "localhost",
		Port:           "8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

// метод конфига сервера для формирования адреса
func (c *ServerConfig) Addr() string {
	return c.Host + ":" + c.Port
}
