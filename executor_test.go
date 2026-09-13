package future

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jizhuozhi/go-future/executors"
)

func TestSetExecutor(t *testing.T) {
	counter := 0
	SetExecutor(executors.ExecutorFunc(func(f func()) {
		counter++
		go f()
	}))

	f := Async(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, counter)

	assert.Panics(t, func() {
		SetExecutor(nil)
	})
}

// TestSetExecutorConcurrent races SetExecutor against the submit path.
//
// Two things have to hold. Under -race, a plain `var executor Executor` is
// reported here, because SetExecutor writes the variable while Async reads it.
// And because the two implementations below have different dynamic types,
// storing them straight into an atomic.Value panics with "store of
// inconsistently typed value" — which is what the executorBox indirection is
// for.
func TestSetExecutorConcurrent(t *testing.T) {
	defer SetExecutor(executors.GoExecutor{})

	var wg sync.WaitGroup

	// Writers: keep swapping between two different Executor implementations.
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				if i%2 == 0 {
					SetExecutor(executors.GoExecutor{})
				} else {
					SetExecutor(executors.ExecutorFunc(func(f func()) { go f() }))
				}
			}
		}()
	}

	// Readers: keep submitting while the swaps happen.
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				got, err := Async(func() (int, error) { return i, nil }).Get()
				if err != nil {
					t.Errorf("Async returned error %v", err)
					return
				}
				if got != i {
					t.Errorf("Async returned %d, want %d", got, i)
					return
				}
			}
		}()
	}

	wg.Wait()
}
