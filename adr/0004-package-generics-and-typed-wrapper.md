# ADR-0004: Package-level Generics and Typed Wrapper

## Status

Accepted

## Context

Go method signatures currently do not support declaring type parameters (e.g. `func (cf *CacheFlow) Get[T]()` is illegal in Go). We want users to be able to retrieve cached objects in a typed manner without type assertions.

## Decision

We decide to provide a hybrid API design:
1. **Package-level Generics**: Define functions like `Get[T]`, `Set[T]`, and `Remember[T]` at the package level that accept the orchestration struct `*CacheFlow`.
2. **`TypedCache[T]` Wrapper**: Implement a typed wrapper struct instantiated via `Typed[T](cache Cache)` that holds a `Cache` and offers type-parameter-free methods (`Get`, `Set`, `Remember`).

## Consequences

- **Developer Experience (DX)**: Callers can use package-level functions directly:
  ```go
  user, err := cacheflow.Remember[User](ctx, cf, key, ttl, loader)
  ```
  Or construct a typed cache client:
  ```go
  users := cacheflow.Typed[User](cf)
  user, err := users.Remember(ctx, key, ttl, loader)
  ```
- **Extensibility**: By making `TypedCache` wrap the generic `Cache` interface (rather than the concrete `*CacheFlow`), we make it possible to transparently apply wrappers (like metrics, logging, tracing, or mocking) underneath the typed interface.
