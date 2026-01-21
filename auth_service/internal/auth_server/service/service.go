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
	StopServices(ctx context.Context)
}

// описание структуры сервисного слоя
type AuthService struct {
	repo repository.AuthRepositoryInterface
}

// Конструктор возвращает интерфейс
func NewAuthService(repo repository.AuthRepositoryInterface) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// Метод регистарции пользователя
func (s *AuthService) Register(ctx context.Context, email, password string) error {
	// реализация
	return nil
}

// Метод логина пользователя
func (s *AuthService) Login(ctx context.Context, email, password string) error {
	// реализация
	return nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {

}
