// описание слоя репозитория сервиса авторизации
package repository

import (
	"auth_service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	postgresdb "shared/postgres_db"

	"github.com/jackc/pgx/v4"
)

// описание интерфейса слоя репозитория
type AuthRepositoryInterface interface {
	CheckIfInBaseByEmail(ctx context.Context, email string) (int64, bool, error)
	AddUser(ctx context.Context, email, hashedPass string) (int64, error)
}

// описание структуры слоя репозитория
type AuthRepository struct {
	pgRepo postgresdb.PgRepoInterface
}

// конструктор для слоя репозиторий
func NewAuthRepository(pgRepo postgresdb.PgRepoInterface) *AuthRepository {
	return &AuthRepository{
		pgRepo: pgRepo,
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
	//заглушка
	return 0, nil
}
