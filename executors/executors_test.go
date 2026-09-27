package executors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutorFunc(t *testing.T) {
	executor := ExecutorFunc(func(f func()) error {
		go f()
		return nil
	})
	i := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	assert.NoError(t, executor.Submit(func() {
		defer wg.Done()
		i = 1
	}))
	wg.Wait()
	assert.Equal(t, 1, i)
}

func TestExecutorFuncPropagatesRefusal(t *testing.T) {
	refused := errors.New("too many tasks")
	executor := ExecutorFunc(func(func()) error { return refused })

	ran := false
	err := executor.Submit(func() { ran = true })
	assert.ErrorIs(t, err, refused)
	assert.False(t, ran)
}

func TestGoExecutor(t *testing.T) {
	done := make(chan int, 1)
	assert.NoError(t, GoExecutor{}.Submit(func() { done <- 1 }))

	select {
	case v := <-done:
		assert.Equal(t, 1, v)
	case <-time.After(time.Second):
		t.Fatal("GoExecutor.Submit did not run the task")
	}
}
