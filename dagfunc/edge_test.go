package dagfunc

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/jizhuozhi/go-future/dagcore"
	"github.com/stretchr/testify/assert"
)

// This file covers the paths that are only reachable with unusual input:
// rejected registrations, rejected bindings, and the branches of the built-in
// options that a happy-path run never takes.

func TestProvideRejectsNilSample(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Provide(nil), ErrInputNotRegistered)
}

func TestProvideRejectsForeignDefault(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Provide(Count(0), Default("not a count")), ErrInvalidOption)
}

func TestGroupRejectsFrozenBuilder(t *testing.T) {
	b := New()
	// A nil definition is a no-op, it is not an error.
	assert.NoError(t, b.Group("g", nil))

	assert.NoError(t, b.Freeze())
	assert.ErrorIs(t, b.Group("g", func(*Builder) error { return nil }), ErrFrozen)
	assert.ErrorIs(t, b.Group("g", nil), ErrFrozen)
}

func TestUseRejectsNilFunction(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Use(nil), ErrNotAFunction)
}

func TestFromRejectsUnknownBinding(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, error) {
		return Count(a.Value), nil
	}, Name("producer")))

	// The node does not exist.
	err := b.Use(func(ctx context.Context, a InputA) (Report, error) {
		return "", nil
	}, From[InputA]("missing"))
	assert.ErrorIs(t, err, ErrNodeNotFound)

	// The node exists but does not produce the bound type.
	err = b.Use(func(ctx context.Context, a InputA) (Answer, error) {
		return "", nil
	}, From[InputA]("producer"))
	assert.ErrorIs(t, err, ErrTypeNotFound)
}

func TestAfterAcceptsEveryForm(t *testing.T) {
	b := New()
	countID := dagcore.NodeID("func:github.com/jizhuozhi/go-future/dagfunc.Count")
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) { return Count(1), nil }))
	// dagcore.NodeID is accepted next to plain strings.
	assert.NoError(t, b.Use(func(ctx context.Context) error { return nil }, Name("audit"), After(countID)))
	// A nil element is rejected instead of being ignored.
	assert.ErrorIs(t, b.Use(func(ctx context.Context) error { return nil }, Name("other"), After(nil)), ErrNodeNotFound)
}

func TestCompileRejectsBadBindings(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}, Name("first")))
	assert.NoError(t, b.Use(fnCFromA))
	assert.NoError(t, b.Freeze())

	// The id is not an input node.
	_, err := b.Compile([]any{Input("func:github.com/jizhuozhi/go-future/dagfunc.ResultC", ResultC{})})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
	// The same input is bound twice.
	_, err = b.Compile([]any{Input("first", InputA{}), Input("first", InputA{})})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
	// The value does not match the type of the input node.
	_, err = b.Compile([]any{Input("first", "not an InputA")})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
	// A nil value carries no type at all.
	_, err = b.Compile([]any{nil})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
}

func TestProgramRejectsNilSample(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(fnCFromA))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)

	_, err = prog.Get(nil)
	assert.ErrorIs(t, err, ErrTypeNotFound)
	_, err = prog.Node(nil)
	assert.ErrorIs(t, err, ErrTypeNotFound)
}

func TestBuiltinTypesUseShortIDs(t *testing.T) {
	b := New()
	// string has no package path, so its id is just the type name.
	assert.NoError(t, b.Provide(""))
	assert.NoError(t, b.Use(func(ctx context.Context, s string) (Count, error) {
		return Count(len(s)), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{"hello"})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	_, ok := prog.NodeByID("input:string")
	assert.True(t, ok)
	_, ok = prog.NodeByID("func:github.com/jizhuozhi/go-future/dagfunc.Count")
	assert.True(t, ok)
}

func TestSinkNodeDerivesIDFromSymbol(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	// No Name: the id comes from the runtime name of the function.
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) error { return nil }))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	found := ""
	for id := range prog.Instance().Nodes() {
		if strings.HasPrefix(string(id), "func:github.com/jizhuozhi/go-future/dagfunc.TestSinkNodeDerivesIDFromSymbol") {
			found = string(id)
		}
	}
	assert.NotEmpty(t, found)
}

func TestNilDependencyValueIsPassedThrough(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Docs, error) { return nil, nil }))
	assert.NoError(t, b.Use(func(ctx context.Context, d Docs) (Count, error) { return Count(len(d)), nil }))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(0), got)
}

func TestPointerResultIsNotAFuture(t *testing.T) {
	b := New()
	// A pointer that is not a *future.Future is an ordinary value.
	assert.NoError(t, b.Use(func(ctx context.Context) (*Count, error) {
		c := Count(3)
		return &c, nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	zero := Count(0)
	got, err := prog.Get(&zero)
	assert.NoError(t, err)
	want := Count(3)
	assert.Equal(t, &want, got)
}

func TestNilFutureResultProducesNil(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (*future.Future[Count], error) { return nil, nil }))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	node, err := prog.Node(Count(0))
	assert.NoError(t, err)
	got, err := node.Future().Get()
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestFallbackRejectsForeignType(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (Count, error) {
		return 0, nil
	}, Recover[Report](func(err error) (Report, error) { return "", nil })), ErrInvalidOption)
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (Count, error) {
		return 0, nil
	}, OrElse[Report]("")), ErrInvalidOption)
}

func TestFallbackPassesSuccessThrough(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		return Count(1), nil
	}, OrElse[Count](Count(9))))
	assert.NoError(t, b.Use(func(ctx context.Context) (Report, error) {
		return Report("ok"), nil
	}, Recover[Report](func(err error) (Report, error) { return Report("recovered"), nil })))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	count, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(1), count)

	report, err := prog.Get(Report(""))
	assert.NoError(t, err)
	assert.Equal(t, Report("ok"), report)
}

func TestRecoverMayReturnAnotherError(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		return 0, errors.New("down")
	}, Recover[Count](func(err error) (Count, error) {
		return 0, errors.New("worse")
	})))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worse")
}

func TestRetryStopsOnCanceledContext(t *testing.T) {
	b := New()
	var attempts int32
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		atomic.AddInt32(&attempts, 1)
		return 0, errors.New("always fails")
	}, Retry(3)))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = prog.Run(ctx)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int32(0), atomic.LoadInt32(&attempts))
}

func TestRetryGivesUpAfterLastAttempt(t *testing.T) {
	b := New()
	var attempts int32
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		atomic.AddInt32(&attempts, 1)
		return Count(0), errors.New("nope")
	}, Retry(1)))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestSubgraphRegistrationRejects(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	// No builder to embed.
	assert.ErrorIs(t, root.Subgraph(nil, Name("qa")), ErrNodeNotFound)
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))))
	// The id is taken.
	assert.ErrorIs(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate(""))), ErrNodeExisted)
	// The graph is frozen.
	assert.NoError(t, root.Freeze())
	assert.ErrorIs(t, root.Subgraph(sub, Name("other"), Outputs(Candidate(""))), ErrFrozen)

	// The subgraph does not produce the declared output.
	root = New()
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Reply(""))))
	assert.ErrorIs(t, root.Freeze(), ErrTypeNotFound)

	// An ordering dependency that does not exist.
	root = New()
	assert.NoError(t, root.Provide(Question("")))
	assert.NoError(t, root.Subgraph(sub, Name("qa"), Outputs(Candidate("")), After("missing")))
	assert.ErrorIs(t, root.Freeze(), ErrNodeNotFound)
}

func TestSubgraphWrap(t *testing.T) {
	sub := New()
	assert.NoError(t, sub.Provide(Question("")))
	assert.NoError(t, sub.Use(retrieveCandidate))

	root := New()
	assert.NoError(t, root.Provide(Question("")))
	wrapped := make([]string, 0, 1)
	assert.NoError(t, root.Subgraph(sub,
		Name("qa"),
		Outputs(Candidate("")),
		Wrap(func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
			return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
				wrapped = append(wrapped, string(n.ID()))
				return run(ctx, deps)
			}
		}),
	))
	assert.NoError(t, root.Freeze())

	prog, err := root.Compile([]any{Question("hi")})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"qa"}, wrapped)
}
