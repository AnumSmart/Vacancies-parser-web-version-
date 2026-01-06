// описание мэжнеджера состояния всех парсеров
// агрегирует состояния всех парсеров для дальнейшего использования этой информации в глобальном circuit breaker
package parsers_status_manager

import (
	"context"
	"fmt"
	"parser/configs"
	"parser/internal/interfaces"
	"sync"
	"time"
)

// ParserStatusManager управляет статусами всех парсеров
// ключ - это имя экземпляра парсера
type ParserStatusManager struct {
	parsersStats map[string]*interfaces.ParserStatus // мапа статусов парсеров
	config       *configs.Config                     // конфиг
	client       interfaces.HealthClient             // клиент для проверки
	initComplete chan struct{}                       // Сигнал завершения инициализации
	stopChan     chan struct{}
	mu           sync.RWMutex
	wg           sync.WaitGroup
}

// конструктор для нового менеджера статусов парсеров
func NewParserStatusManager(conf *configs.Config, parsers ...interfaces.Parser) *ParserStatusManager {
	psm := &ParserStatusManager{
		parsersStats: make(map[string]*interfaces.ParserStatus),
		config:       conf, // конфиг для коиента health check
		client:       NewHttpHealthCheckClient(conf.HealthChech),
		initComplete: make(chan struct{}),
		stopChan:     make(chan struct{}),
	}

	// инициализируем статусы парсеров в менеджере статусов
	for _, parser := range parsers {
		psm.parsersStats[parser.GetName()] = &interfaces.ParserStatus{
			Name:           parser.GetName(),
			LastCheck:      time.Now(),
			Initialized:    false,
			CircuitState:   "closed",
			HealthEndpoint: parser.GetHealthEndPoint(),
			IsHealthy:      false, // Начальное состояние - не здоров
		}
	}

	// Запускаем фоновую горутину для опроса
	psm.startHealthChecker()
	return psm
}

// метод менеджера состояний парсера, который запускает опрос парсеров для получения их состояния
func (psm *ParserStatusManager) startHealthChecker() {
	psm.wg.Add(1)

	go func() {
		defer psm.wg.Done()

		// Выполняем первую синхронную проверку при старте
		psm.performHealthCheck()

		// Сигнализируем, что инициализация завершена
		close(psm.initComplete)

		// запускаем тикер, который через определённое время будет запускать проверку состояния парсеров
		ticker := time.NewTicker(psm.config.HealthChech.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// так как в этом методе запускаются горутины по количеству парсеров
				// и внутри своя waitgroup, значит этот шаг будет блокирующим, пока не завешатся все опросы
				// и не будут обновлены все данные о состояниях
				psm.performHealthCheck()
			case <-psm.stopChan:
				return
			}
		}
	}()
}

// метод, который конкурентно опрашивает сервисы на предмет проверки состояния
func (psm *ParserStatusManager) performHealthCheck() {
	// создаём внутреннюю структуру для сбора результатов инициализации
	type checkResult struct {
		name     string // имя парсера
		healthy  bool   // статус текущего состояния
		initDone bool   // статус, что прошёл первую инициализацию
		err      error  // ошибка для отладки
	}

	// создаём буфферизированный канал, куда будем складывать результаты опросов статусов парсеров
	results := make(chan checkResult, len(psm.parsersStats))
	//необходима внутренняя waitgroup для согласования горутин
	var wg sync.WaitGroup

	psm.mu.RLock()
	for name, status := range psm.parsersStats {
		wg.Add(1)
		go func(n string, endpoint string) {
			defer wg.Done()

			_, healthy, err := psm.client.CheckHealth(context.Background(), endpoint)

			// Определяем статус инициализации
			initDone := true // После первой проверки считаем инициализированным

			// Отправляем ОДИН результат
			results <- checkResult{
				name:     n,
				healthy:  healthy && err == nil, // Здоров только если healthy=true И нет ошибки
				initDone: initDone,
				err:      err,
			}

		}(name, status.HealthEndpoint)
	}
	psm.mu.RUnlock()

	// Закрываем канал после завершения всех проверок
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты
	for result := range results {
		psm.mu.Lock()
		if parser, exists := psm.parsersStats[result.name]; exists {
			parser.IsHealthy = result.healthy
			parser.LastCheck = time.Now()
			parser.Initialized = result.initDone
		}
		psm.mu.Unlock()

	}
	// Все проверки завершены, все статусы обновлены

}

// UpdateStatus обновляет статус парсера в менеджере статуса парсеров (потокобезопасен, есть мьютекс внутри)
// этот метод вызывается отдельно от периодического опроса (допустим, после успешного похода парсера во внешний сервис)
func (psm *ParserStatusManager) UpdateStatus(name string, success bool, err error) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	psm.mu.Lock()
	defer psm.mu.Unlock()

	status, exists := psm.parsersStats[name] // пытаемся получить статус парсера по ключу
	// если его нету, то добавляем новый в менеджер статуса парсеров
	if !exists {
		status = &interfaces.ParserStatus{
			Name:        name,
			Initialized: true,
			LastCheck:   time.Now(),
		}
		psm.parsersStats[name] = status
	}

	status.LastCheck = time.Now()

	if success {
		status.SuccessCount++
		status.ErrorCount = 0
		status.IsHealthy = true
		status.LastError = nil
	} else {
		status.ErrorCount++
		status.SuccessCount = 0
		status.IsHealthy = false
		status.LastError = err
	}
}

// GetHealthyParsers возвращает список здоровых парсеров
func (psm *ParserStatusManager) GetHealthyParsers() []string {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	psm.mu.Lock()
	defer psm.mu.Unlock()

	var healthy []string

	for name, status := range psm.parsersStats {
		if status.IsHealthy && time.Since(status.LastCheck) < 5*time.Minute {
			// проверяем, что статус парсера IsHeathy==true,Lastcheck бы не позднее 5 мин
			healthy = append(healthy, name)
		}

	}
	return healthy
}

// GetStatus возвращает статус конкретного парсера
func (psm *ParserStatusManager) GetParserStatus(name string) (*interfaces.ParserStatus, bool) {
	// так как мэнеджер статуса парсеров основан на мапе, все панипуляции проводит под мьютексом
	psm.mu.RLock()
	defer psm.mu.RUnlock()

	status, exists := psm.parsersStats[name]
	if !exists {
		return nil, false
	}

	return status, true
}

// Метод для остановки менеджера
func (psm *ParserStatusManager) Stop() {
	close(psm.stopChan) // Сигнализируем остановку
	psm.wg.Wait()       // Ждем завершения всех горутин
	fmt.Println("Parsers status manager was stopped correctly")
}
