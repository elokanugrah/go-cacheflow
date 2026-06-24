package cacheflow

import (
	"context"
	"time"

	"github.com/elokanugrah/go-cacheflow/serializer"
)

// TypedCache wraps a raw Cache interface and a Serializer to provide typed cache operations.
//
// TypedCache is safe for concurrent use.
type TypedCache[T any] struct {
	cache      Cache
	serializer serializer.Serializer
}

// Typed creates a new TypedCache wrapper for type T.
//
// It returns a new TypedCache instance on each call. If the cache provides a Serializer()
// method, it uses that serializer; otherwise, it defaults to JSONSerializer.
//
// Example:
//
//	users := cacheflow.Typed[User](cf)
//	user, err := users.Remember(ctx, "user:123", time.Minute, func(ctx context.Context) (User, error) {
//	    return repo.GetUser(ctx, 123)
//	})
func Typed[T any](cache Cache) *TypedCache[T] {
	var ser serializer.Serializer
	if provider, ok := cache.(interface{ Serializer() serializer.Serializer }); ok {
		ser = provider.Serializer()
	} else {
		ser = serializer.NewJSONSerializer()
	}

	return &TypedCache[T]{
		cache:      cache,
		serializer: ser,
	}
}

// Get retrieves a typed value from the cache.
func (c *TypedCache[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var result T
	if err := c.serializer.Unmarshal(data, &result); err != nil {
		return zero, err
	}

	return result, nil
}

// Set stores a typed value in the cache with the given TTL.
func (c *TypedCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, key, data, ttl)
}

// Delete removes a value from the cache.
func (c *TypedCache[T]) Delete(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, key)
}

// Remember retrieves a value from the cache, invoking the loader on cache miss, with SingleFlight deduplication.
func (c *TypedCache[T]) Remember(
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func(ctx context.Context) (T, error),
) (T, error) {
	var zero T
	data, err := c.cache.Remember(ctx, key, ttl, func(ctx context.Context) ([]byte, error) {
		loaded, err := loader(ctx)
		if err != nil {
			return nil, err
		}
		return c.serializer.Marshal(loaded)
	})
	if err != nil {
		return zero, err
	}

	var result T
	if err := c.serializer.Unmarshal(data, &result); err != nil {
		return zero, err
	}

	return result, nil
}
