package maestro_test

import (
	"context"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v6"
	"github.com/go-redis/redis/v9"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/owlint/maestro-client/pkg/cache"
	"github.com/owlint/maestro-client/pkg/maestro"
	"github.com/owlint/maestro-client/test/testutils"
	"github.com/stretchr/testify/assert"
)

func bigPayload() string {
	return fake.Sentence(100)
}

func withRedisCachedClient(t *testing.T, fn func(client *maestro.CachedClient, mock *maestro.MockMaestro), cachedQueues ...string) {
	testutils.WithTestRedis(func(r *redis.Client) {
		cache := cache.NewRedisCache(r)
		maestroMock := maestro.NewMockMaestro(gomock.NewController(t))
		cachedClient := maestro.NewCachedClient(maestroMock, cache, cachedQueues, 900*time.Second)

		fn(cachedClient, maestroMock)
	})
}

func makePendingTask(t *testing.T, queue, payload string, client maestro.Maestro, mock *maestro.MockMaestro) maestro.Task {
	taskID := uuid.NewString()
	task := maestro.Task{
		TaskID:    taskID,
		TaskQueue: queue,
	}

	mock.EXPECT().CreateTask("owner", queue, gomock.Any(), gomock.Any()).Do(func(owner, queue, payload string, opts ...maestro.CreateTaskOptions) {
		task.Payload = payload
	}).Return(taskID, nil)

	_, err := client.CreateTask("owner", queue, payload, *maestro.NewCreateTaskOptions().WithStartTimeout(10))
	assert.NoError(t, err)

	return task
}

func makeCompleteTask(t *testing.T, queue, payload, result string, client maestro.Maestro, mock *maestro.MockMaestro) maestro.Task {
	task := makePendingTask(t, queue, payload, client, mock)

	mock.EXPECT().CompleteTask(task.TaskID, gomock.Any()).
		Do(func(taskID, result string) {
			task.Result = result
		}).Return(nil)

	mock.EXPECT().TaskState(gomock.Any()).Return(&task, nil)

	err := client.CompleteTask(task.TaskID, result)
	assert.NoError(t, err)

	return task
}

func TestCreateTask(t *testing.T) {
	t.Run("queue not in cache must send payload directly", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			mock.EXPECT().CreateTask("owner", "not-cached", "payload", *maestro.NewCreateTaskOptions()).Return("task-id", nil)

			res, err := client.CreateTask("owner", "not-cached", "payload")
			assert.NoError(t, err)
			assert.Equal(t, "task-id", res)
		})
	})

	t.Run("queue in cache must NOT send payload directly", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			mock.EXPECT().CreateTask("owner", "cached", gomock.Not("payload"), gomock.Any()).Return("task-id", nil)

			res, err := client.CreateTask("owner", "cached", "payload", *maestro.NewCreateTaskOptions().WithStartTimeout(time.Hour))
			assert.NoError(t, err)
			assert.Equal(t, "task-id", res)
		}, "cached")
	})

	t.Run("queue in cache no StartTimeout", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			_, err := client.CreateTask("owner", "cached", "payload")
			assert.Error(t, err)
		}, "cached")
	})
}

func TestCompleteTask(t *testing.T) {
	t.Run("cached queue", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			payload, result := bigPayload(), bigPayload()
			completeTask := makeCompleteTask(t, "cached", payload, result, client, mock)
			// task payloads must not be sent directly
			assert.Less(t, len(completeTask.Payload), len(payload))
			assert.Less(t, len(completeTask.Result), len(result))

			mock.EXPECT().TaskState(gomock.Any()).Return(&completeTask, nil)
			task, err := client.TaskState(completeTask.TaskID)
			assert.NoError(t, err)
			assert.Equal(t, payload, task.Payload)
			assert.Equal(t, result, task.Result)
		}, "cached")
	})

	t.Run("queue not cached", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			payload, result := bigPayload(), bigPayload()
			completeTask := makeCompleteTask(t, "not-cached", payload, result, client, mock)
			// task payloads must be sent directly
			assert.Equal(t, payload, completeTask.Payload)
			assert.Equal(t, result, completeTask.Result)

			mock.EXPECT().TaskState(gomock.Any()).Return(&completeTask, nil)
			task, err := client.TaskState(completeTask.TaskID)
			assert.NoError(t, err)
			assert.Equal(t, payload, task.Payload)
			assert.Equal(t, result, task.Result)
		})
	})
}

func TestFailTask(t *testing.T) {
	t.Run("cached queue", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			task := makeCompleteTask(t, "cached", bigPayload(), bigPayload(), client, mock)
			mock.EXPECT().TaskState(task.TaskID).Return(&task, nil)
			mock.EXPECT().FailTask(task.TaskID).Return(nil)

			err := client.FailTask(task.TaskID)
			assert.NoError(t, err)
		}, "cached")
	})

	t.Run("queue not cached", func(t *testing.T) {
		withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
			task := makeCompleteTask(t, "cached", fake.BeerName(), fake.BeerName(), client, mock)
			mock.EXPECT().TaskState(task.TaskID).Return(&task, nil)
			mock.EXPECT().FailTask(task.TaskID).Return(nil)

			err := client.FailTask(task.TaskID)
			assert.NoError(t, err)
		})
	})
}

func TestTaskGetters(t *testing.T) {
	withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
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
			t.Run("cached queue", func(t *testing.T) {
				payload, result := bigPayload(), bigPayload()
				completeTask := makeCompleteTask(t, "cached", payload, result, client, mock)
				// task payloads must not be sent directly
				assert.NotEqual(t, payload, completeTask.Payload)
				assert.NotEqual(t, result, completeTask.Result)

				mockGetter(completeTask.TaskID).Return(&completeTask, nil)

				task, err := (*getter)(completeTask.TaskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    completeTask.TaskID,
					TaskQueue: completeTask.TaskQueue,
					Payload:   payload,
					Result:    result,
				}, task)
			})

			t.Run("cached queue no result", func(t *testing.T) {
				payload := bigPayload()
				pendingTask := makePendingTask(t, "cached", payload, client, mock)
				// task payloads must not be sent directly
				assert.NotEqual(t, payload, pendingTask.Payload)

				mockGetter(pendingTask.TaskID).Return(&pendingTask, nil)

				task, err := (*getter)(pendingTask.TaskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    pendingTask.TaskID,
					TaskQueue: pendingTask.TaskQueue,
					Payload:   payload,
					Result:    "",
				}, task)
			})

			t.Run("queue not cached", func(t *testing.T) {
				payload, result := bigPayload(), bigPayload()
				completeTask := makeCompleteTask(t, "not-cached", payload, result, client, mock)
				// task payloads must be sent directly
				assert.Equal(t, payload, completeTask.Payload)
				assert.Equal(t, result, completeTask.Result)

				mockGetter(completeTask.TaskID).Return(&completeTask, nil)

				task, err := (*getter)(completeTask.TaskID)
				assert.NoError(t, err)
				assert.Equal(t, &maestro.Task{
					TaskID:    completeTask.TaskID,
					TaskQueue: completeTask.TaskQueue,
					Payload:   payload,
					Result:    result,
				}, task)
			})
		}

	}, "cached")
}

func TestDeleteTask(t *testing.T) {
	t.Run("cached queue must delete keys", func(t *testing.T) {
		testutils.WithTestRedis(func(r *redis.Client) {
			cache := cache.NewRedisCache(r)
			mock := maestro.NewMockMaestro(gomock.NewController(t))
			client := maestro.NewCachedClient(mock, cache, []string{"cached"}, 900*time.Second)

			payload, result := bigPayload(), bigPayload()
			completeTask := makeCompleteTask(t, "cached", payload, result, client, mock)

			cacheSize, err := r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Positive(t, cacheSize)

			mock.EXPECT().TaskState(completeTask.TaskID).Return(&completeTask, nil)
			mock.EXPECT().DeleteTask(completeTask.TaskID)

			err = client.DeleteTask(completeTask.TaskID)
			assert.NoError(t, err)

			cacheSize, err = r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Zero(t, cacheSize)
		})
	})

	t.Run("cached queue no result", func(t *testing.T) {
		testutils.WithTestRedis(func(r *redis.Client) {
			cache := cache.NewRedisCache(r)
			mock := maestro.NewMockMaestro(gomock.NewController(t))
			client := maestro.NewCachedClient(mock, cache, []string{"cached"}, 900*time.Second)

			payload := bigPayload()
			pendingTask := makePendingTask(t, "cached", payload, client, mock)

			cacheSize, err := r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Equal(t, int64(1), cacheSize)

			mock.EXPECT().TaskState(pendingTask.TaskID).Return(&maestro.Task{
				TaskID:    pendingTask.TaskID,
				TaskQueue: pendingTask.TaskQueue,
				Payload:   pendingTask.Payload,
				Result:    "",
			}, nil)
			mock.EXPECT().DeleteTask(pendingTask.TaskID)

			err = client.DeleteTask(pendingTask.TaskID)
			assert.NoError(t, err)

			cacheSize, err = r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Zero(t, cacheSize)
		})
	})

	t.Run("queue not cached", func(t *testing.T) {
		testutils.WithTestRedis(func(r *redis.Client) {
			cache := cache.NewRedisCache(r)
			mock := maestro.NewMockMaestro(gomock.NewController(t))
			client := maestro.NewCachedClient(mock, cache, []string{}, 900*time.Second)

			payload, result := bigPayload(), bigPayload()
			completeTask := makeCompleteTask(t, "not-cached", payload, result, client, mock)

			cacheSize, err := r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Zero(t, cacheSize)

			mock.EXPECT().TaskState(completeTask.TaskID).Return(&completeTask, nil)
			mock.EXPECT().DeleteTask(completeTask.TaskID)

			err = client.DeleteTask(completeTask.TaskID)
			assert.NoError(t, err)

			cacheSize, err = r.DBSize(context.Background()).Result()
			assert.NoError(t, err)
			assert.Zero(t, cacheSize)
		})
	})
}

func TestQueueStats(t *testing.T) {
	withRedisCachedClient(t, func(client *maestro.CachedClient, mock *maestro.MockMaestro) {
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
