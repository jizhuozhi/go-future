package future

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func TestPromiseAndFuture(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future()
	p.Set(1, errFoo)
	val, err := f.Get()
	assert.Equal(t, val, 1)
	assert.Equal(t, err, errFoo)
	assert.Equal(t, 2, f.GetOrDefault(2))
}

func TestPromiseAndFutureConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1

	ch := make(chan struct{}, n)
	p := NewPromise[int]()
	go func() {
		for i := 0; i < n; i++ {
			ch <- struct{}{}
		}
		time.Sleep(1 * time.Second)
		p.Set(1, errFoo)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ch
			f := p.Future()
			val, err := f.Get()
			assert.Equal(t, val, 1)
			assert.Equal(t, err, errFoo)
		}()
	}
	wg.Wait()
}

func TestPromiseSetTwice(t *testing.T) {
	p := NewPromise[int]()
	p.Set(1, nil)
	assert.Panics(t, func() {
		p.Set(1, nil)
	})
}

func TestPromiseSetSafetyTwice(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future()
	p.SetSafety(1, nil)
	p.SetSafety(2, nil)
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.NoError(t, err)
}

func TestFutureSubscribe(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future()
	val1 := 0
	val2 := 0
	f.Subscribe(func(val int, err error) {
		val1 = val + 1
	})
	f.Subscribe(func(val int, err error) {
		val2 = val + 2
	})
	p.Set(1, nil)
	assert.Equal(t, val1, 2)
	assert.Equal(t, val2, 3)
}

func TestFutureSubscribeConcurrency(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := 1000

		ch := make(chan struct{}, n)
		p := NewPromise[int]()
		go func() {
			for i := 0; i < n; i++ {
				ch <- struct{}{}
			}
			p.Set(1, errFoo)
		}()

		var counter int32
		wg := sync.WaitGroup{}
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				<-ch
				f := p.Future()

				f.Subscribe(func(val int, err error) {
					defer wg.Done()
					atomic.AddInt32(&counter, 1)
				})
			}()
		}
		wg.Wait()

		assert.Equal(t, int32(n), atomic.LoadInt32(&counter))
	}
}

func TestPromiseFreeAndFutureDone(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future()
	assert.True(t, p.Free())
	assert.False(t, f.Done())

	p.Set(1, nil)
	assert.False(t, p.Free())
	assert.True(t, f.Done())
}

func Benchmark(b *testing.B) {
	b.Run("Promise", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := NewPromise[int]()
			f := p.Future()
			go func() {
				p.Set(1, nil)
			}()
			_, _ = f.Get()
		}
	})
	b.Run("WaitGroup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var val int
			var err error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				val, err = 0, nil
				wg.Done()
			}()
			wg.Wait()
			_, _ = val, err
		}
	})
	// channel does not support multi-consumers. This is used to compare the performance of a single consumer.
	b.Run("Channel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var val int
			var err error
			ch := make(chan struct{})
			go func() {
				val, err = 0, nil
				ch <- struct{}{}
			}()
			<-ch
			_, _ = val, err
		}
	})
}
