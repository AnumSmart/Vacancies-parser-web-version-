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
	"time"

	"golang.org/x/crypto/bcrypt"
)

// описание интерфейса сервисного слоя
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (int64, error)
	IvalidateRefreshToken(ctx context.Context, userEmail, userID string) error
	StopServices(ctx context.Context)
	AddHashRefreshTokenToDb(ctx context.Context, email, refreshToken, tokenJTI string) error
	GetTokens(ctx context.Context, email string) (*jwt_service.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*jwt_service.CustomClaims, bool, error)
}

// описание структуры сервисного слоя
type AuthService struct {
	repo       *repository.AuthRepository // слой репоизтория (прямая зависимость)
	jwtManager jwt_service.JWTManagerInterface
}

// Конструктор возвращает интерфейс
func NewAuthService(repo *repository.AuthRepository, jwt jwt_service.JWTManagerInterface) (AuthServiceInterface, error) {
	if repo == nil {
		return nil, fmt.Errorf("repo must not be nil")
	}

	if jwt == nil {
		return nil, fmt.Errorf("jwt must not be nil")
	}

	return &AuthService{
		repo:       repo,
		jwtManager: jwt,
	}, nil
}

// Метод регистарции пользователя (Возвращает ID пользователя и ошибку)
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {

	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return "Found no userID!", err
	}

	// проверяем, есть ли юзер с такми е-маилом в базе
	userID, isInBase, err := s.repo.DBRepo.CheckIfInBaseByEmail(ctx, email)
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
	userID, err = s.repo.DBRepo.AddUser(ctx, email, string(hashedPassword))
	if err != nil {
		return "", errors.New("failed to add new user to the DB")
	}
	return strconv.Itoa(int(userID)), nil
}

// Метод проверки соответствия пользователя с информацией в базе
func (s *AuthService) Login(ctx context.Context, email, password string) (int64, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Проверяем существует ли пользователь с данным email уже в базе.
	existedUser, err := s.repo.DBRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return 0, err
	}
	// если указатель на пользователя == nil, значит пользователь не был найден
	if existedUser == nil {
		log.Printf("error during search in the DB, user = %v", existedUser)
		return 0, domain.ErrUserWrongCredentials
	}

	//сравниваем хэши паролей, тот, что в базе и тот, что логинится
	err = bcrypt.CompareHashAndPassword([]byte(existedUser.PasswordHash), []byte(password))
	if err != nil {
		return 0, domain.ErrUserWrongCredentials
	}

	return existedUser.ID, nil
}

// метод для генерации jwt токенов
func (a *AuthService) GetTokens(ctx context.Context, email string) (*jwt_service.TokenPair, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// пробуем генерировать JWT токены
	tokenPair, err := a.jwtManager.GenerateTokens(email)
	if err != nil {
		return nil, fmt.Errorf("Error during JWT tokens generation: %v", err)
	}

	return tokenPair, nil
}

// метод работы с repo слоем, добавление refresh токена в DB
func (a *AuthService) AddHashRefreshTokenToDb(ctx context.Context, email, refreshToken, tokenJTI string) error {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return err
	}

	// хэшируем refresh токен перед добавлением в базу данных (peper получаем из jwtConfig)
	hashedRefreshToken := HashRefreshToken(refreshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// добавляем в базу хэшированное значение refresh токена
	err := a.repo.DBRepo.AddRefreshToken(ctx, email, hashedRefreshToken, tokenJTI)
	if err != nil {
		return err
	}
	return nil
}

// метод для обновления токенов (запрос от юзера)
func (a *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 1. валидируем refresh токен
	parsedUserRefToken, err := jwt_service.ParseTokenWithClaims(ctx, refreshToken, a.jwtManager.GetJTWConfig().SecretRefKey)
	if err != nil {
		log.Println("Wrong refresh token")
		return nil, err
	}

	// 2. Проверка срока действия (уже сделано в parsedUserRefToken)
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

	// 5. Получаем оставшееся время жизни user refresh token
	ttlUserRefreshToken := a.jwtManager.CalculateTokenTTL(claims)

	// 6. Получаем хэш userRefreshToken на базе SHA256 для сравнения с хэшом из БД
	userRefTokenHash := HashRefreshToken(refreshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// 5. проверка в Reddis (черный список), находится ли там запись хэша refresh токена
	exists, err := a.repo.BlackListRepo.IsBlacklisted(ctx, claims.ID, claims.TokenType)
	if exists {
		return nil, fmt.Errorf("current refresh token is in Black list!")
	}

	// 6. Полная проверка в БД (проверяем хэш переданного от юзера refresh токена)
	tokenHashFromDB, _, err := a.repo.DBRepo.FindTokenHashByEmail(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("Failed to find token in base by email from claims")
	}

	// 7. сравниваем найденный хэш и полученные хэш от пользователя
	equal := SecureHashCompare(tokenHashFromDB, userRefTokenHash)
	if !equal {
		return nil, fmt.Errorf("Token Hash from DB not equal privided token Hash from user!")
	}

	// 8. получаем пользователя из базы на основании claims
	user, err := a.repo.DBRepo.FindUserByEmail(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 9. Генерируем новую пару токенов
	tokenPair, err := a.jwtManager.GenerateTokens(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 10. добавить старый refresh токен в черный список redis (с оставшимся времененем жизни)
	err = a.repo.BlackListRepo.AddToBlacklist(ctx, claims.ID, userRefTokenHash, claims.UserID, ttlUserRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to add refresh token to black list: %w", err)
	}

	// 11. создаём хэш нового refresh токена
	newRefeshTokenHash := HashRefreshToken(tokenPair.RefreshToken, []byte(a.jwtManager.GetJTWConfig().TokenPepper))

	// 12. Сохранить новый refresh-токен в БД (создаем новый хэш, обновляем в базе)
	err = a.repo.DBRepo.AddRefreshToken(ctx, claims.Email, newRefeshTokenHash, tokenPair.RefreshJTI)
	if err != nil {
		return nil, fmt.Errorf("failed to Update refreshHash in DB: %w", err)
	}

	// 13. Вернуть новую пару токенов
	tokens := domain.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}

	return &tokens, nil
}

// метод для валидации токена от API Getway (access токен)
func (a *AuthService) ValidateToken(ctx context.Context, token string) (*jwt_service.CustomClaims, bool, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}

	//Парсим токен
	parsedToken, err := jwt_service.ParseTokenWithClaims(ctx, token, a.jwtManager.GetJTWConfig().SecretAccKey)
	if err != nil {
		log.Println("Invalid token")
		return nil, false, err
	}

	// Проверяем валидность токена
	if !parsedToken.Valid {
		return nil, false, fmt.Errorf("token is not valid")
	}

	// Проверяем claims
	claims, ok := parsedToken.Claims.(*jwt_service.CustomClaims)
	if !ok {
		return nil, false, fmt.Errorf("wrong token claims structure")
	}

	return claims, true, nil
}

// метод получения хэша токена из
func (a *AuthService) IvalidateRefreshToken(ctx context.Context, userEmail, userID string) error {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return err
	}

	// получаем хэш рефрэш токена из базы данных
	refreshHash, tokenJTI, err := a.repo.DBRepo.FindTokenHashByEmail(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("Failed to find refresh token hash in DB: %v", err)
	}

	// добавили в черный спиок с TTL=1 час, чтобы redis сам очистил
	err = a.repo.BlackListRepo.AddToBlacklist(ctx, tokenJTI, refreshHash, userID, 60*time.Minute)
	if err != nil {
		return fmt.Errorf("Failed to add refresh token hash into Black List: %v", err)
	}

	// обновляем значение в таблице для конкретного юзера на пусты строки
	err = a.repo.DBRepo.AddRefreshToken(ctx, userEmail, "", "")
	if err != nil {
		return fmt.Errorf("Failed to invalidate (tokenHash ---> '') refresh token in DB: %v", err)
	}

	return nil
}

// метод остановки всех сервисов
func (a *AuthService) StopServices(ctx context.Context) {
	// реализация---------------------------------------------------------
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
