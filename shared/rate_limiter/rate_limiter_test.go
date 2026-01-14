package rate_limiter

import (
	"context"
	"shared/interfaces"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	rate = 50 * time.Millisecond
)

// проверяем создание и использование rate limiter
func TestNewChannelRateLimiter(t *testing.T) {
	// проеряем, что конструктор создаёт правлиный rate limiter, что внутренний канал - инициализируется
	t.Run("creates with valid rate", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)
		defer rl.Stop()

		assert.NotNil(t, rl)
		assert.NotNil(t, rl.limiter)
	})

	// проверяем, что экземпляр rate limiter соответствует интерфейсу interfaces.RateLimiter
	t.Run("implements RateLimiter interface", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(time.Second)
		assert.NoError(t, err)
		defer rl.Stop()

		var _ interfaces.RateLimiter = rl
	})
}

// проверяем корректность интервалов врмени у метода Wait(ctx)
// для одиночных запросов, последовательных и асинхронных
func TestChannelRateLimiter_Wait(t *testing.T) {
	// проверяем, что первый запрос через rate limiter проходит за величину времени тикера + погрешность
	t.Run("allows first request immediately", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)
		defer rl.Stop()

		ctx := context.Background()

		start := time.Now()
		err = rl.Wait(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)                            // проверяем, что нет ошибок при выполнении функции rl.Wait(ctx)
		assert.Less(t, elapsed, rate+10*time.Millisecond) // Проверяем, что время работы функции - корректно
	})

	// проверяем, что несколько запросов подряд - выполняются не дольше указанного времени
	t.Run("respects rate limit", func(t *testing.T) {
		rate := 100 * time.Millisecond
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)
		defer rl.Stop()

		ctx := context.Background()

		// Первый запрос
		start := time.Now()
		err = rl.Wait(ctx)
		assert.NoError(t, err)

		// Второй запрос должен ждать
		err = rl.Wait(ctx)
		assert.NoError(t, err)
		elapsed := time.Since(start) // измеряем время 2х запросов

		// Проверяем, что общее время >= 2*rate
		assert.GreaterOrEqual(t, elapsed, 2*rate)

		// И не слишком больше (допуск 20ms)
		assert.Less(t, elapsed, 2*rate+20*time.Millisecond)
	})

	// проверяем, что запросы которые будут идти конкурентно через rate limiter будут выполняться не дольше указанного времени
	t.Run("multiple waits respect rate", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)
		defer rl.Stop()

		const numRequests = 5

		startTimes := make([]time.Time, numRequests)
		endTimes := make([]time.Time, numRequests)
		errors := make([]error, numRequests)

		var wg sync.WaitGroup
		var startMu sync.Mutex

		// Создаем барьер для одновременного старта всех горутин
		startBarrier := make(chan struct{})

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Ждем сигнала на одновременный старт
				<-startBarrier

				// Фиксируем время старта (без блокировки других горутин)
				startMu.Lock()
				startTimes[idx] = time.Now()
				startMu.Unlock()

				// Вызываем Wait БЕЗ мьютекса - это ключевой момент!
				err := rl.Wait(context.Background())

				// Фиксируем время завершения
				startMu.Lock()
				endTimes[idx] = time.Now()
				errors[idx] = err
				startMu.Unlock()
			}(i)
		}

		// Одновременно запускаем все горутины
		close(startBarrier)

		wg.Wait()

		// Проверяем ошибки
		for _, err := range errors {
			assert.NoError(t, err)
		}

		// Сортируем времена завершения (горутины могли завершиться в разном порядке)
		sort.Slice(endTimes, func(i, j int) bool {
			return endTimes[i].Before(endTimes[j])
		})

		// Проверяем интервалы между завершениями
		for i := 1; i < len(endTimes); i++ {
			interval := endTimes[i].Sub(endTimes[i-1])

			// Должно быть не меньше rate (с погрешностью)
			// Первый интервал может быть меньше, если rate limiter разрешает первый запрос сразу
			if i == 1 {
				// Первый запрос мог быть обработан сразу
				assert.GreaterOrEqual(t, interval, rate-10*time.Millisecond)
			} else {
				assert.GreaterOrEqual(t, interval, rate-5*time.Millisecond)
			}

			// Не должно быть слишком больших задержек сверх rate
			assert.Less(t, interval, rate+30*time.Millisecond)
		}

		// Дополнительная проверка: общее время должно быть примерно (n-1)*rate
		totalTime := endTimes[numRequests-1].Sub(endTimes[0])
		expectedTotal := time.Duration(numRequests-1) * rate
		assert.InDelta(t,
			float64(expectedTotal),
			float64(totalTime),
			float64(50*time.Millisecond)) // Погрешность
	})
}

// проверяем механизм остановки rate limiter
func TestChannelRateLimiter_Stop(t *testing.T) {
	// штатная остановка rate limiter
	t.Run("stops rate limiter", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)

		ctx := context.Background()

		// Делаем один успешный запрос
		err = rl.Wait(ctx)
		assert.NoError(t, err)

		// Останавливаем
		rl.Stop()

		// Даём время на остановку горутины
		time.Sleep(20 * time.Millisecond)

		// Последующие запросы должны возвращать ошибку
		err = rl.Wait(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stopped")
	})

	// проверяем, что повторный останов - не паникует
	t.Run("idempotent stop", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)

		ctx := context.Background()

		rl.Stop()
		rl.Stop() // Двойной вызов не должен паниковать
		rl.Stop()

		err = rl.Wait(ctx)
		assert.Error(t, err)
	})

	// проверяем, что метод Wait(ctx) будет возвращать ошибку, если rate limiter остановлен
	t.Run("wait after stop returns error", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(rate)
		assert.NoError(t, err)
		rl.Stop()

		ctx := context.Background()

		err = rl.Wait(ctx)
		assert.Error(t, err)
		assert.Equal(t, "rate limiter stopped", err.Error())
	})
}

// проверяем конкурентное использование rale limiter
// если несколько горутин делают несколько запросов к rate limiter
func TestChannelRateLimiter_Concurrency(t *testing.T) {
	rl, err := NewChannelRateLimiter(rate)
	assert.NoError(t, err)
	defer rl.Stop()

	ctx := context.Background()

	goroutineCount := 10
	reqPerGoroutine := 5
	totalReq := goroutineCount * reqPerGoroutine

	timeStamps := make(chan time.Time, totalReq)
	errors := make(chan error, totalReq)

	var wg sync.WaitGroup

	// запускаем горутины
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// в каждой горутине запускаем запросы через rate limiter
			for j := 0; j < reqPerGoroutine; j++ {
				err := rl.Wait(ctx)
				timeStamps <- time.Now()
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)
	close(timeStamps)

	// Проверяем ошибки
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
		assert.NoError(t, err) // проверяем, что просто есть ошибка
	}
	assert.Equal(t, 0, errorCount) // проверяем, что ошибок ровно 0

	// Собираем временные метки
	var ts []time.Time
	for timestamp := range timeStamps {
		ts = append(ts, timestamp)
	}

	// Проверяем что не было одновременных успешных запросов
	// (минимум rate между последовательными)
	for i := 1; i < len(ts); i++ {
		interval := ts[i].Sub(ts[i-1])
		assert.GreaterOrEqual(t, interval, rate-10*time.Millisecond)
	}
}

// проверяем краевые случаи
// очень быстрые запросы, нулевой rate и  отрицательный rate
func TestChannelRateLimiter_EdgeCases(t *testing.T) {
	// проверяем несколько очень быстрых запросов через rate limiter
	t.Run("very fast rate", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(1 * time.Millisecond)
		assert.NoError(t, err)
		defer rl.Stop()

		ctx := context.Background()
		// Несколько быстрых запросов
		for i := 0; i < 5; i++ {
			err := rl.Wait(ctx)
			assert.NoError(t, err)
		}
	})

	// проверяем, если передали в качестве rate - ноль!
	t.Run("zero rate", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(0)

		// Проверяем, что есть ошибка
		assert.Error(t, err)
		assert.EqualError(t, err, "Rate must be greater than zero")

		// Проверяем, что объект не создан
		assert.Nil(t, rl)

		// НЕ вызываем Stop() - объекта нет!
	})

	// проверяем, если передали в качестве rate - отрицательное значение!
	t.Run("negative rate", func(t *testing.T) {
		rl, err := NewChannelRateLimiter(-10 * time.Millisecond)

		// Проверяем, что есть ошибка
		assert.Error(t, err)
		assert.EqualError(t, err, "Rate must be greater than zero")

		// Проверяем, что объект не создан
		assert.Nil(t, rl)

		// НЕ вызываем Stop() - объекта нет!
	})
}
