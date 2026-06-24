package store

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements the Store interface using Redis as the
// cache backend via the go-redis client library.
//
// RedisStore delegates all concurrency management to the Redis server
// and the go-redis client, which are both safe for concurrent use.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore backed by the given Redis client.
//
// The caller is responsible for configuring and closing the Redis client.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

// Get retrieves the cached value for the given key from Redis.
//
// Returns ErrCacheMiss if the key does not exist (redis.Nil).
// Returns other errors for Redis connection or command failures.
func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	return val, nil
}

// Set stores a value with the given key and time-to-live duration in Redis.
//
// If ttl is 0, the key is set without an expiration (persists until
// explicitly deleted). If ttl is greater than 0, the key expires
// after the specified duration.
func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes the cached value for the given key from Redis.
//
// Delete is idempotent — deleting a non-existent key does not return
// an error.
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
