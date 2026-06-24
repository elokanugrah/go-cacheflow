# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-24

### Added
- **Core Caching Engine**: Multi-backend cache supporting generic type serialization (`Get`, `Set`, `Delete`, `Remember`).
- **Pluggable Storage Layer**:
  - `memory`: High-performance, thread-safe in-memory storage.
  - `redis`: Redis backend using `go-redis/v9`.
- **Cache Stampede Prevention**: Integration of `golang.org/x/sync/singleflight` to collapse concurrent cache-miss loaders on the same key into a single execution.
- **Type-Safe Wrapper**: `TypedCache[T]` helper providing a type-safe generic client API (`Typed[T](cache)`).
- **Custom Serialization**: Extensible serializer interface with standard JSON implementation.
- **Robust Error Contract**: `ErrCacheMiss` sentinel, deterministic error propagation, and transparent loader failure behavior.
- **Deterministic Testing & Mocks**: Redis store integration tests refactored using `miniredis` to run deterministically.
- **Comprehensive Documentation**: Comprehensive README, executable examples, and 6 Architecture Decision Records (ADRs).
