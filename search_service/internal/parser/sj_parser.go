package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"search_service/configs"
	"search_service/internal/domain/models"
	"search_service/internal/parser/model"
	"search_service/internal/search_interfaces"
	"strconv"
)

// создаём стркутуру парсера для SuperJob.ru на базе общего парсера
type SJParser struct {
	*BaseParser
}

// конструктор для парсера SuperJob.ru
func NewSJParser(cfg *configs.ParserInstanceConfig) (search_interfaces.Parser, error) {
	if cfg == nil {
		cfg = configs.DefaultParsersConfig().SuperJob
	}

	baseCfg := BaseConfig{
		Name:                  "SuperJob.ru",
		BaseURL:               cfg.BaseURL,
		HealthEndPoint:        cfg.HealthEndPoint,
		APIKey:                cfg.APIKey,
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

	return &SJParser{
		BaseParser: baseParser,
	}, nil
}

// метод парсера для поиска списка вакансий
func (p *SJParser) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.Vacancy, error) {
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
func (p *SJParser) SearchVacanciesDetailes(ctx context.Context, vacancyID string) (models.SearchVacancyDetailesResult, error) {
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
func (p *SJParser) buildURL(params models.SearchParams) (string, error) {
	// преобразуем строку запроса в структуру URL
	u, err := url.Parse(p.baseURL)
	if err != nil {
		return "", err
	}

	// заводим переменную, где будут хнаниться значения
	query := u.Query()

	// добавляем основной параметр поиска
	if params.Text != "" {
		query.Set("keyword", params.Text)
	}

	// добавляем параметр - локация
	if params.Country != "" {
		query.Set("country", p.convertArea(params.Country))
	}

	// добавляем параетры страниц
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page-1)) // SuperJob использует 0-based
	}

	// формируем строку эндпоинта для запроса
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// метод парсера обработки тела запроса
func (p *SJParser) parseResponseSearchVacancies(body []byte) (interface{}, error) {
	var searchResponse model.SuperJobResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("[Parser name: %s] parse reaponse body - failed: %w", p.name, err)
	}
	return &searchResponse, nil
}

// метод парсера обработки тела запроса при поиске списка вакансий
func (p *SJParser) parseResponseSearchDetails(body []byte) (interface{}, error) {
	var searchResponse model.SearchDetails //--------------------------------------------------------------------???????
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("[Parser name: %s] parse reaponse body - failed: %w", p.name, err)
	}
	return &searchResponse, nil
}

// метод приведения результатов поиска по конкретной ваансии к нужному типу + проверка данных интерфейса
func (p *SJParser) convertDetails(detailsResponse interface{}) (models.SearchVacancyDetailesResult, error) {
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

	var vacDetails models.SearchVacancyDetailesResult
	vacDetails.Description = searchResp.Description

	return vacDetails, nil
}

// метод приведения результатов поиска у унифицированной структуре + проверка данных их интерфейса
func (p *SJParser) convertToUniversal(searchResponse interface{}) ([]models.Vacancy, error) {
	// Проводим type assertion
	searchResp, ok := searchResponse.(*model.SuperJobResponse)
	if !ok {

		// Для более детальной информации можно использовать reflect
		fmt.Printf("----------------->>>[Parser name: %s] DEBUG: Type details: %v\n", p.name, reflect.TypeOf(searchResponse))
		return nil, fmt.Errorf("[Parser name: %s], wrong data type in the response body\n", p.name)
	}

	// сразу инициализируем слайс универсальных вакансий, чтобы уменьшить количество переаалокаций, если выйдем за размер базового массива слайса
	universalVacancies := make([]models.Vacancy, len(searchResp.Items))

	for i, sjv := range searchResp.Items {
		salary := sjv.GetSalaryString()
		universalVacancies[i] = models.Vacancy{
			ID:          strconv.Itoa(sjv.ID),
			Job:         sjv.Profession,
			Company:     sjv.FirmName,
			Currency:    sjv.Currency,
			Salary:      &salary,
			Location:    sjv.Town.Title,
			URL:         sjv.Link,
			Source:      p.GetName(),
			Description: sjv.VacancyRichText,
		}
	}
	return universalVacancies, nil
}

// метод для конвертации локации
func (p *SJParser) convertArea(area string) string {
	// Конвертируем коды регионов HH.ru в названия SuperJob
	areas := map[string]string{
		"1": "Москва",
		"2": "Санкт-Петербург",
	}
	if name, ok := areas[area]; ok {
		return name
	}
	return ""
}

// GetVacancyByID получает детальную информацию о вакансии по ID
func (p *SJParser) GetVacancyByID(vacancyID string) (*model.SJVacancy, error) {
	// Реализация получения деталей вакансии по ID
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

	//анмаршалим успешное тело ответа в в нужную структуру
	var vacancy model.SJVacancy
	if err := json.Unmarshal(body, &vacancy); err != nil {
		return nil, fmt.Errorf("parse SJ-JSON failed: %w", err)
	}

	return &vacancy, nil
}
