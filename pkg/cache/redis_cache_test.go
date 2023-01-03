package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/owlint/maestro-client/pkg/cache"
	"github.com/owlint/maestro-client/test/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)

		key := testutils.RandomStr()
		cache.Put(context.Background(), key, "value", time.Minute)

		t.Run("key exists", func(t *testing.T) {
			v, err := cache.Get(context.Background(), key)

			assert.NoError(t, err)
			assert.Equal(t, "value", v)
		})

		t.Run("get twice", func(t *testing.T) {
			v, err := cache.Get(context.Background(), key)

			assert.NoError(t, err)
			assert.Equal(t, "value", v)
		})

		t.Run("key doesn't exists", func(t *testing.T) {
			v, err := cache.Get(context.Background(), "invalid")

			assert.Error(t, err)
			assert.Empty(t, v)
		})
	})
}

func TestPut(t *testing.T) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)
		key := testutils.RandomStr()

		t.Run("new key", func(t *testing.T) {
			err := cache.Put(context.Background(), key, "value", time.Minute)
			assert.NoError(t, err)

			v, err := cache.Get(context.Background(), key)
			assert.NoError(t, err)
			assert.Equal(t, "value", v)

			assert.Equal(t, time.Minute, r.TTL(context.Background(), key).Val())
		})

		t.Run("update value and TTL", func(t *testing.T) {
			err := cache.Put(context.Background(), key, "value2", time.Hour)
			assert.NoError(t, err)

			v, err := cache.Get(context.Background(), key)
			assert.NoError(t, err)
			assert.Equal(t, "value2", v)

			assert.Equal(t, time.Hour, r.TTL(context.Background(), key).Val())
		})
	})
}

func TestDelete(t *testing.T) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)
		key := testutils.RandomStr()

		t.Run("delete existing", func(t *testing.T) {
			err := cache.Put(context.Background(), key, "value", time.Minute)
			assert.NoError(t, err)

			err = cache.Delete(context.Background(), key)
			assert.NoError(t, err)

			_, err = cache.Get(context.Background(), key)
			assert.Error(t, err)
		})

		t.Run("delete not existing should be noop", func(t *testing.T) {
			err := cache.Delete(context.Background(), "invalid")
			assert.NoError(t, err)
		})
	})
}

func TestSetTTL(t *testing.T) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)
		key := testutils.RandomStr()

		err := cache.Put(context.Background(), key, "value", time.Minute)
		assert.NoError(t, err)

		err = cache.SetTTL(context.Background(), key, time.Hour)
		assert.NoError(t, err)

		assert.Equal(t, time.Hour, r.TTL(context.Background(), key).Val())
	})
}
