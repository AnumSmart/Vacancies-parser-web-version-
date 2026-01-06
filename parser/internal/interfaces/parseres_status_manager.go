package interfaces

import (
	"time"
)

// структура статуса отдельного парсера (DTO для этого интерфейса)
type ParserStatus struct {
	Name           string        // имя парсера
	LastCheck      time.Time     // время последней проверки статуса
	LastSuccess    time.Time     // время последней успешной проверки
	ErrorCount     int           // количество состояний, что парсер в ошибке
	SuccessCount   int           // количество состояний, что парсер - без ошибок
	IsHealthy      bool          // состояние
	LastError      error         // последняя ошибка
	CircuitState   string        // "closed", "open", "half-open" (состояние внутреннего circuit breaker)
	Initialized    bool          // false - просто создан парсер, true - была попытка запроса
	HealthEndpoint string        // URL для health check
	ResponseTime   time.Duration // время ответа от парсера
}

type ParsersStatusManager interface {
	UpdateStatus(name string, success bool, err error)
	GetHealthyParsers() []string
	GetParserStatus(name string) (*ParserStatus, bool)
	Stop()
}
