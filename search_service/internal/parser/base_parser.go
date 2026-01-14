package parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"search_service/internal/domain/models"
	"search_service/pkg"
	"shared/circuitbreaker"
	"shared/config"
	"shared/interfaces"
	"shared/rate_limiter"
	"time"
)

// BaseConfig конфигурация базового парсера
type BaseConfig struct {
	Name                  string                      // имя парсера (к какому источнику будет привязан)
	BaseURL               string                      // базовый URL, через который бдет осуществляться парсинг
	HealthEndPoint        string                      // URL, через который бдет осуществляться health check
	APIKey                string                      // API ключ, если предусмотрен сервисом
	Timeout               time.Duration               // таймаут для http клиента
	RateLimit             time.Duration               // интервал для rate limiter (ограничение частоты обращения к ресурсу)
	MaxConcurrent         int                         // резмер буфера для семафора (ограничение конкурентности), и ограничение для http клиента
	CircuitBreakerCfg     config.CircuitBreakerConfig // конфиг для circuit breaker
	MaxIdleConns          int                         // максимальное количество бездействующих (keep-alive) соединений для http клиента (экономия ресурсов)
	IdleConnTimeout       time.Duration               // интервал, через сколько закрывать неиспользуемое соединение
	TLSHandshakeTimeout   time.Duration               // максимальное время ожидания завершения TLS handshake
	ResponseHeaderTimeout time.Duration               // интервал, сколько ждать ответа сервера после отправки запроса
	ExpectContinueTimeout time.Duration               // интервал, оптимизация для сценариев загрузки больших данных
}

// BaseParser базовая реализация парсера
type BaseParser struct {
	name           string                 // имя парсера (к какому источнику будет привязан)
	baseURL        string                 // базовый URL, через который бдет осуществляться парсинг
	healthEndPoint string                 // URL, через который бдет осуществляться health check
	apiKey         string                 // API ключ, если предусмотрен сервисом
	httpClient     *http.Client           // экземпляр клиента, через который будем проводить парсинг на внешнем источнике
	rateLimiter    interfaces.RateLimiter // экземпляр rate limiter (ограничение частоты обращения к ресурсу)
	circuitBreaker interfaces.CBInterface // экземпляр для circuit breaker (отказоустойчивость)
	semaphore      chan struct{}          // семафор (ограничение конкурентности)
	maxConcurrent  int                    // размер буфера для семафора
}

// Конструктор, который создает базовый парсер
func NewBaseParser(config BaseConfig) (*BaseParser, error) {
	// проверяем, что из конфига приходит валидное значение rate
	rateLimiter, err := rate_limiter.NewChannelRateLimiter(config.RateLimit)
	if err != nil {
		return nil, err
	}

	return &BaseParser{
		name:           config.Name,
		baseURL:        config.BaseURL,
		healthEndPoint: config.HealthEndPoint,
		apiKey:         config.APIKey,
		httpClient:     createHTTPClient(config),
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitbreaker.NewCircutBreaker(config.CircuitBreakerCfg),
		semaphore:      make(chan struct{}, config.MaxConcurrent),
		maxConcurrent:  config.MaxConcurrent,
	}, nil
}

// функция, которая создаёт новый клиент с параметрами
func createHTTPClient(config BaseConfig) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxConnsPerHost:       config.MaxConcurrent,
			MaxIdleConnsPerHost:   config.MaxIdleConns,
			IdleConnTimeout:       config.IdleConnTimeout,
			TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
			ResponseHeaderTimeout: config.ResponseHeaderTimeout,
			ExpectContinueTimeout: config.ExpectContinueTimeout,
		},
	}
}

// ParserFuncs определяет типы специфичных функций парсера
type ParserFuncs struct {
	BuildURL       func(models.SearchParams) (string, error)
	Parse          func([]byte) (interface{}, error)
	Convert        func(interface{}) ([]models.Vacancy, error)
	ConvertDetails func(interface{}) (models.SearchVacancyDetailesResult, error)
}

// SearchVacancies общий метод для поиска вакансий
func (p *BaseParser) SearchVacancies(ctx context.Context, params models.SearchParams, funcs ParserFuncs) ([]models.Vacancy, error) {
	// Строим URL с параметрами
	apiURL, err := funcs.BuildURL(params)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	// заводим переменную, в которую будем складывать результат
	var vacancies []models.Vacancy

	// Используем Circuit Breaker для выполнения запроса к внешнему сервису
	//---------------------------------------------------------------------------------------------------
	err = p.circuitBreaker.Execute(func() error {
		// Обработка семафора
		if err := p.acquireSemaphore(ctx); err != nil {
			return err
		}
		defer p.releaseSemaphore() // после завершения вызова функции, освобождаем семафор

		// Перед осуществлением запроса проверяем rate limiter
		err := p.rateLimiter.Wait(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("ссылка для поиска: %s\n", apiURL)
		// Выполнение HTTP запроса
		resp, err := p.executeRequest(ctx, apiURL)
		if err != nil {
			return err
		}

		// освобождаем ресурсы
		defer p.drainAndClose(resp)

		// Проверка статуса
		if err := p.checkResponseStatus(resp); err != nil {
			return err
		}

		// Чтение и парсинг
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response failed: %w", err)
		}

		// парсим даные
		parsedData, err := funcs.Parse(body)
		if err != nil {
			return fmt.Errorf("parse response failed: %w", err)
		}

		// пробуем сконвертировать результаты поиска к единому формату. Обязательно type assertion, на входе interface{}
		converted, err := funcs.Convert(parsedData)
		if err != nil {
			return fmt.Errorf("convert to universal failed: %w", err)
		}

		vacancies = converted

		return nil
	})
	//---------------------------------------------------------------------------------------------------

	// если ошибки есть, определяем какого они рода
	if err != nil {
		return p.handleCircuitBreakerErrorVacanciesSearch(err)
	}

	return vacancies, nil
}

// SearchVacancyDetailes общий метод для поиска деталей по конкретной вакансии
func (p *BaseParser) SearchVacancyDetailes(ctx context.Context, vacancyID string, funcs ParserFuncs) (models.SearchVacancyDetailesResult, error) {
	// формируем url поиска для базового парсера
	searchUrl := pkg.UrlBuilder(p.baseURL, vacancyID)

	// заводим переменную, в которую будем складывать результат
	var vacancyDetails models.SearchVacancyDetailesResult

	// Используем Circuit Breaker для выполнения запроса к внешнему сервису
	//---------------------------------------------------------------------------------------------------
	err := p.circuitBreaker.Execute(func() error {
		// Обработка семафора
		if err := p.acquireSemaphore(ctx); err != nil {
			return err
		}
		defer p.releaseSemaphore() // после завершения вызова функции, освобождаем семафор

		// Перед осуществлением запроса проверяем rate limiter
		err := p.rateLimiter.Wait(ctx)
		if err != nil {
			return err
		}

		// Выполнение HTTP запроса
		resp, err := p.executeRequest(ctx, searchUrl)
		if err != nil {
			return err
		}

		// освобождаем ресурсы
		defer p.drainAndClose(resp)

		// Проверка статуса
		if err := p.checkResponseStatus(resp); err != nil {
			return err
		}

		// Чтение и парсинг
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response failed: %w", err)
		}

		// парсим даные
		parsedData, err := funcs.Parse(body)
		if err != nil {
			return fmt.Errorf("parse response failed: %w", err)
		}

		// пробуем сконвертировать результаты поиска к единому формату. Обязательно type assertion, на входе interface{}
		converted, err := funcs.ConvertDetails(parsedData)
		if err != nil {
			return fmt.Errorf("convert of details - failed: %w", err)
		}

		vacancyDetails = converted

		return nil
	})

	// если ошибки есть, определяем какого они рода
	if err != nil {
		return p.handleCircuitBreakerErrorVacancyDetails(err)
	}

	return vacancyDetails, nil

}

// метод проверки доступности семафора
func (p *BaseParser) acquireSemaphore(ctx context.Context) error {
	select {
	case p.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context canceled while waiting for semaphore: %w", ctx.Err())
	case <-time.After(2 * time.Second):
		return fmt.Errorf("semaphore timeout: %s API is busy", p.name)
	}
}

// метод освобождения семафора
func (p *BaseParser) releaseSemaphore() {
	<-p.semaphore
}

// метод для выполнения HTTP запроса через клиент
func (p *BaseParser) executeRequest(ctx context.Context, url string) (*http.Response, error) {
	//формируем запрос
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// делаем запрос, получаем ответ
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	return resp, nil
}

// метод для дренирования и закрытия тела ответа, освобождения ресурсов.
// защищает, если идёт отмена чтения или body полностью не читается
// вычитывает данные в мусорный ридер с лимитом, освобождает ресурсы
func (p *BaseParser) drainAndClose(resp *http.Response) {
	// Читаем с лимитом (защита от DoS)
	const maxBodySlurp = 1 << 20 // 1MB
	io.CopyN(io.Discard, resp.Body, maxBodySlurp)

	// Игнорируем ошибку закрытия (обычно это EOF или уже закрыто)
	_ = resp.Body.Close()
}

// метод проверки статуса ответа на запрос к API
func (p *BaseParser) checkResponseStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			return fmt.Errorf("API server error %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetName возвращает имя парсера
func (p *BaseParser) GetName() string {
	return p.name
}

// GetHTTPClient возвращает HTTP клиент
func (p *BaseParser) GetHTTPClient() *http.Client {
	return p.httpClient
}

func (p *BaseParser) GetHealthEndPoint() string {
	return p.healthEndPoint
}

// Отдельная функция с дженериками для определния : обычная ошибка или ошибка circuitBreaker
func handleCircuitBreakerErrorUniversal[T any](name string, cb interfaces.CBInterface, err error) (T, error) {
	var zero T

	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		tR, tS, tF := cb.GetStats()
		fmt.Printf("%s circuit breaker open - totalReq=%d, totalSuccess=%d, totalFailures=%d\n",
			name, tR, tS, tF)
		return zero, fmt.Errorf("%s is temporarily unavailable (circuit breaker open)", name)
	}
	return zero, fmt.Errorf("operation failed: %w", err)
}

// метод - обёртка для обработки ошибок. Выясняем это ошибки circuit breaker или внешние ошибки.
// для поиска списка вакнсий
func (p *BaseParser) handleCircuitBreakerErrorVacanciesSearch(err error) ([]models.Vacancy, error) {
	return handleCircuitBreakerErrorUniversal[[]models.Vacancy](p.name, p.circuitBreaker, err)
}

// метод - обёртка для обработки ошибок. Выясняем это ошибки circuit breaker или внешние ошибки.
// для поиска деталей по конкретной вакансии
func (p *BaseParser) handleCircuitBreakerErrorVacancyDetails(err error) (models.SearchVacancyDetailesResult, error) {
	return handleCircuitBreakerErrorUniversal[models.SearchVacancyDetailesResult](p.name, p.circuitBreaker, err)
}
