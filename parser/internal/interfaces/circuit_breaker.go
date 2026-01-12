package interfaces

// интерфейс для circuit breaker
type CBInterface interface {
	Execute(fn func() error) error
	GetStats() (total, success, failure uint32)
}
