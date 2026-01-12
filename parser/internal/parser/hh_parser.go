package parser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"parser/configs"
	"parser/internal/domain/models"
	"parser/internal/interfaces"
	"parser/internal/parser/model"
	"reflect"
	"strconv"
)

// создаём стркутуру парсера для HH.ru на базе общего парсера
type HHParser struct {
	*BaseParser
}

// конструктор для парсера HH.ru
func NewHHParser(cfg *configs.ParserInstanceConfig) (interfaces.Parser, error) {
	if cfg == nil {
		cfg = configs.DefaultParsersConfig().HH
	}

	baseCfg := BaseConfig{
		Name:                  "HH.ru",
		BaseURL:               cfg.BaseURL,
		HealthEndPoint:        cfg.HealthEndPoint,
		Timeout:               cfg.Timeout,
		RateLimit:             cfg.RateLimit,
		MaxConcurrent:         cfg.MaxConcurrent,
		CircuitBreakerCfg:     cfg.CircuitBreaker,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
	}

	baseParser, err := NewBaseParser(baseCfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка в конфигурации rate limiter для парсера %s\n", baseCfg.Name)
	}

	return &HHParser{
		BaseParser: baseParser,
	}, nil
}

// метод парсера для поиска списка вакансий
func (p *HHParser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
	return p.BaseParser.SearchVacancies(
		ctx,
		params,
		ParserFuncs{
			BuildURL: p.buildURL,
			Parse:    p.parseResponseSearchVacancies,
			Convert:  p.convertToUniversal,
		},
	)
}

// метод парсера для поиска деталей по конкретной вакансии
func (p *HHParser) SearchVacanciesDetailes(ctx context.Context, vacancyID string) (models.SearchVacancyDetailesResult, error) {
	fmt.Println("vacancy ID для поиска: ", vacancyID)
	return p.BaseParser.SearchVacancyDetailes(
		ctx,
		vacancyID,
		ParserFuncs{
			Parse:          p.parseResponseSearchDetails,
			ConvertDetails: p.convertDetails,
		},
	)
}

// buildURL строит URL для API запроса для поиска списка вакансий
func (p *HHParser) buildURL(params models.SearchParams) (string, error) {
	// преобразуем строку запроса в структуру URL
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	// заводим переменную, где будут хнаниться значения
	query := u.Query()

	// добавляем основной параметр поиска
	if params.Text != "" {
		query.Set("text", params.Text)
	}

	// добавляем параметр - локация
	if params.Location != "" {
		query.Set("location", params.Location)
	}

	// добавляем параетры страниц
	perPage := params.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20 // Значение по умолчанию
	}
	query.Set("per_page", strconv.Itoa(perPage))

	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}

	// формируем строку эндпоинта для запроса
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// метод парсера обработки тела запроса при поиске списка вакансий
func (p *HHParser) parseResponseSearchVacancies(body []byte) (interface{}, error) {
	var searchResponse model.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("[Parser name: %s] parse reaponse body - failed: %w", p.name, err)
	}
	return &searchResponse, nil
}

// метод парсера обработки тела запроса при поиске списка вакансий
func (p *HHParser) parseResponseSearchDetails(body []byte) (interface{}, error) {
	var searchResponse model.SearchDetails //--------------------------------------------------------------------???????
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("[Parser name: %s] parse reaponse body - failed: %w", p.name, err)
	}
	return &searchResponse, nil
}

// метод приведения результатов поиска у унифицированной структуре + проверка данных их интерфейса
func (p *HHParser) convertToUniversal(searchResponse interface{}) ([]models.Vacancy, error) {
	// проверка интерфейса на nil
	if searchResponse == nil {
		return []models.Vacancy{}, fmt.Errorf("[%s] searchResponse is nil", p.name)
	}

	// Проводим type assertion
	searchResp, ok := searchResponse.(*model.SearchResponse)
	if !ok {

		// Для более детальной информации можно использовать reflect
		fmt.Printf("----------------->>>[Parser name: %s] DEBUG: Type details: %v\n", p.name, reflect.TypeOf(searchResponse))
		return []models.Vacancy{}, fmt.Errorf("[Parser name: %s], wrong data type in the response body\n", p.name) // возвращаем пустой слайс
	}

	// Проверка на релевантность работы алгоритмов поиска на HH.ru
	// если запрос не достаточно точен, то могут быть найденные вакансии, но они не попадут в ответе в слайс items[]
	if len(searchResp.Items) == 0 && searchResp.Found > 0 {
		return []models.Vacancy{}, errors.New("Need to change querry request, not enought information for search") // Нет ошибки, просто нет данных
	}

	// сразу инициализируем слайс универсальных вакансий, чтобы уменьшить количество переаалокаций, если выйдем за размер базового массива слайса
	universalVacancies := make([]models.Vacancy, len(searchResp.Items))

	for i, hhvacancy := range searchResp.Items {
		// получаем строку-описания вилки зарплаты для каждой найденной записи
		salary := hhvacancy.GetSalaryString()

		universalVacancies[i] = models.Vacancy{
			ID:          hhvacancy.ID,
			Job:         hhvacancy.Name,
			Company:     hhvacancy.Employer.Name,
			Currency:    hhvacancy.Salary.Currency,
			Salary:      &salary,
			Location:    hhvacancy.Area.Name,
			URL:         hhvacancy.URL,
			Source:      p.GetName(),
			Description: hhvacancy.Description,
		}
	}

	return universalVacancies, nil
}

// метод приведения результатов поиска по конкретной ваансии к нужному типу + проверка данных интерфейса
func (p *HHParser) convertDetails(detailsResponse interface{}) (models.SearchVacancyDetailesResult, error) {
	// проверка интерфейса на nil
	if detailsResponse == nil {
		return models.SearchVacancyDetailesResult{}, fmt.Errorf("[%s] searchResponse is nil", p.name)
	}

	// Проводим type assertion
	searchResp, ok := detailsResponse.(*model.SearchDetails)
	if !ok {

		// Для более детальной информации можно использовать reflect
		fmt.Printf("----------------->>>[Parser name: %s] DEBUG: Type details: %v\n", p.name, reflect.TypeOf(searchResp))
		return models.SearchVacancyDetailesResult{}, fmt.Errorf("[Parser name: %s], wrong data type in the response body\n", p.name)
	}

	vacDetails := models.SearchVacancyDetailesResult{
		Employer:    models.Employer(searchResp.Employer),
		Location:    models.Area(searchResp.Area),
		Salary:      models.Salary(searchResp.Salary),
		Description: searchResp.Description,
		Name:        searchResp.Name,
		ID:          searchResp.ID,
		Url:         searchResp.Url,
	}

	return vacDetails, nil
}

// GetVacancyByID получает детальную информацию о вакансии по ID
func (p *HHParser) GetVacancyByID(vacancyID string) (*model.HHVacancy, error) {
	if vacancyID == "" {
		return nil, fmt.Errorf("vacancy ID cannot be empty")
	}

	apiURL := p.baseURL + "/" + vacancyID
	resp, err := p.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// если API - вернул ошибку, прерываем функцию
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var vacancy model.HHVacancy
	//анмаршалим успешное тело ответа в в нужную структуру
	if err := json.Unmarshal(body, &vacancy); err != nil {
		return nil, fmt.Errorf("parse HH-JSON failed: %w", err)
	}

	return &vacancy, nil
}

// получаем имя парсера
func (p *HHParser) GetName() string {
	return "HH.ru"
}
