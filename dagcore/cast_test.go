//go:build go1.27

package dagcore

import (
	"context"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

// TestDAG_NodeCast exercises the generic method NodeInstance.Cast (Go 1.27+).
func TestDAG_NodeCast(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return 42, nil
	}))
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)

	// Cast is non-blocking, it can be obtained before the DAG is started.
	f := inst.Nodes()["A"].Cast[int]()

	_, err = inst.Run(context.Background())
	assert.NoError(t, err)

	val, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	// A mismatching type is reported instead of panicking.
	_, err = inst.Nodes()["A"].Cast[string]().Get()
	assert.ErrorIs(t, err, future.ErrTypeMismatch)
}
