package search_interfaces

import (
	"context"
	"time"
)

// HealthClient интерфейс для health checks
type HealthClient interface {
	CheckHealth(ctx context.Context, endpoint string) (time.Duration, bool, error)
}
