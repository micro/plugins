package cache

import (
	"time"

	"go-micro.dev/v5/registry/cache"
)

// WithTTL sets the cache TTL.
func WithTTL(t time.Duration) cache.Option {
	return cache.WithTTL(t)
}
