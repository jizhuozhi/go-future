package executors

type GoExecutor struct{}

func (GoExecutor) Submit(f func()) {
	go f()
}

type ExecutorFunc func(func())

func (e ExecutorFunc) Submit(f func()) {
	e(f)
}
