//go:build go1.27

package future

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ThenGo submits its callback to the executor, so it has to report a refusal the
// same way the plain entry points do rather than leave the chained Future
// unresolved.
func TestExecutorRejectionFailsThenGo(t *testing.T) {
	withExecutor(t, refusing())

	ran := false
	_, err := Done(1).ThenGo(func(int, error) (int, error) { ran = true; return 2, nil }).Get()

	assert.ErrorIs(t, err, ErrExecutorRejected)
	assert.False(t, ran)
}
