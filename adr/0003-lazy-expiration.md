# ADR-0003: Lazy Expiration in MemoryStore

## Status

Accepted

## Context

When implementing an in-memory cache, expired items must be removed to free up memory and prevent returning stale results. We can clean up expired items using:
1. **Active Cleanup**: A background worker goroutine that runs periodically (e.g. using a ticker) and prunes expired items.
2. **Lazy Cleanup**: Checking the item's expiration timestamp during a `Get` operation. If expired, we delete it immediately and return `ErrCacheMiss`.

## Decision

We decide to use a **Lazy Expiration** strategy in `MemoryStore` to avoid background routines.

## Consequences

- **Resource Safety**: Avoids launching background goroutines, preventing potential goroutine leaks and minimizing background CPU overhead.
- **Simplicity**: Code remains extremely simple, focused on concurrent map operations guarded by `sync.RWMutex`.
- **Memory Overhead**: Expired items that are never accessed again will persist in memory until rewritten or until the application restarts. For the v0.1 MVP, this trade-off is accepted. A periodic active sweeper can be added as an option in the future if memory growth becomes an issue.
