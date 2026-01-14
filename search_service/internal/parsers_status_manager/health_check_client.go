// реализация HealthClient для HTTP парсеров
// необходим для периодического опроса парсеров на предмет их состояние для работы
package parsers_status_manager

import (
	"context"
	"fmt"
	"net/http"
	"search_service/configs"
	"time"
)

// HttpHealthCheckClient реализация HealthClient для HTTP парсеров
type HttpHealthCheckClient struct {
	client  *http.Client  // клиент для реализации запроса
	timeout time.Duration // таймаут запроса healthCheck
}

func NewHttpHealthCheckClient(conf *configs.HealthCheckConfig) *HttpHealthCheckClient {
	if conf == nil {
		conf = configs.DefaultHealthCheckConfig()
	}

	return &HttpHealthCheckClient{
		client: &http.Client{
			Timeout: conf.HealthCheckClientConfig.TimeOut, // общий таймаут клиента
			Transport: &http.Transport{
				MaxConnsPerHost:       conf.HealthCheckClientConfig.MaxConnPerHost,
				MaxIdleConnsPerHost:   conf.HealthCheckClientConfig.MaxIdleConns,
				IdleConnTimeout:       conf.HealthCheckClientConfig.IdleConnTimeout,
				TLSHandshakeTimeout:   conf.HealthCheckClientConfig.TLSHandshakeTimeout,
				ExpectContinueTimeout: conf.HealthCheckClientConfig.ExpectContinueTimeout,
			},
		},
		timeout: conf.RequestTimeOut, // таймаут на запрос
	}
}

// метод для выполнения тестовго запроса для проверки healthCheck
func (h *HttpHealthCheckClient) CheckHealth(ctx context.Context, endpoint string) (time.Duration, bool, error) {
	// создаём кнтекст с таймаутом для контроля времени запроса
	reqCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	start := time.Now()

	// формируем запрос с контектом
	req, err := http.NewRequestWithContext(reqCtx, "GET", endpoint, nil)
	if err != nil {
		return 0, false, fmt.Errorf("Failed to create request with context for health check: %v", err)
	}

	// устанавливаем заголовок
	req.Header.Set("User-Agent", "ParserHealthCheck/1.0")

	// делаем запрос с помощью клиента
	resp, err := h.client.Do(req)

	select {
	case <-ctx.Done():
		return h.timeout, false, ctx.Err()
	default:
		// высчитываем время запроса
		reqDuration := time.Since(start)
		if err != nil {
			return reqDuration, false, fmt.Errorf("Err during health request: %v", err)
		}

		// очищаем ресурсы (после проверки ошибок)
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return reqDuration, true, nil
		}
		// возвращаем ошибку проверки работоспособности
		return reqDuration, false, fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
}
