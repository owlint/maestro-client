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
	store            payloadStore
	completedTaskTTL time.Duration
}

func NewCachedClient(maestro Maestro, cache cache.Cache, cachedQueues []string, completedTaskTTL time.Duration) *CachedClient {
	queueNameSet := map[string]struct{}{}
	for _, queue := range cachedQueues {
		queueNameSet[queue] = struct{}{}
	}

	return &CachedClient{
		maestro: maestro,
		store: payloadStore{
			cache:        cache,
			cachedQueues: queueNameSet,
		},
		completedTaskTTL: completedTaskTTL,
	}
}

func (m *CachedClient) CreateTask(owner, queue, payload string, options ...CreateTaskOptions) (string, error) {
	opt := MergeCreateTaskOptions(options...)
	if m.store.IsCached(queue) && opt.StartTimeout() <= 0 {
		return "", errors.New("start timeout must be > 0 for cached task")
	}

	ttl := opt.ExecutesIn() + m.completedTaskTTL + (opt.StartTimeout()+opt.Timeout())*time.Duration(opt.Retries()+1)

	payload, err := m.store.Persist(context.TODO(), queue, payload, ttl)
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

	return m.store.Reload(context.TODO(), task)
}

func (m *CachedClient) DeleteTask(taskID string) error {
	ctx := context.TODO()

	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	err = m.store.Delete(ctx, task)
	if err != nil {
		return err
	}

	return m.maestro.DeleteTask(taskID)
}

func (m *CachedClient) FailTask(taskID string) error {
	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	err = m.store.SetTTL(context.TODO(), task.TaskQueue, task.Payload, m.completedTaskTTL)
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

	return m.store.Reload(context.TODO(), task)
}

func (m *CachedClient) CompleteTask(taskID, result string) error {
	task, err := m.maestro.TaskState(taskID)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	err = m.store.SetTTL(ctx, task.TaskQueue, task.Payload, m.completedTaskTTL)
	if err != nil {
		return err
	}

	result, err = m.store.Persist(ctx, task.TaskQueue, result, m.completedTaskTTL)
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

	return m.store.Reload(context.TODO(), task)
}

func (m *CachedClient) QueueStats(queue string) (map[string][]string, error) {
	return m.maestro.QueueStats(queue)
}

type payloadStore struct {
	cache        cache.Cache
	cachedQueues map[string]struct{}
}

func (s *payloadStore) IsCached(queue string) bool {
	_, exists := s.cachedQueues[queue]
	return exists
}

func (s *payloadStore) Delete(ctx context.Context, task *Task) error {
	if !s.IsCached(task.TaskQueue) {
		return nil
	}

	err := s.cache.Delete(ctx, task.Payload)
	if err != nil {
		return err
	}

	if task.Result != "" {
		err = s.cache.Delete(ctx, task.Result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *payloadStore) SetTTL(ctx context.Context, queue, key string, ttl time.Duration) error {
	if !s.IsCached(queue) {
		return nil
	}

	return s.cache.SetTTL(ctx, key, ttl)
}

func (s *payloadStore) Persist(ctx context.Context, queue string, payload string, timeout time.Duration) (string, error) {
	if !s.IsCached(queue) {
		return payload, nil
	}

	cacheKey := s.uniqueKey(queue)
	err := s.cache.Put(ctx, cacheKey, payload, timeout)
	if err != nil {
		return "", err
	}
	return cacheKey, err
}

func (s *payloadStore) Reload(ctx context.Context, task *Task) (*Task, error) {
	if task == nil {
		return nil, nil
	}

	payload, err := s.payloadFromCache(ctx, task.TaskQueue, task.Payload)
	if err != nil {
		return nil, err
	}

	task.Payload = payload

	if task.Result != "" {
		result, err := s.payloadFromCache(ctx, task.TaskQueue, task.Result)
		if err != nil {
			return nil, err
		}

		task.Result = result
	}

	return task, nil
}

func (s *payloadStore) uniqueKey(queueName string) string {
	return "maestro-cache-" + queueName + "-" + uuid.NewString()
}

func (s *payloadStore) payloadFromCache(ctx context.Context, queue, payload string) (string, error) {
	if !s.IsCached(queue) {
		return payload, nil
	}

	return s.cache.Get(ctx, payload)
}
