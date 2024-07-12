package future

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsync(t *testing.T) {
	f := Async(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestLazy(t *testing.T) {
	f := Lazy(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestLazyConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1

	var counter int32
	f := Lazy(func() (int, error) {
		c := atomic.AddInt32(&counter, 1)
		return int(c), nil
	})

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := f.Get()
			assert.Equal(t, val, 1)
			assert.Equal(t, err, nil)
		}()
	}
	wg.Wait()
}
