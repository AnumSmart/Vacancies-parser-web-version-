// этот раздел отвечает за добавление задачи по поиску списка вакансий в очередь
// получение результатов из очереди
// описание метода, которым пользуется воркер, если определил джобу по поиску вакансий
package parsers_manager

import (
	"context"
	"fmt"
	"search_service/internal/domain/models"
	"time"
)

// метод менджера парсеров, который формирует джобу для поиска списка вакансий, добавляет эту джобу в очередь и получает результат поиска в канал
// возвращает результат поиска или ошибку
func (pm *ParsersManager) SearchVacancies(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {
	// создаём новую джобу необходимого типа (в данном случае джоба поиска списка вакансий)
	job := pm.newSearchJob(params)

	// Пытаемся добавить в очередь с таймаутом и повторными попытками
	success := pm.tryEnqueueJob(ctx, job, 5*time.Second)

	// проверяем успешность добавления в очередь
	if !success {
		return []models.SearchVacanciesResult{}, fmt.Errorf("❌ Джоба не была добавлена в очередь")
	}

	// дожидаемся результатов из очереди с учётом таймаута
	result, err := pm.waitForJobSearchVacansiesResult(ctx, job.ResultChan, 30*time.Second)

	// специально тут не обрабатываем ошибку, они уже обработаны выше
	return result, err
}

// Основная логика поиска списка вакансий по всем доступным парсерам
func (pm *ParsersManager) executeSearch(ctx context.Context, params models.SearchParams) ([]models.SearchVacanciesResult, error) {

	// Проверяем кэш
	if cachedResults, found := pm.tryGetFromCache(params); found {
		// Только возвращаем кэшированные данные
		// Статус парсеров не трогаем — они не участвовали
		return cachedResults, nil
	}

	// Получаем список парсеров для использования
	parsersToUse := pm.selectParsersForSearch()
	if len(parsersToUse) == 0 {
		return nil, fmt.Errorf("❌ Нет доступных парсеров для поиска")
	}

	// Выполняем поиск через парсеры
	searchResults, err := pm.searchWithParsers(ctx, params, parsersToUse)

	if err != nil {
		return nil, fmt.Errorf("❌ Конкурентный поиск по парсерам - не удался!")
	}

	// Фильтруем результаты: берем только успешные, т.е. те, у которых в models.SearchResult.Error == nil
	successfulResults := pm.filterSuccessfulResults(searchResults)

	// Кэшируем только если есть хотя бы один успешный результат
	if len(successfulResults) > 0 {
		pm.cacheSearchResults(params, successfulResults)
	} else {
		// Ни один парсер не вернул результатов
		// НЕ кэшируем, пробуем снова при следующем запросе
	}

	return searchResults, nil // успех для глобального CB и получение данных парсинга
}

// Формируем слайс стркутур, где поиск прошёл без ошибок
func (pm *ParsersManager) filterSuccessfulResults(results []models.SearchVacanciesResult) []models.SearchVacanciesResult {
	var successful []models.SearchVacanciesResult
	for _, result := range results {
		if result.Error == nil && len(result.Vacancies) > 0 {
			successful = append(successful, result)
		}
	}
	return successful
}
