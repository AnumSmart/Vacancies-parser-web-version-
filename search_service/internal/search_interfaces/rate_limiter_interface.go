package search_interfaces

import "context"

type RateLimiter interface {
	Wait(ctx context.Context) error
	Stop()
}
