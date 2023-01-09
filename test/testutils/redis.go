package testutils

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v9"
)

func WithTestRedis(fn func(r *redis.Client)) {
	r := acquireDB()
	defer releaseDB(r)

	fn(r)
}

var totalDBs = 0

func newTestRedis() *redis.Client {
	options := redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       totalDBs,
	}
	client := redis.NewClient(&options)

	cmd := client.Ping(context.Background())
	if cmd.Err() != nil {
		panic(cmd.Err())
	}

	totalDBs += 1
	return client
}

var availableDBs = make(chan *redis.Client, 16)

func acquireDB() *redis.Client {
	db := acquireDBFromPool()
	if db != nil {
		return db
	}

	return newTestRedis()
}

func releaseDB(db *redis.Client) {
	availableDBs <- db
}

func acquireDBFromPool() *redis.Client {
	select {
	case db := <-availableDBs:
		if err := resetDB(db); err != nil {
			panic(err)
		}
		return db
	default:
		return nil
	}
}

func resetDB(db *redis.Client) error {
	return db.FlushDB(context.Background()).Err()
}
