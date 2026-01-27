// описание сервисного слоя сервера авторизации
package service

import (
	"auth_service/internal/auth_server/repository"
	"context"
	"errors"
	"shared/jwt_service"
	"strconv"

	"golang.org/x/crypto/bcrypt"
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

// Метод регистарции пользователя (Возвращает ID пользователя и ошибку)
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {

	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return "Found no userID!", err
	}

	// проверяем, есть ли юзер с такми е-маилом в базе
	userID, isInBase, err := s.repo.CheckIfInBaseByEmail(ctx, email)
	if err != nil {
		return "Found no userID!", err
	}

	// если такой пользователь уже зарегестрирован, возвращем его ID и ошибку-сообщение
	if isInBase {
		return strconv.Itoa(int(userID)), errors.New("user with such Email is in base")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка при хешировании пароля")
	}

	// пробуем добавить нового юзера в базу данных, возвращем ID юзера и ошибку
	userID, err = s.repo.AddUser(ctx, email, string(hashedPassword))
	if err != nil {
		return "", errors.New("failed to add new user to the DB")
	}
	return strconv.Itoa(int(userID)), nil
}

// Метод логина пользователя
func (s *AuthService) Login(ctx context.Context, email, password string) error {
	// реализация
	return nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {
	// реализация
}
