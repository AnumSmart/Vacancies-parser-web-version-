// описываем конструкторы для разных типов джоб
package parsers_manager

import (
	"context"
	"fmt"
	"search_service/internal/domain/models"
	"search_service/internal/jobs"
	"search_service/internal/search_interfaces"
	"search_service/pkg"
	"time"
)

// newSearchJob - создает джобу для поиска вакансий
func (pm *ParsersManager) newSearchJob(params models.SearchParams) *jobs.SearchJob {
	return &jobs.SearchJob{
		BaseJob: jobs.BaseJob{
			ID:         pkg.QuickUUID(),
			ResultChan: make(chan *jobs.JobOutput, 1), // обязательно - буферизированный канал
			CreatedAt:  time.Now(),
		},
		Params: params,
	}
}

// NewFetchVacancyJob - создает джобу для получения деталей вакансии
func (pm *ParsersManager) NewFetchVacancyJob(source, vacancyID string) *jobs.FetchDetailsJob {
	return &jobs.FetchDetailsJob{
		BaseJob: jobs.BaseJob{
			ID:         pkg.QuickUUID(),
			ResultChan: make(chan *jobs.JobOutput, 1), // обязательно - буферизированный канал
			CreatedAt:  time.Now(),
		},
		Source:    source,
		VacancyID: vacancyID,
	}
}

// метод для добавления джобы в очередь, с возможностью повторных попыток в течение таймаута
func (pm *ParsersManager) tryEnqueueJob(ctx context.Context, job search_interfaces.Job, timeout time.Duration) bool {

	start := time.Now()

	for {
		// Пробуем добавить в очередь
		if pm.jobSearchQueue.Enqueue(job) {
			return true
		}

		// Проверяем таймаут
		if time.Since(start) > timeout {
			return false
		}

		// Проверяем отмену контекста
		select {
		case <-ctx.Done():
			return false
		default:
			// Небольшая пауза перед следующей попыткой
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// метод - обёртка, чтобы внутри вызвать функцию-джерерик для нужного типа (тип: список вакансий)
func (pm *ParsersManager) waitForJobSearchVacansiesResult(ctx context.Context, resChan <-chan *jobs.JobOutput, timeout time.Duration) ([]models.SearchVacanciesResult, error) {
	// вызываем дженерик с нужным типом
	result, err := waitJobResult[[]models.SearchVacanciesResult](ctx, resChan, timeout)

	// не проверяем ошибки, так как ошибки и краевые случаи уже обработаны в функции-дженерике
	return result, err
}

// метод - обёртка, чтобы внутри вызвать функцию-джерерик для нужного типа (тип: список вакансий)
func (pm *ParsersManager) waitForJobSearchVacancyDeyailsResult(ctx context.Context, resChan <-chan *jobs.JobOutput, timeout time.Duration) (models.SearchVacancyDetailesResult, error) {
	// вызываем дженерик с нужным типом
	result, err := waitJobResult[models.SearchVacancyDetailesResult](ctx, resChan, timeout)

	// не проверяем ошибки, так как ошибки и краевые случаи уже обработаны в функции-дженерике
	return result, err
}

// функуция на базе дженериков, для получения результатов
func waitJobResult[T any](ctx context.Context, resChan <-chan *jobs.JobOutput, timeout time.Duration) (T, error) {
	// объявляем нулевое значние переменной типа T
	var zero T

	select {
	case result, ok := <-resChan:
		// проверяем открыт ли канал
		if !ok {
			return zero, fmt.Errorf("канал результата закрыт\n")
		}

		// проверяем наличие ошибки
		if result.Error != nil {
			return zero, result.Error
		}

		// проводим type assertion
		typedResult, ok := result.Data.(T)
		if !ok {
			return zero, fmt.Errorf("неверный тип результата\n")
		}

		return typedResult, nil

		// проверяем таймаут
	case <-time.After(timeout):
		return zero, fmt.Errorf("таймаут выполнения поиска\n")

		// проверяем контекст
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}
