package repository

import (
	"context"
	"fmt"
	"log"
	"shared/config"
	"time"

	"github.com/go-redis/redis/v8"
)

// интерфейс базовых возможностей redis (для общего пользования из других сервисов)
type RedisRepositoryInterface interface {
	// Основные CRUD операции
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetBytes(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// TTL операции
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Управление соединением
	Close() error
}

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(cfg *config.RedisConfig) (RedisRepositoryInterface, error) {
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

// метод для добавления значения с TTL в redis
func (r *RedisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// метод получения значения из redis по ключу
func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// метод получения значения из redis по ключу (результат в виде байтового среза)
func (r *RedisRepository) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

// метод удаления элемента по ключу из redis
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// метод проверки существования элемента в redis по ключу
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// метод устанавливает время жизни ключа в Redis.
func (r *RedisRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// метод возвращает оставшееся время жизни ключа в Redis.
// Возвращает -1 если время жизни не установлено, -2 если ключ не существует.
func (r *RedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}
