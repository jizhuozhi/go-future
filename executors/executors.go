// Package executors provides the Executor implementations go-future ships with:
// one that runs each task on its own goroutine, and an adapter that turns a
// plain function into an Executor.
//
// future.Async uses GoExecutor by default. Replace it with future.SetExecutor,
// for example to bound concurrency or to reuse goroutines.
package executors

// GoExecutor runs every task on a new goroutine.
//
// It is the default executor, and it has no concurrency limit: a submitted task
// starts immediately. That is the right default for short tasks and the wrong
// one for tasks that block, which would otherwise pile up unbounded.
type GoExecutor struct{}

// Submit starts f on a new goroutine and returns without waiting for it.
func (GoExecutor) Submit(f func()) {
	go f()
}

// ExecutorFunc adapts a plain function to the Executor interface. It is the
// shortest way to plug in a pool or any other scheduling policy:
//
//	future.SetExecutor(executors.ExecutorFunc(func(f func()) {
//		pool.Submit(f)
//	}))
type ExecutorFunc func(func())

// Submit calls e with f.
func (e ExecutorFunc) Submit(f func()) {
	e(f)
}
