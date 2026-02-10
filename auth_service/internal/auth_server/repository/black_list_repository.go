package repository

import (
	"context"
	"fmt"
	redis "shared/redis"
	"strconv"
	"time"
)

// Специализированный интерфейс - наследует базовый + добавляет свои методы
type BlackListRepositoryInterface interface {
	// Встраиваем базовый интерфейс Redis
	redis.RedisRepositoryInterface // ← это ключевой момент!
	// Добавляем кастомные методы для работы с токенами
	AddToBlacklist(ctx context.Context, tokenJTI, tokenType, userID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenJTI string) (bool, error)
}

// структура репозитория для токенов
type tokenRepository struct {
	// Встраиваем базовый репозиторий
	redis.RedisRepositoryInterface // ← получаем все его методы "бесплатно"
	prefix                         string
}

// конструктор для репозитория токенов (возвращает интерфейс)
func NewTokenRepository(baseRepo redis.RedisRepositoryInterface, prefix string) (BlackListRepositoryInterface, error) {
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

// метод репозитория токенов для добавления токена в черный список
// ключ: blacklist:хэш_токена, значение: время истечения
func (t *tokenRepository) AddToBlacklist(ctx context.Context, tokenJTI, tokenType, userID string, ttl time.Duration) error {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return err
	}
	// Просто сохраняем токен с timestamp истечения
	key := fmt.Sprintf("blacklist:%s:%s", tokenType, tokenJTI)

	// Значение - время истечения в unix timestamp
	expiresAt := time.Now().UTC().Add(ttl).Unix()
	value := strconv.FormatInt(expiresAt, 10)

	err := t.Set(ctx, key, value, ttl)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// метод репозитория токенов для проверки, есть ли такой ключ в черном списке
func (t *tokenRepository) IsBlacklisted(ctx context.Context, tokenJTI string) (bool, error) {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return false, err
	}

	key := fmt.Sprintf("blacklist:%s", tokenJTI)

	// Проверяем существование ключа
	exists, err := t.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}
