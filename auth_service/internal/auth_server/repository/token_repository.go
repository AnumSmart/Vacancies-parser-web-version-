package repository

import (
	"context"
	"fmt"
	redis "shared/redis"
	"time"
)

// Специализированный интерфейс - наследует базовый + добавляет свои методы
type TokenRepositoryInterface interface {
	// Встраиваем базовый интерфейс Redis
	redis.RedisRepositoryInterface // ← это ключевой момент!
	// Добавляем кастомные методы для работы с токенами
	AddToBlacklist(ctx context.Context, token, userID string, ttl time.Duration) error
}

// структура репозитория для токенов
type tokenRepository struct {
	// Встраиваем базовый репозиторий
	redis.RedisRepositoryInterface // ← получаем все его методы "бесплатно"
	prefix                         string
}

// конструктор для репозитория токенов (возвращает интерфейс)
func NewTokenRepository(baseRepo redis.RedisRepositoryInterface, prefix string) (TokenRepositoryInterface, error) {
	// Проверяем обязательные зависимости
	if baseRepo == nil {
		return nil, fmt.Errorf("baseRepo cannot be nil")
	}

	// Проверяем префикс
	if prefix == "" {
		return nil, fmt.Errorf("prefix cannot be empty")
	}

	return &tokenRepository{
		RedisRepositoryInterface: baseRepo,
		prefix:                   prefix,
	}, nil
}

func (t *tokenRepository) AddToBlacklist(ctx context.Context, token, userID string, ttl time.Duration) error {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return err
	}
	// пока в разработке--------------------------------------------------------------------
	return nil
}
