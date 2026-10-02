package cache

import (
	"context"
	"time"
)

// Cache defines the cache interface.
type Cache interface {
	// Set stores the value under the key.
	Set(ctx context.Context, key, value string, ttl int) error
	// Get loads the value under the key, empty when missing.
	Get(ctx context.Context, key string) (string, error)
	// Delete removes the key.
	Delete(ctx context.Context, key string) error
	// Exists reports whether the key exists.
	Exists(ctx context.Context, key string) (bool, error)
	// Close releases the cache resources.
	Close() error
}

// Options holds the cache settings.
type Options struct {
	TTL int // expiration in seconds (default to 300)
}

func New(opts Options) (Cache, error) {
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = 300
	}
	return newMemoryCache(time.Duration(ttl) * time.Second), nil
}
