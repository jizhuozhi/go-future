package future

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var ErrPanic = errors.New("async panic")
var ErrTimeout = errors.New("future timeout")

// ErrTypeMismatch is returned by Future.Cast when the resolved value cannot be
// asserted to the requested type.
var ErrTypeMismatch = errors.New("future type mismatch")

type Result[T any] struct {
	Val T
	Err error
}

type AnyResult[T any] struct {
	Index int
	Val   T
	Err   error
}

func Async[T any](f func() (T, error)) *Future[T] {
	return Submit(executor, f)
}

func CtxAsync[T any](ctx context.Context, f func(ctx context.Context) (T, error)) *Future[T] {
	return CtxSubmit(ctx, executor, f)
}

func Submit[T any](e Executor, f func() (T, error)) *Future[T] {
	s := &state[T]{}
	e.Submit(func() {
		var val T
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w, err=%s, stack=%s", ErrPanic, r, debug.Stack())
			}
			s.set(val, err)
		}()
		val, err = f()
	})
	return &Future[T]{state: s}
}

func CtxSubmit[T any](ctx context.Context, e Executor, f func(ctx context.Context) (T, error)) *Future[T] {
	s := &state[T]{}
	e.Submit(func() {
		var val T
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w, err=%s, stack=%s", ErrPanic, r, debug.Stack())
			}
			s.set(val, err)
		}()
		val, err = f(ctx)
	})
	return &Future[T]{state: s}
}

func Done[T any](val T) *Future[T] {
	return Done2(val, nil)
}

func Done2[T any](val T, err error) *Future[T] {
	s := &state[T]{}
	s.set(val, err)
	return &Future[T]{state: s}
}

func AnyOf[T any](fs ...*Future[T]) *Future[AnyResult[T]] {
	if len(fs) == 0 {
		return Done(AnyResult[T]{Index: -1})
	}

	var counter int32
	var done uint32
	var errIndex int32 = -1
	s := &state[AnyResult[T]]{}
	for i, f := range fs {
		i := i
		f.state.subscribe(func(val T, err error) {
			if err == nil {
				if atomic.CompareAndSwapUint32(&done, 0, 1) {
					s.set(AnyResult[T]{Index: i, Val: val, Err: err}, nil)
				}
			} else {
				atomic.CompareAndSwapInt32(&errIndex, -1, int32(i))
				if atomic.AddInt32(&counter, 1) == int32(len(fs)) {
					idx := atomic.LoadInt32(&errIndex)
					fval, ferr := fs[idx].Get()
					s.set(AnyResult[T]{Index: int(idx), Val: fval, Err: ferr}, nil)
				}
			}
		})
	}
	return &Future[AnyResult[T]]{state: s}
}

func AllOf[T any](fs ...*Future[T]) *Future[[]T] {
	if len(fs) == 0 {
		return Done[[]T](nil)
	}

	var done uint32
	s := &state[[]T]{}
	c := int32(len(fs))
	results := make([]T, len(fs))
	for i, f := range fs {
		i := i
		f.state.subscribe(func(val T, err error) {
			if err != nil {
				if atomic.CompareAndSwapUint32(&done, 0, 1) {
					s.set(nil, err)
				}
			} else {
				results[i] = val
				if atomic.AddInt32(&c, -1) == 0 {
					s.set(results, nil)
				}
			}
		})
	}
	return &Future[[]T]{state: s}
}

// Deprecated single-Future transforms, kept for source compatibility. They
// delegate to the methods on *Future[T], which hold the only implementation;
// see the package documentation for the layering rule.

// Await blocks until f is resolved.
//
// Deprecated: it is a pure alias for f.Get(), use that instead.
func Await[T any](f *Future[T]) (T, error) {
	return f.Get()
}

// Then chains a synchronous computation onto f.
//
// Deprecated: use f.Then(cb).
func Then[T any, R any](f *Future[T], cb func(T, error) (R, error)) *Future[R] {
	return f.Then(cb)
}

// ThenAsync chains an asynchronous computation onto f.
//
// Deprecated: use f.ThenAsync(cb).
func ThenAsync[T any, R any](f *Future[T], cb func(T, error) *Future[R]) *Future[R] {
	return f.ThenAsync(cb)
}

// ToAny widens f into a *Future[any].
//
// Deprecated: use f.ToAny().
func ToAny[T any](f *Future[T]) *Future[any] {
	return f.ToAny()
}

// ToChan converts f into a single read-only channel of Result[T].
//
// Deprecated: use f.ToChan().
func ToChan[T any](f *Future[T]) <-chan Result[T] {
	return f.ToChan()
}

// Timeout wraps f so that it fails with ErrTimeout when it is not resolved
// within d.
//
// Deprecated: use f.Timeout(d).
func Timeout[T any](f *Future[T], d time.Duration) *Future[T] {
	return f.Timeout(d)
}

// Until wraps f so that it fails with ErrTimeout when it is not resolved before
// t.
//
// Deprecated: use f.Until(t).
func Until[T any](f *Future[T], t time.Time) *Future[T] {
	return f.Until(t)
}
