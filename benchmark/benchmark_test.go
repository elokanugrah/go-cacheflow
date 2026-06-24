package benchmark

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elokanugrah/go-cacheflow"
)

type benchUser struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func BenchmarkRemember_CacheHit(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()

	// Pre-populate cache
	_ = cacheflow.Set(ctx, cf, "user:1", benchUser{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cacheflow.Remember(ctx, cf, "user:1", time.Minute,
			func(ctx context.Context) (benchUser, error) {
				return benchUser{Name: "Alice", Age: 30, Email: "alice@example.com"}, nil
			},
		)
	}
}

func BenchmarkRemember_CacheMiss(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()
	loader := func(ctx context.Context) (benchUser, error) {
		return benchUser{Name: "Bob", Age: 25, Email: "bob@example.com"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use a unique key for each iteration to force a cache miss on the shared instance
		key := "user:miss:" + strconv.Itoa(i)
		_, _ = cacheflow.Remember(ctx, cf, key, time.Minute, loader)
	}
}

func BenchmarkRemember_ConcurrentStampede(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		const concurrency = 1000
		wg.Add(concurrency)

		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				_, _ = cacheflow.Remember(ctx, cf, "stampede:key", time.Minute,
					func(ctx context.Context) (benchUser, error) {
						time.Sleep(10 * time.Millisecond) // simulate db latency
						return benchUser{Name: "Stampede", Age: 30, Email: "stampede@example.com"}, nil
					},
				)
			}()
		}
		wg.Wait()

		// Clean up the key so the next iteration forces a new stampede
		_ = cacheflow.Delete(ctx, cf, "stampede:key")
	}
}

func BenchmarkRemember_SingleFlight(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cacheflow.Remember(ctx, cf, "sf:contended", time.Minute,
				func(ctx context.Context) (benchUser, error) {
					// Simulate a slow DB call
					time.Sleep(time.Microsecond)
					return benchUser{Name: "SF", Age: 42, Email: "sf@example.com"}, nil
				},
			)
		}
	})
}

func BenchmarkGet_CacheHit(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()

	_ = cacheflow.Set(ctx, cf, "bench:get", benchUser{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cacheflow.Get[benchUser](ctx, cf, "bench:get")
	}
}

func BenchmarkSet(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()
	user := benchUser{Name: "Alice", Age: 30, Email: "alice@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cacheflow.Set(ctx, cf, "bench:set", user, time.Minute)
	}
}

// BenchmarkStampede_NoSingleFlight simulates 1000 concurrent requests
// WITHOUT any SingleFlight deduplication. Every goroutine calls the loader
// independently. This demonstrates the "naive" cache-aside approach where
// all concurrent cache misses result in separate database calls.
func BenchmarkStampede_NoSingleFlight(b *testing.B) {
	ctx := context.Background()
	loader := func(ctx context.Context) (benchUser, error) {
		time.Sleep(10 * time.Millisecond) // simulate db latency
		return benchUser{Name: "Stampede", Age: 30, Email: "stampede@example.com"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		var loaderCalls atomic.Int32
		const concurrency = 1000
		wg.Add(concurrency)

		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				loaderCalls.Add(1)
				_, _ = loader(ctx) // No deduplication — every goroutine calls the loader
			}()
		}
		wg.Wait()

		b.ReportMetric(float64(loaderCalls.Load()), "loader_calls/op")
	}
}

// BenchmarkStampede_WithSingleFlight simulates 1000 concurrent requests
// WITH Go-CacheFlow's built-in SingleFlight deduplication. Only one goroutine
// calls the loader; all others share the result. This demonstrates the
// stampede prevention that Go-CacheFlow provides out of the box.
func BenchmarkStampede_WithSingleFlight(b *testing.B) {
	cf := cacheflow.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		var loaderCalls atomic.Int32
		const concurrency = 1000
		wg.Add(concurrency)

		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()
				_, _ = cacheflow.Remember(ctx, cf, "stampede:cmp", time.Minute,
					func(ctx context.Context) (benchUser, error) {
						loaderCalls.Add(1)
						time.Sleep(10 * time.Millisecond) // simulate db latency
						return benchUser{Name: "Stampede", Age: 30, Email: "stampede@example.com"}, nil
					},
				)
			}()
		}
		wg.Wait()

		b.ReportMetric(float64(loaderCalls.Load()), "loader_calls/op")

		// Clean up so next iteration starts fresh
		_ = cacheflow.Delete(ctx, cf, "stampede:cmp")
	}
}

