package future

import (
	"sync/atomic"
)

const (
	stateFree uint64 = iota
	stateGray
	stateDone
)

type state[T any] struct {
	noCopy noCopy

	state uint64 // high 32 bits are state, low 32 bits are waiter count.
	sema  uint32

	val T
	err error
}

type Promise[T any] struct {
	state state[T]
}

type Future[T any] struct {
	state *state[T]
}

func (s *state[T]) set(val T, err error) {
	for {
		st := atomic.LoadUint64(&s.state)
		if (st >> 32) > stateFree {
			panic("promise already satisfied")
		}
		if atomic.CompareAndSwapUint64(&s.state, st, st+(1<<32)) {
			s.val = val
			s.err = err
			st = atomic.AddUint64(&s.state, 1<<32)
			for w := st & (1<<32 - 1); w > 0; w-- {
				runtime_Semrelease(&s.sema, false, 0)
			}
			return
		}
	}
}

func (s *state[T]) get() (T, error) {
	for {
		st := atomic.LoadUint64(&s.state)
		if (st >> 32) == stateDone {
			return s.val, s.err
		}
		if atomic.CompareAndSwapUint64(&s.state, st, st+1) {
			runtime_Semacquire(&s.sema)
			return s.val, s.err
		}
	}
}

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{}
}

func (p *Promise[T]) Set(val T, err error) {
	p.state.set(val, err)
}

func (p *Promise[T]) Future() *Future[T] {
	return &Future[T]{state: &p.state}
}

func (f *Future[T]) Get() (T, error) {
	return f.state.get()
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
