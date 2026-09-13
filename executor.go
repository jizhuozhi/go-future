package future

import (
	"sync/atomic"

	"github.com/jizhuozhi/go-future/executors"
)

// Executor defines an abstraction for executing asynchronous tasks in go-future.
//
// By default, go-future uses the standard Go goroutines (executors.GoExecutor{}) to execute tasks.
// This provides lightweight asynchronous execution without pooling or concurrency limits.
//
// You can override the default executor using any implementation of the Executor interface with SetExecutor.
// A common pattern is to use executors.ExecutorFunc to wrap a goroutine pool, for example:
//
//	pool := ants.NewPool(100)
//	SetExecutor(executors.ExecutorFunc(func(f func()) {
//	    pool.Submit(f)
//	}))
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
	Submit(func())
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
