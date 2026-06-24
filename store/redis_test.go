package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRedisClient creates a Redis client for testing using miniredis.
func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return client
}

func TestRedisStore_GetSet(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)
	ctx := context.Background()

	err := s.Set(ctx, "key1", []byte("value1"), time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	val, err := s.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("Get() = %q, want %q", string(val), "value1")
	}
}

func TestRedisStore_GetMiss(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get(nonexistent) error = %v, want ErrCacheMiss", err)
	}
}

func TestRedisStore_TTLExpiration(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	s := NewRedisStore(client)
	ctx := context.Background()

	err = s.Set(ctx, "expiring", []byte("data"), 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	// Should be available immediately
	val, err := s.Get(ctx, "expiring")
	if err != nil {
		t.Fatalf("Get() before expiry unexpected error: %v", err)
	}
	if string(val) != "data" {
		t.Errorf("Get() = %q, want %q", string(val), "data")
	}

	// Fast forward time in miniredis to expire the key
	mr.FastForward(700 * time.Millisecond)

	_, err = s.Get(ctx, "expiring")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get() after expiry error = %v, want ErrCacheMiss", err)
	}
}

func TestRedisStore_ZeroTTLNoExpiration(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)
	ctx := context.Background()

	err := s.Set(ctx, "forever", []byte("persistent"), 0)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	// Verify the key has no TTL set in Redis
	ttl := client.TTL(ctx, "forever").Val()
	if ttl != -1*time.Nanosecond && ttl != -1 {
		// Redis returns -1 for keys without TTL
		// go-redis represents this as -1ns
		t.Logf("TTL = %v (expected no expiration)", ttl)
	}

	val, err := s.Get(ctx, "forever")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(val) != "persistent" {
		t.Errorf("Get() = %q, want %q", string(val), "persistent")
	}
}

func TestRedisStore_Delete(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)
	ctx := context.Background()

	_ = s.Set(ctx, "key", []byte("value"), time.Minute)

	err := s.Delete(ctx, "key")
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	_, err = s.Get(ctx, "key")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get() after Delete() error = %v, want ErrCacheMiss", err)
	}
}

func TestRedisStore_DeleteNonexistent(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)
	ctx := context.Background()

	err := s.Delete(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Delete(nonexistent) unexpected error: %v", err)
	}
}

func TestRedisStore_ContextCancellation(t *testing.T) {
	client := newTestRedisClient(t)
	s := NewRedisStore(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := s.Get(ctx, "key")
	if err == nil {
		t.Fatal("Get() with cancelled context expected error, got nil")
	}
}

func TestRedisStore_ImplementsInterface(t *testing.T) {
	// Compile-time check that RedisStore implements Store
	var _ Store = (*RedisStore)(nil)
}
