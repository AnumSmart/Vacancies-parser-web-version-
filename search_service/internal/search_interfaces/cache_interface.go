package search_interfaces

import "time"

type CacheInterface interface {
	GetItem(key string) (interface{}, bool)
	AddItemWithTTL(key string, value interface{}, ttl time.Duration)
	DeleteItem(key string)
}
