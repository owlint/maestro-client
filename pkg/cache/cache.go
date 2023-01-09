package cache

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
}
