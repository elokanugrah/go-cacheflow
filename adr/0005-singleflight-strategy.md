# ADR-0005: SingleFlight Strategy

## Status

Accepted

## Context

A key value proposition of Go-CacheFlow is **stampede prevention** via SingleFlight. The initial MVP draft used a custom `sfGroup` implementation to keep dependency count minimal. However, correctness is paramount for concurrency-critical parts.

## Decision

We decide to:
1. Replace the custom SingleFlight implementation with the official Go package `golang.org/x/sync/singleflight`.
2. Hide `singleflight.Group` completely internal to `CacheFlow` to keep the public API surface clean.

## Rationale

- **Correctness and Reliability**: `x/sync/singleflight` is battle-tested, standard in the Go ecosystem, and maintained by the Go team.
- **Dependency Rules**: The official Go packages (`golang.org/x/...`) are prioritized directly below the standard library (rule E-06).
- **Encapsulation**: Exposing SingleFlight increases public API footprint and couples users to implementation details. By encapsulating it within `CacheFlow.Remember`, users just buy the "stampede prevention" benefit transparently.
