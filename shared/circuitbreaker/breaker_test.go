package circuitbreaker

import (
	"errors"
	"shared/config"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewCircuitBreaker проверяет создание Circuit Breaker с дефолтными значениями
func TestNewCircuitBreaker(t *testing.T) {
	// проверяем создание circuit breaker с пустым конфигом
	t.Run("with empty config", func(t *testing.T) {
		config := config.NewCircuitBreakerConfig(0, 0, 0, 0, 0)
		cb := NewCircutBreaker(config)

		if cb == nil {
			t.Fatal("CircuitBreaker should not be nil")
		}

		if cb.failureThreshold != 5 {
			t.Errorf("Expected failureThreshold 5, got %d", cb.failureThreshold)
		}

		if cb.successThreshold != 3 {
			t.Errorf("Expected successThreshold 3, got %d", cb.successThreshold)
		}

		if cb.halfOpenMaxRequests != 2 {
			t.Errorf("Expected halfOpenMaxRequests 2, got %d", cb.halfOpenMaxRequests)
		}

		if cb.resetTimeout != 10*time.Second {
			t.Errorf("Expected resetTimeout 10s, got %v", cb.resetTimeout)
		}

		if cb.windowDuration != 10*time.Second {
			t.Errorf("Expected windowDuration 10s, got %v", cb.windowDuration)
		}

		if cb.state != StateClosed {
			t.Errorf("Expected initial state Closed, got %d", cb.state)
		}
	})

	// проверяем создание circuit breaker с кастомным конфигом
	t.Run("with custom config", func(t *testing.T) {
		config := config.NewCircuitBreakerConfig(
			10,             // failureThreshold
			5,              // successThreshold
			3,              // halfOpenMaxRequests
			5*time.Second,  // resetTimeout
			30*time.Second, // windowDuration
		)
		cb := NewCircutBreaker(config)

		if cb.failureThreshold != 10 {
			t.Errorf("Expected failureThreshold 10, got %d", cb.failureThreshold)
		}

		if cb.successThreshold != 5 {
			t.Errorf("Expected successThreshold 5, got %d", cb.successThreshold)
		}

		if cb.halfOpenMaxRequests != 3 {
			t.Errorf("Expected halfOpenMaxRequests 3, got %d", cb.halfOpenMaxRequests)
		}

		if cb.resetTimeout != 5*time.Second {
			t.Errorf("Expected resetTimeout 5s, got %v", cb.resetTimeout)
		}

		if cb.windowDuration != 30*time.Second {
			t.Errorf("Expected windowDuration 30s, got %v", cb.windowDuration)
		}
	})
}

// TestExecuteSuccess проверяет успешное выполнение в закрытом состоянии
func TestExecuteSuccess(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		2,                    // failureThreshold
		3,                    // successThreshold
		2,                    // halfOpenMaxRequests
		100*time.Millisecond, // resetTimeout
		1*time.Second,        // windowDuration
	)
	cb := NewCircutBreaker(config)

	counter := 0
	err := cb.Execute(func() error {
		counter++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if counter != 1 {
		t.Errorf("Expected function to be called once, got %d", counter)
	}

	if state := cb.getState(); state != StateClosed {
		t.Errorf("Expected state Closed, got %d", state)
	}
}

// TestExecuteFailure проверяет обработку ошибок
func TestExecuteFailure(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		2,
		3,
		2,
		100*time.Millisecond,
		1*time.Second,
	)
	cb := NewCircutBreaker(config)

	testErr := errors.New("test error")
	err := cb.Execute(func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	total, success, failure := cb.GetStats()
	if total != 1 || success != 0 || failure != 1 {
		t.Errorf("Expected stats (1,0,1), got (%d,%d,%d)", total, success, failure)
	}
}

// TestStateTransitionClosedToOpen проверяет переход из Closed в Open
func TestStateTransitionClosedToOpen(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		2,                    // failureThreshold
		3,                    // successThreshold
		2,                    // halfOpenMaxRequests
		100*time.Millisecond, // resetTimeout
		1*time.Second,        // windowDuration
	)
	cb := NewCircutBreaker(config)

	// Две ошибки подряд
	for i := 0; i < 2; i++ {
		err := cb.Execute(func() error {
			return errors.New("error")
		})
		if err == nil {
			t.Error("Expected error")
		}
	}

	// После второго вызова должен быть StateOpen
	if state := cb.getState(); state != StateOpen {
		t.Errorf("Expected state Open, got %d", state)
	}

	// Третий вызов должен вернуть ErrCircuitOpen
	err := cb.Execute(func() error {
		t.Error("Function should not be called in Open state")
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

// TestStateTransitionOpenToHalfOpen проверяет переход из Open в Half-Open
func TestStateTransitionOpenToHalfOpen(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		1,                   // failureThreshold
		2,                   // successThreshold
		2,                   // halfOpenMaxRequests
		50*time.Millisecond, // resetTimeout
		1*time.Second,       // windowDuration
	)
	cb := NewCircutBreaker(config)

	// Переводим в Open
	cb.Execute(func() error {
		return errors.New("error")
	})

	if state := cb.getState(); state != StateOpen {
		t.Errorf("Expected state Open, got %d", state)
	}

	// Ждем resetTimeout
	time.Sleep(60 * time.Millisecond)

	// Следующий вызов должен перевести в Half-Open
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if state := cb.getState(); state != StateHalfOpen {
		t.Errorf("Expected state HalfOpen, got %d", state)
	}
}

// TestStateTransitionHalfOpenToClosed проверяет переход из Half-Open в Closed
func TestStateTransitionHalfOpenToClosed(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		1,                   // failureThreshold
		2,                   // successThreshold
		5,                   // halfOpenMaxRequests
		50*time.Millisecond, // resetTimeout
		1*time.Second,       // windowDuration
	)
	cb := NewCircutBreaker(config)

	// Переводим в Open, затем ждем и выполняем успешные запросы
	cb.Execute(func() error {
		return errors.New("error")
	})

	time.Sleep(60 * time.Millisecond)

	// Два успешных запроса в Half-Open
	for i := 0; i < 2; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error on attempt %d, got %v", i, err)
		}
	}

	// После второго успешного запроса должен быть Closed
	if state := cb.getState(); state != StateClosed {
		t.Errorf("Expected state Closed, got %d", state)
	}
}

// TestStateTransitionHalfOpenToOpen проверяет переход из Half-Open в Open при ошибке
func TestStateTransitionHalfOpenToOpen(t *testing.T) {
	config := config.NewCircuitBreakerConfig(
		1,                   // failureThreshold
		2,                   // successThreshold
		5,                   // halfOpenMaxRequests
		50*time.Millisecond, // resetTimeout
		1*time.Second,       // windowDuration
	)
	cb := NewCircutBreaker(config)

	// Переводим в Open
	cb.Execute(func() error {
		return errors.New("error")
	})

	time.Sleep(60 * time.Millisecond)

	// Ошибка в Half-Open состоянии
	err := cb.Execute(func() error {
		return errors.New("half-open error")
	})

	if err == nil {
		t.Error("Expected error")
	}

	// Должен вернуться в Open
	if state := cb.getState(); state != StateOpen {
		t.Errorf("Expected state Open, got %d", state)
	}
}

// TestHalfOpenMaxRequestsConcurrent проверяет ограничение запросов в Half-Open при конкурентном доступе
func TestHalfOpenMaxRequestsConcurrent(t *testing.T) {
	// Используем очень маленький resetTimeout для быстрого перехода в Half-Open
	config := config.NewCircuitBreakerConfig(
		1,                   // failureThreshold - одна ошибка переводит в Open
		2,                   // successThreshold - два успеха для перехода в Closed
		3,                   // halfOpenMaxRequests - максимум 3 запроса в Half-Open
		50*time.Millisecond, // resetTimeout
		1*time.Second,       // windowDuration
	)
	cb := NewCircutBreaker(config)

	// 1. Переводим Circuit Breaker в Open состояние
	err := cb.Execute(func() error {
		return errors.New("initial error")
	})
	if err == nil {
		t.Fatal("Expected error to trigger Open state")
	}

	// Ждем resetTimeout чтобы перейти в Half-Open
	time.Sleep(60 * time.Millisecond)

	// 2. Запускаем множество горутин, которые пытаются выполнить запросы
	const numGoroutines = 10
	var wg sync.WaitGroup

	// Атомарные счетчики для отслеживания результатов
	var (
		successCount   uint32
		tooManyErrors  uint32
		circuitErrors  uint32
		functionCalled uint32
	)

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			err := cb.Execute(func() error {
				// Увеличиваем счетчик вызовов функции
				atomic.AddUint32(&functionCalled, 1)

				// Первые 2 запроса возвращают успех, остальные - ошибку
				// для тестирования перехода в Closed
				if atomic.LoadUint32(&successCount) < 2 {
					// Симулируем небольшую задержку
					time.Sleep(time.Millisecond)
					return nil
				}
				return errors.New("function error")
			})

			// Классифицируем результат
			switch {
			case err == nil:
				atomic.AddUint32(&successCount, 1)
			case err == ErrTooManyRequests:
				atomic.AddUint32(&tooManyErrors, 1)
			case err == ErrCircuitOpen:
				atomic.AddUint32(&circuitErrors, 1)
			default:
				// Ошибка из функции
			}
		}(i)
	}

	wg.Wait()

	// 3. Проверяем результаты
	t.Logf("Results: success=%d, tooManyErrors=%d, circuitErrors=%d, functionCalled=%d",
		atomic.LoadUint32(&successCount),
		atomic.LoadUint32(&tooManyErrors),
		atomic.LoadUint32(&circuitErrors),
		atomic.LoadUint32(&functionCalled))

	// Проверяем основные инварианты:

	// a) Функция должна быть вызвана не более чем halfOpenMaxRequests раз
	if atomic.LoadUint32(&functionCalled) > uint32(config.HalfOpenMaxRequests) {
		t.Errorf("Function called %d times, but halfOpenMaxRequests is %d",
			atomic.LoadUint32(&functionCalled), config.HalfOpenMaxRequests)
	}

	// b) Должны быть ошибки ErrTooManyRequests
	if atomic.LoadUint32(&tooManyErrors) == 0 {
		t.Error("Expected at least one ErrTooManyRequests in concurrent scenario")
	}

	// c) Общее количество запросов = количество горутин
	// (все должны получить какой-то результат)
	totalResults := atomic.LoadUint32(&successCount) + atomic.LoadUint32(&tooManyErrors) + atomic.LoadUint32(&circuitErrors)
	if totalResults != uint32(numGoroutines) {
		t.Errorf("Not all goroutines got results: expected %d, got %d",
			numGoroutines, totalResults)
	}

	// d) Проверяем состояние Circuit Breaker после всех операций
	state := cb.getState()
	if atomic.LoadUint32(&successCount) >= uint32(config.SuccessThreshold) {
		// Если было достаточно успехов, должен быть Closed
		if state != StateClosed {
			t.Errorf("Expected state Closed after %d successes, got %d",
				atomic.LoadUint32(&successCount), state)
		}
	} else {
		// Иначе должен остаться в Half-Open или вернуться в Open если была ошибка
		if atomic.LoadUint32(&functionCalled) > 2 {
			// Если функция вызывалась больше 2 раз, могла быть ошибка
			// которая перевела бы в Open
			if state != StateOpen && state != StateHalfOpen {
				t.Errorf("Expected state Open or HalfOpen, got %d", state)
			}
		}
	}
}
