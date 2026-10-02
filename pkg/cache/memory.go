package cache

import (
	"context"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// memoryCache implements the Cache interface, using an in-memory store.
type memoryCache struct {
	store *gocache.Cache // underlying in-memory store
	ttl   time.Duration  // default expiration
}

func newMemoryCache(ttl time.Duration) *memoryCache {
	return &memoryCache{store: gocache.New(ttl, 2*ttl), ttl: ttl}
}

// Set stores the value under the key.
func (c *memoryCache) Set(ctx context.Context, key, value string, ttl int) error {
	if ttl <= 0 {
		ttl = int(c.ttl.Seconds())
	}
	c.store.Set(key, value, time.Duration(ttl)*time.Second)
	return nil
}

// Get loads the value under the key, empty when missing.
func (c *memoryCache) Get(ctx context.Context, key string) (string, error) {
	value, found := c.store.Get(key)
	if !found {
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("[cache] key %s holds an unexpected value type", key)
	}
	return s, nil
}

// Delete removes the key.
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.store.Delete(key)
	return nil
}

// Exists reports whether the key exists.
func (c *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := c.store.Get(key)
	return found, nil
}

// Close releases the cache resources.
func (c *memoryCache) Close() error {
	return nil
}
