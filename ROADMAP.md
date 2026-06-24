# Go-CacheFlow Roadmap

This document outlines the strategic milestones and features planned for the future versions of Go-CacheFlow.

---

## 🌟 Milestone 1: Observability & Instrumentation (v0.2)
Focuses on providing deep insights into cache performance, hit/miss ratios, and latency.

- [ ] **Prometheus Metrics Wrapper**: Implement an optional decorator/wrapper over the `Cache` interface to export cache metrics (Hits, Misses, Sets, Deletes, Latencies).
- [ ] **OpenTelemetry Integration (v0.3)**: Implement structured tracing wrappers to track caching operations across distributed requests.
- [ ] **Custom Event Listeners**: Add hooks/events for Cache operations (e.g., `OnHit`, `OnMiss`, `OnEviction`).

---

## 🛠️ Milestone 2: Developer Experience & Testing Utilities (v0.4)
Simplifies testing for client applications using Go-CacheFlow.

- [ ] **Mock Cache Implementation**: A robust, pre-built mock store that makes it easy for downstream services to mock caching behavior without setting up memory/Redis instances in unit tests.
- [ ] **ADR-0006: Cache Interface Boundary**: Evaluate refining the boundary between the store, Cache interface, and TypedCache to allow easier mocking and custom decoration.

---

## 🚀 Milestone 3: Advanced Architectures (v0.5)
Enables multi-tiered caching strategies to optimize latency and network overhead.

- [ ] **Multi-Level Caching (L1/L2)**: Introduce a hybrid cache coordinator (`L1` local memory cache with shorter TTL, `L2` Redis shared cache with longer TTL).
- [ ] **Cache Invalidations (Pub/Sub)**: In a multi-replica setup, broadcast local L1 cache invalidations via Redis Pub/Sub when a key is modified.
- [ ] **Compression Layer**: Optional compression (e.g., zstd, gzip, snappy) for large JSON payloads before writing to Redis.

---

## 📊 Performance & Marketing Hardening (Ongoing)
- [ ] **Stampede Comparison Benchmarks**: Create side-by-side benchmarks demonstrating performance *with* and *without* SingleFlight (`BenchmarkStampede_NoSingleFlight` vs `BenchmarkStampede_WithSingleFlight`).
- [ ] **Scalability Stress Benchmarks**: Profile latency and contention with concurrent stampedes up to 10,000 goroutines.
- [ ] **Store Benchmarks**: Establish base performance metrics for the Redis store under highly concurrent loads.
- [ ] **Visual Architecture Diagrams**: Add a visual architecture block diagram to the README to make the workflow easier to understand at a glance.
