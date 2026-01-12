// result_handler.go
package parsers_manager

import (
	"errors"
	"fmt"
	"parser/internal/circuitbreaker"
	"parser/internal/domain/models"
	"strings"
)

// метод определния типов ошибок, если они были. Или это ошибки circut breaker или другие, или их не было и можно вернуть результат
func (pm *ParsersManager) handleSearchResult(results []models.SearchVacanciesResult, err error, params models.SearchParams) ([]models.SearchVacanciesResult, error) {
	// Случай 1: Всё идеально
	if err == nil {
		return results, nil
	}

	// Случай 2: Частичный успех - есть результаты, но и есть ошибка
	if len(results) > 0 {
		// тут возможно логирование....
		// Проверяем, не связана ли ошибка с circuit breaker
		if pm.isCircuitBreakerError(err) {
			// Если circuit breaker открыт, но есть хоть какие-то результаты -
			// возвращаем их с warning
			return results, fmt.Errorf("частичные результаты (Parser manager circuit breaker): %w", err)
		}

		// Возвращаем результаты, которые удалось получить
		return results, err
	}

	// Случай 3: Полный провал - нет результатов вообще
	// Пробуем стратегии fallback
	return pm.tryFallbackStrategies(params, err)
}

// метод - попытка получить данные из кэша (тут понимаем, что это ошибка НЕ от circuit breaker)
func (pm *ParsersManager) tryFallbackStrategies(params models.SearchParams, originalErr error) ([]models.SearchVacanciesResult, error) {
	if results, found := pm.tryGetFromCache(params); found {
		msg := "данные из кэша"
		return results, fmt.Errorf("%s: %w", msg, originalErr)
	}

	return nil, fmt.Errorf("%s: %w", "Не удалось найти данные в кэше, ошибка: ", originalErr)
}

// метод определения ошибки. Это ошибка circuit breaker или нет
func (pm *ParsersManager) isCircuitBreakerError(err error) bool {
	return errors.Is(err, circuitbreaker.ErrCircuitOpen) ||
		errors.Is(err, circuitbreaker.ErrTooManyRequests) ||
		strings.Contains(err.Error(), "circuit breaker")
}
