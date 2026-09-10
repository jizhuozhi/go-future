package dagviz

import (
	"context"
	"testing"

	"github.com/jizhuozhi/go-future/dagfunc"
	"github.com/stretchr/testify/assert"
)

type Query string

type Candidate string

type Reply string

// TestMermaidFromDagFunc renders a graph built with dagfunc, including a
// subgraph, which is what Program.Instance hands over to dagviz.
func TestMermaidFromDagFunc(t *testing.T) {
	sub := dagfunc.New()
	assert.NoError(t, sub.Provide(Query("")))
	assert.NoError(t, sub.Use(func(ctx context.Context, q Query) (Candidate, error) {
		return Candidate("c:" + string(q)), nil
	}))

	root := dagfunc.New()
	assert.NoError(t, root.Provide(Query("")))
	assert.NoError(t, root.Subgraph(sub, dagfunc.Name("qa"), dagfunc.Outputs(Candidate(""))))
	assert.NoError(t, root.Use(func(ctx context.Context, c Candidate) (Reply, error) {
		return Reply("r:" + string(c)), nil
	}))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile([]any{Query("hi")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	out := ToMermaid(prog.Instance())
	assert.Contains(t, out, "subgraph qa [Subgraph qa]")
	assert.Contains(t, out, "qa --> func:github.com/jizhuozhi/go-future/dagviz.Reply")
	assert.Contains(t, out, "input:github.com/jizhuozhi/go-future/dagviz.Query --> qa")
}
