package future

import (
	"sync/atomic"
	"unsafe"
)

const (
	stateFree uint64 = iota
	stateDoing
	stateDone
)

const stateDelta = 1 << 32

const (
	maskCounter = 1<<32 - 1
	maskState   = 1<<34 - 1
)

const flagLazy uint64 = 1 << 63

// Promise The Promise provides a facility to store a value or an error that is later acquired asynchronously via a Future
// created by the Promise. Note that the Promise object is meant to be set only once.
//
// Each Promise is associated with a shared state, which contains some state information and a result which may be not yet evaluated,
// evaluated to a value (possibly nil) or evaluated to an error.
//
// The Promise is the "push" end of the promise-future communication channel: the operation that stores a value in the shared state
// synchronizes-with (as defined in Go's memory model) the successful return from any function that is waiting on the shared state
// (such as Future.Get).
//
// A Promise must not be copied after first use.
type Promise[T any] struct {
	state state[T]
}

// Future The Future provides a mechanism to access the result of asynchronous operations:
//
// 1. An asynchronous operation (Async, Lazy or Promise) can provide a Future to the creator of that asynchronous operation.
//
// 2. The creator of the asynchronous operation can then use a variety of methods to query, wait for, or extract a value from the Future.
// These methods may block if the asynchronous operation has not yet provided a value.
//
// 3. When the asynchronous operation is ready to send a result to the creator, it can do so by modifying shared state (e.g. Promise.Set)
// that is linked to the creator's std::future.
//
// The Future also has the ability to register a callback to be called when the asynchronous operation is ready to send a result to the creator.
type Future[T any] struct {
	state *state[T]
}

func (s *state[T]) set(val T, err error) bool {
	for {
		st := atomic.LoadUint64(&s.state)
		if !isFree(st) {
			return false
		}
		if atomic.CompareAndSwapUint64(&s.state, st, st+stateDelta) {
			s.val = val
			s.err = err
			st = atomic.AddUint64(&s.state, stateDelta)
			for w := st & maskCounter; w > 0; w-- {
				runtime_Semrelease(&s.sema, false, 0)
			}
			for {
				head := (*callback[T])(atomic.LoadPointer(&s.stack))
				if head == nil {
					break
				}
				if atomic.CompareAndSwapPointer(&s.stack, unsafe.Pointer(head), unsafe.Pointer(head.next)) {
					head.f(val, err)
					head.next = nil
				}
			}
			return true
		}
	}
}

func (s *state[T]) get() (T, error) {
	if atomic.LoadUint64(&s.state)&flagLazy == flagLazy {
		for {
			st := atomic.LoadUint64(&s.state)
			if st&flagLazy != flagLazy {
				break
			}
			if atomic.CompareAndSwapUint64(&s.state, st, st&(^flagLazy)) {
				val, err := s.f()
				s.set(val, err)
				return val, err
			}
		}
	}
	for {
		st := atomic.LoadUint64(&s.state)
		if isDone(st) {
			return s.val, s.err
		}
		if atomic.CompareAndSwapUint64(&s.state, st, st+1) {
			runtime_Semacquire(&s.sema)
			if !isDone(atomic.LoadUint64(&s.state)) {
				panic("sync: notified before state has done")
			}
			return s.val, s.err
		}
	}
}

func (s *state[T]) subscribe(cb func(T, error)) {
	newCb := &callback[T]{f: cb}
	for {
		oldCb := (*callback[T])(atomic.LoadPointer(&s.stack))

		if isDone(atomic.LoadUint64(&s.state)) {
			cb(s.val, s.err)
			return
		}

		newCb.next = oldCb
		if atomic.CompareAndSwapPointer(&s.stack, unsafe.Pointer(oldCb), unsafe.Pointer(newCb)) {
			for {
				// Double-check the state to ensure the callback is not missed
				if isDone(atomic.LoadUint64(&s.state)) {
					if atomic.CompareAndSwapPointer(&s.stack, unsafe.Pointer(newCb), unsafe.Pointer(newCb.next)) {
						cb(s.val, s.err)
						return
					}
				} else {
					return
				}
			}
		}
	}
}

// NewPromise creates a new Promise object.
func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{}
}

// Set sets the value and error of the Promise.
func (p *Promise[T]) Set(val T, err error) {
	if !p.state.set(val, err) {
		panic("promise already satisfied")
	}
}

// SetSafety sets the value and error of the Promise, and it will return false if already set.
func (p *Promise[T]) SetSafety(val T, err error) bool {
	return p.state.set(val, err)
}

// Future returns a Future object associated with the Promise.
func (p *Promise[T]) Future() *Future[T] {
	return &Future[T]{state: &p.state}
}

// Free returns true if the Promise is not set.
func (p *Promise[T]) Free() bool {
	return isFree(atomic.LoadUint64(&p.state.state))
}

// Get returns the value and error of the Future.
func (f *Future[T]) Get() (T, error) {
	return f.state.get()
}

// GetOrDefault returns the value of the Future. If error has been set, it returns the default value.
func (f *Future[T]) GetOrDefault(defaultVal T) T {
	val, err := f.state.get()
	if err != nil {
		return defaultVal
	}
	return val
}

// Subscribe registers a callback to be called when the Future is done.
//
// NOTE: The callback will be called in goroutine that is the same as the goroutine which changed Future state.
// The callback should not contain any blocking operations.
func (f *Future[T]) Subscribe(cb func(val T, err error)) {
	f.state.subscribe(cb)
}

// Done returns true if the Future is done.
func (f *Future[T]) Done() bool {
	return isDone(atomic.LoadUint64(&f.state.state))
}

func isFree(st uint64) bool {
	return ((st & maskState) >> 32) == stateFree
}

func isDone(st uint64) bool {
	return ((st & maskState) >> 32) == stateDone
}

type state[T any] struct {
	noCopy noCopy

	state uint64 // high 30 bits are flags, mid 2 bits are state, low 32 bits are waiter count.
	sema  uint32

	val T
	err error
	f   func() (T, error)

	stack unsafe.Pointer // *callback[T]
}

type callback[T any] struct {
	f    func(T, error)
	next *callback[T]
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
