package cacheflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elokanugrah/go-cacheflow/store"
)

type testUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestGet_CacheHit(t *testing.T) {
	cf := New()
	ctx := context.Background()

	user := testUser{Name: "Alice", Age: 30}
	err := Set(ctx, cf, "user:1", user, time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	result, err := Get[testUser](ctx, cf, "user:1")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if result.Name != "Alice" || result.Age != 30 {
		t.Errorf("Get() = %+v, want %+v", result, user)
	}
}

func TestGet_CacheMiss(t *testing.T) {
	cf := New()
	ctx := context.Background()

	_, err := Get[testUser](ctx, cf, "nonexistent")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get(nonexistent) error = %v, want ErrCacheMiss", err)
	}
}

func TestSet_StoresValue(t *testing.T) {
	cf := New()
	ctx := context.Background()

	err := Set(ctx, cf, "key", "hello", time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	result, err := Get[string](ctx, cf, "key")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("Get() = %q, want %q", result, "hello")
	}
}

func TestSet_PropagatesTTL(t *testing.T) {
	cf := New()
	ctx := context.Background()

	err := Set(ctx, cf, "expiring", "data", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	_, err = Get[string](ctx, cf, "expiring")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get() after TTL expiry error = %v, want ErrCacheMiss", err)
	}
}

func TestGet_SliceType(t *testing.T) {
	cf := New()
	ctx := context.Background()

	input := []testUser{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	err := Set(ctx, cf, "users", input, time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	result, err := Get[[]testUser](ctx, cf, "users")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("Get() len = %d, want 2", len(result))
	}
	if result[0].Name != "Alice" {
		t.Errorf("Get()[0].Name = %q, want %q", result[0].Name, "Alice")
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	cf := New()
	ctx := context.Background()

	_ = Set(ctx, cf, "key", "value", time.Minute)

	err := Delete(ctx, cf, "key")
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	_, err = Get[string](ctx, cf, "key")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get() after Delete() error = %v, want ErrCacheMiss", err)
	}
}

func TestErrCacheMiss_ReExported(t *testing.T) {
	// Verify that the re-exported ErrCacheMiss is the same as store.ErrCacheMiss
	if !errors.Is(ErrCacheMiss, store.ErrCacheMiss) {
		t.Error("ErrCacheMiss should match store.ErrCacheMiss")
	}
}

// --- Error path helpers ---

// failSerializer is a Serializer that always fails.
type failSerializer struct{}

func (failSerializer) Marshal(v any) ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func (failSerializer) Unmarshal(data []byte, v any) error {
	return errors.New("unmarshal failed")
}

// failStore is a Store that returns a non-CacheMiss error on Get,
// and an error on Set.
type failStore struct {
	store.Store
}

func (failStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, errors.New("store connection error")
}

func (failStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return errors.New("store write error")
}

func (failStore) Delete(ctx context.Context, key string) error {
	return nil
}

// --- Error path tests for cache.go ---

func TestGet_UnmarshalError(t *testing.T) {
	cf := New(WithSerializer(failSerializer{}))
	ctx := context.Background()

	// Store raw bytes directly
	_ = cf.store.Set(ctx, "key", []byte("not-valid-for-failSerializer"), time.Minute)

	_, err := Get[testUser](ctx, cf, "key")
	if err == nil {
		t.Fatal("Get() expected unmarshal error, got nil")
	}
}

func TestSet_MarshalError(t *testing.T) {
	cf := New(WithSerializer(failSerializer{}))
	ctx := context.Background()

	err := Set(ctx, cf, "key", testUser{Name: "Alice"}, time.Minute)
	if err == nil {
		t.Fatal("Set() expected marshal error, got nil")
	}
}

// --- Error path tests for CacheFlow.Remember (cacheflow.go) ---

func TestCacheFlowRemember_StoreError(t *testing.T) {
	// When Get returns a non-CacheMiss error, Remember should surface it
	cf := New(WithStore(failStore{}))
	ctx := context.Background()

	_, err := cf.Remember(ctx, "key", time.Minute, func(ctx context.Context) ([]byte, error) {
		t.Fatal("loader should not be called when store returns a non-CacheMiss error")
		return nil, nil
	})

	if err == nil {
		t.Fatal("CacheFlow.Remember() expected store error, got nil")
	}
}

func TestCacheFlowRemember_SetFailureStillReturnsValue(t *testing.T) {
	// When the loader succeeds but Set fails, Remember should still return the loaded value
	ms := store.NewMemoryStore()
	ctx := context.Background()

	// Use a wrapper that fails on Set but delegates Get/Delete to the real store
	wrapper := &setFailStore{Store: ms}
	cf := New(WithStore(wrapper))

	data, err := cf.Remember(ctx, "key", time.Minute, func(ctx context.Context) ([]byte, error) {
		return []byte("loaded-value"), nil
	})

	if err != nil {
		t.Fatalf("Remember() unexpected error: %v", err)
	}
	if string(data) != "loaded-value" {
		t.Errorf("Remember() = %q, want %q", string(data), "loaded-value")
	}
}

// setFailStore wraps a real Store but fails on Set.
type setFailStore struct {
	store.Store
}

func (s *setFailStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return errors.New("set failed")
}

// --- Error path tests for Remember[T] (remember.go) ---

func TestRemember_UnmarshalError(t *testing.T) {
	// Pre-populate cache with bytes that won't unmarshal to testUser
	cf := New()
	ctx := context.Background()

	// Store raw invalid bytes directly in the store
	_ = cf.store.Set(ctx, "bad:data", []byte("not-json{{{"), time.Minute)

	_, err := Remember(ctx, cf, "bad:data", time.Minute, func(ctx context.Context) (testUser, error) {
		t.Fatal("loader should not be called on cache hit")
		return testUser{}, nil
	})

	if err == nil {
		t.Fatal("Remember() expected unmarshal error from corrupted cache data, got nil")
	}
}

