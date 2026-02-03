// описание слоя репозитория сервиса авторизации
package repository

import (
	"auth_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	postgresdb "shared/postgres_db"
	"time"

	"github.com/jackc/pgx/v4"
)

// описание интерфейса слоя репозитория
type AuthRepositoryInterface interface {
	CheckIfInBaseByEmail(ctx context.Context, email string) (int64, bool, error)
	AddUser(ctx context.Context, email, hashedPass string) (int64, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	AddRefreshToken(ctx context.Context, email, refreshToken string) error
}

// описание структуры слоя репозитория
type AuthRepository struct {
	pgRepo    postgresdb.PgRepoInterface
	tokenRepo TokenRepositoryInterface
}

// конструктор для слоя репозиторий
func NewAuthRepository(pgRepo postgresdb.PgRepoInterface, tokenRepo TokenRepositoryInterface) *AuthRepository {
	return &AuthRepository{
		pgRepo:    pgRepo,
		tokenRepo: tokenRepo,
	}
}

// метод репоизтория для проверки наличия записи о пользователе по email
func (a *AuthRepository) CheckIfInBaseByEmail(ctx context.Context, email string) (int64, bool, error) {
	// проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}

	const query = `
		SELECT id
		FROM users 
		WHERE email = $1
		LIMIT 1
	`

	var user domain.User
	// вызываем метод поиска строки у базы данных
	err := a.pgRepo.GetPool().QueryRow(ctx, query, email).Scan(&user.ID)

	// проверяем: или это ошибка БД или, действительно, нет такого юзера
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println(err.Error())
			return 0, false, err
		}
		return 0, false, fmt.Errorf("failed to query user by email: %w", err)
	}

	return user.ID, true, nil
}

// метод репозитория добавления нового пользователя в базу
func (a *AuthRepository) AddUser(ctx context.Context, email, hashedPass string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	var userID int64
	query := `
        INSERT INTO users (email, password_hash, created_at) 
        VALUES ($1, $2, $3) 
        ON CONFLICT (email) DO NOTHING
        RETURNING id
    `

	err := a.pgRepo.GetPool().QueryRow(
		ctx,
		query,
		email,
		hashedPass,
		time.Now(),
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return -1, domain.ErrUserAlreadyExists
	}
	if err != nil {
		return -1, fmt.Errorf("failed to insert user: %w", err)
	}

	return userID, nil
}

// метод поиска пользователя в базе по email
func (a *AuthRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	const query = `
		SELECT id, email, hashed_pass, created_at
		FROM users 
		WHERE email = $1
		LIMIT 1
	`
	var user domain.User
	err := a.pgRepo.GetPool().QueryRow(ctx, query, email).Scan(
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		// если не нашлось такого юзера в базе, возвдращаем nil и nil ошибку (это услови обработатеся на уровне сервиса)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		log.Printf("function [FindByEmail], failed to query user by email: %v", err)
		return nil, err
	}

	return &user, nil
}

// добавляем поле refreshToken в базу по email (нужно держать refreshToken в БД)
func (a *AuthRepository) AddRefreshToken(ctx context.Context, email, refreshToken string) error {
	// Проверяем не отменен ли контекст
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `
		UPDATE users 
		SET refresh_token = $1 
		WHERE email = $2;
	`
	_, err := a.pgRepo.GetPool().Exec(ctx, query, refreshToken, email)
	if err != nil {
		return err
	}

	return nil
}
