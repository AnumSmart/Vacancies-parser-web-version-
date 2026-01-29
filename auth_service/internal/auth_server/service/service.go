// описание сервисного слоя сервера авторизации
package service

import (
	"auth_service/internal/auth_server/repository"
	"auth_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"shared/jwt_service"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// описание интерфейса сервисного слоя
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) error
	StopServices(ctx context.Context)
	AddRefreshTokenToDb(ctx context.Context, email, refreshToken string) error
	GetTokens(ctx context.Context, email string) (string, string, error)
}

// описание структуры сервисного слоя
type AuthService struct {
	repo       repository.AuthRepositoryInterface
	jwtManager jwt_service.JWTManagerInterface
}

// Конструктор возвращает интерфейс
func NewAuthService(repo repository.AuthRepositoryInterface, jwt jwt_service.JWTManagerInterface) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtManager: jwt,
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
		return "", errors.New("user with such Email is in base")
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

// Метод проверки соответствия пользователя с информацией в базе
func (s *AuthService) Login(ctx context.Context, email, password string) error {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return err
	}

	// Проверяем существует ли пользователь с данным email уже в базе.
	existedUser, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	// если указатель на пользователя == nil, значит пользователь не был найден
	if existedUser == nil {
		log.Printf("error during search in the DB, user = %v", existedUser)
		return domain.ErrUserWrongCredentials
	}

	//сравниваем хэши паролей, тот, что в базе и тот, что логинится
	err = bcrypt.CompareHashAndPassword([]byte(existedUser.PasswordHash), []byte(password))
	if err != nil {
		return domain.ErrUserWrongCredentials
	}

	return nil
}

// метод для генерации jwt токенов
func (a *AuthService) GetTokens(ctx context.Context, email string) (string, string, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	// пробуем генерировать JWT токены
	accessToken, refreshToken, err := a.jwtManager.GenerateTokens(email)
	if err != nil {
		return "", "", fmt.Errorf("Error during JWT tokens generation: %v", err)
	}

	return accessToken, refreshToken, nil
}

// метод работы с repo слоем, добавление refresh токена в DB
func (a *AuthService) AddRefreshTokenToDb(ctx context.Context, email, refreshToken string) error {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return err
	}

	err := a.repo.AddRefreshToken(ctx, email, refreshToken)
	if err != nil {
		return err
	}
	return nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {
	// реализация
}
