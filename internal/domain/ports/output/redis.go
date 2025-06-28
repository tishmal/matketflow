package output

import (
	"context"
	"time"
)

// domain/ports/output/redis.go
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error) // ← возвращает уже обработанные данные
}
