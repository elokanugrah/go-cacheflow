# ADR-0002: TTL = 0 Represents No Expiration

## Status

Accepted

## Context

Caching entries need expiration parameters. In Go-CacheFlow, a `time.Duration` parameter represents the Time-to-Live (TTL) of cache entries. We need to decide how to handle a TTL value of `0`.

## Decision

A TTL of `0` represents no expiration. Entries stored with TTL=0 will persist in the cache indefinitely until they are explicitly deleted.

## Consequences

- **Convenience**: Users can cache static or long-lived data indefinitely without fabricating large timeouts (like `100 * time.Hour`).
- **Implementation Alignment**: Both `MemoryStore` and `RedisStore` handle a TTL of `0` differently than positive TTLs:
  - Redis: `client.Set(ctx, key, value, 0)` sets the key without an expiration.
  - MemoryStore: Saves an empty/zero `time.Time` expiration field (`time.Time{}`), which is skipped during lazy checks.
