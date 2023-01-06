package maestro

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/owlint/maestro-client/pkg/cache"
)

type CachedClient struct {
	maestro          Maestro
	cachedQueues     map[string]struct{}
	completedTaskTTL time.Duration
	cache            cache.Cache
}

func NewCachedClient(maestro Maestro, cache cache.Cache, cachedQueues []string, completedTaskTTL time.Duration) *CachedClient {
	queuesMap := map[string]struct{}{}
	for _, queue := range cachedQueues {
		queuesMap[queue] = struct{}{}
	}

	return &CachedClient{
		maestro:          maestro,
		cache:            cache,
		cachedQueues:     queuesMap,
		completedTaskTTL: completedTaskTTL,
	}
}

func (m *CachedClient) CreateTask(owner, queue, payload string, options ...createTaskOptions) (string, error) {
	opt := MergeCreateTaskOptions(options...)
	if _, exists := m.cachedQueues[queue]; exists && opt.StartTimeout() <= 0 {
		return "", errors.New("start timeout must be > 0 for cached task")
	}

	ttl := opt.ExecutesIn() + m.completedTaskTTL + (opt.StartTimeout()+opt.Timeout())*time.Duration(opt.Retries()+1)

	payload, err := m.cachePayload(context.TODO(), queue, payload, ttl)
	if err != nil {
		return "", err
	}

	return m.maestro.CreateTask(owner, queue, payload, *opt)
}

func (m *CachedClient) TaskState(taskID string) (*Task, error) {
	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return task, err
	}

	return m.taskFromCache(context.TODO(), task)
}

func (m *CachedClient) DeleteTask(taskID string) error {
	ctx := context.TODO()

	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	err = m.delete(ctx, task.TaskQueue, task.Payload)
	if err != nil {
		return err
	}

	if task.Result != "" {
		err = m.delete(ctx, task.TaskQueue, task.Result)
		if err != nil {
			return err
		}
	}

	return m.maestro.DeleteTask(taskID)
}

func (m *CachedClient) FailTask(taskID string) error {
	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	err = m.setTTL(context.TODO(), task.TaskQueue, task.Payload, m.completedTaskTTL)
	if err != nil {
		return err
	}

	return m.maestro.FailTask(taskID)
}

func (m *CachedClient) NextInQueue(queueName string) (*Task, error) {
	task, err := m.maestro.NextInQueue(queueName)
	if err != nil {
		return task, err
	}

	return m.taskFromCache(context.TODO(), task)
}

func (m *CachedClient) CompleteTask(taskID, result string) error {
	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	err = m.setTTL(ctx, task.TaskQueue, task.Payload, m.completedTaskTTL)
	if err != nil {
		return err
	}

	result, err = m.cachePayload(ctx, task.TaskQueue, result, m.completedTaskTTL)
	if err != nil {
		return err
	}

	return m.maestro.CompleteTask(taskID, result)
}

func (m *CachedClient) Consume(queue string) (*Task, error) {
	task, err := m.maestro.Consume(queue)
	if err != nil {
		return nil, err
	}

	return m.taskFromCache(context.TODO(), task)
}

func (m *CachedClient) QueueStats(queue string) (map[string][]string, error) {
	return m.maestro.QueueStats(queue)
}

func (m *CachedClient) delete(ctx context.Context, queue, key string) error {
	if _, exists := m.cachedQueues[queue]; !exists {
		return nil
	}

	return m.cache.Delete(ctx, key)
}

func (m *CachedClient) setTTL(ctx context.Context, queue, key string, ttl time.Duration) error {
	if _, exists := m.cachedQueues[queue]; !exists {
		return nil
	}

	return m.cache.SetTTL(ctx, key, ttl)
}

func (m *CachedClient) cachePayload(ctx context.Context, queue string, payload string, timeout time.Duration) (string, error) {
	if _, exists := m.cachedQueues[queue]; !exists {
		return payload, nil
	}

	cacheKey := m.uniqueKey(queue)
	err := m.cache.Put(ctx, cacheKey, payload, timeout)
	if err != nil {
		return "", err
	}
	return cacheKey, err
}

func (m *CachedClient) uniqueKey(queueName string) string {
	return "maestro-cache-" + queueName + "-" + uuid.NewString()
}

func (m *CachedClient) taskFromCache(ctx context.Context, task *Task) (*Task, error) {
	if task == nil {
		return nil, nil
	}

	payload, err := m.payloadFromCache(ctx, task.TaskQueue, task.Payload)
	if err != nil {
		return nil, err
	}

	task.Payload = payload

	if task.Result != "" {
		result, err := m.payloadFromCache(ctx, task.TaskQueue, task.Result)
		if err != nil {
			return nil, err
		}

		task.Result = result
	}

	return task, nil
}

func (m *CachedClient) payloadFromCache(ctx context.Context, queue, payload string) (string, error) {
	if _, exists := m.cachedQueues[queue]; !exists {
		return payload, nil
	}

	return m.cache.Get(ctx, payload)
}
