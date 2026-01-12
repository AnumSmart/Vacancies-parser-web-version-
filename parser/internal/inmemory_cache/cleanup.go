package inmemory_cache

import "time"

// метод для вызова интервальной очистки кэша или его остановки
func (c *InmemoryShardedCache) cleanUp(ttl time.Duration) {

	// создаём тикер, который буедт через интервал времени ttl посылать в свой канал ticker.C текущую дату
	ticker := time.NewTicker(ttl)
	// останавливаем тикер по выходу из функции
	defer ticker.Stop()

	// в этом селекте будем ждать одно из 2х событий
	// 1. читаем из канала тикера --> запускаем метод очистки устаревших записей из кэша
	select {
	case <-ticker.C:
		c.cleanUpExpired()
	// 2. читаем из stopChan самого кэша, это значит мы останавливаем кэш и останавливаем логику очистки
	case <-c.stopChan:
		return
	}
}

//метод для очистки кэша от устаревших данных
func (c *InmemoryShardedCache) cleanUpExpired() {
	// создаём переменную, в которой будет содержаться текущее время на момент вызова этой функции
	start := time.Now()
	// пробегаемся циклом по всм шардам
	for _, shard := range c.shards {
		shard.mu.Lock()
		for key, value := range shard.Items {
			// если текущее время - это время после времени жизни элемента кэша, то удаляем его
			if start.After(value.expTime) {
				delete(shard.Items, key)
			}
		}
		shard.mu.Unlock()
	}
}
