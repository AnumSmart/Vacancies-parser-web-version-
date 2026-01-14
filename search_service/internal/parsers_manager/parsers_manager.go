// описание структуры мэнеджера парсеров и его конструктора
package parsers_manager

import (
	"errors"
	"math"
	"search_service/configs"
	"search_service/internal/search_interfaces"
	"shared/circuitbreaker"
	"shared/inmemory_cache"
	"shared/interfaces"
	"shared/queue"
	"sync"
	"time"
)

// структура менеджера парсеров
type ParsersManager struct {
	parsers              []search_interfaces.Parser             // парсеры, которыми оперирует мэнеджер
	config               *configs.Config                        // общий конфиг
	SearchCache          *inmemory_cache.InmemoryShardedCache   // поисковый кэш
	VacancyIndex         *inmemory_cache.InmemoryShardedCache   // кэш для обратного индекса
	vacancyDetails       *inmemory_cache.InmemoryShardedCache   // кэш для деталей вакансии
	parsersStatusManager search_interfaces.ParsersStatusManager // менеджер сотсояний парверов внутри менеджера
	circuitBreaker       interfaces.CBInterface                 // глобальный circut breaker (используем интерфейс)

	// Поля для управления нагрузкой --------------------------------------------------------------------------
	semaphore          chan struct{}                                        // Семафор для ограничения одновременных запросов
	jobSearchQueue     interfaces.FIFOQueueInterface[search_interfaces.Job] // Очередь заданий (в качестве типа используем интерфейс с дженеником)
	workers            int                                                  // Количество воркеров
	stopWorkers        chan struct{}                                        // Сигнал остановки воркеров (когда захотим завершить все воркеры - зкрываем канал)
	semaSlotGetTimeout time.Duration                                        // таймаут ожидания свободного слота глобального семафора менеджера парсеров
	wg                 sync.WaitGroup                                       // Для graceful shutdown
	mu                 sync.RWMutex                                         // Для потокобезопасности
	// --------------------------------------------------------------------------------------------------------
}

// структура параметров для системы управления нагрузкой менеджера парсеров
type PMLoad struct {
	numOfWorkers   int           // количество воркеров для очереди менеджера парсеров
	semaphoreSize  int           // размер глобального семафора для мэнеджера парсеров
	queueSize      int           // размер очереди для менеджера парсеров
	semSlotTimeout time.Duration // таймаут ожидания свододного слота семафора для воркера
}

// конструктор для системы управления нагрузкой менеджера парсеров
func NewPMLoad(numCPUCores int) *PMLoad {
	numOfWorkers := numCPUCores * 2                                // рассчитываем количество воркеров относительно доступных ядер на машине
	semaphoreSize := int(math.Ceil(0.7 * (float64(numOfWorkers)))) // рассчитываем размер семафора (70-80% от количества workers)
	queueSize := semaphoreSize * 3                                 // рассчитываем размер очереди: размер семафора * 3

	return &PMLoad{
		numOfWorkers:   numOfWorkers,
		semaphoreSize:  semaphoreSize,
		queueSize:      queueSize,
		semSlotTimeout: 200 * time.Millisecond, // таймаут для получения свободного слота у семафора отдельным воркером
	}
}

// Конструктор для мэнеджера парсинга из разных источников
func NewParserManager(config *configs.Config,
	numCPUCores int,
	searchCache *inmemory_cache.InmemoryShardedCache,
	vacancyIndex *inmemory_cache.InmemoryShardedCache,
	vacancyDetails *inmemory_cache.InmemoryShardedCache,
	pStatManager search_interfaces.ParsersStatusManager,
	parsers ...search_interfaces.Parser) (*ParsersManager, error) {

	// запускаем конструктор параметров для управления нагрузкой в парсер менеджере
	pmLoad := NewPMLoad(numCPUCores)

	// Валидация
	if pStatManager == nil {
		return nil, errors.New("ParsersStatusManager обязателен")
	}
	if len(parsers) == 0 {
		return nil, errors.New("нужен хотя бы один парсер")
	}
	if searchCache == nil || vacancyIndex == nil || vacancyDetails == nil {
		return nil, errors.New("кэши обязательны")
	}

	pm := &ParsersManager{
		parsers:              parsers,
		config:               config,
		SearchCache:          searchCache,    // кэш для поиска
		VacancyIndex:         vacancyIndex,   // кэш для обратного индекса
		vacancyDetails:       vacancyDetails, // кэш для деталей отдельной вакансии
		parsersStatusManager: pStatManager,
		circuitBreaker:       circuitbreaker.NewCircutBreaker(config.Manager.CircuitBreakerCfg),
		workers:              pmLoad.numOfWorkers,
		semaphore:            make(chan struct{}, pmLoad.semaphoreSize),
		jobSearchQueue:       queue.NewFIFOQueue[search_interfaces.Job](pmLoad.queueSize), // создаём очередь через конструктор
		stopWorkers:          make(chan struct{}),
		semaSlotGetTimeout:   pmLoad.semSlotTimeout,
		// wg и mu автоматически инициализируются нулевыми значениями
	}

	// Запускаем воркеры для обработки очереди
	pm.startSearchWorkers()

	return pm, nil
}

// GetAllParsers возвращает список доступных парсеров
func (pm *ParsersManager) GetParserNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}
