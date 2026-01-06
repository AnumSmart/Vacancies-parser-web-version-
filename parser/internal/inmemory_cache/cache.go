package inmemory_cache

import (
	"fmt"
	"hash/fnv"
	"log"
	"time"
)

// конструктор для создания кэша с указаным количеством шардов и интервалом очистки кэша
func NewInmemoryShardedCache(numShards int, cleanUpInterval time.Duration) (*InmemoryShardedCache, error) {
	// Валидация входных параметров
	if numShards <= 0 {
		return nil, fmt.Errorf("numShards must be positive, got %d", numShards)
	}

	if cleanUpInterval < 0 {
		return nil, fmt.Errorf("cleanUpInterval must be non-negative, got %v", cleanUpInterval)
	}

	// Дополнительно можно проверять очень большие значения
	if numShards > 1000 { // или другое разумное ограничение
		return nil, fmt.Errorf("numShards is too large: %d", numShards)
	}

	// инициализируем базовую структуру кэша
	cache := &InmemoryShardedCache{
		shards:    make([]*Shard, numShards), // инициализируем слайс указателй на шарды
		numShards: numShards,                 // указывавем количество шардов
		stopChan:  make(chan bool),           // инициализируем канал для остановки
	}

	// для каждого шарда инициализируем внутреннюю мапу
	for i := 0; i < numShards; i++ {
		cache.shards[i] = &Shard{
			Items: map[string]CashItem{},
		}
	}

	// асинхронно запускаем метод очистки кэша через поределённый интервал времени
	// Запускаем очистку только если интервал > 0
	if cleanUpInterval > 0 {
		go cache.cleanUp(cleanUpInterval)
	}

	return cache, nil
}

// метод получения значения из кэша по заданному ключу (это хэшированный запрос поиска)
// чтобы реализовать этот метод - нужна функция, которая будет находить нужный шард по заданному ключу (внутри будет хэш-функция)
// результатом будет значение в CashItem и флаг
func (c *InmemoryShardedCache) GetItem(key string) (interface{}, bool) {
	// получаем необходимый шард
	shard := c.getShard(key)
	now := time.Now()
	// лочимся на чтение, так как читаем из мапы
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, ok := shard.Items[key] // проверяем наличие значения по ключу в мапе шарда
	if !ok {
		//fmt.Printf("No data in cache with key:%s\n", key) -- место для логирования
		return nil, false
	}
	// проверяем, не истёк ли TTL у значения
	if now.After(val.expTime) {
		fmt.Printf("Data in cache with key:%s are not valid\n", key)
		return nil, false
	}

	return val.value, true
}

// метод, чтобы находить нужный шард по заданному ключу
func (c *InmemoryShardedCache) getShard(key string) *Shard {
	// создаём экземпляр хэша
	hashf := fnv.New32a()
	//записываем в хэш наш ключ в виде байтового среза
	_, err := hashf.Write([]byte(key))
	if err != nil {
		log.Println(err.Error())
	}
	//вычисляем индекс нужного нам шарда по ключу
	// если мы хэш по ключу % количество шардов = индекс шарда в диапазоне от 0 до shardNum-1
	// для каждого ключа будет вой шард, там мы будем рапределять данные по шардам
	shardIndex := int(hashf.Sum32()) % c.numShards

	return c.shards[shardIndex]
}

// метод, чтобы записать значение в кэш с заданным TTL
func (c *InmemoryShardedCache) AddItemWithTTL(key string, value interface{}, ttl time.Duration) {
	// получаем необходимый шард
	shard := c.getShard(key)
	now := time.Now()

	// берём лок на запись, так как обращаемся к мапе
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.Items[key] = CashItem{
		value:   value,
		expTime: now.Add(ttl), // время жизни для нового занчения - высчитывавем: время на момоент вызова функции + ttl
	}
}

// метод удаления элемента из кэша по ключу
func (c *InmemoryShardedCache) DeleteItem(key string) {
	for _, shard := range c.shards {
		shard.mu.Lock()
		for key := range shard.Items {
			if key == key {
				delete(shard.Items, key)
			}
		}
		shard.mu.Unlock()
	}
}
