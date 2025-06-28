package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// отвязываем MarketService от сторонней библиотеки, и сможешь в тестах подменять Redis-зависимость (например, моками).
type RedisAdapter struct {
	client *redis.Client
}

func (r *RedisAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
