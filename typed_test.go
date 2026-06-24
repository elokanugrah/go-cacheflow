package cacheflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elokanugrah/go-cacheflow/store"
)

type typedUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestTyped_GetSet(t *testing.T) {
	cf := New()
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	user := typedUser{Name: "Alice", Age: 30}
	err := tc.Set(ctx, "user:1", user, time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	got, err := tc.Get(ctx, "user:1")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.Name != "Alice" || got.Age != 30 {
		t.Errorf("Get() = %+v, want %+v", got, user)
	}
}

func TestTyped_GetMiss(t *testing.T) {
	cf := New()
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	_, err := tc.Get(ctx, "nonexistent")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get() error = %v, want %v", err, store.ErrCacheMiss)
	}
}

func TestTyped_Remember(t *testing.T) {
	cf := New()
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	loaderCalled := 0
	loader := func(ctx context.Context) (typedUser, error) {
		loaderCalled++
		return typedUser{Name: "Bob", Age: 25}, nil
	}

	// First call: miss, calls loader
	got1, err := tc.Remember(ctx, "user:2", time.Minute, loader)
	if err != nil {
		t.Fatalf("Remember() first call unexpected error: %v", err)
	}
	if got1.Name != "Bob" || got1.Age != 25 {
		t.Errorf("got = %+v, want Bob/25", got1)
	}
	if loaderCalled != 1 {
		t.Errorf("loader called %d times, want 1", loaderCalled)
	}

	// Second call: hit, bypasses loader
	got2, err := tc.Remember(ctx, "user:2", time.Minute, loader)
	if err != nil {
		t.Fatalf("Remember() second call unexpected error: %v", err)
	}
	if got2.Name != "Bob" || got2.Age != 25 {
		t.Errorf("got = %+v, want Bob/25", got2)
	}
	if loaderCalled != 1 {
		t.Errorf("loader called %d times, want 1", loaderCalled)
	}
}

func TestTyped_Delete(t *testing.T) {
	cf := New()
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	user := typedUser{Name: "Charlie", Age: 35}
	_ = tc.Set(ctx, "user:3", user, time.Minute)

	err := tc.Delete(ctx, "user:3")
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	_, err = tc.Get(ctx, "user:3")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get() after Delete() error = %v, want %v", err, store.ErrCacheMiss)
	}
}

// --- Typed error-path and fallback tests ---

// minimalCache is a Cache implementation WITHOUT a Serializer() method,
// exercising the fallback branch in Typed[T].
type minimalCache struct {
	data map[string][]byte
}

func (m *minimalCache) Get(_ context.Context, key string) ([]byte, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return nil, store.ErrCacheMiss
}

func (m *minimalCache) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *minimalCache) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *minimalCache) Remember(ctx context.Context, key string, ttl time.Duration, loader func(context.Context) ([]byte, error)) ([]byte, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	loaded, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	m.data[key] = loaded
	return loaded, nil
}

func TestTyped_FallbackSerializer(t *testing.T) {
	// Typed should use JSONSerializer when Cache has no Serializer() method
	mc := &minimalCache{data: make(map[string][]byte)}
	tc := Typed[typedUser](mc)
	ctx := context.Background()

	err := tc.Set(ctx, "user:fb", typedUser{Name: "Fallback", Age: 99}, time.Minute)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	got, err := tc.Get(ctx, "user:fb")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if got.Name != "Fallback" || got.Age != 99 {
		t.Errorf("Get() = %+v, want Fallback/99", got)
	}
}

func TestTyped_GetUnmarshalError(t *testing.T) {
	mc := &minimalCache{data: map[string][]byte{
		"bad": []byte("not-json{{{"),
	}}
	tc := Typed[typedUser](mc)
	ctx := context.Background()

	_, err := tc.Get(ctx, "bad")
	if err == nil {
		t.Fatal("Get() expected unmarshal error, got nil")
	}
}

func TestTyped_SetMarshalError(t *testing.T) {
	cf := New(WithSerializer(failSerializer{}))
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	err := tc.Set(ctx, "key", typedUser{Name: "X"}, time.Minute)
	if err == nil {
		t.Fatal("Set() expected marshal error, got nil")
	}
}

func TestTyped_RememberLoaderError(t *testing.T) {
	cf := New()
	tc := Typed[typedUser](cf)
	ctx := context.Background()

	expectedErr := errors.New("loader boom")
	_, err := tc.Remember(ctx, "err:key", time.Minute, func(ctx context.Context) (typedUser, error) {
		return typedUser{}, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Errorf("Remember() error = %v, want %v", err, expectedErr)
	}
}

func TestTyped_RememberUnmarshalError(t *testing.T) {
	// Pre-populate cache with invalid bytes, then call Remember
	cf := New()
	ctx := context.Background()

	_ = cf.Set(ctx, "bad:typed", []byte("not-json{{{"), time.Minute)

	tc := Typed[typedUser](cf)
	_, err := tc.Remember(ctx, "bad:typed", time.Minute, func(ctx context.Context) (typedUser, error) {
		t.Fatal("loader should not be called on cache hit")
		return typedUser{}, nil
	})
	if err == nil {
		t.Fatal("Remember() expected unmarshal error, got nil")
	}
}

