package maestro_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/owlint/maestro-client/pkg/cache"
	"github.com/owlint/maestro-client/pkg/maestro"
	"github.com/owlint/maestro-client/test/testutils"
	"github.com/stretchr/testify/assert"
)

type testCache struct {
	cache map[string]string
}

func newTestCache() *testCache {
	return &testCache{
		cache: map[string]string{},
	}
}
func (c *testCache) Get(ctx context.Context, key string) (string, error) {
	return c.cache[key], nil
}
func (c *testCache) Put(ctx context.Context, key, value string, ttl time.Duration) error {
	c.cache[key] = value
	return nil
}
func (c *testCache) Delete(ctx context.Context, key string) error {
	delete(c.cache, key)
	return nil
}
func (c *testCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (c *testCache) KeyForValue(value string) string {
	for key, cachedValue := range c.cache {
		if value == cachedValue {
			return key
		}
	}
	return ""
}

func WithRedisCachedClient(t *testing.T, fn func(client *maestro.CachedClient, mock *maestro.MockMaestro), cachedQueues ...string) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)
		maestroMock := maestro.NewMockMaestro(gomock.NewController(t))
		cachedClient := maestro.NewCachedClient(maestroMock, cache, cachedQueues, 900*time.Second)

		fn(cachedClient, maestroMock)
	})
}

func WithTestCachedClient(t *testing.T, fn func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache), cachedQueues ...string) {
	cache := newTestCache()
	maestroMock := maestro.NewMockMaestro(gomock.NewController(t))
	cachedClient := maestro.NewCachedClient(maestroMock, cache, cachedQueues, 900*time.Second)

	fn(cachedClient, maestroMock, cache)
}

func makeCompleteCachedTask(t *testing.T, client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) string {
	taskID := uuid.NewString()

	mock.EXPECT().CreateTask("owner", "cached", gomock.Not("payload"), gomock.Any()).Return(taskID, nil)
	taskID, err := client.CreateTask("owner", "cached", "payload", *maestro.NewCreateTaskOptions().WithStartTimeout(10))
	assert.NoError(t, err)

	mock.EXPECT().CompleteTask(taskID, gomock.Not("result")).Return(nil)
	mock.EXPECT().TaskState(gomock.Any()).Return(&maestro.Task{
		TaskID:    taskID,
		TaskQueue: "cached",
		Payload:   cache.KeyForValue("payload"),
	}, nil)

	err = client.CompleteTask(taskID, "result")
	assert.NoError(t, err)

	assert.NotEmpty(t, cache.KeyForValue("payload"))
	assert.NotEmpty(t, cache.KeyForValue("result"))

	return taskID
}

func TestCreateTask(t *testing.T) {
	t.Run("queue not in cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			mock.EXPECT().CreateTask("owner", "not-cached", "payload", *maestro.NewCreateTaskOptions()).Return("task-id", nil)

			res, err := client.CreateTask("owner", "not-cached", "payload")
			assert.NoError(t, err)
			assert.Equal(t, "task-id", res)
			assert.Empty(t, cache.KeyForValue("payload"))
		})
	})

	t.Run("queue in cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			mock.EXPECT().CreateTask("owner", "cached", gomock.Not("payload"), gomock.Any()).Return("task-id", nil)

			res, err := client.CreateTask("owner", "cached", "payload", *maestro.NewCreateTaskOptions().WithStartTimeout(time.Hour))
			assert.NoError(t, err)
			assert.Equal(t, "task-id", res)
			assert.NotEmpty(t, cache.KeyForValue("payload"))
		}, "cached")
	})

	t.Run("queue in cache no StartTimeout", func(t *testing.T) {
		WithRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			_, err := client.CreateTask("owner", "cached", "payload")
			assert.Error(t, err)
		}, "cached")
	})
}

func TestCompleteTask(t *testing.T) {
	t.Run("with cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := uuid.NewString()

			mock.EXPECT().CreateTask("owner", "cached", gomock.Not("payload"), gomock.Any()).Return(taskID, nil)
			taskID, err := client.CreateTask("owner", "cached", "payload", *maestro.NewCreateTaskOptions().WithStartTimeout(10))
			assert.NoError(t, err)

			mock.EXPECT().CompleteTask(taskID, gomock.Not("result"))
			mock.EXPECT().TaskState(taskID).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "cached",
				Payload:   cache.KeyForValue("payload"),
			}, nil)

			err = client.CompleteTask(taskID, "result")
			assert.NoError(t, err)
			assert.NotEmpty(t, cache.KeyForValue("result"))
		}, "cached")
	})

	t.Run("without cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := uuid.NewString()

			mock.EXPECT().CompleteTask(taskID, "result").Return(nil)
			mock.EXPECT().TaskState(taskID).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "not-cached",
				Payload:   "payload",
			}, nil)

			err := client.CompleteTask(taskID, "result")
			assert.NoError(t, err)
			assert.Empty(t, cache.KeyForValue("result"))
		})
	})
}

func TestFailTask(t *testing.T) {
	t.Run("with cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := uuid.NewString()

			mock.EXPECT().CreateTask("owner", "cached", gomock.Not("payload"), gomock.Any()).Return(taskID, nil)
			taskID, err := client.CreateTask("owner", "cached", "payload", *maestro.NewCreateTaskOptions().WithStartTimeout(10))
			assert.NoError(t, err)

			mock.EXPECT().FailTask(taskID)
			mock.EXPECT().TaskState(taskID).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "cached",
				Payload:   cache.KeyForValue("payload"),
			}, nil)

			err = client.FailTask(taskID)
			assert.NoError(t, err)
		}, "cached")
	})

	t.Run("without cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := uuid.NewString()

			mock.EXPECT().FailTask(taskID)
			mock.EXPECT().TaskState(taskID).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "not-cached",
				Payload:   "payload",
			}, nil)

			err := client.FailTask(taskID)
			assert.NoError(t, err)
		})
	})
}

func TestTaskGetters(t *testing.T) {
	WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
		// test multiple getters at the same time
		taskState, mockTaskState := client.TaskState, mock.EXPECT().TaskState
		next, mockNext := client.NextInQueue, mock.EXPECT().NextInQueue
		consume, mockConsume := client.Consume, mock.EXPECT().Consume
		getters := map[*func(string) (*maestro.Task, error)]func(interface{}) *gomock.Call{
			&taskState: mockTaskState,
			&next:      mockNext,
			&consume:   mockConsume,
		}

		for getter, mockGetter := range getters {
			t.Run("with cache", func(t *testing.T) {
				taskID := makeCompleteCachedTask(t, client, mock, cache)

				mockGetter(gomock.Any()).Return(&maestro.Task{
					TaskID:    taskID,
					TaskQueue: "cached",
					Payload:   cache.KeyForValue("payload"),
					Result:    cache.KeyForValue("result"),
				}, nil)

				task, err := (*getter)(taskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    taskID,
					TaskQueue: "cached",
					Payload:   "payload",
					Result:    "result",
				}, task)
			})

			t.Run("with cache no result", func(t *testing.T) {
				taskID := makeCompleteCachedTask(t, client, mock, cache)

				mockGetter(gomock.Any()).Return(&maestro.Task{
					TaskID:    taskID,
					TaskQueue: "cached",
					Payload:   cache.KeyForValue("payload"),
				}, nil)

				task, err := (*getter)(taskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    taskID,
					TaskQueue: "cached",
					Payload:   "payload",
					Result:    "",
				}, task)
			})

			t.Run("without cache", func(t *testing.T) {
				taskID := uuid.NewString()

				mockGetter(gomock.Any()).Return(&maestro.Task{
					TaskID:    taskID,
					TaskQueue: "not-cached",
					Payload:   "payload",
					Result:    "result",
				}, nil)

				task, err := (*getter)(taskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    taskID,
					TaskQueue: "not-cached",
					Payload:   "payload",
					Result:    "result",
				}, task)
			})
		}

	}, "cached")
}

func TestDeleteTask(t *testing.T) {
	t.Run("with cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := makeCompleteCachedTask(t, client, mock, cache)

			mock.EXPECT().TaskState(gomock.Any()).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "cached",
				Payload:   cache.KeyForValue("payload"),
				Result:    cache.KeyForValue("result"),
			}, nil)
			mock.EXPECT().DeleteTask(taskID)

			err := client.DeleteTask(taskID)
			assert.NoError(t, err)
			assert.Empty(t, cache.KeyForValue("payload"))
			assert.Empty(t, cache.KeyForValue("result"))
		}, "cached")
	})

	t.Run("with cache no result", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := makeCompleteCachedTask(t, client, mock, cache)

			mock.EXPECT().TaskState(gomock.Any()).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "cached",
				Payload:   cache.KeyForValue("payload"),
			}, nil)
			mock.EXPECT().DeleteTask(taskID)

			err := client.DeleteTask(taskID)
			assert.NoError(t, err)
			assert.Empty(t, cache.KeyForValue("payload"))
		}, "cached")
	})

	t.Run("without cache", func(t *testing.T) {
		WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
			taskID := uuid.NewString()

			mock.EXPECT().TaskState(gomock.Any()).Return(&maestro.Task{
				TaskID:    taskID,
				TaskQueue: "not-cached",
				Payload:   "payload",
				Result:    "result",
			}, nil)
			mock.EXPECT().DeleteTask(taskID)

			err := client.DeleteTask(taskID)
			assert.NoError(t, err)
		})
	})
}

func TestQueueStats(t *testing.T) {
	WithTestCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro, cache *testCache) {
		queue := uuid.NewString()
		expected := map[string][]string{
			"pending": {"task-1", "task-2"},
		}

		mock.EXPECT().QueueStats(queue).Return(expected, nil)

		actual, err := client.QueueStats(queue)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
