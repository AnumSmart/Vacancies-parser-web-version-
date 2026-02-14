// описание и инициализация всех общих зависимостей
package core

import (
	"fmt"
	"runtime"
	"search_service/configs"
	"search_service/internal/parser"
	"search_service/internal/parsers_manager"
	"search_service/internal/parsers_status_manager"
	"search_service/internal/search_interfaces"
	"search_service/internal/search_server/handlers"
	"search_service/internal/search_server/service"
	"shared/inmemory_cache"
)

// SearchServiceDependencies содержит все общие зависимости
type SearchServiceDependencies struct {
	Config              *configs.SearchServiceConfig
	SearchCache         search_interfaces.CacheInterface
	VacancyIndex        search_interfaces.CacheInterface
	VacancyDetails      search_interfaces.CacheInterface
	ParserFactory       *parser.ParserFactory
	ParserStatusManager *parsers_status_manager.ParserStatusManager
	ParserManager       *parsers_manager.ParsersManager
	SearchHandler       *handlers.SearchHandler
}

// InitDependencies инициализирует общие зависимости для search_service
func InitDependencies() (*SearchServiceDependencies, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	// Получаем конфигурацию
	conf, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	//создаём экземпляр inmemory cache для результатов поиска вакансий
	searchCache, err := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.SearchCacheConfig.SearchCacheCleanUp)
	if err != nil {
		return nil, fmt.Errorf("failed to create search cache: %w", err)
	}

	//создаём экземпляр inmemory cache для обратного индекса для вакансий
	vacancyIndex, err := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.VacancyCacheConfig.VacancyCacheCleanUp)
	if err != nil {
		return nil, fmt.Errorf("failed to create vacancy index cache: %w", err)
	}

	// создаём экземпляр inmemory cache для деталей конкретной вакансии (ключ: ID вакансии)
	vacancyDetails, err := inmemory_cache.NewInmemoryShardedCache(conf.Cache.NumOfShards, conf.Cache.VacancyCacheConfig.VacancyCacheCleanUp)
	if err != nil {
		return nil, fmt.Errorf("failed to create vacancy details cache: %w", err)
	}

	//создаём фабрику парсеров
	parserFactory := parser.NewParserFactory()

	// регистрируем парсеры в фабрике
	// НЕ ВЫЗЫВАЕМ функцию, а передаем ее как значение!
	parserFactory.Register("hh", conf.Parsers.HH, parser.NewHHParser)
	parserFactory.Register("superjob", conf.Parsers.SuperJob, parser.NewSJParser)

	// создаём список парсеров для создания (пока хард-код, но в будущем это будут переменные)
	enabledParsers := []parser.ParserType{"hh", "superjob"}

	// создаём только те парсеры, у которых в конфиге указано Enabled
	parsers, err := parserFactory.CreateEnabled(enabledParsers)
	if err != nil {
		return nil, fmt.Errorf("failed to create enabled parsers: %w", err)
	}

	// создаём мэнеджера состояния парсеров и инициализируем начальными значениями
	parserStatusManager := parsers_status_manager.NewParserStatusManager(conf, parsers...)

	// Создаём менеджер парсеров
	parserManager, err := parsers_manager.NewParserManager(conf, currentMaxProcs, searchCache, vacancyIndex, vacancyDetails, parserStatusManager, parsers...)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser manager: %w", err)
	}

	// создаём поисковый сервис
	searchService := service.NewSearchService(parserManager)

	// создаём хэндлер поиска
	searchHandler := handlers.NewSearchHandler(searchService)

	// возвращаем указатель на структуру зависимостей
	return &SearchServiceDependencies{
		Config:              conf,
		SearchCache:         searchCache,
		VacancyIndex:        vacancyIndex,
		VacancyDetails:      vacancyDetails,
		ParserFactory:       parserFactory,
		ParserStatusManager: parserStatusManager,
		ParserManager:       parserManager,
		SearchHandler:       searchHandler,
	}, nil
}
