package future

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jizhuozhi/go-future/executors"
)

var errPoolFull = errors.New("pool is full")

// withExecutor installs e for the duration of one test.
func withExecutor(t *testing.T, e Executor) {
	t.Helper()
	prev := currentExecutor()
	SetExecutor(e)
	t.Cleanup(func() { SetExecutor(prev) })
}

// refusing refuses every task.
func refusing() Executor {
	return executors.ExecutorFunc(func(func()) error { return errPoolFull })
}

// A refused task must fail the Future instead of leaving it unresolved: a Future
// that is never resolved blocks Get forever, which is a worse outcome than an
// error the caller can act on.
func TestExecutorRejectionFailsTheFuture(t *testing.T) {
	withExecutor(t, refusing())

	ran := false
	_, err := Async(func() (int, error) { ran = true; return 1, nil }).Get()

	assert.ErrorIs(t, err, ErrExecutorRejected)
	assert.Contains(t, err.Error(), errPoolFull.Error(), "the executor's own error survives")
	assert.False(t, ran, "a refused task must not run")
}

// The same holds for the context-aware entry points.
func TestExecutorRejectionFailsCtxAsync(t *testing.T) {
	withExecutor(t, refusing())

	_, err := CtxAsync(context.Background(), func(context.Context) (int, error) { return 1, nil }).Get()
	assert.ErrorIs(t, err, ErrExecutorRejected)
}

// An executor that accepts the task reports success, whether it runs the task on
// another goroutine or inline.
func TestExecutorInlineAcceptanceIsNotAnError(t *testing.T) {
	withExecutor(t, executors.ExecutorFunc(func(f func()) error { f(); return nil }))

	v, err := Async(func() (int, error) { return 42, nil }).Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, v)
}
