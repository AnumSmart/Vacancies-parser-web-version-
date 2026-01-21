// описание слоя репозитория сервиса авторизации
package repository

import postgresdb "shared/postgres_db"

// описание интерфейса слоя репозитория
type AuthRepositoryInterface interface {
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
