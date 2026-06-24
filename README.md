# Go-CacheFlow

[![CI](https://github.com/elokanugrah/go-cacheflow/actions/workflows/ci.yml/badge.svg)](https://github.com/elokanugrah/go-cacheflow/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/elokanugrah/go-cacheflow)](https://goreportcard.com/report/github.com/elokanugrah/go-cacheflow)
[![Go Reference](https://pkg.go.dev/badge/github.com/elokanugrah/go-cacheflow.svg)](https://pkg.go.dev/github.com/elokanugrah/go-cacheflow)

Go-CacheFlow is an opinionated, production-grade cache orchestration library for Go that makes cache-aside implementation simple, safe, and robust through a single `Remember()` API with built-in SingleFlight deduplication.

---

## Why Go-CacheFlow?

Caching is simple in theory, but implementing it correctly in production is hard. A typical "cache-aside" implementation often looks like this:

```go
// 🛑 Verbose, boilerplate-heavy, and vulnerable to cache stampedes
val, err := redisClient.Get(ctx, "user:123").Result()
if err == redis.Nil {
    // 1000 concurrent requests will hit the DB at the exact same time! (Cache Stampede)
    user, err := db.GetUser(ctx, 123)
    if err != nil {
        return nil, err
    }
    
    bytes, _ := json.Marshal(user)
    redisClient.Set(ctx, "user:123", bytes, time.Minute)
    return user, nil
} else if err != nil {
    return nil, err
}

var user User
json.Unmarshal([]byte(val), &user)
return &user, nil
```

With **Go-CacheFlow**, the entire flow is consolidated into a single, type-safe, stampede-proof call:

```go
// ✅ Clean, generic-first, and stampede-proof
user, err := cacheflow.Remember(ctx, cf, "user:123", time.Minute, func(ctx context.Context) (User, error) {
    return db.GetUser(ctx, 123)
})
```

By leveraging built-in **SingleFlight deduplication**, Go-CacheFlow guarantees that under massive concurrent traffic, **only one database call** is made to fetch the data. All other concurrent requests wait for the first execution and receive the identical result.

---

## How It Works

Go-CacheFlow acts as a smart orchestration layer between your application code, your cache backend, and your data source (database/API):

```text
Request for Key
      ↓
┌───────────┐
│ Cache Hit │ ── Yes ──> Return Value (Fast Path)
└───────────┘
      │ No
      ↓
┌──────────────┐
│ SingleFlight │ ── Joined by concurrent waiters
└──────────────┘
      │ (First request only)
      ↓
┌──────────────┐
│ Call Loader  │ ──> Fetches from DB / external API
└──────────────┘
      │
      ├─ Success ─> [Save to Cache] ─> Return Value to all waiters
      │
      └─ Failure ─> [Skip Caching]  ─> Propagate Error to all waiters
```

---

## Features

- **Simplest Cache-aside pattern**: Retrieve or fetch-and-store in one call.
- **Prevent Cache Stampedes**: Built-in, transparent SingleFlight deduplication for concurrent requests fetching the same key.
- **Generic-first API**: Type-safe cache retrieval and persistence out of the box.
- **Flexible DX Options**: Choose package-level generic functions or clean type-level wrappers.
- **Interchangeable Backends**: Out-of-the-box support for in-memory and Redis stores.

---

## Supported Stores

Go-CacheFlow defines a simple, backend-agnostic storage boundary interface (`store.Store`). The following storage adapters are supported out of the box:

- **MemoryStore** (`store/memory`): Thread-safe in-memory cache using a native Go map under a `sync.RWMutex`.
- **RedisStore** (`store/redis`): Distributed caching backend powered by [go-redis/v9](https://github.com/redis/go-redis).

To configure a custom store:
```go
import (
    "github.com/elokanugrah/go-cacheflow"
    "github.com/elokanugrah/go-cacheflow/store"
    "github.com/redis/go-redis/v9"
)

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
cf := cacheflow.New(
    cacheflow.WithStore(store.NewRedisStore(rdb)),
)
```

---

## Installation

```bash
go get github.com/elokanugrah/go-cacheflow
```

---

## Quick Start

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/elokanugrah/go-cacheflow"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// UserRepository simulates a real-world database repository
type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (User, error) {
	// Simulated database query or API call
	return User{ID: id, Name: "Alice"}, nil
}

func main() {
	// 1. Initialize Go-CacheFlow (uses memory store by default)
	cf := cacheflow.New()
	ctx := context.Background()
	userRepo := &UserRepository{}

	// 2. Remember: Fetch from cache, or fallback to loader and cache it automatically.
	// Built-in SingleFlight guarantees only 1 call to GetByID under concurrent stampedes.
	user, err := cacheflow.Remember(ctx, cf, "user:123", time.Minute, func(ctx context.Context) (User, error) {
		return userRepo.GetByID(ctx, 123)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched user: %+v\n", user)

	// 3. Or use Typed wrapper for clean, type-safe caching of a specific entity
	userCache := cacheflow.Typed[User](cf)
	user, err = userCache.Remember(ctx, "user:123", time.Minute, func(ctx context.Context) (User, error) {
		return userRepo.GetByID(ctx, 123)
	})
}
```

---

## Running the Examples

Go-CacheFlow contains executable examples in the `example/` directory.

### Basic Memory Example
```bash
go run ./example/basic
```

### Redis Example
The Redis example expects a Redis instance running at `localhost:6379`. You can spin one up using Docker:
```bash
docker run -d --name cacheflow-redis -p 6379:6379 redis
go run ./example/redis
```

---

## Performance

The following benchmarks were run on Go 1.25 on a standard development machine (`AMD Ryzen 7 4800H`):

| Benchmark Scenario | Operations | Time | Memory / Op | Allocations / Op |
|---|---|---|---|---|
| `Get (Cache Hit)` | 1,000,000 | **1.20 µs / op** | 296 B / op | 7 allocs / op |
| `Remember (Cache Hit)` | 1,000,000 | **1.14 µs / op** | 296 B / op | 7 allocs / op |
| `Remember (Cache Miss)` | 529,345 | **2.68 µs / op** | 843 B / op | 13 allocs / op |
| `Remember (SingleFlight)` | 4,103,275 | **0.28 µs / op** | 287 B / op | 7 allocs / op |
| `Set (Cache Write)` | 2,822,770 | **0.37 µs / op** | 112 B / op | 2 allocs / op |

### Concurrency & Stampede Resistance
We simulated a massive cache stampede with **1,000 concurrent goroutines** querying the same expired key with a **10ms database latency penalty**:

- **Without SingleFlight**: All 1,000 requests hit the database simultaneously, incurring significant database connection overhead, query latency, and resource starvation.
- **With Go-CacheFlow**: Only **1** loader execution was triggered. The remaining 999 goroutines shared the exact same fetched result. The benchmark completes in **~11 ms** total (the single database latency + tiny framework overhead).

---


## Error Handling

Go-CacheFlow has a deterministic error contract designed to make your applications resilient.

### Detecting Cache Misses
A cache miss returns `store.ErrCacheMiss` (also re-exported as `cacheflow.ErrCacheMiss`). Always check using `errors.Is`:
```go
val, err := cacheflow.Get[User](ctx, cf, "user:123")
if errors.Is(err, cacheflow.ErrCacheMiss) {
    // Key does not exist or has expired
}
```

### Error Sources

| Source | When | Behavior |
|---|---|---|
| **Cache miss** (`ErrCacheMiss`) | Key not found or expired | Returned from `Get`. In `Remember`, triggers the loader instead. |
| **Loader error** | Loader function returns an error | Error is propagated directly to all waiting callers. **Not cached.** Subsequent requests retry the loader. |
| **Serializer error** | `Marshal` or `Unmarshal` fails | Returned directly. For example, storing an unmarshalable type or reading corrupted cache data. |
| **Store error** | Backend connection or command failure | Returned directly. For example, Redis connection refused or context cancelled. |

### Loader Error Behavior in `Remember()`
- Cache hit → cached value returned, loader is never called.
- Cache miss → loader is called via SingleFlight.
  - Loader succeeds → value is cached and returned.
  - Loader fails → error is **propagated directly** to the caller(s) and is **not cached**.
- In stampede scenarios, if the single loader call fails, the error is broadcast to all waiting concurrent callers.

### Error Propagation Rules
- All errors are returned **directly** (unwrapped). Use `errors.Is` to match sentinel errors.
- If caching fails (store write error) after the loader succeeds, the loaded value is **still returned** to the caller. The data is not lost — it just wasn't cached.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.