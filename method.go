package future

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// Then chains a synchronous computation onto f.
//
// cb runs on the goroutine that completes f, therefore it must not block,
// exactly like the callback registered through Subscribe.
//
//	f := Async(func() (int, error) { return 1, nil })
//	s := f.Then(func(v int, err error) (string, error) {
//	    if err != nil {
//	        return "", err
//	    }
//	    return strconv.Itoa(v), nil
//	})
func (f *Future[T]) Then[R any](cb func(val T, err error) (R, error)) *Future[R] {
	s := &state[R]{}
	f.state.subscribe(func(val T, err error) {
		rval, rerr := cb(val, err)
		s.set(rval, rerr)
	})
	return &Future[R]{state: s}
}

// ThenAsync chains an asynchronous computation onto f.
//
// cb returns a *Future[R]; the nested future is flattened so the caller only
// ever observes a single *Future[R].
func (f *Future[T]) ThenAsync[R any](cb func(val T, err error) *Future[R]) *Future[R] {
	s := &state[R]{}
	f.state.subscribe(func(val T, err error) {
		cb(val, err).state.subscribe(func(rval R, rerr error) {
			s.set(rval, rerr)
		})
	})
	return &Future[R]{state: s}
}

// ThenGo behaves like Then, except that cb is dispatched to the configured
// Executor instead of running inline on the goroutine that completes f.
//
// Use it when cb is expensive or may block: Then runs on the producer's
// goroutine and would delay that goroutine as well as every other callback
// registered on the same Future.
//
// A panic raised by cb is captured and turned into an ErrPanic error, exactly
// like Async does.
func (f *Future[T]) ThenGo[R any](cb func(val T, err error) (R, error)) *Future[R] {
	s := &state[R]{}
	f.state.subscribe(func(val T, err error) {
		executor.Submit(func() {
			var rval R
			var rerr error
			defer func() {
				if r := recover(); r != nil {
					rerr = fmt.Errorf("%w, err=%s, stack=%s", ErrPanic, r, debug.Stack())
				}
				s.set(rval, rerr)
			}()
			rval, rerr = cb(val, err)
		})
	})
	return &Future[R]{state: s}
}

// Map transforms the success value of f with fn.
//
// If f fails, fn is skipped and the error is propagated unchanged.
func (f *Future[T]) Map[R any](fn func(val T) R) *Future[R] {
	return f.Then(func(val T, err error) (R, error) {
		var zero R
		if err != nil {
			return zero, err
		}
		return fn(val), nil
	})
}

// FlatMap transforms the success value of f into another Future and flattens
// the result. If f fails, fn is skipped and the error is propagated unchanged.
func (f *Future[T]) FlatMap[R any](fn func(val T) *Future[R]) *Future[R] {
	return f.ThenAsync(func(val T, err error) *Future[R] {
		var zero R
		if err != nil {
			return Done2(zero, err)
		}
		return fn(val)
	})
}

// Cast reinterprets f as a *Future[R] by performing a dynamic type assertion on
// the resolved value (any(val).(R)).
//
// It is the typed bridge for values that crossed an `any` boundary, e.g. the
// per-node results produced by dagcore and dagfunc:
//
//	nodes := inst.Nodes()
//	count, err := nodes["func:TokenCount"].Future().Cast[TokenCount]().Get()
//
// An error carried by f is propagated as-is; if the value is not assignable to
// R the returned Future fails with ErrTypeMismatch.
func (f *Future[T]) Cast[R any]() *Future[R] {
	return f.Then(func(val T, err error) (R, error) {
		var zero R
		if err != nil {
			return zero, err
		}
		if r, ok := any(val).(R); ok {
			return r, nil
		}
		return zero, fmt.Errorf("%w: %T is not assignable to %v", ErrTypeMismatch, val, reflect.TypeFor[R]())
	})
}

// Recover turns a failed Future into a successful one by handing the error to
// fn. A successful Future is passed through untouched.
func (f *Future[T]) Recover(fn func(err error) (T, error)) *Future[T] {
	return f.Then(func(val T, err error) (T, error) {
		if err != nil {
			return fn(err)
		}
		return val, nil
	})
}

// OrElse replaces the error of a failed Future with defaultVal.
func (f *Future[T]) OrElse(defaultVal T) *Future[T] {
	return f.Then(func(val T, err error) (T, error) {
		if err != nil {
			return defaultVal, nil
		}
		return val, nil
	})
}

// ToAny widens f into a *Future[any].
func (f *Future[T]) ToAny() *Future[any] {
	return f.Then(func(val T, err error) (any, error) {
		return val, err
	})
}

// ToChan converts f into a single read-only channel of Result[T].
//
// When f completes a Result[T] holding both value and error is sent through the
// channel, which is then closed. The channel is buffered (size 1) so that a
// callback firing synchronously never blocks.
func (f *Future[T]) ToChan() <-chan Result[T] {
	ch := make(chan Result[T], 1)
	f.Subscribe(func(val T, err error) {
		ch <- Result[T]{Val: val, Err: err}
		close(ch)
	})
	return ch
}

// Timeout wraps f so that it fails with ErrTimeout when it is not resolved
// within d.
func (f *Future[T]) Timeout(d time.Duration) *Future[T] {
	var done uint32
	s := &state[T]{}
	timer := time.AfterFunc(d, func() {
		if atomic.CompareAndSwapUint32(&done, 0, 1) {
			var zero T
			s.set(zero, ErrTimeout)
		}
	})
	f.state.subscribe(func(val T, err error) {
		if atomic.CompareAndSwapUint32(&done, 0, 1) {
			s.set(val, err)
			timer.Stop()
		}
	})
	return &Future[T]{state: s}
}

// Until wraps f so that it fails with ErrTimeout when it is not resolved before
// t.
func (f *Future[T]) Until(t time.Time) *Future[T] {
	return f.Timeout(time.Until(t))
}
