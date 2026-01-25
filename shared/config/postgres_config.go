// описание конфига для подключения к базе PostgresQL
package config

import "time"

// структура конфига для базы
type PostgresDBConfig struct {
	DSN string

	// Configure connection pool settings
	MaxConns int32
	MinConns int32

	// Configure connection health checks
	HealthCheckPeriod time.Duration
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration

	// Configure connection timeouts
	ConnectTimeout time.Duration
}

// функция использования конфига по-дэфолту
func UseDefautPostgresDBConfig() *PostgresDBConfig {
	return &PostgresDBConfig{
		DSN:               "host=localhost user=admin password=my_pass dbname=users port=5432 sslmode=disable",
		MaxConns:          10,
		MinConns:          2,
		HealthCheckPeriod: 1 * time.Minute,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}
