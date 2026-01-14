package parsers_manager

import (
	"context"
	"fmt"
	"search_service/internal/domain/models"
)

// метод получения списка парсеров, доступных для поиска, согласно статусам в мэнеджере состояния парсеров
func (pm *ParsersManager) selectParsersForSearch() []string {
	// Сначала берем здоровые парсеры
	healthyParsers := pm.getHealthyParsers()
	if len(healthyParsers) > 0 {
		fmt.Printf("нашлись здоровые парсеры: %v\n", healthyParsers)
		fmt.Printf("Количество здоровых парсеров: %v\n", len(healthyParsers))
		return healthyParsers
	}

	// Если все парсеры нездоровы, берем все
	fmt.Println("⚠️  Все парсеры в нездоровом состоянии, пробуем перезапуск...")
	fmt.Printf("доступные парсеры с неизвестным состоянием: %v\n", pm.getAllParsersNames())
	return pm.getAllParsersNames()
}

// метод, который позволит асинхронно провести поиск по заданным параметрам среди списка переданных парсеров, учитывается контектс с таймаутом
func (pm *ParsersManager) searchWithParsers(ctx context.Context, params models.SearchParams, parserNames []string) ([]models.SearchVacanciesResult, error) {
	searchCtx, cancel := context.WithTimeout(ctx, pm.config.API.ConcSearchTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("❌ Таймайут конкурентного поиска: %v\n", ctx.Err())
	default:
		return pm.concurrentSearchWithTimeout(searchCtx, params, parserNames)
	}

}
