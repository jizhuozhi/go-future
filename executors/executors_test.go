package executors

import (
	"sync"
	"testing"

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
