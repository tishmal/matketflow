package redis

import (
	"context"
	"marketflow/internal/domain/ports/output"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func NewRedisAdapter(client *redis.Client) output.RedisClient {
	return &RedisAdapter{
		client: client,
	}
}

// отвязываем MarketService от сторонней библиотеки, и сможешь в тестах подменять Redis-зависимость (например, моками).
type RedisAdapter struct {
	client *redis.Client
}

func (r *RedisAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result() // ← преобразует *StringCmd в string
}

// добавляет элемент (member) с числовым значением (score) в отсортированное множество (Sorted Set) Redis по указанному key.
func (r *RedisAdapter) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	cmd := r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	})
	return cmd.Err()
}

// получение данных за последнюю минуту
func (r *RedisAdapter) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// удаляем устаревшие записи, чтобы Redis не разрастался бесконечно, старше 60 секунд
func (r *RedisAdapter) ZRemRangeByScore(ctx context.Context, key string, min, max string) error {
	cmd := r.client.ZRemRangeByScore(ctx, key, min, max)
	return cmd.Err()
}
