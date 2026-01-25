package core

import (
	"auth_service/configs"
	"auth_service/internal/auth_server/handlers"
	"auth_service/internal/auth_server/repository"
	"auth_service/internal/auth_server/service"
	"context"
	"fmt"
	"runtime"
	"shared/jwt_service"
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
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL repository: %w", err)
	}

	// создаём слой репозитория
	repo := repository.NewAuthRepository(pgRepo)

	// создаём сервис jwt
	jwtManager := jwt_service.NewJWTService(conf.JWTConfig)
	if jwtManager == nil {
		return nil, fmt.Errorf("failed to create jwt service")
	}

	// создаём сервис аторизации
	authService := service.NewAuthService(repo, jwtManager)
	if authService == nil {
		return nil, fmt.Errorf("failed to create auth service")
	}

	// создаём хэндлер поиска
	authHandler := handlers.NewAuthHandler(authService)
	if authHandler == nil {
		return nil, fmt.Errorf("failed to create auth handler")
	}

	return &AuthServiceDepenencies{
		AuthConfig:  conf,
		AuthHandler: authHandler,
	}, nil
}
