package postgresdb

import (
	"context"
	"fmt"
	"shared/config"

	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PgRepoInterface interface {
	Close()
	GetPool() *pgxpool.Pool
}

type PgRepo struct {
	closeOnce sync.Once
	pool      *pgxpool.Pool
}

// NewPgRepo creates a new PostgreSQL repository with properly configured connection pool
func NewPgRepo(ctx context.Context, conf *config.PostgresDBConfig) (*PgRepo, error) {
	// Parse the connection string into a pgxpool.Config
	poolConfig, err := pgxpool.ParseConfig(conf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DB DSN: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = conf.MaxConns
	poolConfig.MinConns = conf.MinConns

	// Configure connection health checks
	poolConfig.HealthCheckPeriod = conf.HealthCheckPeriod
	poolConfig.MaxConnLifetime = conf.MaxConnLifetime
	poolConfig.MaxConnIdleTime = conf.MaxConnIdleTime

	// Configure connection timeouts
	poolConfig.ConnConfig.ConnectTimeout = conf.ConnectTimeout

	// Create the connection pool with context
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(connCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PgRepo{
		pool: pool,
	}, nil
}

// Close gracefully closes the connection pool (only once)
func (r *PgRepo) Close() {
	r.closeOnce.Do(func() {
		if r.pool != nil {
			r.pool.Close()
		}
	})
}

// GetPool returns the connection pool (useful for transactions)
func (r *PgRepo) GetPool() *pgxpool.Pool {
	return r.pool
}
