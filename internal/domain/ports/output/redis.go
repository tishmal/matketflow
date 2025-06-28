package output

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// domain/ports/output/redis.go
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}
