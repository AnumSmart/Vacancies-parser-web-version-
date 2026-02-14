// описание слоя репозитория сервиса авторизации
package repository

import (
	authinterfaces "auth_service/internal/auth_interfaces"
	"fmt"
)

// описание структуры слоя репозитория
type AuthRepository struct {
	DBRepo        authinterfaces.DBRepoInterface
	BlackListRepo authinterfaces.BlackListRepository
}

// конструктор для слоя репозиторий
func NewAuthRepository(dbRepo authinterfaces.DBRepoInterface, blackListRepo authinterfaces.BlackListRepository) (*AuthRepository, error) {
	// Проверяем обязательные зависимости
	if dbRepo == nil {
		return nil, fmt.Errorf("dbRepo is required")
	}
	if blackListRepo == nil {
		return nil, fmt.Errorf("blackListRepo is required")
	}
	return &AuthRepository{
		DBRepo:        dbRepo,
		BlackListRepo: blackListRepo,
	}, nil
}
