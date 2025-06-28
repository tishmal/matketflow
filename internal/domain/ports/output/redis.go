package output

import (
	"context"
	"time"
)

// domain/ports/output/redis.go
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error) // ← возвращает уже обработанные данные
	ZAdd(ctx context.Context, key string, score float64, member interface{}) error
	ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error)
}
