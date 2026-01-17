// описание сервисного слоя сервера авторизации
package service

import (
	"auth_service/internal/auth_server/repository"
	"context"
)

// описание интерфейса сервисного слоя
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) error
}

// описание структуры сервисного слоя
type AuthService struct {
	repo repository.AuthRepositoryInterface
}
