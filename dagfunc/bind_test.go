package dagfunc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

type NotProduced struct{}

func TestFromSelectsOneResultOfAMultiOutputNode(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, Report, error) {
		return Count(a.Value), Report("from pair"), nil
	}, Name("pair")))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Report, error) {
		return Report("from solo"), nil
	}, Name("solo")))
	// Report is produced twice, and From pins the one of the multi-output node.
	assert.NoError(t, b.Use(func(ctx context.Context, r Report) (Answer, error) {
		return Answer(r), nil
	}, From[Report]("pair")))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 1}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Answer(""))
	assert.NoError(t, err)
	assert.Equal(t, Answer("from pair"), got)
}

func TestCompileRejectsAmbiguousAndDuplicateInputs(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}, Name("first")))
	assert.NoError(t, b.Provide(InputA{}, Name("second")))
	assert.NoError(t, b.Freeze())

	// Two inputs share the type, so it has to be bound by id.
	_, err := b.Compile([]any{InputA{}})
	assert.ErrorIs(t, err, ErrAmbiguousType)

	// The same type is supplied twice.
	single := New()
	assert.NoError(t, single.Provide(InputA{}))
	assert.NoError(t, single.Use(fnCFromA))
	assert.NoError(t, single.Freeze())
	_, err = single.Compile([]any{InputA{Value: 1}, InputA{Value: 2}})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
}

func TestAfterRejectsUnknownType(t *testing.T) {
	b := New()
	assert.ErrorIs(t, b.Use(func(ctx context.Context) error { return nil },
		Name("audit"), After(NotProduced{})), ErrTypeNotFound)
}

func TestRetryStopsWhenContextExpiresDuringBackoff(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) {
		return 0, errors.New("always fails")
	}, RetryWith(3, 200*time.Millisecond)))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = prog.Run(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPromiseResultIsNotAFuture(t *testing.T) {
	b := New()
	// Same package as future.Future, different name: an ordinary value.
	assert.NoError(t, b.Use(func(ctx context.Context) (*future.Promise[Count], error) {
		return future.NewPromise[Count](), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	p, err := prog.Get(future.NewPromise[Count]())
	assert.NoError(t, err)
	assert.IsType(t, &future.Promise[Count]{}, p)
}

func TestInterfaceDependencyMayBeNil(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (any, error) { return nil, nil }))
	assert.NoError(t, b.Use(func(ctx context.Context, v any) (Count, error) {
		if v == nil {
			return Count(-1), nil
		}
		return Count(1), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(-1), got)
}
