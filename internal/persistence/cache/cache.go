package cache

import (
	"context"
	"time"

	"github.com/KronusRodion/task-tracker/internal/api/tasks/usecase"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) usecase.Cache {
	return &redisCache{client: client}
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (r *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}