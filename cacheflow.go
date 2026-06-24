// Package cacheflow provides a cache orchestration library for Go that
// simplifies cache-aside pattern implementation through a single Remember() API.
//
// CacheFlow manages cache lookup, population, serialization, and SingleFlight
// deduplication, allowing developers to focus on business logic rather than
// caching infrastructure.
//
// Basic usage:
//
//	cf := cacheflow.New()
//
//	user, err := cacheflow.Remember(ctx, cf, "user:123", time.Minute,
//	    func(ctx context.Context) (*User, error) {
//	        return repo.GetUser(ctx, 123)
//	    },
//	)
package cacheflow

import (
	"context"
	"errors"
	"time"

	"github.com/elokanugrah/go-cacheflow/serializer"
	"github.com/elokanugrah/go-cacheflow/store"
	"golang.org/x/sync/singleflight"
)

// Cache defines the raw (bytes-only) cache operations.
//
// All implementations must be safe for concurrent use by multiple goroutines.
type Cache interface {
	// Get retrieves raw bytes from the cache.
	//
	// Returns store.ErrCacheMiss if the key does not exist or has expired.
	// Returns other errors for backend failures.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores raw bytes in the cache with the given TTL.
	//
	// If ttl is 0, the entry does not expire.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from the cache.
	//
	// Delete is idempotent — deleting a non-existent key does not return an error.
	Delete(ctx context.Context, key string) error

	// Remember executes the loader function if cache miss occurs, deduplicating concurrent calls.
	//
	// If loader returns an error, the cache is not populated.
	Remember(ctx context.Context, key string, ttl time.Duration, loader func(ctx context.Context) ([]byte, error)) ([]byte, error)
}

// CacheFlow is the core orchestration struct that coordinates cache
// operations across a Store and Serializer with SingleFlight deduplication.
//
// CacheFlow is safe for concurrent use by multiple goroutines.
type CacheFlow struct {
	store      store.Store
	serializer serializer.Serializer
	sfGroup    *singleflight.Group
}

// New creates a new CacheFlow instance with the given options.
//
// Without any options, CacheFlow defaults to:
//   - Store: MemoryStore (in-memory, concurrent-safe)
//   - Serializer: JSONSerializer (encoding/json)
//
// Example:
//
//	// Default configuration
//	cf := cacheflow.New()
//
//	// With custom store
//	cf := cacheflow.New(
//	    cacheflow.WithStore(store.NewRedisStore(redisClient)),
//	)
func New(opts ...Option) *CacheFlow {
	cf := &CacheFlow{
		store:      store.NewMemoryStore(),
		serializer: serializer.NewJSONSerializer(),
		sfGroup:    &singleflight.Group{},
	}

	for _, opt := range opts {
		opt(cf)
	}

	return cf
}

// Get retrieves raw bytes from the store.
func (cf *CacheFlow) Get(ctx context.Context, key string) ([]byte, error) {
	return cf.store.Get(ctx, key)
}

// Set stores raw bytes in the store with the given TTL.
func (cf *CacheFlow) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return cf.store.Set(ctx, key, value, ttl)
}

// Delete removes a key from the store.
func (cf *CacheFlow) Delete(ctx context.Context, key string) error {
	return cf.store.Delete(ctx, key)
}

// Serializer returns the configured serializer.
func (cf *CacheFlow) Serializer() serializer.Serializer {
	return cf.serializer
}

// Remember implements the raw cache-aside pattern with SingleFlight deduplication.
func (cf *CacheFlow) Remember(
	ctx context.Context,
	key string,
	ttl time.Duration,
	loader func(ctx context.Context) ([]byte, error),
) ([]byte, error) {
	// 1. Cache lookup
	data, err := cf.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	if !errors.Is(err, store.ErrCacheMiss) {
		return nil, err
	}

	// 2. Cache miss -> SingleFlight
	val, err, _ := cf.sfGroup.Do(key, func() (any, error) {
		// Double-check cache inside SingleFlight
		if cached, err := cf.Get(ctx, key); err == nil {
			return cached, nil
		}

		// Invoke loader
		loaded, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		// Store the result
		if setErr := cf.Set(ctx, key, loaded, ttl); setErr != nil {
			// Return loaded even if caching fails
			return loaded, nil
		}

		return loaded, nil
	})

	if err != nil {
		return nil, err
	}

	return val.([]byte), nil
}