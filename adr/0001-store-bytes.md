# ADR-0001: Store Operates on Raw Bytes

## Status

Accepted

## Context

We need to define the data type for keys and values stored in our caching store implementations (e.g. MemoryStore, RedisStore). In Go, we can either:
1. Store typed objects using `any` (interface{}) and handle serialization inside the store.
2. Delegate serialization/deserialization to the orchestration layer, and have the stores deal strictly with raw byte slices (`[]byte`).

## Decision

We decide that the `Store` interface operates on raw `[]byte` values.

```go
type Store interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

## Consequences

- **Separation of Concerns**: The cache stores are completely decoupled from serialization formats and libraries. They only care about storage.
- **Portability**: Different store backends (Memory, Redis, Memcached, etc.) can be swapped easily since they all consume and emit `[]byte`.
- **Efficiency**: Serialization happens once at the orchestration layer before hitting the store, and deserialization happens once on retrieval.
