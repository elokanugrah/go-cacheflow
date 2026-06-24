package store

import (
	"context"
	"sync"
	"time"
)

// entry represents a cached item with its value and optional expiration time.
type entry struct {
	value     []byte
	expiresAt time.Time
	noExpiry  bool
}

// isExpired reports whether the entry has expired.
func (e entry) isExpired() bool {
	if e.noExpiry {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// MemoryStore implements the Store interface using an in-memory map
// protected by a sync.RWMutex.
//
// TTL expiration is checked lazily on Get — expired entries are removed
// when they are accessed. This avoids the overhead of background
// goroutines and timers.
//
// MemoryStore is safe for concurrent use by multiple goroutines.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]entry
}

// NewMemoryStore creates a new MemoryStore ready for use.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]entry),
	}
}

// Get retrieves the cached value for the given key.
//
// Returns ErrCacheMiss if the key does not exist or has expired.
// Expired entries are lazily deleted on access.
func (s *MemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	e, ok := s.items[key]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrCacheMiss
	}

	if e.isExpired() {
		s.mu.Lock()
		// Double-check after acquiring write lock to avoid deleting
		// a freshly-set entry from another goroutine.
		if current, exists := s.items[key]; exists && current.isExpired() {
			delete(s.items, key)
		}
		s.mu.Unlock()
		return nil, ErrCacheMiss
	}

	return e.value, nil
}

// Set stores a value with the given key and time-to-live duration.
//
// If ttl is 0, the entry does not expire and lives until explicitly
// deleted. If ttl is greater than 0, the entry expires after the
// specified duration.
func (s *MemoryStore) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	e := entry{
		value: value,
	}

	if ttl == 0 {
		e.noExpiry = true
	} else {
		e.expiresAt = time.Now().Add(ttl)
	}

	s.mu.Lock()
	s.items[key] = e
	s.mu.Unlock()

	return nil
}

// Delete removes the cached value for the given key.
//
// Delete is idempotent — deleting a non-existent key does not return
// an error.
func (s *MemoryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	delete(s.items, key)
	s.mu.Unlock()

	return nil
}
