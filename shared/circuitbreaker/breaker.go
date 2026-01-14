package circuitbreaker

import (
	"errors"
	"shared/config"
	"sync"
	"time"
)

// Состояния Circuit Breaker
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// Структура Circuit Breaker
type CircuitBreaker struct {
	mu sync.RWMutex

	//Конфигурация
	failureThreshold    uint32        // Макс кол-во ошибок до перехода в Open
	successThreshold    uint32        // Кол-во успешных запросов для перехода в Closed
	halfOpenMaxRequests uint32        // Макс запросов в Half-Open состоянии
	resetTimeout        time.Duration // Время ожидания перед Half-Open
	windowDuration      time.Duration // Время для подсчета статистики

	// Состояние
	state            State
	failures         uint32
	successes        uint32
	lastFailureTime  time.Time
	halfOpenAttempts uint32

	// Статистика
	totalRequests  uint32
	totalSuccesses uint32
	totalFailures  uint32
}

func NewCircutBreaker(config config.CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 3
	}
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = 2
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 10 * time.Second
	}
	if config.WindowDuration <= 0 {
		config.WindowDuration = 10 * time.Second
	}

	return &CircuitBreaker{
		failureThreshold:    config.FailureThreshold,
		successThreshold:    config.SuccessThreshold,
		halfOpenMaxRequests: config.HalfOpenMaxRequests,
		resetTimeout:        config.ResetTimeout,
		windowDuration:      config.WindowDuration,
		state:               StateClosed,
	}
}
