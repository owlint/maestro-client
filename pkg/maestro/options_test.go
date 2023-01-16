package maestro_test

import (
	"testing"
	"time"

	"github.com/owlint/maestro-client/pkg/maestro"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateTaskOptions(t *testing.T) {
	opt := maestro.NewCreateTaskOptions()

	assert.Equal(t, 0, opt.Retries())
	assert.Equal(t, 900*time.Second, opt.Timeout())
	assert.Empty(t, opt.StartTimeout())
	assert.Empty(t, opt.ExecutesIn())
}

func TestWithRetries(t *testing.T) {
	opt := maestro.NewCreateTaskOptions().WithRetries(10)

	assert.Equal(t, 10, opt.Retries())
}

func TestWithTimeout(t *testing.T) {
	opt := maestro.NewCreateTaskOptions().WithTimeout(time.Minute)

	assert.Equal(t, time.Minute, opt.Timeout())
}

func TestWithStartTimeout(t *testing.T) {
	opt := maestro.NewCreateTaskOptions().WithStartTimeout(time.Minute)

	assert.Equal(t, time.Minute, opt.StartTimeout())
}

func TestWithExecutesIn(t *testing.T) {
	opt := maestro.NewCreateTaskOptions().WithExecutesIn(time.Minute)

	assert.Equal(t, time.Minute, opt.ExecutesIn())
}

func TestMergeCreateOptions(t *testing.T) {
	defaultOpt := maestro.NewCreateTaskOptions()

	setOpt := maestro.NewCreateTaskOptions().
		WithRetries(2).WithTimeout(time.Minute).
		WithStartTimeout(2 * time.Minute).
		WithExecutesIn(3 * time.Minute)

	t.Run("empty", func(t *testing.T) {
		res := maestro.MergeCreateTaskOptions()
		assert.Equal(t, res, defaultOpt)
	})

	t.Run("single", func(t *testing.T) {
		res := maestro.MergeCreateTaskOptions(*setOpt)
		assert.Equal(t, *res, *setOpt)
	})

	t.Run("default with set", func(t *testing.T) {
		res := maestro.MergeCreateTaskOptions(*defaultOpt, *setOpt)
		assert.Equal(t, *res, *setOpt)
	})

	t.Run("set with default", func(t *testing.T) {
		res := maestro.MergeCreateTaskOptions(*setOpt, *defaultOpt)
		assert.Equal(t, res, defaultOpt)
	})
}
