package dagcore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jizhuozhi/go-future"
	"github.com/jizhuozhi/go-future/executors"
)

// counting returns an executor that counts submissions and runs each task on its
// own goroutine.
func counting(n *int32) future.Executor {
	return executors.ExecutorFunc(func(f func()) error {
		atomic.AddInt32(n, 1)
		go f()
		return nil
	})
}

// WithExecutor routes just that node to its own executor; every other node keeps
// using the global one.
func TestDAG_NodeExecutorIsUsed(t *testing.T) {
	var globalRuns, ownRuns int32

	future.SetExecutor(counting(&globalRuns))
	defer future.SetExecutor(executors.GoExecutor{})

	d := NewDAG()
	assert.NoError(t, d.AddNode("global", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "g", nil
	}))
	assert.NoError(t, d.AddNode("own", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "o", nil
	}, WithExecutor(counting(&ownRuns))))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "g", res["global"])
	assert.Equal(t, "o", res["own"])
	assert.Equal(t, int32(1), atomic.LoadInt32(&globalRuns))
	assert.Equal(t, int32(1), atomic.LoadInt32(&ownRuns))
}

// refusing refuses every submission.
func refusing() future.Executor {
	return executors.ExecutorFunc(func(func()) error { return errRefused })
}

var errRefused = errors.New("pool is full")

// A refused submission has to surface as a failure of that node. Dropping the
// closure would leave the node's future unresolved, and the run would block
// instead of failing.
func TestDAG_RejectedSubmitFailsTheNode(t *testing.T) {
	var ran int32

	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		atomic.AddInt32(&ran, 1)
		return "ran", nil
	}, WithExecutor(refusing())))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	_, err = inst.Run(context.Background())
	assert.ErrorIs(t, err, future.ErrExecutorRejected)
	assert.Zero(t, atomic.LoadInt32(&ran), "a refused node must not run")
}

// A refused submission fails the node, and neither skip nor recovery can absorb
// it: both are evaluated inside the submitted closure, which never started.
//
// So an executor that refuses makes saturation fatal for that node. One that
// wants saturation to be survivable has to run the task some other way — inline,
// for instance — rather than refuse it.
func TestDAG_RejectedSubmitBypassesSkipAndRecovery(t *testing.T) {
	var skipped, recovered int32

	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "ran", nil
	},
		WithExecutor(refusing()),
		WithSkipFunc(func(context.Context, map[NodeID]any) (bool, any) {
			atomic.AddInt32(&skipped, 1)
			return true, "skipped"
		}),
		WithRecoverFunc(func(_ context.Context, _ map[NodeID]any, err error) (any, error) {
			atomic.AddInt32(&recovered, 1)
			return "degraded", nil
		})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	_, err = inst.Run(context.Background())
	assert.ErrorIs(t, err, future.ErrExecutorRejected)
	assert.Zero(t, atomic.LoadInt32(&skipped), "the skip predicate lives inside the closure")
	assert.Zero(t, atomic.LoadInt32(&recovered), "the recovery handler lives inside the closure")
	assert.False(t, inst.nodes["N"].Recovered())
}
