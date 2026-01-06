package inmemory_cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// проверяем конструктор
func TestNewInmemoryShardedCache(t *testing.T) {
	// Проверяем создание кэша с разным количеством шардов
	tests := []struct {
		name            string
		numShards       int
		cleanUpInterval time.Duration
		wantErr         bool
	}{
		{"valid cache", 8, time.Minute, false},
		{"single shard", 1, time.Second, false},
		{"zero shards", 0, time.Minute, true}, // должно падать
		{"negative shards", -1, time.Minute, true},
		{"zero interval", 8, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			cache, err := NewInmemoryShardedCache(tt.numShards, tt.cleanUpInterval)
			if tt.numShards <= 0 && err != nil {
				assert.EqualError(t, err, fmt.Sprintf("numShards must be positive, got %d", tt.numShards))
			}

			if tt.cleanUpInterval < 0 && err != nil {
				assert.EqualError(t, err, fmt.Sprintf("cleanUpInterval must be non-negative, got %v", tt.cleanUpInterval))
			}

			if cache == nil && !tt.wantErr {
				t.Error("expected cache, got nil")
			}
			if cache != nil && len(cache.shards) != tt.numShards {
				t.Errorf("expected %d shards, got %d", tt.numShards, len(cache.shards))
			}
		})
	}
}

// проверяем распределение по шардам
func TestGetShardDistribution(t *testing.T) {
	cache, _ := NewInmemoryShardedCache(4, time.Minute)

	testKeys := []string{"key1", "key2", "key3", "key4", "key5"}
	shardMap := make(map[int][]string)

	// Проверяем распределение ключей
	for _, key := range testKeys {
		shard := cache.getShard(key)
		shardIndex := -1
		for i, s := range cache.shards {
			if s == shard {
				shardIndex = i
				break
			}
		}
		shardMap[shardIndex] = append(shardMap[shardIndex], key)
	}

	// Проверяем что ключи распределились (не все в один шард)
	if len(shardMap) == 1 && len(testKeys) > 1 {
		t.Error("all keys mapped to single shard, bad distribution")
	}

	// Проверяем детерминированность распределения
	for _, key := range testKeys {
		shard1 := cache.getShard(key)
		shard2 := cache.getShard(key)
		if shard1 != shard2 {
			t.Errorf("shard distribution not deterministic for key %s", key)
		}
	}
}

func TestCacheOperations(t *testing.T) {
	cache, _ := NewInmemoryShardedCache(4, time.Hour)

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := "test-value"

		cache.AddItemWithTTL(key, value, time.Minute)

		got, ok := cache.GetItem(key)
		if !ok {
			t.Error("expected to find key in cache")
		}
		if got != value {
			t.Errorf("expected %v, got %v", value, got)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, ok := cache.GetItem("non-existent")
		if ok {
			t.Error("expected not to find key in cache")
		}
	})

	t.Run("Overwrite value", func(t *testing.T) {
		key := "same-key"
		cache.AddItemWithTTL(key, "value1", time.Minute)
		cache.AddItemWithTTL(key, "value2", time.Minute)

		got, ok := cache.GetItem(key)
		if !ok || got != "value2" {
			t.Error("value not overwritten correctly")
		}
	})
}
