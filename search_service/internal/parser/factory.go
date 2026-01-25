// описание структуры для создания механизма фабрики для парсеров
// на вход фабрики подаётся тип парсера, конфиг для нужного парсера и конструктор для нужного парсера
package parser

import (
	"fmt"
	"search_service/configs"
	"search_service/internal/search_interfaces"

	"sync"
)

//Фабрика с регистрацией парсеров

// ParserType тип парсера
type ParserType string

const (
	ParserTypeHH ParserType = "hh"
	ParserTypeSJ ParserType = "superjob"
	// можно добавить: ParserTypeRabotaRu ParserType = "rabota.ru"
)

// ParserConstructor функция-конструктор парсера
type ParserConstructor func(config *configs.ParserInstanceConfig) (search_interfaces.Parser, error)

// ParserFactory фабрика парсеров
type ParserFactory struct {
	constructors map[ParserType]ParserConstructor
	configs      map[ParserType]*configs.ParserInstanceConfig
	mu           sync.RWMutex
}

// NewParserFactory создает новую фабрику
func NewParserFactory() *ParserFactory {
	return &ParserFactory{
		constructors: make(map[ParserType]ParserConstructor),
		configs:      make(map[ParserType]*configs.ParserInstanceConfig),
	}
}

// Register регистрирует конструктор парсера и конфиг
func (f *ParserFactory) Register(parserType ParserType, config *configs.ParserInstanceConfig, constructor ParserConstructor) {
	// так как есть конкурентный доступ к мапе - делаем черезе мьютекс
	f.mu.Lock()
	defer f.mu.Unlock()

	f.constructors[parserType] = constructor
	f.configs[parserType] = config
}

// Create - создает парсер, если вся инфа до этого была зарегестрирована в фабрике
func (f *ParserFactory) Create(parserType ParserType) (search_interfaces.Parser, error) {
	f.mu.RLock()
	// под защитой мьютекса проверяем, есть ли в фабрике зарегестрированный конструктор для данного типа парсера
	constructor, ok := f.constructors[parserType]
	// под защитой мьютекса проверяем, есть ли в фабрике зарегестрированный конфиг для данного типа парсера
	config, configOk := f.configs[parserType]
	f.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("parser type not registered: %s", parserType)
	}
	if !configOk {
		return nil, fmt.Errorf("config not found for parser: %s", parserType)
	}
	return constructor(config)
}

// CreateEnabled создает только включенные парсеры (причем логика такая, что должны создаться только те у которых флаг enabled: true)
// если хоть 1 из таких парсеров - не создан (была ошибка) - то и этот метод вернёт ошибку
func (f *ParserFactory) CreateEnabled(enabled []ParserType) ([]search_interfaces.Parser, error) {

	// проверяем, если вообще нет парсеров с разрешённым флагом
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled parsers specified")
	}

	parsers := make([]search_interfaces.Parser, len(enabled))

	for i, parserType := range enabled {
		parser, err := f.Create(parserType)
		if err != nil {
			return nil, fmt.Errorf("failed to create parser %s: %w", parserType, err)
		}
		parsers[i] = parser
	}

	return parsers, nil
}
