package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func newRateLimiter(url string) (*RateLimiter, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Println("connected to redis")
	return &RateLimiter{
		client: client,
		limit:  10,
		window: 1 * time.Minute,
	}, nil
}

func (r *RateLimiter) isAllowed(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("rate_limit:upload:%s", userID)

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("increment counter: %w", err)
	}

	if count == 1 {
		if err := r.client.Expire(ctx, key, r.window).Err(); err != nil {
			return false, fmt.Errorf("set expiry: %w", err)
		}
	}

	return count <= int64(r.limit), nil
}

func (r *RateLimiter) close() {
	if err := r.client.Close(); err != nil {
		log.Printf("close redis client: %v", err)
	}
}
