package core

import (
	"auth_service/configs"
	"auth_service/internal/auth_server/handlers"
	"auth_service/internal/auth_server/repository"
	"auth_service/internal/auth_server/service"
	"context"
	"fmt"
	"global_models/global_cache"
	"global_models/global_db"
	"runtime"
	"shared/cookie"
	"shared/jwt_service"
	postgresdb "shared/postgres_db"
	redis "shared/redis"
	"sync"
)

// Dependencies содержит все общие зависимости
type AuthServiceDepenencies struct {
	AuthConfig  *configs.AuthServiceConfig
	AuthHandler handlers.AuthHandlerInterface

	// добавляем поля для логики освобождения ресурсов
	pgPool         global_db.Pool     // для особождения ресурсов DB
	redisCacherepo global_cache.Cache // для освобождения ресурсов redis
	closeOnce      sync.Once          // для того, чтобы функция освобождения ресурсов выполнилась только 1 раз
	closeErr       error
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
	// адаптер к глобальному интерфейсу используется внутри NewPoolWithConfig
	pgPool, err := postgresdb.NewPoolWithConfig(ctx, conf.PostgresDBConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL repository: %w", err)
	}

	// создаём репозиторий для авторизации пользователя
	userRepo := repository.NewAuthUserRepository(pgPool)

	// создаём экземпляр redis
	redisCacherepo, err := redis.NewRedisCacheRepository(conf.RedisConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Black List repository (based om Redis): %w", err)
	}

	// создаём репозиторий черного списка
	blackListrepo, err := repository.NewBlackListRepo(redisCacherepo, "auth")

	// создаём слой репозитория (на базе репозитория Postgres и репозитория токенов (на базе redis))
	repo, err := repository.NewAuthRepository(userRepo, blackListrepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create Auth Repository Layer: %w", err)
	}

	// создаём сервис jwt
	jwtManager := jwt_service.NewJWTService(conf.JWTConfig)
	if jwtManager == nil {
		return nil, fmt.Errorf("failed to create jwt service")
	}

	// создаем менеджера куки
	cookieManager := cookie.NewManager(conf.CookieManagerConfig)
	if cookieManager == nil {
		return nil, fmt.Errorf("failed to create cookieManager")
	}

	// создаём сервис аторизации
	authService, err := service.NewAuthService(repo, jwtManager)

	if err != nil {
		return nil, fmt.Errorf("failed to create auth service")
	}

	if authService == nil {
		return nil, fmt.Errorf("failed to create auth service")
	}

	// создаём хэндлер поиска
	authHandler := handlers.NewAuthHandler(authService, cookieManager)
	if authHandler == nil {
		return nil, fmt.Errorf("failed to create auth handler")
	}

	return &AuthServiceDepenencies{
		AuthConfig:  conf,
		AuthHandler: authHandler,
	}, nil
}

// метод структуры зависимостей для осбобождения ресурсов
func (d *AuthServiceDepenencies) Close() error {
	d.closeOnce.Do(func() {
		var errs []error

		// Закрываем Redis
		if d.redisCacherepo != nil {
			if err := d.redisCacherepo.Close(); err != nil {
				errs = append(errs, fmt.Errorf("redis: %w", err))
			}
		}

		// Закрываем PostgreSQL
		if d.pgPool != nil {
			if err := d.pgPool.Close(); err != nil {
				errs = append(errs, fmt.Errorf("postgres: %w", err))
			}
		}

		// проверяем аггрегированные ошибки
		if len(errs) > 0 {
			d.closeErr = fmt.Errorf("close errors: %v", errs)
		}
	})

	if d.closeErr == nil {
		fmt.Println("Ресурсы - освобождены")
	}

	// если все хоршо, то возвращаем nil
	return d.closeErr
}
