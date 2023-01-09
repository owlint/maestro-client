package maestro

import "time"

type CreateTaskOptions struct {
	retries      int
	timeout      time.Duration
	executesIn   time.Duration
	startTimeout time.Duration
}

func NewCreateTaskOptions() *CreateTaskOptions {
	return &CreateTaskOptions{
		retries:      0,
		timeout:      900 * time.Second,
		executesIn:   0,
		startTimeout: 0,
	}
}

func (c *CreateTaskOptions) Retries() int {
	return c.retries
}

func (c *CreateTaskOptions) Timeout() time.Duration {
	return c.timeout
}

func (c *CreateTaskOptions) ExecutesIn() time.Duration {
	return c.executesIn
}

func (c *CreateTaskOptions) StartTimeout() time.Duration {
	return c.startTimeout
}

func (c *CreateTaskOptions) WithRetries(retries int) *CreateTaskOptions {
	c.retries = retries
	return c
}

func (c *CreateTaskOptions) WithTimeout(timeout time.Duration) *CreateTaskOptions {
	c.timeout = timeout
	return c
}

func (c *CreateTaskOptions) WithExecutesIn(executesIn time.Duration) *CreateTaskOptions {
	c.executesIn = executesIn
	return c
}

func (c *CreateTaskOptions) WithStartTimeout(startTimeout time.Duration) *CreateTaskOptions {
	c.startTimeout = startTimeout
	return c
}

// Merge always merge the last given option or return the default option if none. There is no partial merge.
func MergeCreateTaskOptions(options ...CreateTaskOptions) *CreateTaskOptions {
	if len(options) == 0 {
		return NewCreateTaskOptions()
	}

	lastOption := options[len(options)-1]

	return NewCreateTaskOptions().WithRetries(lastOption.retries).
		WithTimeout(lastOption.timeout).
		WithStartTimeout(lastOption.startTimeout).
		WithExecutesIn(lastOption.executesIn)
}
