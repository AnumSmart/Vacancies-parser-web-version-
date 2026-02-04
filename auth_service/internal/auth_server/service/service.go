// описание сервисного слоя сервера авторизации
package service

import (
	"auth_service/internal/auth_server/repository"
	"auth_service/internal/domain"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"shared/jwt_service"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// описание интерфейса сервисного слоя
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) error
	StopServices(ctx context.Context)
	AddRefreshTokenToDb(ctx context.Context, email, refreshToken string) error
	GetTokens(ctx context.Context, email string) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
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
	existedUser, err := s.repo.FindUserByEmail(ctx, email)
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

	// хэшируем refresh токен перед добавлением в базу данных (peper получаем из jwtConfig)
	hashedRefreshToken := HashRefreshToken(refreshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// добавляем в базу хэшированное значение refresh токена
	err := a.repo.AddRefreshToken(ctx, email, hashedRefreshToken)
	if err != nil {
		return err
	}
	return nil
}

// метод для обновления токенов (запрос от юзера)
func (a *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// 1. валидируем refresh токен
	parsedUserRefToken, err := jwt_service.ParseTokenWithClaims(ctx, refreshToken, a.jwtManager.GetJTWConfig().SecretRefKey)
	if err != nil {
		log.Println("Wrong refresh token")
		return nil, err
	}

	// 2. Проверка срока действия (уже сделано в ParseRefreshToken)
	if !parsedUserRefToken.Valid {
		return nil, fmt.Errorf("token expired")
	}

	// 3. Извлечение claims
	claims, ok := parsedUserRefToken.Claims.(*jwt_service.CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// 4. Проверка типа токена
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}

	// 5. Получаем оставшееся время жизни old refresh token
	ttlUserRefreshToken := a.jwtManager.CalculateTokenTTL(claims)

	// 6. Получаем хэш oldRefreshToken на базе SHA256 для добавления в redis
	userRefTokenHash := HashRefreshToken(refreshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// 5. проверка в Reddis (черный список), находится ли там запись хэша refresh токена
	exists, err := a.repo.CheckTokenHashInBalckList(ctx, userRefTokenHash)
	if exists {
		return nil, fmt.Errorf("current refresh token is in Black list!")
	}

	// 6. Полная проверка в БД (проверяем хэш переданного от юзера refresh токена)
	tokenHashFromDB, err := a.repo.FindTokenHashByEmail(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("Failed to find token in base by email from claims")
	}

	equal := SecureHashCompare(tokenHashFromDB, userRefTokenHash)
	if !equal {
		return nil, fmt.Errorf("Token Hash from DB not equal privided token Hash from user!")
	}

	// 7. Создание новой пары токенов
	user, err := a.repo.FindUserByEmail(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	newAccessToken, newRefeshToken, err := a.jwtManager.GenerateTokens(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 8. добавить старый refresh токен в черный список redis (с оставшимся времененем жизни)
	err = a.repo.AddRefreshTokenToBlackList(ctx, userRefTokenHash, claims.UserID, ttlUserRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to add refresh token to black list: %w", err)
	}

	// 9. создаём хэш нового refresh токена
	newRefeshTokenHash := HashRefreshToken(newRefeshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// 10. Сохранить новый refresh-токен в БД (создаем новый хэш, обновляем в базе)
	err = a.repo.AddRefreshToken(ctx, claims.Email, newRefeshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to Update refreshHash in DB: %w", err)
	}

	// 11. Вернуть новую пару токенов
	tokens := domain.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefeshToken,
	}

	return &tokens, nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {
	// реализация
}

// функция для хэширования токена
// HashRefreshToken создает HMAC-SHA256 хэш refresh-токена.
// Используется для безопасного хранения в Redis вместо исходного токена.
// Pepper ключ должен отличаться от JWT секрета и быть длиной ≥32 байта.
func HashRefreshToken(token string, pepper []byte) string {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

// функция безопасного сравнения хэшэй
func SecureHashCompare(expectedHash, providedHash string) bool {
	// Тримминг на всякий случай (если хэш из внешнего источника)
	expectedTrimmed := strings.TrimSpace(expectedHash)
	providedTrimmed := strings.TrimSpace(providedHash)

	// Быстрая проверка длины (хэши должны быть одинаковой длины)
	if len(expectedTrimmed) != len(providedTrimmed) {
		return false
	}

	// Константное по времени сравнение
	return subtle.ConstantTimeCompare(
		[]byte(expectedTrimmed),
		[]byte(providedTrimmed),
	) == 1
}
