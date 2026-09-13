package executors

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutorFunc(t *testing.T) {
	executor := ExecutorFunc(func(f func()) {
		go f()
	})
	i := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	executor.Submit(func() {
		defer wg.Done()
		i = 1
	})
	wg.Wait()
	assert.Equal(t, 1, i)
}

func TestGoExecutor(t *testing.T) {
	done := make(chan int, 1)
	GoExecutor{}.Submit(func() { done <- 1 })

	select {
	case v := <-done:
		assert.Equal(t, 1, v)
	case <-time.After(time.Second):
		t.Fatal("GoExecutor.Submit did not run the task")
	}
}
