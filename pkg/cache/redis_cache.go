package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v9"
)

type RedisCache struct {
	redis *redis.Client
}

func NewRedisCache(redis *redis.Client) *RedisCache {
	return &RedisCache{
		redis: redis,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	cmd := r.redis.Get(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return "", ErrKeyNotFound
	}
	if cmd.Err() != nil {
		return "", cmd.Err()
	}

	return cmd.Val(), nil
}

func (r *RedisCache) Put(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.redis.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.redis.Del(ctx, key).Err()
}

func (r *RedisCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return r.redis.Expire(ctx, key, ttl).Err()
}
