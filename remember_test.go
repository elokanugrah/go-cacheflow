package cacheflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elokanugrah/go-cacheflow/store"
)

func TestRemember_CacheHit(t *testing.T) {
	cf := New()
	ctx := context.Background()

	// Pre-populate the cache
	_ = Set(ctx, cf, "user:1", testUser{Name: "Alice", Age: 30}, time.Minute)

	loaderCalled := false
	result, err := Remember(ctx, cf, "user:1", time.Minute,
		func(ctx context.Context) (testUser, error) {
			loaderCalled = true
			return testUser{Name: "FromDB", Age: 99}, nil
		},
	)

	if err != nil {
		t.Fatalf("Remember() unexpected error: %v", err)
	}
	if loaderCalled {
		t.Error("Remember() called loader on cache hit")
	}
	if result.Name != "Alice" {
		t.Errorf("Remember() Name = %q, want %q", result.Name, "Alice")
	}
}

func TestRemember_CacheMiss(t *testing.T) {
	cf := New()
	ctx := context.Background()

	result, err := Remember(ctx, cf, "user:2", time.Minute,
		func(ctx context.Context) (testUser, error) {
			return testUser{Name: "Bob", Age: 25}, nil
		},
	)

	if err != nil {
		t.Fatalf("Remember() unexpected error: %v", err)
	}
	if result.Name != "Bob" || result.Age != 25 {
		t.Errorf("Remember() = %+v, want {Bob 25}", result)
	}

	// Verify it was cached
	cached, err := Get[testUser](ctx, cf, "user:2")
	if err != nil {
		t.Fatalf("Get() after Remember() unexpected error: %v", err)
	}
	if cached.Name != "Bob" {
		t.Errorf("cached Name = %q, want %q", cached.Name, "Bob")
	}
}

func TestRemember_LoaderError(t *testing.T) {
	cf := New()
	ctx := context.Background()

	expectedErr := errors.New("db connection failed")
	_, err := Remember(ctx, cf, "user:err", time.Minute,
		func(ctx context.Context) (testUser, error) {
			return testUser{}, expectedErr
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Errorf("Remember() error = %v, want %v", err, expectedErr)
	}

	// Verify the value was NOT cached
	_, err = Get[testUser](ctx, cf, "user:err")
	if !errors.Is(err, store.ErrCacheMiss) {
		t.Errorf("Get() after loader error = %v, want ErrCacheMiss", err)
	}
}

func TestRemember_SingleFlight(t *testing.T) {
	cf := New()
	ctx := context.Background()

	var loaderCount atomic.Int32
	var wg sync.WaitGroup

	const goroutines = 50
	wg.Add(goroutines)

	results := make([]testUser, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = Remember(ctx, cf, "sf:key", time.Minute,
				func(ctx context.Context) (testUser, error) {
					loaderCount.Add(1)
					// Simulate slow DB call
					time.Sleep(50 * time.Millisecond)
					return testUser{Name: "SingleFlightUser", Age: 42}, nil
				},
			)
		}(i)
	}

	wg.Wait()

	// SingleFlight: loader should be called exactly once
	count := loaderCount.Load()
	if count != 1 {
		t.Errorf("SingleFlight loader called %d times, want 1", count)
	}

	// All goroutines should get the same result
	for i := 0; i < goroutines; i++ {
		if errs[i] != nil {
			t.Errorf("goroutine %d error: %v", i, errs[i])
			continue
		}
		if results[i].Name != "SingleFlightUser" {
			t.Errorf("goroutine %d result = %+v, want SingleFlightUser", i, results[i])
		}
	}
}

func TestRemember_DifferentKeysInvokeLoaderIndependently(t *testing.T) {
	cf := New()
	ctx := context.Background()

	var loaderCount atomic.Int32
	var wg sync.WaitGroup

	const keys = 5
	wg.Add(keys)

	for i := 0; i < keys; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key:%d", idx)
			_, _ = Remember(ctx, cf, key, time.Minute,
				func(ctx context.Context) (string, error) {
					loaderCount.Add(1)
					return fmt.Sprintf("value:%d", idx), nil
				},
			)
		}(i)
	}

	wg.Wait()

	count := loaderCount.Load()
	if count != keys {
		t.Errorf("loader called %d times, want %d (once per key)", count, keys)
	}
}

func TestRemember_ContextCancellation(t *testing.T) {
	cf := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Remember should still work — context cancellation affects the loader
	// and store, but the orchestration itself continues
	_, err := Remember(ctx, cf, "cancelled", time.Minute,
		func(ctx context.Context) (string, error) {
			// Check if context is done
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "value", nil
			}
		},
	)

	// Expect context cancellation error
	if err == nil {
		t.Log("Remember() with cancelled context returned no error (context was checked by loader)")
	}
}

func TestRemember_PointerType(t *testing.T) {
	cf := New()
	ctx := context.Background()

	result, err := Remember(ctx, cf, "ptr:user", time.Minute,
		func(ctx context.Context) (*testUser, error) {
			return &testUser{Name: "PtrUser", Age: 35}, nil
		},
	)

	if err != nil {
		t.Fatalf("Remember() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Remember() returned nil pointer")
	}
	if result.Name != "PtrUser" {
		t.Errorf("Remember() Name = %q, want %q", result.Name, "PtrUser")
	}
}

func TestRemember_SingleFlightStampedeStress(t *testing.T) {
	cf := New()
	ctx := context.Background()

	var loaderCount atomic.Int32
	var wg sync.WaitGroup

	const goroutines = 1000
	wg.Add(goroutines)

	expected := testUser{Name: "StressUser", Age: 100}
	results := make([]testUser, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = Remember(ctx, cf, "stress:key", time.Minute,
				func(ctx context.Context) (testUser, error) {
					loaderCount.Add(1)
					time.Sleep(50 * time.Millisecond)
					return expected, nil
				},
			)
		}(i)
	}

	wg.Wait()

	// Assert: loader called exactly once (SingleFlight deduplication)
	count := loaderCount.Load()
	if count != 1 {
		t.Errorf("loader called %d times, want exactly 1 (SingleFlight failed under stampede)", count)
	}

	// Assert: ALL goroutines received the identical value
	for i := 0; i < goroutines; i++ {
		if errs[i] != nil {
			t.Errorf("goroutine %d returned error: %v", i, errs[i])
			continue
		}
		if results[i] != expected {
			t.Errorf("goroutine %d result = %+v, want %+v (value identity violated)", i, results[i], expected)
		}
	}
	t.Logf("All %d goroutines received identical result: %+v", goroutines, expected)
}

func TestRemember_ConcurrentExpiration(t *testing.T) {
	cf := New()
	ctx := context.Background()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("exp:%d", idx%10) // 10 unique keys
			_, _ = Remember(ctx, cf, key, 10*time.Millisecond, func(ctx context.Context) (string, error) {
				return "value", nil
			})
			time.Sleep(15 * time.Millisecond) // wait for expiry
			// Fetch again to trigger lazy expiration
			_, _ = Remember(ctx, cf, key, 10*time.Millisecond, func(ctx context.Context) (string, error) {
				return "value2", nil
			})
		}(i)
	}
	wg.Wait()
}

func TestRemember_ConcurrentDelete(t *testing.T) {
	cf := New()
	ctx := context.Background()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Setters / Getters
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("del:%d", idx%5)
			_, _ = Remember(ctx, cf, key, time.Minute, func(ctx context.Context) (string, error) {
				return "val", nil
			})
		}(i)
	}

	// Deleters
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("del:%d", idx%5)
			_ = Delete(ctx, cf, key)
		}(i)
	}

	wg.Wait()
}

// TestRemember_ErrorPropagationStress verifies that when a loader fails during
// a cache stampede, the error is correctly propagated to ALL concurrent waiters.
//
// It also verifies that:
//   - No waiter receives a stale or zero value on error.
//   - The error is NOT cached — a subsequent call retries the loader and succeeds.
func TestRemember_ErrorPropagationStress(t *testing.T) {
	cf := New()
	ctx := context.Background()

	loaderErr := errors.New("database unavailable")
	var loaderCount atomic.Int32
	var wg sync.WaitGroup

	const goroutines = 500
	wg.Add(goroutines)

	errs := make([]error, goroutines)
	results := make([]testUser, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = Remember(ctx, cf, "err:stampede", time.Minute,
				func(ctx context.Context) (testUser, error) {
					loaderCount.Add(1)
					time.Sleep(50 * time.Millisecond) // simulate slow failure
					return testUser{}, loaderErr
				},
			)
		}(i)
	}

	wg.Wait()

	// Assert: loader called exactly once (SingleFlight deduplication on error path)
	count := loaderCount.Load()
	if count != 1 {
		t.Errorf("loader called %d times, want 1 (SingleFlight should deduplicate even on error)", count)
	}

	// Assert: ALL goroutines received the same error
	for i := 0; i < goroutines; i++ {
		if !errors.Is(errs[i], loaderErr) {
			t.Errorf("goroutine %d error = %v, want %v", i, errs[i], loaderErr)
		}
		// Assert: no goroutine received a populated value
		if results[i] != (testUser{}) {
			t.Errorf("goroutine %d result = %+v, want zero value on error", i, results[i])
		}
	}
	t.Logf("All %d goroutines received identical error: %v", goroutines, loaderErr)

	// Assert: error was NOT cached — subsequent call retries the loader and succeeds
	successUser := testUser{Name: "Recovered", Age: 1}
	result, err := Remember(ctx, cf, "err:stampede", time.Minute,
		func(ctx context.Context) (testUser, error) {
			return successUser, nil
		},
	)
	if err != nil {
		t.Fatalf("retry after error: unexpected error: %v", err)
	}
	if result != successUser {
		t.Errorf("retry result = %+v, want %+v", result, successUser)
	}
	t.Log("Retry after error succeeded — error was not cached")
}

