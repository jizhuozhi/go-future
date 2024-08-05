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

type state struct {
	noCopy noCopy

	state uint64 // high 30 bits are flags, mid 2 bits are state, low 32 bits are waiter count.
	sema  uint32

	val interface{}
	err error
	f   func() (interface{}, error)

	stack unsafe.Pointer // *callback
}

type callback struct {
	f    func(interface{}, error)
	next *callback
}

type Promise struct {
	state state
}

type Future struct {
	state *state
}

func (s *state) set(val interface{}, err error) {
	for {
		st := atomic.LoadUint64(&s.state)
		if !isFree(st) {
			panic("promise already satisfied")
		}
		if atomic.CompareAndSwapUint64(&s.state, st, st+stateDelta) {
			s.val = val
			s.err = err
			st = atomic.AddUint64(&s.state, stateDelta)
			for w := st & maskCounter; w > 0; w-- {
				runtime_Semrelease(&s.sema, false, 0)
			}
			for {
				head := (*callback)(atomic.LoadPointer(&s.stack))
				if head == nil {
					break
				}
				if atomic.CompareAndSwapPointer(&s.stack, unsafe.Pointer(head), unsafe.Pointer(head.next)) {
					head.f(val, err)
					head.next = nil
				}
			}
			return
		}
	}
}

func (s *state) get() (interface{}, error) {
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

func (s *state) subscribe(cb func(interface{}, error)) {
	newCb := &callback{f: cb}
	for {
		oldCb := (*callback)(atomic.LoadPointer(&s.stack))

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

func NewPromise() *Promise {
	return &Promise{}
}

func (p *Promise) Set(val interface{}, err error) {
	p.state.set(val, err)
}

func (p *Promise) Future() *Future {
	return &Future{state: &p.state}
}

func (p *Promise) Free() bool {
	return isFree(atomic.LoadUint64(&p.state.state))
}

func (f *Future) Get() (interface{}, error) {
	return f.state.get()
}

func (f *Future) GetOrDefault(defaultVal interface{}) interface{} {
	val, err := f.state.get()
	if err != nil {
		return defaultVal
	}
	return val
}

func (f *Future) Subscribe(cb func(val interface{}, err error)) {
	f.state.subscribe(cb)
}

func (f *Future) Done() bool {
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
