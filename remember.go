package cacheflow

import (
	"context"
	"time"
)

// Remember is the primary API of CacheFlow. It implements the full
// cache-aside pattern with SingleFlight deduplication.
//
// Flow:
//  1. Cache Lookup — check if key exists in the Store
//  2. Cache Hit — return the deserialized value
//  3. Cache Miss — proceed to SingleFlight
//  4. SingleFlight — deduplicate concurrent calls for the same key
//  5. Loader — invoke the loader function to fetch the value
//  6. Store — serialize and cache the result with the given TTL
//  7. Return — return the value to all waiting callers
//
// If the loader returns an error, the cache is NOT populated and the
// error is returned to the caller. This prevents caching error states.
//
// Example:
//
//	user, err := cacheflow.Remember(ctx, cf, "user:123", time.Minute,
//	    func(ctx context.Context) (*User, error) {
//	        return repo.GetUser(ctx, 123)
//	    },
//	)
func Remember[T any](
	ctx context.Context,
	cf *CacheFlow,
	key string,
	ttl time.Duration,
	loader func(ctx context.Context) (T, error),
) (T, error) {
	var zero T
	data, err := cf.Remember(ctx, key, ttl, func(ctx context.Context) ([]byte, error) {
		loaded, err := loader(ctx)
		if err != nil {
			return nil, err
		}
		return cf.serializer.Marshal(loaded)
	})
	if err != nil {
		return zero, err
	}

	var result T
	if err := cf.serializer.Unmarshal(data, &result); err != nil {
		return zero, err
	}

	return result, nil
}
