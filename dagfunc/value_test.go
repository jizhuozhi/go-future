//go:build go1.27

package dagfunc

import (
	"context"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

type Unregistered struct{}

// TestDagFuncGenericGet exercises the generic methods introduced with Go 1.27.
func TestDagFuncGenericGet(t *testing.T) {
	builder := New()

	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))
	assert.NoError(t, builder.Use(fnD))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 10}, InputB{Text: "hello"}})
	assert.NoError(t, err)

	// ValueAsync does not block, it can be subscribed before the DAG is started.
	fc := prog.ValueAsync[ResultC]()
	fd := prog.ValueAsync[ResultD]()

	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	c, err := fc.Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 15}, c)

	d, err := fd.Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultD{Message: "Sum is \x0f"}, d)

	// After the DAG completed, Value returns the very same values.
	c2, err := prog.Value[ResultC]()
	assert.NoError(t, err)
	assert.Equal(t, c, c2)

	// An unregistered type is reported eagerly, without executing anything.
	_, err = prog.ValueAsync[Unregistered]().Get()
	assert.ErrorIs(t, err, ErrTypeNotFound)

	// Casting a node result to an unrelated type fails with ErrTypeMismatch.
	_, err = prog.ValueAsync[ResultC]().Cast[ResultD]().Get()
	assert.ErrorIs(t, err, future.ErrTypeMismatch)
}

func TestDagFuncGenericAmbiguous(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(fnCFromA, Name("slow")))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA) (ResultC, error) {
		return ResultC{Sum: 1}, nil
	}, Name("fast")))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 1}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	// Two nodes produce ResultC, so the type alone is not enough.
	_, err = prog.Value[ResultC]()
	assert.ErrorIs(t, err, ErrAmbiguousType)

	node, ok := prog.NodeByID("fast")
	assert.True(t, ok)
	got, err := node.Cast[ResultC]().Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 1}, got)
}

func TestDagFuncGenericMultiOutput(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA) (Count, Report, error) {
		return Count(a.Value * 2), Report("done"), nil
	}))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 2}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	count, err := prog.Value[Count]()
	assert.NoError(t, err)
	assert.Equal(t, Count(4), count)

	report, err := prog.Value[Report]()
	assert.NoError(t, err)
	assert.Equal(t, Report("done"), report)
}

func TestDagFuncGenericSubgraph(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile([]any{Question("hello")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	candidate, err := prog.Value[Candidate]()
	assert.NoError(t, err)
	assert.Equal(t, Candidate("candidate:hello"), candidate)
}
