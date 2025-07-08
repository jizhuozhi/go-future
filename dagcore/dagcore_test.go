package dagcore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDAG_SimpleExecution(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))
	assert.NoError(t, dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"].(string) + "b", nil
	}))

	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "a", res["A"])
	assert.Equal(t, "ab", res["B"])

	assert.Equal(t, dag, inst.Spec())
}

func TestDAG_SimpleWrapExecution(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))
	assert.NoError(t, dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"].(string) + "b", nil
	}))

	assert.NoError(t, dag.Freeze())
	wraps := make(map[NodeID]struct{})
	inst, err := dag.Instantiate(nil, func(n *NodeInstance, run NodeFunc) NodeFunc {
		return func(ctx context.Context, deps map[NodeID]any) (any, error) {
			wraps[n.ID()] = struct{}{}
			return run(ctx, deps)
		}
	})
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "a", res["A"])
	assert.Equal(t, "ab", res["B"])
	assert.Equal(t, wraps, map[NodeID]struct{}{
		"A": {},
		"B": {},
	})
}

func TestDAG_DepFailed(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		panic("not implemented")
	}))
	assert.NoError(t, dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"].(string) + "b", nil
	}))

	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)

	_, err = inst.Run(context.Background())
	assert.Error(t, err)
}

func TestDAG_NodeExisted(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))
	assert.ErrorIs(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}), ErrDAGNodeExisted)
	assert.NoError(t, dag.AddInput("B"))
	assert.ErrorIs(t, dag.AddInput("B"), ErrDAGNodeExisted)
}

func TestDAG_NodeNotRunnable(t *testing.T) {
	dag := NewDAG()
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))
	assert.ErrorIs(t, dag.AddNode("B", nil, nil), ErrDAGNodeNotRunnable)
}

func TestDAG_MissingDependency(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", []NodeID{"Missing"}, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return nil, nil
	})
	assert.ErrorIs(t, dag.Freeze(), ErrDAGIncomplete)
}

func TestDAG_Cyclic(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", []NodeID{"B"}, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return nil, nil
	})
	dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return nil, nil
	})
	assert.ErrorIs(t, dag.Freeze(), ErrDAGCyclic)
}

func TestDAG_ExecutionFailure(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return nil, errors.New("fail")
	})
	dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"], nil
	})
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)
	_, err = inst.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fail")
}

func TestDAG_ParallelExecution(t *testing.T) {
	dag := NewDAG()
	var counter int32

	dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
		return "a", nil
	})
	dag.AddNode("B", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
		return "b", nil
	})
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)
	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, int32(2), counter)
}

func TestDAG_NodeNotExecuted(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return nil, errors.New("fail")
	})
	dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return "b", nil
	})
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)
	_, err = inst.Run(context.Background())
	assert.Error(t, err)

	for _, node := range inst.Nodes() {
		_, getErr := node.Future().Get()
		assert.NotNil(t, getErr)
	}
}

func TestDAG_NodeDuration(t *testing.T) {
	dag := NewDAG()
	dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		time.Sleep(20 * time.Millisecond)
		return "ok", nil
	})
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)
	_, err = inst.Run(context.Background())
	assert.NoError(t, err)
	dur := inst.Nodes()["A"].Duration()
	assert.GreaterOrEqual(t, dur.Milliseconds(), int64(15))
}

func TestDAG_ForwardReference(t *testing.T) {
	dag := NewDAG()
	// Forward reference: B depends on A, but A is added later
	assert.NoError(t, dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"].(string) + "b", nil
	}))
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))

	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(nil)
	assert.NoError(t, err)
	res, err := inst.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ab", res["B"])
}

func TestDAG_Input(t *testing.T) {
	dag := NewDAG()
	dag.AddInput("A")
	dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		time.Sleep(20 * time.Millisecond)
		assert.Equal(t, "a", deps["A"])
		return "ok", nil
	})
	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(map[NodeID]any{"A": "a"})
	assert.NoError(t, err)
	_, err = inst.Run(context.Background())
	assert.NoError(t, err)
}

func TestDAG_MissInput(t *testing.T) {
	dag := NewDAG()
	dag.AddInput("A")
	dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		time.Sleep(20 * time.Millisecond)
		assert.Equal(t, "a", deps["A"])
		return "ok", nil
	})
	assert.NoError(t, dag.Freeze())
	_, err := dag.Instantiate(nil)
	assert.ErrorIs(t, err, ErrDAGNodeNotRunnable)
}

func TestDAG_InputDuplicated(t *testing.T) {
	dag := NewDAG()
	// Forward reference: B depends on A, but A is added later
	assert.NoError(t, dag.AddNode("B", []NodeID{"A"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["A"].(string) + "b", nil
	}))
	assert.NoError(t, dag.AddNode("A", nil, func(ctx context.Context, _ map[NodeID]any) (any, error) {
		return "a", nil
	}))
	assert.NoError(t, dag.Freeze())
	_, err := dag.Instantiate(map[NodeID]any{"A": "a"})
	assert.ErrorIs(t, err, ErrDAGNodeNotInput)
}

func TestDAG_SimpleExecutionMermaid(t *testing.T) {
	dag := NewDAG()

	// Inputs
	assert.NoError(t, dag.AddInput("inputA"))
	assert.NoError(t, dag.AddInput("inputB"))

	// Nodes
	err := dag.AddNode("node1", []NodeID{"inputA"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return "ok", nil
	})
	assert.NoError(t, err)

	err = dag.AddNode("node2", []NodeID{"inputA", "inputB"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return "ok", nil
	})
	assert.NoError(t, err)

	err = dag.AddNode("node3", []NodeID{"node1", "node2"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return "ok", nil
	})
	assert.NoError(t, err)

	assert.NoError(t, dag.Freeze())
	inst, err := dag.Instantiate(map[NodeID]any{
		"inputA": nil,
		"inputB": nil,
	})
	assert.NoError(t, err)

	out := ToMermaid(inst)

	expected := `graph LR
	inputA["inputA"]
	inputB["inputB"]
	node1(("node1"))
	node2(("node2"))
	node3(("node3"))
	inputA --> node1
	inputA --> node2
	inputB --> node2
	node1 --> node3
	node2 --> node3
`

	assert.Equal(t, expected, out)
}

func TestDAG_Freeze(t *testing.T) {
	dag := NewDAG()
	_, err := dag.Instantiate(nil)
	assert.ErrorIs(t, err, ErrDAGNotFrozen)

	assert.False(t, dag.Frozen())
	assert.NoError(t, dag.Freeze())
	assert.ErrorIs(t, dag.Freeze(), ErrDAGFrozen)
	assert.True(t, dag.Frozen())
	assert.ErrorIs(t, dag.AddInput("inputA"), ErrDAGFrozen)
	assert.ErrorIs(t, dag.AddNode("inputA", nil, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return nil, nil
	}), ErrDAGFrozen)
	_, err = dag.Instantiate(nil)
	assert.NoError(t, err)
}
