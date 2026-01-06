package circuitbreaker

import (
	"sync/atomic"
	"time"
)

// Execute выполняет операцию с защитой Circuit Breaker
// Execute выполняет операцию с защитой Circuit Breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Обрабатываем все состояния
	switch cb.state {
	// состояние, когда CB - открыт (заблокирован)
	case StateOpen:
		// проверяем таймер
		if time.Since(cb.lastFailureTime) < cb.resetTimeout {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		// Переходим в Half-Open
		cb.state = StateHalfOpen
		atomic.StoreUint32(&cb.halfOpenAttempts, 0)
		atomic.StoreUint32(&cb.successes, 0)
		// Продолжаем выполнение как Half-Open
		fallthrough // если доходим до сюда, то сразу проваливаемся в условие StateHalfOpen

	// состояние, когда CB - полуоткрыт (пробный режим)
	case StateHalfOpen:
		// Атомарно проверяем и резервируем слот
		attempts := atomic.AddUint32(&cb.halfOpenAttempts, 1)

		// Если превысили лимит - отменяем резервирование
		if attempts > cb.halfOpenMaxRequests {
			atomic.AddUint32(&cb.halfOpenAttempts, ^uint32(0)) // откат
			cb.mu.Unlock()
			return ErrTooManyRequests
		}

		// Увеличиваем общие счетчики
		atomic.AddUint32(&cb.totalRequests, 1)

		// Сохраняем номер текущей попытки
		currentAttempt := attempts
		cb.mu.Unlock()

		// Выполняем операцию для Half-Open
		err := fn()

		cb.mu.Lock()
		defer cb.mu.Unlock()

		// Проверяем, не изменилось ли состояние во время выполнения
		if cb.state != StateHalfOpen {
			// Откатываем резервирование, так как состояние изменилось
			atomic.AddUint32(&cb.halfOpenAttempts, ^uint32(0))

			// Обновляем счетчики
			if err != nil {
				atomic.AddUint32(&cb.totalFailures, 1)
			} else {
				atomic.AddUint32(&cb.totalSuccesses, 1)
			}
			return err
		}

		// Дополнительная проверка на случай race condition
		if currentAttempt > cb.halfOpenMaxRequests {
			atomic.AddUint32(&cb.halfOpenAttempts, ^uint32(0))
			if err != nil {
				atomic.AddUint32(&cb.totalFailures, 1)
			} else {
				atomic.AddUint32(&cb.totalSuccesses, 1)
			}
			return err
		}

		if err != nil {
			atomic.AddUint32(&cb.totalFailures, 1)
			cb.onFailure()
			return err
		}

		atomic.AddUint32(&cb.totalSuccesses, 1)
		cb.onSuccess()
		return nil

	// состояние, когда CB - закрыт (работает нормально)
	case StateClosed:
		atomic.AddUint32(&cb.totalRequests, 1)
		cb.mu.Unlock()

		// Выполняем операцию для Closed
		err := fn()

		cb.mu.Lock()
		defer cb.mu.Unlock()

		if err != nil {
			atomic.AddUint32(&cb.totalFailures, 1)
			cb.onFailure()
		} else {
			atomic.AddUint32(&cb.totalSuccesses, 1)
			cb.onSuccess()
		}
		return err
	}

	// Этот код никогда не выполнится, но компилятор требует return
	return nil
}

// onFailure обрабатывает неудачное выполнение
func (cb *CircuitBreaker) onFailure() {
	// проверяем статус circuit breaker
	// мьютекс уже захвачен вызывающим кодом
	switch cb.state {
	// если circuit breaker - закрыт (работает нормально)
	case StateClosed:
		atomic.AddUint32(&cb.failures, 1)
		if atomic.LoadUint32(&cb.failures) >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
			// Сбрасываем счетчик ошибок при переходе в Open
			atomic.StoreUint32(&cb.failures, 0)
		}

	// если circuit breaker - полуоткрыт (пробный режим)
	case StateHalfOpen:
		// При ошибке в Half-Open возвращаемся в Open
		cb.state = StateOpen
		cb.lastFailureTime = time.Now()
		atomic.StoreUint32(&cb.halfOpenAttempts, 0)
		atomic.StoreUint32(&cb.successes, 0)

	// если circuit breaker - открыт (заблокирован)
	case StateOpen:
		// В Open состоянии счетчики не обновляются
		// Можно добавить логирование или метрики
	}
}

// onSuccess обрабатывает удачное выполнение
func (cb *CircuitBreaker) onSuccess() {
	// проверяем статус circuit breaker
	// мьютекс уже захвачен вызывающим кодом
	switch cb.state {
	// если circuit breaker - закрыт (работает нормально)
	case StateClosed:
		// Сбрасываем счетчик ошибок при успешном выполнении
		atomic.StoreUint32(&cb.failures, 0)

	// если circuit breaker - полуоткрыт (пробный режим)
	case StateHalfOpen:
		successes := atomic.AddUint32(&cb.successes, 1)
		if successes >= cb.successThreshold {
			// Переходим в Closed состояние
			cb.state = StateClosed
			atomic.StoreUint32(&cb.failures, 0)
			atomic.StoreUint32(&cb.successes, 0)
			atomic.StoreUint32(&cb.halfOpenAttempts, 0)
		}

	// если circuit breaker - открыт (заблокирован)
	case StateOpen:
		// В Open состоянии не должно быть успешных вызовов
		// Это может указывать на логическую ошибку
	}
}

// GetState возвращает текущее состояние
func (cb *CircuitBreaker) getState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats возвращает статистику
func (cb *CircuitBreaker) GetStats() (total, success, failure uint32) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.totalRequests, cb.totalSuccesses, cb.totalFailures
}
