package future

func Async[T any](f func() (T, error)) *Future[T] {
	p := NewPromise[T]()
	go func() {
		val, err := f()
		p.Set(val, err)
	}()
	return p.Future()
}

func Lazy[T any](f func() (T, error)) *Future[T] {
	p := NewPromise[T]()
	p.state.state |= flagLazy
	p.state.f = f
	return p.Future()
}
