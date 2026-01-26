// описание конфига для подключения к базе PostgresQL
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

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

// NewPostgresDBConfigFromEnv создает конфиг PostgreSQL из переменных окружения
// Возвращает ошибку, если обязательные поля не заполнены или значения некорректны
func NewPostgresDBConfigFromEnv() (*PostgresDBConfig, error) {
	var errors []string

	// Получаем обязательные поля с проверкой
	host, err := getRequiredEnv("DB_HOST")
	if err != nil {
		errors = append(errors, err.Error())
	}

	user, err := getRequiredEnv("DB_USER")
	if err != nil {
		errors = append(errors, err.Error())
	}

	password, err := getRequiredEnv("DB_PASSWORD")
	if err != nil {
		errors = append(errors, err.Error())
	}

	dbName, err := getRequiredEnv("DB_NAME")
	if err != nil {
		errors = append(errors, err.Error())
	}

	// Если есть ошибки в обязательных полях - возвращаем сразу
	if len(errors) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(errors, ", "))
	}

	// Получаем опциональные поля со значениями по умолчанию
	port := getEnvWithDefault("DB_PORT", "5432")
	sslMode := getEnvWithDefault("DB_SSL_MODE", "disable")

	// Собираем DSN
	dsn := buildDSN(host, port, user, password, dbName, sslMode)

	// Получаем числовые значения с валидацией
	maxConns, err := getEnvAsInt32WithValidation("DB_MAX_CONNS", 10, 1, 100)
	if err != nil {
		errors = append(errors, err.Error())
	}

	minConns, err := getEnvAsInt32WithValidation("DB_MIN_CONNS", 2, 0, 50)
	if err != nil {
		errors = append(errors, err.Error())
	}

	// Проверяем что minConns <= maxConns
	if minConns > maxConns {
		errors = append(errors, fmt.Sprintf("DB_MIN_CONNS (%d) cannot be greater than DB_MAX_CONNS (%d)", minConns, maxConns))
	}

	// Получаем значения duration с валидацией
	healthCheckPeriod, err := getEnvAsDurationWithValidation("DB_HEALTH_CHECK_PERIOD", 60*time.Second, 1*time.Second, 300*time.Second)
	if err != nil {
		errors = append(errors, err.Error())
	}

	maxConnLifetime, err := getEnvAsDurationWithValidation("DB_MAX_CONN_LIFETIME", 3600*time.Second, 1*time.Second, 24*time.Hour)
	if err != nil {
		errors = append(errors, err.Error())
	}

	maxConnIdleTime, err := getEnvAsDurationWithValidation("DB_MAX_CONN_IDLE_TIME", 1800*time.Second, 1*time.Second, 24*time.Hour)
	if err != nil {
		errors = append(errors, err.Error())
	}

	connectTimeout, err := getEnvAsDurationWithValidation("DB_CONNECT_TIMEOUT", 5*time.Second, 1*time.Second, 60*time.Second)
	if err != nil {
		errors = append(errors, err.Error())
	}

	// Проверяем что maxConnIdleTime <= maxConnLifetime
	if maxConnIdleTime > maxConnLifetime {
		errors = append(errors, fmt.Sprintf("DB_MAX_CONN_IDLE_TIME (%v) cannot be greater than DB_MAX_CONN_LIFETIME (%v)", maxConnIdleTime, maxConnLifetime))
	}

	// Если есть ошибки валидации - возвращаем их
	if len(errors) > 0 {
		return nil, fmt.Errorf("configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return &PostgresDBConfig{
		DSN:               dsn,
		MaxConns:          maxConns,
		MinConns:          minConns,
		HealthCheckPeriod: healthCheckPeriod,
		MaxConnLifetime:   maxConnLifetime,
		MaxConnIdleTime:   maxConnIdleTime,
		ConnectTimeout:    connectTimeout,
	}, nil
}

// buildDSN собирает DSN строку из компонентов
func buildDSN(host, port, user, password, dbName, sslMode string) string {
	parts := []string{
		"host=" + host,
		"port=" + port,
		"user=" + user,
		"password=" + password,
		"dbname=" + dbName,
		"sslmode=" + sslMode,
	}
	return strings.Join(parts, " ")
}

// getRequiredEnv получает обязательную переменную окружения
func getRequiredEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return val, nil
}

// getEnvWithDefault получает переменную окружения или значение по умолчанию
func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvAsInt32WithValidation получает переменную окружения как int32 с валидацией
func getEnvAsInt32WithValidation(key string, defaultValue, min, max int32) (int32, error) {
	if val := os.Getenv(key); val != "" {
		// Atoi возвращает int, что на большинстве систем = int64
		i, err := strconv.Atoi(val)
		if err != nil {
			return defaultValue, fmt.Errorf("%s: must be an integer, got %q", key, val)
		}

		result := int32(i)

		// Проверяем, что значение помещается в int32
		if int64(i) != int64(result) {
			return defaultValue, fmt.Errorf("%s: value %d is too large for int32", key, i)
		}

		if result < min || result > max {
			return defaultValue, fmt.Errorf("%s: value %d is out of range [%d, %d]", key, result, min, max)
		}

		return result, nil
	}
	return defaultValue, nil
}

// getEnvAsDurationWithValidation получает переменную окружения как time.Duration с валидацией
func getEnvAsDurationWithValidation(key string, defaultValue, min, max time.Duration) (time.Duration, error) {
	if val := os.Getenv(key); val != "" {
		// Пробуем распарсить как duration строку
		d, err := time.ParseDuration(val)
		if err != nil {
			// Пробуем как число (предполагаем секунды)
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return defaultValue, fmt.Errorf("%s: must be a duration (like '1m', '1h') or number of seconds, got %q", key, val)
			}
			d = time.Duration(i) * time.Second
		}

		if d < min || d > max {
			return defaultValue, fmt.Errorf("%s: duration %v is out of range [%v, %v]", key, d, min, max)
		}

		return d, nil
	}
	return defaultValue, nil
}
