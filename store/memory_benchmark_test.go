package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func BenchmarkMemoryStore_Get_Hit(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()
	_ = s.Set(ctx, "key", []byte("value"), time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(ctx, "key")
	}
}

func BenchmarkMemoryStore_Get_Miss(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Get(ctx, "nonexistent")
	}
}

func BenchmarkMemoryStore_Set(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()
	value := []byte("value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Set(ctx, "key", value, time.Minute)
	}
}

func BenchmarkMemoryStore_Delete(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Delete(ctx, "key")
	}
}

func BenchmarkMemoryStore_Set_UniqueKeys(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()
	value := []byte("value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key:%d", i)
		_ = s.Set(ctx, key, value, time.Minute)
	}
}

func BenchmarkMemoryStore_Parallel_GetSet(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()
	_ = s.Set(ctx, "key", []byte("value"), time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.Get(ctx, "key")
			_ = s.Set(ctx, "key", []byte("value"), time.Minute)
		}
	})
}

func BenchmarkMemoryStore_Parallel_Get(b *testing.B) {
	s := NewMemoryStore()
	ctx := context.Background()
	_ = s.Set(ctx, "key", []byte("value"), time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.Get(ctx, "key")
		}
	})
}
