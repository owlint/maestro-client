package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
)

type redisCache struct {
	redis *redis.Client
}

func NewRedisCache(redis *redis.Client) Cache {
	return &redisCache{
		redis: redis,
	}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	cmd := r.redis.Get(ctx, key)
	if cmd.Err() != nil {
		return "", cmd.Err()
	}

	return cmd.Val(), nil
}

func (r *redisCache) Put(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.redis.Set(ctx, key, value, ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.redis.Del(ctx, key).Err()
}

func (r *redisCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return r.redis.Expire(ctx, key, ttl).Err()
}
