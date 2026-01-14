// parser_executor.go
package parsers_manager

import (
	"context"
	"fmt"
	"search_service/internal/domain/models"
	"search_service/internal/search_interfaces"
	"sync"
	"time"
)

// concurrentSearchWithTimeout выполняет поиск во всех парсерах одновременно с таймаутом
func (pm *ParsersManager) concurrentSearchWithTimeout(ctx context.Context, params models.SearchParams, parsers []string) ([]models.SearchVacanciesResult, error) {

	var wg sync.WaitGroup
	// создаём переменную для результатов
	results := make(chan models.SearchVacanciesResult, len(parsers))

	// получаем список экземпляров "живых парсеров"
	aliveParsers := pm.getAliveParsers(parsers)

	// получаем хэш для поиска
	searchHash, err := genHashFromSearchParam(params)
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка при генерации поискового хэша: %v\n", err)
	}

	for _, parser := range aliveParsers {
		wg.Add(1)
		go func(p search_interfaces.Parser) {
			defer wg.Done()

			// Создаем канал для результата и создаём ещё одну горутину, где производим поиск
			// 2я - горутина нужна, чтобы потом использовать select и контролировать отмену контекста
			resultChan := make(chan models.SearchVacanciesResult, 1)

			go func() {
				start := time.Now()
				vacancies, err := p.SearchVacancies(ctx, params)
				duration := time.Since(start)

				// обновляем статус парсера, в зависимости от результата поиска
				if err != nil {
					pm.parsersStatusManager.UpdateStatus(p.GetName(), false, err)
				} else {
					pm.parsersStatusManager.UpdateStatus(p.GetName(), true, nil)
				}

				resultChan <- models.SearchVacanciesResult{
					ParserName: p.GetName(),
					Vacancies:  vacancies,
					SearchHash: searchHash,
					Error:      err,
					Duration:   duration,
				}
			}()

			select {
			case <-ctx.Done():
				// Таймаут или отмена
				results <- models.SearchVacanciesResult{
					ParserName: p.GetName(),
					Error:      fmt.Errorf("timeout exceeded"),
				}
			case result := <-resultChan:
				results <- result
			}
		}(parser)
	}

	// в этой горутине дожидаемся окончания обработки от всех парсеров и закрываем канал результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// обьявляем переменную для выходных данных
	var searchResults []models.SearchVacanciesResult

	for result := range results {
		searchResults = append(searchResults, result)
	}

	return searchResults, nil
}

// метод получения списка "живых" парсеров, согласно менеджеру состояния парсеров
func (pm *ParsersManager) getAliveParsers(names []string) []search_interfaces.Parser {
	var result []search_interfaces.Parser

	for _, name := range names {
		for _, parser := range pm.parsers {
			if parser.GetName() == name {
				result = append(result, parser)
			}
		}
	}

	return result
}
