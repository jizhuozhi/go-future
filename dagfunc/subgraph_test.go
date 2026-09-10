package dagfunc

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Question string

type Candidate string

type Score int

type Reply string

func retrieveCandidate(ctx context.Context, q Question) (Candidate, error) {
	return Candidate("candidate:" + string(q)), nil
}

func TestSubgraphSingleOutput(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))))
	assert.NoError(t, root.Use(func(ctx context.Context, c Candidate) (Reply, error) {
		return Reply("reply:" + string(c)), nil
	}))
	assert.NoError(t, root.Freeze())
	// The parent freezes the subgraph for us.
	assert.True(t, sub.Frozen())

	prog, err := root.Compile([]any{Question("hello")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	reply, err := prog.Get(Reply(""))
	assert.NoError(t, err)
	assert.Equal(t, Reply("reply:candidate:hello"), reply)

	node, ok := prog.NodeByID("qa")
	assert.True(t, ok)
	assert.NotNil(t, node.Subgraph())
}

func TestSubgraphMultiOutput(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))
	assert.NoError(t, sub.Use(func(ctx context.Context, q Question, c Candidate) (Score, error) {
		return Score(len(q) + len(c)), nil
	}))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""), Score(0))))
	assert.NoError(t, root.Use(func(ctx context.Context, c Candidate, s Score) (Reply, error) {
		return Reply(string(c) + ":" + strconv.Itoa(int(s))), nil
	}))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile([]any{Question("hi")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	candidate, err := prog.Get(Candidate(""))
	assert.NoError(t, err)
	assert.Equal(t, Candidate("candidate:hi"), candidate)

	score, err := prog.Get(Score(0))
	assert.NoError(t, err)
	assert.Equal(t, Score(14), score)

	reply, err := prog.Get(Reply(""))
	assert.NoError(t, err)
	assert.Equal(t, Reply("candidate:hi:14"), reply)
}

func TestSubgraphNested(t *testing.T) {
	inner := New()
	assert.NoError(t, inner.Provide(Question("")))
	assert.NoError(t, inner.Use(retrieveCandidate))

	middle := New()
	assert.NoError(t, middle.Provide(Question("")))
	assert.NoError(t, middle.Subgraph(inner, Name("retrieve"), Outputs(Candidate(""))))
	assert.NoError(t, middle.Use(func(ctx context.Context, q Question, c Candidate) (Score, error) {
		return Score(len(q) * len(c)), nil
	}))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	assert.NoError(t, root.Subgraph(middle, Name("pipeline"), Outputs(Score(0))))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile([]any{Question("abc")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	score, err := prog.Get(Score(0))
	assert.NoError(t, err)
	// len("abc") * len("candidate:abc")
	assert.Equal(t, Score(39), score)
}

func TestSubgraphDefaultInput(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question(""), Default(Question("fallback"))))
	assert.NoError(t, sub.Use(retrieveCandidate))

	root := New()
	// Question is not produced anywhere in the root, the default is used.
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	candidate, err := prog.Get(Candidate(""))
	assert.NoError(t, err)
	assert.Equal(t, Candidate("candidate:fallback"), candidate)
}

func TestSubgraphErrors(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))

	// A subgraph has no result type to derive an id from.
	root := New()
	assert.ErrorIs(t, root.Subgraph(sub), ErrSubgraphName)
	assert.ErrorIs(t, root.Subgraph(sub, Outputs(Candidate(""))), ErrSubgraphName)

	// Without Outputs the parent cannot know what the node produces.
	root = New()
	assert.NoError(t, root.Subgraph(sub, Name("qa")))
	assert.ErrorIs(t, root.Freeze(), ErrSubgraphOutputs)

	// An input of the subgraph that neither the parent nor a default provides.
	root = New()
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))))
	assert.ErrorIs(t, root.Freeze(), ErrMissingDependency)
}
