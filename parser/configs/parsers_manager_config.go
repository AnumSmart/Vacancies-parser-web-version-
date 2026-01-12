package configs

import (
	"parser/internal/circuitbreaker"
	"time"
)

// структура конфига для менеджера парсеров
type ParserManagerConfig struct {
	MaxConcurrentParsers int                                 `yaml:"max_concurrent_parsers"` // глобальный семафор на использование парсеров
	CircuitBreakerCfg    circuitbreaker.CircuitBreakerConfig `yaml:"circuit_breaker"`        // глобальный circuit breaker
	HealthCheckInterval  time.Duration                       `yaml:"health_check_interval"`  // интервал проверки систояния менеджера парсеров
}

// функция, которая возвращает указатель на дэфолтный конфиг мэнеджера парсеров
func DefaultParsersManagerConfig() *ParserManagerConfig {
	return &ParserManagerConfig{
		CircuitBreakerCfg: circuitbreaker.CircuitBreakerConfig{
			FailureThreshold:    5,
			SuccessThreshold:    3,
			HalfOpenMaxRequests: 2,
			ResetTimeout:        10 * time.Second,
			WindowDuration:      10 * time.Second,
		},
	}
}
