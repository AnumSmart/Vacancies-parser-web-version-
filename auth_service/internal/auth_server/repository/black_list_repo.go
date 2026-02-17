package repository

import (
	authinterfaces "auth_service/internal/auth_interfaces"
	"context"
	"fmt"
	"global_models/global_cache"
	"strconv"
	"time"
)

// создаём репозиторий кэша (тут редис) на базе глобального интерфейса

// Реализуем ТОЛЬКО то, что нужно auth_service
type AuthBlackListRepository struct {
	blackCache global_cache.Cache // создаём на базе глобального интерфейса
	prefix     string
}

// конструктор для репозитория черного списка (использует интерфейс для глобального кэша)
func NewBlackListRepo(cache global_cache.Cache, prefix string) (authinterfaces.BlackListRepository, error) {
	// Проверяем обязательные зависимости
	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}

	// Проверяем префикс
	if prefix == "" {
		return nil, fmt.Errorf("prefix cannot be empty")
	}
	return &AuthBlackListRepository{
		blackCache: cache,
		prefix:     prefix,
	}, nil
}

// Добавление в черный список
// ключ: blacklist:тип токена:JTI токена, значение: время истечения
func (b *AuthBlackListRepository) AddToBlacklist(ctx context.Context, tokenJTI, tokenHash, userID string, ttl time.Duration) error {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return err
	}
	// Просто сохраняем токен с timestamp истечения
	key := fmt.Sprintf("blacklist:%s:%s", tokenHash, tokenJTI)

	// Значение - время истечения в unix timestamp
	expiresAt := time.Now().UTC().Add(ttl).Unix()
	value := strconv.FormatInt(expiresAt, 10)

	err := b.blackCache.Set(ctx, key, []byte(value), ttl)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

// метод для проверки, есть ли такой ключ в черном списке
func (b *AuthBlackListRepository) IsBlacklisted(ctx context.Context, tokenJTI, tokenHash string) (bool, error) {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return false, err
	}

	key := fmt.Sprintf("blacklist:%s:%s", tokenHash, tokenJTI)

	// Проверяем существование ключа
	exists, err := b.blackCache.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}
