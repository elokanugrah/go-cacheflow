package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elokanugrah/go-cacheflow"
	"github.com/elokanugrah/go-cacheflow/store"
	"github.com/redis/go-redis/v9"
)

type Product struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

func main() {
	// Initialize go-redis client
	// This assumes a Redis instance is running on localhost:6379
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	// Verify connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Redis is not available: %v\n", err)
		fmt.Println("Please run: docker run -d -p 6379:6379 redis")
		return
	}
	defer rdb.Close()

	// Initialize CacheFlow using RedisStore
	cf := cacheflow.New(
		cacheflow.WithStore(store.NewRedisStore(rdb)),
	)

	// Create typed cache wrapper
	productsCache := cacheflow.Typed[Product](cf)

	// Fetch product (cache miss)
	fmt.Println("Fetching product first time (cache miss)...")
	p1, err := productsCache.Remember(ctx, "prod:999", time.Minute, func(ctx context.Context) (Product, error) {
		return Product{SKU: "IPHONE15", Price: 999.99}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Result: %+v\n", p1)

	// Fetch product again (cache hit)
	fmt.Println("Fetching product second time (cache hit)...")
	p2, err := productsCache.Remember(ctx, "prod:999", time.Minute, func(ctx context.Context) (Product, error) {
		fmt.Println("Should not be called!")
		return Product{}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Result: %+v\n", p2)
}
