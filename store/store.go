// Package store provides cache storage abstractions for CacheFlow.
//
// Store defines a simple interface for cache backends that operate
// on raw bytes. Implementations must be safe for concurrent use.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned by Store.Get when the requested key
// does not exist in the cache or has expired.
var ErrCacheMiss = errors.New("cacheflow: cache miss")

// Store defines the interface for cache storage backends.
//
// All implementations must be safe for concurrent use by multiple
// goroutines. Store operates on raw []byte values — serialization
// is handled by the orchestration layer, not the store.
type Store interface {
	// Get retrieves the cached value for the given key.
	//
	// Returns ErrCacheMiss if the key does not exist or has expired.
	// Returns other errors for backend failures (e.g., network issues).
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with the given key and time-to-live duration.
	//
	// If ttl is 0, the entry does not expire and lives until explicitly
	// deleted. If ttl is greater than 0, the entry expires after the
	// specified duration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes the cached value for the given key.
	//
	// Delete is idempotent — deleting a non-existent key does not
	// return an error.
	Delete(ctx context.Context, key string) error
}
