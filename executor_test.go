package future

import (
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
