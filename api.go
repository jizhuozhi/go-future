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

func Async(f func() (interface{}, error)) *Future {
	s := &state{}
	go func() {
		var val interface{}
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w, err=%s, stack=%s", ErrPanic, r, debug.Stack())
			}
			s.set(val, err)
		}()
		val, err = f()
	}()
	return &Future{state: s}
}

func Lazy(f func() (interface{}, error)) *Future {
	s := &state{}
	s.state |= flagLazy
	s.f = f
	return &Future{state: s}
}

func Done(val interface{}) *Future {
	s := &state{}
	s.set(val, nil)
	return &Future{state: s}
}

func Await(f *Future) (interface{}, error) {
	return f.Get()
}

func Then(f *Future, cb func(interface{}, error) (interface{}, error)) *Future {
	s := &state{}
	f.state.subscribe(func(val interface{}, err error) {
		rval, rerr := cb(val, err)
		s.set(rval, rerr)
	})
	return &Future{state: s}
}

func ThenAsync(f *Future, cb func(interface{}, error) *Future) *Future {
	s := &state{}
	f.state.subscribe(func(val interface{}, err error) {
		cb(val, err).state.subscribe(func(rval interface{}, rerr error) {
			s.set(rval, rerr)
		})
	})
	return &Future{state: s}
}

type AnyResult struct {
	Index int
	Val   interface{}
	Err   error
}

func AnyOf(fs ...*Future) *Future {
	var counter int32
	var done uint32
	var errIndex int32 = -1
	s := &state{}
	for i, f := range fs {
		i := i
		f.state.subscribe(func(val interface{}, err error) {
			if err == nil {
				if atomic.CompareAndSwapUint32(&done, 0, 1) {
					s.set(AnyResult{Index: i, Val: val, Err: err}, nil)
				}
			} else {
				atomic.CompareAndSwapInt32(&errIndex, -1, int32(i))
				if atomic.AddInt32(&counter, 1) == int32(len(fs)) {
					idx := atomic.LoadInt32(&errIndex)
					fval, ferr := fs[idx].Get()
					s.set(AnyResult{Index: int(idx), Val: fval, Err: ferr}, nil)
				}
			}
		})
	}
	return &Future{state: s}
}

func ToAny(f *Future) *Future {
	return Then(f, func(val interface{}, err error) (interface{}, error) {
		return val, err
	})
}

func AllOf(fs ...*Future) *Future {
	var done uint32
	s := &state{}
	c := int32(len(fs))
	for _, f := range fs {
		f.state.subscribe(func(val interface{}, err error) {
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
	return &Future{state: s}
}

func Timeout(f *Future, d time.Duration) *Future {
	var done uint32
	s := &state{}
	timer := time.AfterFunc(d, func() {
		if atomic.CompareAndSwapUint32(&done, 0, 1) {
			var zero interface{}
			s.set(zero, ErrTimeout)
		}
	})
	f.state.subscribe(func(val interface{}, err error) {
		if atomic.CompareAndSwapUint32(&done, 0, 1) {
			s.set(val, err)
			timer.Stop()
		}
	})
	return &Future{state: s}
}

func Until(f *Future, t time.Time) *Future {
	return Timeout(f, t.Sub(time.Now()))
}
