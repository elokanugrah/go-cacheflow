package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_GetSet(t *testing.T) {
	s := NewMemoryStore()
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

func TestMemoryStore_GetMiss(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get(nonexistent) error = %v, want ErrCacheMiss", err)
	}
}

func TestMemoryStore_TTLExpiration(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	err := s.Set(ctx, "expiring", []byte("data"), 50*time.Millisecond)
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

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, err = s.Get(ctx, "expiring")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("Get() after expiry error = %v, want ErrCacheMiss", err)
	}
}

func TestMemoryStore_ZeroTTLNoExpiration(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	err := s.Set(ctx, "forever", []byte("persistent"), 0)
	if err != nil {
		t.Fatalf("Set() unexpected error: %v", err)
	}

	// Should still be available after a short sleep
	time.Sleep(50 * time.Millisecond)

	val, err := s.Get(ctx, "forever")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(val) != "persistent" {
		t.Errorf("Get() = %q, want %q", string(val), "persistent")
	}
}

func TestMemoryStore_OverwriteExistingKey(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	_ = s.Set(ctx, "key", []byte("v1"), time.Minute)
	_ = s.Set(ctx, "key", []byte("v2"), time.Minute)

	val, err := s.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if string(val) != "v2" {
		t.Errorf("Get() = %q, want %q", string(val), "v2")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := NewMemoryStore()
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

func TestMemoryStore_DeleteNonexistent(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	err := s.Delete(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Delete(nonexistent) unexpected error: %v", err)
	}
}

func TestMemoryStore_ConcurrentReadWrite(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := "key"
				val := []byte("value")
				_ = s.Set(ctx, key, val, time.Minute)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = s.Get(ctx, "key")
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryStore_ImplementsInterface(t *testing.T) {
	// Compile-time check that MemoryStore implements Store
	var _ Store = (*MemoryStore)(nil)
}
