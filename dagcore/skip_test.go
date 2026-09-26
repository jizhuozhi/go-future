package dagcore

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

var (
	errTransient = errors.New("backend unavailable")
	errDefect    = errors.New("model and feature id spaces differ")
)

// Test-side atomics over the untyped functions: atomic.Bool and atomic.Int32
// are Go 1.19, while this module declares go 1.18 and is built against it.
type atomicFlag struct{ v int32 }

func (f *atomicFlag) Store(v bool) {
	var n int32
	if v {
		n = 1
	}
	atomic.StoreInt32(&f.v, n)
}

func (f *atomicFlag) Load() bool { return atomic.LoadInt32(&f.v) != 0 }

type atomicCounter struct{ v int32 }

func (c *atomicCounter) Add(delta int32) { atomic.AddInt32(&c.v, delta) }
func (c *atomicCounter) Load() int32     { return atomic.LoadInt32(&c.v) }

// WithSkipFunc and WithRecoverFunc stay independent: one callback for both
// makes "may be skipped, but a real failure must fail the run" inexpressible,
// and silently switches on error swallowing.
func TestDAG_SkipAndRecoveryAreIndependent(t *testing.T) {
	failing := func(context.Context, map[NodeID]any) (any, error) { return nil, errFoo }

	// A predicate is configured but declines to skip.
	notSkipping := WithSkipFunc(func(context.Context, map[NodeID]any) (bool, any) {
		return false, "ignored"
	})
	skipping := WithSkipFunc(func(context.Context, map[NodeID]any) (bool, any) {
		return true, "skipped"
	})
	recovering := WithRecoverFunc(func(context.Context, map[NodeID]any, error) (any, error) {
		return "recovered", nil
	})

	build := func(t *testing.T, opts ...NodeOpt) *DAGInstance {
		t.Helper()
		d := NewDAG()
		assert.NoError(t, d.AddNode("N", nil, failing, opts...))
		assert.NoError(t, d.Freeze())
		inst, err := d.Instantiate(nil)
		assert.NoError(t, err)
		return inst
	}

	t.Run("configured but declining does not recover failures", func(t *testing.T) {
		inst := build(t, notSkipping)
		_, err := inst.Run(context.Background())
		assert.ErrorIs(t, err, errFoo,
			"configuring a skip predicate must not implicitly make a node optional")
		assert.ErrorIs(t, inst.nodes["N"].Err(), errFoo)
		assert.False(t, inst.nodes["N"].Recovered())
	})

	t.Run("recovery only", func(t *testing.T) {
		inst := build(t, recovering)
		res, err := inst.Run(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "recovered", res["N"])
		assert.False(t, inst.nodes["N"].Skipped())
		assert.True(t, inst.nodes["N"].Recovered())
		assert.ErrorIs(t, inst.nodes["N"].Err(), errFoo,
			"a recovered error must stay observable; it is invisible in the run result")
	})

	t.Run("skip takes precedence over recovery", func(t *testing.T) {
		inst := build(t, skipping, recovering)
		res, err := inst.Run(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "skipped", res["N"],
			"the value from the skip predicate must win; the node never ran")
		assert.True(t, inst.nodes["N"].Skipped())
		assert.False(t, inst.nodes["N"].Recovered())
		assert.NoError(t, inst.nodes["N"].Err())
	})

	t.Run("skip only, no recovery", func(t *testing.T) {
		inst := build(t, skipping)
		res, err := inst.Run(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "skipped", res["N"])
		assert.False(t, inst.nodes["N"].Recovered())
	})

	t.Run("neither: mandatory", func(t *testing.T) {
		inst := build(t)
		_, err := inst.Run(context.Background())
		assert.ErrorIs(t, err, errFoo)
		assert.False(t, inst.nodes["N"].Recovered())
	})
}

// Recovered is the only flag that separates an absorbed failure from a
// propagated one: both leave Err populated.
func TestDAG_RecoveredSeparatesAbsorbedFromPropagated(t *testing.T) {
	absorbed, res, err := runFailing(t, errFoo, WithRecoverFunc(
		func(context.Context, map[NodeID]any, error) (any, error) { return "degraded", nil }))
	assert.NoError(t, err)
	assert.Equal(t, "degraded", res["N"])
	assert.True(t, absorbed.nodes["N"].Recovered())
	assert.ErrorIs(t, absorbed.nodes["N"].Err(), errFoo)

	propagated, _, err := runFailing(t, errFoo, WithRecoverFunc(
		func(_ context.Context, _ map[NodeID]any, err error) (any, error) { return nil, err }))
	assert.ErrorIs(t, err, errFoo)
	assert.False(t, propagated.nodes["N"].Recovered(),
		"a declined error must not be reported as recovered")
	assert.ErrorIs(t, propagated.nodes["N"].Err(), errFoo,
		"Err is populated in both cases; it cannot tell them apart")
}

// A successful node is neither skipped nor recovered.
func TestDAG_RecoveredIsFalseOnSuccess(t *testing.T) {
	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "ok", nil
	}, WithRecoverFunc(func(context.Context, map[NodeID]any, error) (any, error) {
		return "unused", nil
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ok", res["N"])
	assert.False(t, inst.nodes["N"].Recovered())
	assert.False(t, inst.nodes["N"].Skipped())
	assert.NoError(t, inst.nodes["N"].Err())
}

// runFailing builds a single-node DAG whose node always fails with err, runs
// it, and returns the instance together with the run outcome.
func runFailing(t *testing.T, err error, opts ...NodeOpt) (*DAGInstance, map[NodeID]any, error) {
	t.Helper()
	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		return nil, err
	}, opts...))
	assert.NoError(t, d.Freeze())

	inst, instantiateErr := d.Instantiate(nil)
	assert.NoError(t, instantiateErr)

	res, runErr := inst.Run(context.Background())
	return inst, res, runErr
}

// Recovery is a per-failure decision: a handler may decline an error, and the
// run then fails exactly as if no handler had been registered.
func TestDAG_RecoveryMayDecline(t *testing.T) {
	inst, _, err := runFailing(t, errFoo, WithRecoverFunc(
		func(_ context.Context, _ map[NodeID]any, err error) (any, error) {
			return nil, err
		}))

	assert.ErrorIs(t, err, errFoo, "a declining handler must let the error fail the run")
	assert.ErrorIs(t, inst.nodes["N"].Err(), errFoo)
}

// The point of passing the error in: recover one class of error, propagate
// another.
func TestDAG_RecoveryIsSelective(t *testing.T) {
	handler := WithRecoverFunc(func(_ context.Context, _ map[NodeID]any, err error) (any, error) {
		if errors.Is(err, errTransient) {
			return "degraded", nil
		}
		return nil, err
	})

	t.Run("transient error is recovered", func(t *testing.T) {
		inst, res, err := runFailing(t, fmt.Errorf("gpu pool exhausted: %w", errTransient), handler)
		assert.NoError(t, err)
		assert.Equal(t, "degraded", res["N"])
		assert.ErrorIs(t, inst.nodes["N"].Err(), errTransient)
	})

	t.Run("defect is not recovered", func(t *testing.T) {
		_, _, err := runFailing(t, fmt.Errorf("inference: %w", errDefect), handler)
		assert.ErrorIs(t, err, errDefect,
			"a defect must not be swallowed by a handler that only recovers transient errors")
	})
}

// A declining handler may add context to the error. That annotated error is
// what fails the run, while Err keeps the error the node function returned.
func TestDAG_DeclinedRecoveryMayAnnotate(t *testing.T) {
	inst, _, err := runFailing(t, errFoo, WithRecoverFunc(
		func(_ context.Context, _ map[NodeID]any, err error) (any, error) {
			return nil, fmt.Errorf("node N declined to recover: %w", err)
		}))

	assert.ErrorIs(t, err, errFoo)
	assert.Contains(t, err.Error(), "declined to recover",
		"the annotated error is what fails the run")
	assert.ErrorIs(t, inst.nodes["N"].Err(), errFoo)

	// Err holds the original error, not the annotation.
	assert.NotContains(t, inst.nodes["N"].Err().Error(), "declined to recover")
}

// A skipped node's function never runs, so it costs no downstream resources and
// leaves no near-zero sample in the latency distribution.
func TestDAG_SkipDoesNotInvokeNodeFunc(t *testing.T) {
	var called atomicFlag

	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		called.Store(true)
		return "ran", nil
	}, WithSkipFunc(func(context.Context, map[NodeID]any) (bool, any) {
		return true, "skipped"
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "skipped", res["N"])
	assert.False(t, called.Load(), "the node function must not run when skipped")
	assert.True(t, inst.nodes["N"].Skipped())
}

// The skip value is computed per evaluation, so it may depend on why the node
// was skipped. A statically configured value cannot distinguish between
// "irrelevant for this request" and "not enough budget left".
func TestDAG_SkipValueMayDependOnReason(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddInput("budget_ms"))
	assert.NoError(t, dag.AddNode("N", []NodeID{"budget_ms"}, func(context.Context, map[NodeID]any) (any, error) {
		return "computed", nil
	}, WithSkipFunc(func(_ context.Context, deps map[NodeID]any) (bool, any) {
		switch deps["budget_ms"].(int) {
		case 0:
			return true, "no-budget-fallback"
		case 1:
			return true, "irrelevant-for-this-request"
		default:
			return false, nil
		}
	})))
	assert.NoError(t, dag.Freeze())

	cases := map[int]string{
		0: "no-budget-fallback",
		1: "irrelevant-for-this-request",
		9: "computed",
	}
	for budget, want := range cases {
		inst, err := dag.Instantiate(map[NodeID]any{"budget_ms": budget})
		assert.NoError(t, err)

		res, err := inst.Run(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, want, res["N"], "budget=%d", budget)
	}
}

// A failed dependency short-circuits before the skip predicate is consulted.
// Skipping is not error handling; to tolerate an upstream failure, give the
// upstream node a WithRecoverFunc.
func TestDAG_SkipCannotObserveFailedDependency(t *testing.T) {
	var predicateCalled atomicFlag

	d := NewDAG()
	assert.NoError(t, d.AddNode("upstream", nil, func(context.Context, map[NodeID]any) (any, error) {
		return nil, errFoo
	}))
	assert.NoError(t, d.AddNode("downstream", []NodeID{"upstream"}, func(context.Context, map[NodeID]any) (any, error) {
		return "ran", nil
	}, WithSkipFunc(func(context.Context, map[NodeID]any) (bool, any) {
		predicateCalled.Store(true)
		return false, nil
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	_, err = inst.Run(context.Background())
	assert.ErrorIs(t, err, errFoo)
	assert.False(t, predicateCalled.Load(),
		"a skip predicate must not be consulted when a dependency failed; "+
			"give the upstream node a WithRecoverFunc to tolerate its failure instead")
}

// A recovered upstream failure lets the downstream predicate run, so it can
// decide whether the node is still worth executing.
func TestDAG_RecoveredDependencyLetsDownstreamDecide(t *testing.T) {
	d := NewDAG()
	assert.NoError(t, d.AddNode("upstream", nil, func(context.Context, map[NodeID]any) (any, error) {
		return nil, errFoo
	}, WithRecoverFunc(func(context.Context, map[NodeID]any, error) (any, error) {
		return -1, nil
	})))
	assert.NoError(t, d.AddNode("downstream", []NodeID{"upstream"}, func(context.Context, map[NodeID]any) (any, error) {
		return "ran", nil
	}, WithSkipFunc(func(_ context.Context, deps map[NodeID]any) (bool, any) {
		// The upstream value was recovered to -1, so stop here.
		return deps["upstream"].(int) < 0, "upstream-degraded"
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, -1, res["upstream"])
	assert.Equal(t, "upstream-degraded", res["downstream"])
	assert.True(t, inst.nodes["downstream"].Skipped())
}

// Err must be nil for nodes that neither failed nor were skipped, otherwise
// counting recovered failures via Err would also count healthy nodes.
func TestDAG_ErrIsNilOnSuccess(t *testing.T) {
	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "ok", nil
	}, WithRecoverFunc(func(context.Context, map[NodeID]any, error) (any, error) {
		return "unused", nil
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ok", res["N"])
	assert.NoError(t, inst.nodes["N"].Err())
	assert.False(t, inst.nodes["N"].Skipped())
}

// The handler is only called on failure; the success path is unaffected.
func TestDAG_RecoverNotCalledOnSuccess(t *testing.T) {
	var called atomicCounter

	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		return "ok", nil
	}, WithRecoverFunc(func(context.Context, map[NodeID]any, error) (any, error) {
		called.Add(1)
		return "never", nil
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ok", res["N"])
	assert.Zero(t, called.Load())
}

// A panic stops short of the handler: future.CtxAsync converts it into an
// ErrPanic error outside the node closure, so the unwind skips the handler and
// the run fails with neither Recovered nor Err reporting it. An optional node
// is therefore not panic-tolerant — change this test if that should change.
func TestDAG_RecoveryDoesNotSeePanics(t *testing.T) {
	const boom = "assignment to entry in nil map"

	var called atomicCounter

	d := NewDAG()
	assert.NoError(t, d.AddNode("N", nil, func(context.Context, map[NodeID]any) (any, error) {
		panic(boom)
	}, WithRecoverFunc(func(_ context.Context, _ map[NodeID]any, _ error) (any, error) {
		called.Add(1)
		return "degraded", nil
	})))
	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())

	assert.ErrorIs(t, err, future.ErrPanic)
	assert.Contains(t, err.Error(), boom)
	assert.Nil(t, res)
	assert.Zero(t, called.Load(), "the handler must not be reached")
	assert.False(t, inst.nodes["N"].Recovered())
	assert.NoError(t, inst.nodes["N"].Err(), "the error is never assigned to the node")
}
