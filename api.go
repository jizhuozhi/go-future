package future

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var ErrPanic = errors.New("async panic")
var ErrTimeout = errors.New("future timeout")

func Async[T any](f func() (T, error)) *Future[T] {
	s := &state[T]{}
	go func() {
		var val T
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w, err=%s, stack=%s", ErrPanic, r, debug.Stack())
			}
			s.set(val, err)
		}()
		val, err = f()
	}()
	return &Future[T]{state: s}
}

func Lazy[T any](f func() (T, error)) *Future[T] {
	s := &state[T]{}
	s.state |= flagLazy
	s.f = f
	return &Future[T]{state: s}
}

func Then[T any, R any](f *Future[T], cb func(T, error) (R, error)) *Future[R] {
	s := &state[R]{}
	f.state.subscribe(func(val T, err error) {
		rval, rerr := cb(val, err)
		s.set(rval, rerr)
	})
	return &Future[R]{state: s}
}

func ThenAsync[T any, R any](f *Future[T], cb func(T, error) *Future[R]) *Future[R] {
	s := &state[R]{}
	f.state.subscribe(func(val T, err error) {
		cb(val, err).state.subscribe(func(rval R, rerr error) {
			s.set(rval, rerr)
		})
	})
	return &Future[R]{state: s}
}

type AnyResult[T any] struct {
	Index int
	Val   T
	Err   error
}

func AnyOf[T any](fs ...*Future[T]) *Future[AnyResult[T]] {
	var done uint32
	s := &state[AnyResult[T]]{}
	for i, f := range fs {
		i := i
		f.state.subscribe(func(val T, err error) {
			if atomic.CompareAndSwapUint32(&done, 0, 1) {
				s.set(AnyResult[T]{Index: i, Val: val, Err: err}, nil)
			}
		})
	}
	return &Future[AnyResult[T]]{state: s}
}

func AllOf[T any](fs ...*Future[T]) *Future[struct{}] {
	var done uint32
	s := &state[struct{}]{}
	c := int32(len(fs))
	for _, f := range fs {
		f.state.subscribe(func(val T, err error) {
			if err != nil {
				if atomic.CompareAndSwapUint32(&done, 0, 1) {
					s.set(struct{}{}, err)
				}
			} else {
				if atomic.AddInt32(&c, -1) == 0 {
					s.set(struct{}{}, nil)
				}
			}
		})
	}
	return &Future[struct{}]{state: s}
}

func Timeout[T any](f *Future[T], d time.Duration) *Future[T] {
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

func Until[T any](f *Future[T], t time.Time) *Future[T] {
	return Timeout(f, t.Sub(time.Now()))
}
