package cacheflow

import (
	"github.com/elokanugrah/go-cacheflow/serializer"
	"github.com/elokanugrah/go-cacheflow/store"
)

// Option is a function that configures a CacheFlow instance.
//
// Options are applied in order during New(). Later options
// override earlier ones for the same configuration field.
type Option func(*CacheFlow)

// WithStore sets the cache storage backend for the CacheFlow instance.
//
// If not provided, CacheFlow defaults to a MemoryStore.
//
// Example:
//
//	cf := cacheflow.New(
//	    cacheflow.WithStore(store.NewRedisStore(redisClient)),
//	)
func WithStore(s store.Store) Option {
	return func(cf *CacheFlow) {
		cf.store = s
	}
}

// WithSerializer sets the serializer for the CacheFlow instance.
//
// If not provided, CacheFlow defaults to a JSONSerializer.
//
// Example:
//
//	cf := cacheflow.New(
//	    cacheflow.WithSerializer(myCustomSerializer),
//	)
func WithSerializer(s serializer.Serializer) Option {
	return func(cf *CacheFlow) {
		cf.serializer = s
	}
}
