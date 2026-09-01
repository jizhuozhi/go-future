package dagcore

import (
	"context"
	"errors"
	"fmt"
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

func TestDAG_SubgraphExecution(t *testing.T) {
	sub := NewDAG()
	assert.NoError(t, sub.AddInput("x"))
	assert.NoError(t, sub.AddNode("double", []NodeID{"x"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["x"].(int) * 2, nil
	}))
	assert.NoError(t, sub.Freeze())

	mainDAG := NewDAG()
	assert.NoError(t, mainDAG.AddInput("input"))

	inputMapping := func(inputs map[NodeID]any) map[NodeID]any {
		return map[NodeID]any{
			"x": inputs["input"],
		}
	}
	outputMapping := func(outputs map[NodeID]any) any {
		return outputs["double"]
	}

	assert.NoError(t, mainDAG.AddSubgraph("subnode", []NodeID{"input"}, sub, inputMapping, outputMapping))
	assert.NoError(t, mainDAG.Freeze())

	inst, err := mainDAG.Instantiate(map[NodeID]any{"input": 3})
	assert.NoError(t, err)

	res, err := inst.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 6, res["subnode"])
}

func TestDAG_ComplexNestedSubgraphs(t *testing.T) {
	// === 子图 A: 层级型 ===
	levelDAG := NewDAG()
	assert.NoError(t, levelDAG.AddInput("input1"))
	assert.NoError(t, levelDAG.AddInput("input2"))
	assert.NoError(t, levelDAG.AddNode("child1", []NodeID{"input1"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["input1"].(int) + 1, nil
	}))
	assert.NoError(t, levelDAG.AddNode("child2", []NodeID{"child1", "input2"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return deps["child1"].(int) * deps["input2"].(int), nil
	}))
	assert.NoError(t, levelDAG.Freeze())

	// === 子图 B: 并行校验型 ===
	parallelChecks := NewDAG()
	assert.NoError(t, parallelChecks.AddInput("raw"))
	assert.NoError(t, parallelChecks.AddNode("check1", []NodeID{"raw"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return fmt.Sprintf("check1:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.AddNode("check2", []NodeID{"raw"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return fmt.Sprintf("check2:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.AddNode("check3", []NodeID{"raw"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return fmt.Sprintf("check3:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.Freeze())

	d := NewDAG()
	assert.NoError(t, d.AddInput("A"))
	assert.NoError(t, d.AddInput("B"))

	assert.NoError(t, d.AddSubgraph("levelSubgraph", []NodeID{"A", "B"}, levelDAG,
		func(deps map[NodeID]any) map[NodeID]any {
			return map[NodeID]any{
				"input1": deps["A"],
				"input2": deps["B"],
			}
		},
		func(results map[NodeID]any) any {
			return results["child2"]
		}))

	assert.NoError(t, d.AddSubgraph("parallelChecks", []NodeID{"A"}, parallelChecks,
		func(deps map[NodeID]any) map[NodeID]any {
			return map[NodeID]any{
				"raw": deps["A"],
			}
		},
		func(results map[NodeID]any) any {
			return []any{
				results["check1"],
				results["check2"],
				results["check3"],
			}
		}))

	assert.NoError(t, d.AddNode("merge", []NodeID{"levelSubgraph", "parallelChecks"}, func(ctx context.Context, deps map[NodeID]any) (any, error) {
		return fmt.Sprintf("level=%v, checks=%v", deps["levelSubgraph"], deps["parallelChecks"]), nil
	}))

	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(map[NodeID]any{
		"A": 2,
		"B": 3,
	})
	assert.NoError(t, err)

	result, err := inst.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 9, result["levelSubgraph"].(int))
	assert.Len(t, result["parallelChecks"].([]any), 3)
	assert.Contains(t, result["merge"].(string), "level=9")
	assert.Contains(t, result["merge"].(string), "check1")
	assert.Contains(t, result["merge"].(string), "check2")
	assert.Contains(t, result["merge"].(string), "check3")
}
