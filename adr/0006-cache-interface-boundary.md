# ADR-0006: Cache Interface Boundary

## Status

Accepted

## Context

Go-CacheFlow's `CacheFlow` struct is both the concrete orchestrator (holding a `store.Store`, `serializer.Serializer`, and `singleflight.Group`) and the entity that downstream code interacts with. As the project grows toward v0.2+ (metrics wrappers, tracing wrappers, mock caches), it becomes important to define a clear **interface boundary** so that:

1. Downstream code can depend on an interface, not a concrete struct.
2. Wrapper patterns (metrics, tracing, logging) can compose using decoration.
3. Test doubles can implement the same interface without importing internal components.

## Decision

We define a `Cache` interface in the root `cacheflow` package:

```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Remember(ctx context.Context, key string, ttl time.Duration, loader func(ctx context.Context) ([]byte, error)) ([]byte, error)
}
```

- `CacheFlow` struct satisfies `Cache`.
- `TypedCache[T]` accepts `Cache` (not `*CacheFlow`) to enable decoration.
- The `Cache` interface operates on raw `[]byte`, keeping serialization concerns in the generic layer (`Remember[T]`, `TypedCache[T]`).

## Rationale

### Why an interface at the raw bytes level?

Serialization is type-specific and belongs in the generic wrapper layer. The `Cache` interface stays simple and decoration-friendly because wrappers (metrics, tracing) don't need to know about `T`.

### Why `TypedCache` takes `Cache` instead of `*CacheFlow`?

This enables the following patterns in future versions:

```go
// v0.2: Metrics wrapper
metricsCache := metrics.Wrap(cf)
users := cacheflow.Typed[User](metricsCache)

// v0.3: Tracing wrapper
tracedCache := tracing.Wrap(cf)
users := cacheflow.Typed[User](tracedCache)

// v0.4: Mock cache for unit testing
mockCache := &MockCache{}
users := cacheflow.Typed[User](mockCache)
```

### Why not a separate `Rememberer` interface?

Splitting `Get/Set/Delete` from `Remember` into separate interfaces was considered, but rejected because:

- `Remember` is the primary API — most users never call `Get/Set/Delete` directly.
- Splitting would force wrappers to implement multiple interfaces.
- A single interface keeps the mental model simple.

## Consequences

- All future wrappers (metrics, tracing, mock) implement the `Cache` interface.
- `TypedCache[T]` remains a thin, allocation-free wrapper that delegates to `Cache`.
- The `Cache` interface is frozen for v0.x — breaking changes only at v1.0.
