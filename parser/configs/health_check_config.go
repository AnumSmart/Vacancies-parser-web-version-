package configs

import "time"

type HealthCheckConfig struct {
	RequestTimeOut          time.Duration           `yaml:"request_timeout"`        // Таймаут на один health check запрос
	Initialization_timeout  time.Duration           `yaml:"initialization_timeout"` // Максимальное время инициализации
	HealthCheckInterval     time.Duration           `yaml:"check_interval"`         // интервал периодической проверки состояния сервисов, куда ходят парсеры
	HealthCheckClientConfig HealthCheckClientConfig `yaml:"http_client"`            // healthCheck client
}

type HealthCheckClientConfig struct {
	TimeOut               time.Duration `yaml:"timeout"`                 // Общий таймаут клиента
	MaxIdleConns          int           `yaml:"max_idle_conns"`          // максимальное количество бездействующих (keep-alive) соединений для http клиента (экономия ресурсов)
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`       // интервал, через сколько закрывать неиспользуемое соединение
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout"`   // максимальное время ожидания завершения TLS handshake
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"` // интервал, оптимизация для сценариев загрузки больших данных
	MaxConnPerHost        int           `yaml:"max_conns_per_host"`
}

func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		RequestTimeOut:         5 * time.Second,
		Initialization_timeout: 10 * time.Second,
		HealthCheckInterval:    15 * time.Second,
		HealthCheckClientConfig: HealthCheckClientConfig{
			TimeOut:               5 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: time.Second,
			MaxConnPerHost:        2,
		},
	}
}
