// описание сервисного слоя сервера авторизации
package service

import (
	"auth_service/internal/auth_server/repository"
	"context"
	"fmt"
	"shared/jwt_service"
)

// описание интерфейса сервисного слоя
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) error
	StopServices(ctx context.Context)
}

// описание структуры сервисного слоя
type AuthService struct {
	repo repository.AuthRepositoryInterface
	jwt  jwt_service.JWTManager
}

// Конструктор возвращает интерфейс
func NewAuthService(repo repository.AuthRepositoryInterface, jwt jwt_service.JWTManager) *AuthService {
	return &AuthService{
		repo: repo,
		jwt:  jwt,
	}
}

// Метод регистарции пользователя
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	// тестовая заглушка
	result := fmt.Sprintf("Server has validated the request! Email:%s, pass:%s\n", email, password)
	return result, nil
}

// Метод логина пользователя
func (s *AuthService) Login(ctx context.Context, email, password string) error {
	// реализация
	return nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {

}
