package dagviz

import (
	"context"
	"fmt"
	"github.com/jizhuozhi/go-future/dagcore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDAG_ComplexNestedSubgraphs(t *testing.T) {
	// === 子图 A: 层级型 ===
	levelDAG := dagcore.NewDAG()
	assert.NoError(t, levelDAG.AddInput("input1"))
	assert.NoError(t, levelDAG.AddInput("input2"))
	assert.NoError(t, levelDAG.AddNode("child1", []dagcore.NodeID{"input1"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return deps["input1"].(int) + 1, nil
	}))
	assert.NoError(t, levelDAG.AddNode("child2", []dagcore.NodeID{"child1", "input2"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return deps["child1"].(int) * deps["input2"].(int), nil
	}))
	assert.NoError(t, levelDAG.Freeze())

	// === 子图 B: 并行校验型 ===
	parallelChecks := dagcore.NewDAG()
	assert.NoError(t, parallelChecks.AddInput("raw"))
	assert.NoError(t, parallelChecks.AddNode("check1", []dagcore.NodeID{"raw"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return fmt.Sprintf("check1:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.AddNode("check2", []dagcore.NodeID{"raw"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return fmt.Sprintf("check2:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.AddNode("check3", []dagcore.NodeID{"raw"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return fmt.Sprintf("check3:%v", deps["raw"]), nil
	}))
	assert.NoError(t, parallelChecks.Freeze())

	d := dagcore.NewDAG()
	assert.NoError(t, d.AddInput("A"))
	assert.NoError(t, d.AddInput("B"))

	assert.NoError(t, d.AddSubgraph("levelSubgraph", []dagcore.NodeID{"A", "B"}, levelDAG,
		func(deps map[dagcore.NodeID]any) map[dagcore.NodeID]any {
			return map[dagcore.NodeID]any{
				"input1": deps["A"],
				"input2": deps["B"],
			}
		},
		func(results map[dagcore.NodeID]any) any {
			return results["child2"]
		}))

	assert.NoError(t, d.AddSubgraph("parallelChecks", []dagcore.NodeID{"A"}, parallelChecks,
		func(deps map[dagcore.NodeID]any) map[dagcore.NodeID]any {
			return map[dagcore.NodeID]any{
				"raw": deps["A"],
			}
		},
		func(results map[dagcore.NodeID]any) any {
			return []any{
				results["check1"],
				results["check2"],
				results["check3"],
			}
		}))

	assert.NoError(t, d.AddNode("merge", []dagcore.NodeID{"levelSubgraph", "parallelChecks"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
		return fmt.Sprintf("level=%v, checks=%v", deps["levelSubgraph"], deps["parallelChecks"]), nil
	}))

	assert.NoError(t, d.Freeze())

	inst, err := d.Instantiate(map[dagcore.NodeID]any{
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

	out := ToMermaid(inst)

	expected := `graph LR
	A["A"]
	B["B"]
	subgraph levelSubgraph [Subgraph levelSubgraph]
		levelSubgraph.child1(("levelSubgraph.child1"))
		levelSubgraph.child2(("levelSubgraph.child2"))
		levelSubgraph.input1["levelSubgraph.input1"]
		levelSubgraph.input2["levelSubgraph.input2"]
		levelSubgraph.input1 --> levelSubgraph.child1
		levelSubgraph.child1 --> levelSubgraph.child2
		levelSubgraph.input2 --> levelSubgraph.child2
	end
	merge(("merge"))
	subgraph parallelChecks [Subgraph parallelChecks]
		parallelChecks.check1(("parallelChecks.check1"))
		parallelChecks.check2(("parallelChecks.check2"))
		parallelChecks.check3(("parallelChecks.check3"))
		parallelChecks.raw["parallelChecks.raw"]
		parallelChecks.raw --> parallelChecks.check1
		parallelChecks.raw --> parallelChecks.check2
		parallelChecks.raw --> parallelChecks.check3
	end
	A --> levelSubgraph
	B --> levelSubgraph
	levelSubgraph --> merge
	parallelChecks --> merge
	A --> parallelChecks
`
	assert.Equal(t, expected, out)
}
