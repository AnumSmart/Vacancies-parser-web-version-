package authinterfaces

import (
	"context"
	globalmodels "global_models"
	"time"
)

// интерфейс слоя базы данных для авторизации юзеров
type DBRepoInterface interface {
	CheckIfInBaseByEmail(ctx context.Context, email string) (int64, bool, error)
	AddUser(ctx context.Context, email, hashedPass string) (int64, error)
	FindUserByEmail(ctx context.Context, email string) (*globalmodels.User, error)
	AddRefreshToken(ctx context.Context, email, refreshToken, tokenJTI string) error
	FindTokenHashByEmail(ctx context.Context, email string) (string, string, error)
}

// интерфейс для черного списка
type BlackListRepository interface {
	// Только кастомные методы для работы с черным списком!
	AddToBlacklist(ctx context.Context, tokenJTI, tokenHash, userID string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, tokenJTI, tokenHash string) (bool, error)
}
