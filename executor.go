package future

import "github.com/jizhuozhi/go-future/executors"

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

var executor Executor = executors.GoExecutor{}

func SetExecutor(e Executor) {
	if e == nil {
		panic("executor is nil")
	}
	executor = e
}
