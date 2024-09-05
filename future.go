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

type Promise[T any] struct {
	state state[T]
}

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

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{}
}

func (p *Promise[T]) Set(val T, err error) {
	if !p.state.set(val, err) {
		panic("promise already satisfied")
	}
}

func (p *Promise[T]) SetSafety(val T, err error) bool {
	return p.state.set(val, err)
}

func (p *Promise[T]) Future() *Future[T] {
	return &Future[T]{state: &p.state}
}

func (p *Promise[T]) Free() bool {
	return isFree(atomic.LoadUint64(&p.state.state))
}

func (f *Future[T]) Get() (T, error) {
	return f.state.get()
}

func (f *Future[T]) GetOrDefault(defaultVal T) T {
	val, err := f.state.get()
	if err != nil {
		return defaultVal
	}
	return val
}

func (f *Future[T]) Subscribe(cb func(val T, err error)) {
	f.state.subscribe(cb)
}

func (f *Future[T]) Done() bool {
	return isDone(atomic.LoadUint64(&f.state.state))
}

func isFree(st uint64) bool {
	return ((st & maskState) >> 32) == stateFree
}

func isDone(st uint64) bool {
	return ((st & maskState) >> 32) == stateDone
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
