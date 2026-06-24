package main

import (
	"context"
	"fmt"
	"time"

	"github.com/elokanugrah/go-cacheflow"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	cf := cacheflow.New()
	ctx := context.Background()

	// 1. Using package-level generic functions
	fmt.Println("--- Package-level API ---")
	user1, err := cacheflow.Remember(ctx, cf, "user:1", time.Second*5, func(ctx context.Context) (User, error) {
		fmt.Println("Fetching user 1 from database...")
		return User{ID: 1, Name: "Alice"}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched: %+v\n", user1)

	// Fetch again (should hit cache)
	user1Cached, err := cacheflow.Remember(ctx, cf, "user:1", time.Second*5, func(ctx context.Context) (User, error) {
		fmt.Println("Fetching user 1 from database (should not be called)...")
		return User{ID: 1, Name: "Alice"}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched (cached): %+v\n", user1Cached)

	// 2. Using TypedWrapper API (DX optimization)
	fmt.Println("\n--- TypedWrapper API ---")
	usersCache := cacheflow.Typed[User](cf)

	user2, err := usersCache.Remember(ctx, "user:2", time.Second*5, func(ctx context.Context) (User, error) {
		fmt.Println("Fetching user 2 from database...")
		return User{ID: 2, Name: "Bob"}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched: %+v\n", user2)

	// Fetch again (should hit cache)
	user2Cached, err := usersCache.Remember(ctx, "user:2", time.Second*5, func(ctx context.Context) (User, error) {
		fmt.Println("Fetching user 2 from database (should not be called)...")
		return User{ID: 2, Name: "Bob"}, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetched (cached): %+v\n", user2Cached)
}
