package dagfunc

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jizhuozhi/go-future/dagcore"
	"github.com/stretchr/testify/assert"
)

func TestOptionNameAndWrap(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))

	var wrapped []string
	var mu = make(chan struct{}, 1)
	mu <- struct{}{}

	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, error) {
		return Count(a.Value * 3), nil
	},
		Name("triple"),
		Wrap(func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
			return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
				<-mu
				wrapped = append(wrapped, string(n.ID()))
				mu <- struct{}{}
				return run(ctx, deps)
			}
		}),
	))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 2}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, []string{"triple"}, wrapped)
	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(6), got)
}

func TestOptionGroup(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Group("metrics", func(g *Builder) error {
		return g.Use(func(ctx context.Context, a InputA) (Count, error) {
			return Count(a.Value + 100), nil
		})
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 2}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	// The id is namespaced...
	_, ok := prog.NodeByID("metrics.func:github.com/jizhuozhi/go-future/dagfunc.Count")
	assert.True(t, ok)
	// ...but the type registry is shared, so the result resolves across groups.
	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(102), got)
}

func TestOptionAfter(t *testing.T) {
	b := New()
	order := make(chan string, 2)
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		time.Sleep(20 * time.Millisecond)
		order <- "slow"
		return Count(1), nil
	}))
	// A node without parameters cannot depend on anything by type; After wires
	// the ordering explicitly.
	assert.NoError(t, b.Use(func(ctx context.Context) error {
		order <- "audit"
		return nil
	}, Name("audit"), After(Count(0))))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, "slow", <-order)
	assert.Equal(t, "audit", <-order)
}

func TestOptionAfterUnknownNode(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Use(func(ctx context.Context) error { return nil }, Name("audit"), After("missing")), ErrNodeNotFound)
}

func TestOptionFrom(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, error) {
		return Count(1), nil
	}, Name("one")))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, error) {
		return Count(2), nil
	}, Name("two")))
	// Count is produced twice, so the parameter has to be pinned.
	assert.NoError(t, b.Use(func(ctx context.Context, c Count) (Answer, error) {
		return Answer(strconv.Itoa(int(c))), nil
	}, From[Count]("two")))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Answer(""))
	assert.NoError(t, err)
	assert.Equal(t, Answer("2"), got)
}

func TestOptionTimeout(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Second):
			return Count(1), nil
		}
	}, Timeout(20*time.Millisecond)))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOptionRetry(t *testing.T) {
	b := New()
	var attempts int32
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			return 0, errors.New("flaky")
		}
		return Count(7), nil
	}, RetryWith(3, time.Millisecond)))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(7), got)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestOptionRecoverAndOrElse(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		return 0, errors.New("down")
	}, Recover[Count](func(err error) (Count, error) { return Count(-1), nil })))
	assert.NoError(t, b.Use(func(ctx context.Context) (Report, error) {
		return "", errors.New("down")
	}, OrElse[Report]("fallback")))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	count, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(-1), count)

	report, err := prog.Get(Report(""))
	assert.NoError(t, err)
	assert.Equal(t, Report("fallback"), report)
}

func TestOptionFallbackOnMultiOutput(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (Count, Report, error) {
		return 0, "", nil
	}, OrElse[Count](0)), ErrInvalidOption)
}
