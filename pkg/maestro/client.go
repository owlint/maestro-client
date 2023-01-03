package maestro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Task struct {
	TaskID     string `json:"task_id"`
	Owner      string `json:"owner"`
	TaskQueue  string `json:"task_queue"`
	Payload    string `json:"payload"`
	State      string `json:"state"`
	Timeout    int32  `json:"timeout"`
	Retries    int32  `json:"retries"`
	MaxRetries int32  `json:"max_retries"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	Result     string `json:"result,omitempty"`
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

type Maestro interface {
	CreateTask(owner, queue, payload string) (string, error)
	CreateScheduledTask(owner, queue, payload string, executesAfter time.Duration) (string, error)

	TaskState(taskID string) (*Task, error)
	DeleteTask(taskID string) error
	FailTask(taskID string) error
	NextInQueue(queueName string) (*Task, error)
	CompleteTask(taskID, result string) error
	Consume(queue string) (*Task, error)
	QueueStats(queue string) (map[string][]string, error)
}

type Client struct {
	endpoint string
	client   *http.Client
}

func (m Client) CreateTask(owner, queue, payload string, options ...CreateTaskOptions) (string, error) {
	httpPayload := struct {
		Owner   string `json:"owner"`
		Queue   string `json:"queue"`
		Retries int    `json:"retries"`
		Timeout int    `json:"timeout"`
		Payload string `json:"payload"`
	}{
		owner,
		queue,
		0,
		900,
		payload,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/create", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return "", err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := struct {
		Error  string `json:"error,omitempty"`
		TaskID string `json:"task_id,omitempty"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	if respBody.Error != "" {
		return "", fmt.Errorf("Maestro error %s", respBody.Error)
	}

	return respBody.TaskID, nil
}

func (m Client) CreateScheduledTask(owner, queue, payload string, executesAfter time.Duration) (string, error) {
	httpPayload := struct {
		Owner     string `json:"owner"`
		Queue     string `json:"queue"`
		Retries   int    `json:"retries"`
		Timeout   int    `json:"timeout"`
		Payload   string `json:"payload"`
		NotBefore int64  `json:"not_before"`
	}{
		owner,
		queue,
		0,
		180,
		payload,
		time.Now().Add(executesAfter).Unix(),
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/create", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return "", err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := struct {
		Error  string `json:"error,omitempty"`
		TaskID string `json:"task_id,omitempty"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	if respBody.Error != "" {
		return "", fmt.Errorf("Maestro error %s", respBody.Error)
	}

	return respBody.TaskID, nil
}

func (m Client) TaskState(taskID string) (*Task, error) {
	httpPayload := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: taskID,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/get", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := struct {
		Task Task
	}{Task: Task{}}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}

	return &respBody.Task, nil
}

func (m Client) DeleteTask(taskID string) error {
	httpPayload := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: taskID,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/delete", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	return nil
}

func (m Client) FailTask(taskID string) error {
	httpPayload := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: taskID,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/fail", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	return nil
}

func (m Client) NextInQueue(queueName string) (*Task, error) {
	httpPayload := struct {
		QueueName string `json:"queue"`
	}{
		QueueName: queueName,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queue/next", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := struct {
		Task Task
	}{Task: Task{}}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}

	return &respBody.Task, nil
}

func (m Client) CompleteTask(taskID, result string) error {
	httpPayload := struct {
		TaskID string `json:"task_id"`
		Result string `json:"result"`
	}{
		TaskID: taskID,
		Result: result,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/task/complete", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	return nil
}

func (m Client) Consume(queue string) (*Task, error) {
	httpPayload := struct {
		Queue string `json:"queue"`
	}{
		Queue: queue,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queue/results/consume", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := struct {
		Task Task
	}{Task: Task{}}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}

	return &respBody.Task, nil
}

func (m Client) QueueStats(queue string) (map[string][]string, error) {
	httpPayload := struct {
		Queue string `json:"queue"`
	}{
		Queue: queue,
	}

	bytePayload, err := json.Marshal(&httpPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/queue/stats", m.endpoint), bytes.NewReader(bytePayload))
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Maestro responded with invalid status code %d : %s", resp.StatusCode, string(msg))
	}

	respBody := make(map[string][]string)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
