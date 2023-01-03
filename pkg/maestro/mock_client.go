package maestro

import (
	"time"

	"github.com/google/uuid"
)

type MaestroMock struct {
	NbCreated int
}

func (m *MaestroMock) CreateTask(owner, queue, payload string) (string, error) {
	m.NbCreated += 1
	return uuid.NewString(), nil
}
func (m *MaestroMock) CreateScheduledTask(owner, queue, payload string, executesAfter time.Duration) (string, error) {
	m.NbCreated += 1
	return uuid.NewString(), nil
}
func (m *MaestroMock) TaskState(taskID string) (*Task, error) {
	return nil, nil
}
func (m *MaestroMock) DeleteTask(taskID string) error {
	return nil
}
func (m *MaestroMock) FailTask(taskID string) error {
	return nil
}
func (m *MaestroMock) NextInQueue(queueName string) (*Task, error) {
	return nil, nil
}
func (m *MaestroMock) CompleteTask(taskID string, result string) error {
	return nil
}
func (m *MaestroMock) Consume(queueName string) (*Task, error) {
	return nil, nil
}
func (m *MaestroMock) QueueStats(queue string) (map[string][]string, error) {
	return nil, nil
}
