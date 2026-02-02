package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"shared/config"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRepositoryInterface interface {
	// нужно реализовать!
	/*
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error // добавить значение в redis с поределённым ключом
		Get(ctx context.Context, key string) (string, error)                                    // получить значение по ключу
		Exists(ctx context.Context, key string) (bool, error)                                   // проверить существование значения в reddis по ключу
		Delete(ctx context.Context, key string) error                                           // удалить значение по ключу
		Close() error
	*/
}

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(cfg *config.RedisConfig) (*RedisRepository, error) {
	// проверяем, что конфиг редиса не nil
	if cfg == nil {
		return nil, fmt.Errorf("Error in redis config")
	}

	// создаёем экземпляр опций, на базе которых построим клиента
	redisOptions := cfg.ToRedisOptions()

	// создаём клиента redis на основе опций *redis.Options, которые получились из клиента redis
	client := redis.NewClient(redisOptions)

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Printf("Connected to Redis at %s (DB: %d)", redisOptions.Addr, redisOptions.DB)

	return &RedisRepository{
		client: client,
	}, nil
}

// метод для завершения работы экземпляра redis
func (r *RedisRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Set добавляет значение в Redis с определенным ключом
// Поддерживает любые типы данных через JSON сериализацию
func (r *RedisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var err error
	var finalValue string

	// Преобразуем value в строку в зависимости от типа
	switch v := value.(type) {
	case string:
		finalValue = v
	case []byte:
		finalValue = string(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		finalValue = fmt.Sprintf("%d", v)
	case float32, float64:
		finalValue = fmt.Sprintf("%f", v)
	case bool:
		finalValue = fmt.Sprintf("%t", v)
	default:
		// Для сложных структур используем JSON
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON: %w", err)
		}
		finalValue = string(jsonData)
	}

	// Устанавливаем значение с TTL
	if expiration > 0 {
		err = r.client.SetEX(ctx, key, finalValue, expiration).Err()
	} else {
		err = r.client.Set(ctx, key, finalValue, 0).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}
