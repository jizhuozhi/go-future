package future

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/jizhuozhi/go-future/executors"
)

// ErrExecutorRejected reports that an executor refused a task.
//
// The Future returned by Submit, Async, CtxSubmit or CtxAsync fails with this
// error, and the task never runs.
var ErrExecutorRejected = errors.New("future: executor rejected the task")

// rejected wraps the error an executor returned so callers can test for
// ErrExecutorRejected with errors.Is.
func rejected(err error) error { return fmt.Errorf("%w: %v", ErrExecutorRejected, err) }

// Executor defines an abstraction for executing asynchronous tasks in go-future.
//
// By default, go-future uses the standard Go goroutines (executors.GoExecutor{}) to execute tasks.
// This provides lightweight asynchronous execution without pooling or concurrency limits.
//
// You can override the default executor using any implementation of the Executor interface with SetExecutor.
// A common pattern is to use executors.ExecutorFunc to wrap a goroutine pool, for example:
//
//	pool, _ := ants.NewPool(100, ants.WithNonblocking(true))
//	SetExecutor(executors.ExecutorFunc(func(f func()) error {
//	    return pool.Submit(f)
//	}))
//
// Submit returns nil when the task was accepted — on another goroutine or
// inline in the caller, both are valid — and a non-nil error when it was
// refused, in which case the task will not run. Reporting the refusal instead
// of dropping the task silently is what keeps a bounded pool from turning a
// rejection into a hang: the Future fails with ErrExecutorRejected rather than
// waiting for a result that will never arrive.
//
// Most cases do NOT require changing the executor. Replacing the default executor can be useful
// to limit concurrency, reuse goroutines, or reduce GC pressure.
//
// Caution:
//   - For RPC tasks or other potentially blocking operations, using a pooled executor may
//     cause task queuing and negative performance impact. Only override the executor if you
//     understand the workload and have performed thorough performance testing.
//   - Passing nil to SetExecutor will panic.
type Executor interface {
	Submit(func()) error
}

// executorBox keeps the dynamic type held by the atomic.Value constant. Storing
// two Executor implementations of different dynamic types straight into an
// atomic.Value panics with "store of inconsistently typed value", so the value
// is boxed and the box pointer is what gets stored.
type executorBox struct{ e Executor }

// executor holds an *executorBox. Routing through the box lets SetExecutor swap
// the executor while Async and the generic methods read it, without a data race
// on the variable itself.
var executor atomic.Value

func init() {
	SetExecutor(executors.GoExecutor{})
}

// currentExecutor returns the executor currently in use.
func currentExecutor() Executor { return executor.Load().(*executorBox).e }

// SetExecutor replaces the executor that Async, CtxAsync and the *Go generic
// methods submit to.
//
// It is safe to call concurrently with those functions: the swap is atomic, so
// a caller observes either the previous executor or the new one and never a torn
// value. Passing nil panics.
func SetExecutor(e Executor) {
	if e == nil {
		panic("executor is nil")
	}
	executor.Store(&executorBox{e: e})
}
