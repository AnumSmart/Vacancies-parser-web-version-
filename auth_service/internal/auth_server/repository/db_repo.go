package repository

import (
	authinterfaces "auth_service/internal/auth_interfaces"
	"context"
	"errors"
	"fmt"
	globalmodels "global_models"
	"global_models/global_db"
	"time"

	"github.com/jackc/pgx/v4"
)

// создаём репозиторий базы данных для сервиса авторизации на базе адаптера к pgxpool

// Реализуем ТОЛЬКО то, что нужно auth_service
type AuthUserDBRepository struct {
	pool global_db.Pool // строится на базе глобального интерфейса
}

// создаём конструктор для слоя базы данных
func NewAuthUserRepository(pool global_db.Pool) authinterfaces.DBRepoInterface {
	return &AuthUserDBRepository{pool: pool}
}

// метод для проверки, есть ли такой пользователь в базе (возвращаем ID пользователя, флаг и ошибку)
func (a *AuthUserDBRepository) CheckIfInBaseByEmail(ctx context.Context, email string) (int64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}

	const query = `SELECT id FROM users WHERE email = $1 LIMIT 1`

	var id int64
	err := a.pool.QueryRow(ctx, query, email).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil // ✅ false, но без ошибки
		}
		return 0, false, fmt.Errorf("failed to query user by email: %w", err)
	}

	return id, true, nil
}

// метод для добавления нового пользователя в базу данных (возвращаем ID пользователя из базы и ошибку)
func (a *AuthUserDBRepository) AddUser(ctx context.Context, email, hashedPass string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	const query = `
        INSERT INTO users (email, password_hash, created_at) 
        VALUES ($1, $2, $3) 
        RETURNING id
    `

	var userID int64
	err := a.pool.QueryRow(ctx, query, email, hashedPass, time.Now()).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return -1, globalmodels.ErrUserAlreadyExists
	}
	if err != nil {
		return -1, fmt.Errorf("failed to insert user: %w", err)
	}

	return userID, nil
}

// метод для поиска пользователя с заданным email в базе (получаем пользователя и ошибку)
func (a *AuthUserDBRepository) FindUserByEmail(ctx context.Context, email string) (*globalmodels.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	const query = `
        SELECT id, email, password_hash, created_at
        FROM users 
        WHERE email = $1
        LIMIT 1
    `

	var user globalmodels.User
	err := a.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // ✅ nil, nil - пользователь не найден
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// метод поиска хэша токена в БД по email (получаем хэш токена, токенJTI, ошибку)
func (a *AuthUserDBRepository) FindTokenHashByEmail(ctx context.Context, email string) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}

	const query = `SELECT refresh_token_hash, token_jti FROM users WHERE email = $1 LIMIT 1`

	var tokenHash, tokenJTI string // Обычные строки, т.к. NOT NULL

	err := a.pool.QueryRow(ctx, query, email).Scan(&tokenHash, &tokenJTI)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", nil // Пользователь не найден
		}
		return "", "", fmt.Errorf("failed to query token hash: %w", err)
	}

	return tokenHash, tokenJTI, nil
}

// метод добавления (обновления) рефрэш токена в БД
func (a *AuthUserDBRepository) AddRefreshToken(ctx context.Context, email, refreshToken, tokenJTI string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const query = `
        UPDATE users 
        SET refresh_token_hash = $1, token_jti = $2
        WHERE email = $3
    `

	_, err := a.pool.Exec(ctx, query, refreshToken, tokenJTI, email)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	return nil
}
