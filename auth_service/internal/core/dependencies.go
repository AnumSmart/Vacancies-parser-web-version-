package core

import (
	"auth_service/configs"
	"auth_service/internal/auth_server/handlers"
	"auth_service/internal/auth_server/repository"
	"auth_service/internal/auth_server/service"
	"context"
	"fmt"
	"runtime"
	postgresdb "shared/postgres_db"
)

// Dependencies содержит все общие зависимости
type AuthServiceDepenencies struct {
	AuthConfig  *configs.AuthServiceConfig
	AuthHandler handlers.AuthHandlerInterface
}

// InitDependencies инициализирует общие зависимости для auth_service
func InitDependencies(ctx context.Context) (*AuthServiceDepenencies, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	// Получаем конфигурацию
	conf, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// создаём экземпляр пула соединений для postgresQL
	pgRepo, err := postgresdb.NewPgRepo(ctx, conf.PostgresDBConf)

	// создаём слой репозитория
	repo := repository.NewAuthRepository(pgRepo)

	// создаём сервис аторизации
	authService := service.NewAuthService(repo)

	// создаём хэндлер поиска
	authHandler := handlers.NewAuthHandler(authService)

	return &AuthServiceDepenencies{
		AuthConfig:  conf,
		AuthHandler: authHandler,
	}, nil
}
