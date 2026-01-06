package circuitbreaker

import "time"

// CircuitBreakerConfig - конфигурация Circuit Breaker
type CircuitBreakerConfig struct {
	FailureThreshold    uint32
	SuccessThreshold    uint32
	HalfOpenMaxRequests uint32
	ResetTimeout        time.Duration
	WindowDuration      time.Duration
}

// создаём конструктор для конфига circuit breaker
// будем возвращать копию структуры, так как будут разные конфиги
func NewCircuitBreakerConfig(fTreshold, sTreshold, halfTreshold uint32, resetTimeout, winDuration time.Duration) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:    fTreshold,
		SuccessThreshold:    sTreshold,
		HalfOpenMaxRequests: halfTreshold,
		ResetTimeout:        resetTimeout,
		WindowDuration:      winDuration,
	}
}
