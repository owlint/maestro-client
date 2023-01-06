package maestro

import "time"

type createTaskOptions struct {
	retries      int
	timeout      time.Duration
	executesIn   time.Duration
	startTimeout time.Duration
}

func NewCreateTaskOptions() *createTaskOptions {
	return &createTaskOptions{
		retries:      0,
		timeout:      900 * time.Second,
		executesIn:   0,
		startTimeout: 0,
	}
}

func (c *createTaskOptions) Retries() int {
	return c.retries
}

func (c *createTaskOptions) Timeout() time.Duration {
	return c.timeout
}

func (c *createTaskOptions) ExecutesIn() time.Duration {
	return c.executesIn
}

func (c *createTaskOptions) StartTimeout() time.Duration {
	return c.startTimeout
}

func (c *createTaskOptions) WithRetries(retries int) *createTaskOptions {
	c.retries = retries
	return c
}

func (c *createTaskOptions) WithTimeout(timeout time.Duration) *createTaskOptions {
	c.timeout = timeout
	return c
}

func (c *createTaskOptions) WithExecutesIn(executesIn time.Duration) *createTaskOptions {
	c.executesIn = executesIn
	return c
}

func (c *createTaskOptions) WithStartTimeout(startTimeout time.Duration) *createTaskOptions {
	c.startTimeout = startTimeout
	return c
}

// Merge always merge the last given option or return the default option if none. There is no partial merge.
func MergeCreateTaskOptions(options ...createTaskOptions) *createTaskOptions {
	if len(options) == 0 {
		return NewCreateTaskOptions()
	}

	lastOption := options[len(options)-1]

	return NewCreateTaskOptions().WithRetries(lastOption.retries).
		WithTimeout(lastOption.timeout).
		WithStartTimeout(lastOption.startTimeout).
		WithExecutesIn(lastOption.executesIn)
}
