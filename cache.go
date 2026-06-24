package cacheflow

import (
	"context"
	"time"

	"github.com/elokanugrah/go-cacheflow/store"
)

// Get retrieves a typed value from the cache by key.
//
// Get deserializes the raw bytes from the Store into T using the
// configured Serializer. Returns store.ErrCacheMiss transparently
// when the key does not exist or has expired.
//
// Example:
//
//	user, err := cacheflow.Get[User](ctx, cf, "user:123")
//	if errors.Is(err, store.ErrCacheMiss) {
//	    // handle miss
//	}
func Get[T any](ctx context.Context, cf *CacheFlow, key string) (T, error) {
	var zero T

	data, err := cf.store.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var result T
	if err := cf.serializer.Unmarshal(data, &result); err != nil {
		return zero, err
	}

	return result, nil
}

// Set stores a typed value in the cache with the given key and TTL.
//
// Set serializes the value to bytes using the configured Serializer
// and delegates storage to the Store. If ttl is 0, the entry does
// not expire and lives until explicitly deleted.
//
// Example:
//
//	err := cacheflow.Set(ctx, cf, "user:123", user, time.Minute)
func Set[T any](ctx context.Context, cf *CacheFlow, key string, value T, ttl time.Duration) error {
	data, err := cf.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return cf.store.Set(ctx, key, data, ttl)
}

// Delete removes a cached value by key.
//
// Delete is idempotent — deleting a non-existent key does not
// return an error.
//
// Example:
//
//	err := cacheflow.Delete(ctx, cf, "user:123")
func Delete(ctx context.Context, cf *CacheFlow, key string) error {
	return cf.store.Delete(ctx, key)
}

// re-export ErrCacheMiss at the package level for convenience.
var ErrCacheMiss = store.ErrCacheMiss
