package testutils

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v9"
)

func NewTestRedis() *redis.Client {
	options := redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	client := redis.NewClient(&options)

	cmd := client.Ping(context.Background())
	if cmd.Err() != nil {
		panic(cmd.Err())
	}

	return client
}

func WithTestRedis(fn func(r *redis.Client)) {
	r := NewTestRedis()
	defer r.Close()

	fn(r)
}
