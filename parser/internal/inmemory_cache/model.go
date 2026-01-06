package inmemory_cache

import (
	"sync"
	"time"
)

// основная структура inmemory cache для кэширования результатов поиска. Кэш - шардирован
type InmemoryShardedCache struct {
	shards    []*Shard
	numShards int
	stopChan  chan bool
}

// структура отдельного шарда
// у него есть мапа с CashItems и мьютекс для доступа к мапе
// ключем в этой мапе будет строка ----> хэшированный запрос поиска
type Shard struct {
	Items map[string]CashItem
	mu    sync.RWMutex
}

// структура отдельного элемента inmemory cache
type CashItem struct {
	//value   []models.SearchResult
	value   interface{}
	expTime time.Time
}
