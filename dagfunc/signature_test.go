package dagfunc

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

func typeNameOf(v any) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return ""
	}
	return fullTypeName(t)
}

type Count int

type Report string

type Answer string

// Docs is a slice on purpose: it is not comparable and therefore cannot be a
// key of the map returned by Run.
type Docs []string

func TestSignatureWithoutContextAndError(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(a InputA) Count { return Count(a.Value * 2) }))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 4}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(8), got)
}

func TestSignatureSource(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) { return Count(42), nil }))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(42), got)
}

func TestSignatureSink(t *testing.T) {
	b := New()
	var seen Count
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(a InputA) Count { return Count(a.Value) }))
	assert.NoError(t, b.Use(func(ctx context.Context, c Count) error {
		seen = c
		return nil
	}, Name("audit")))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 5}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, Count(5), seen)

	// A sink produces no value, but it is still a node with an id.
	node, ok := prog.NodeByID("audit")
	assert.True(t, ok)
	v, err := node.Future().Get()
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestSignatureSinkError(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) error {
		return errors.New("audit failed")
	}, Name("audit")))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit failed")
}

func TestSignatureMultiOutput(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	// Two results plus an error, no dependency on an earlier Count.
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, Report, error) {
		return Count(a.Value), Report("n=" + strconv.Itoa(a.Value)), nil
	}))
	assert.NoError(t, b.Use(func(ctx context.Context, c Count, r Report) (Answer, error) {
		return Answer(string(r) + ":" + strconv.Itoa(int(c))), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 3}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	c, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(3), c)

	r, err := prog.Get(Report(""))
	assert.NoError(t, err)
	assert.Equal(t, Report("n=3"), r)

	a, err := prog.Get(Answer(""))
	assert.NoError(t, err)
	assert.Equal(t, Answer("n=3:3"), a)
}

// MyError implements error but is not declared as error, so dagfunc treats it
// as an ordinary value: the signature decides, not the method set.
type MyError struct {
	Msg string
}

func (e MyError) Error() string { return e.Msg }

func TestSignatureErrorAliasIsAValue(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	// (Count, MyError): two results, no error at all.
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, MyError) {
		if a.Value == 0 {
			return 0, MyError{Msg: "zero"}
		}
		return Count(a.Value), MyError{}
	}))
	assert.NoError(t, b.Use(func(ctx context.Context, e MyError) (Report, error) {
		return Report("err=" + e.Msg), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 0}})
	assert.NoError(t, err)
	// The node never fails, MyError is just another value on the wire.
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	count, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(0), count)

	report, err := prog.Get(Report(""))
	assert.NoError(t, err)
	assert.Equal(t, Report("err=zero"), report)
}

func TestSignatureErrorMustBeLastAndUnique(t *testing.T) {
	// Two errors: there is no way to tell which one fails the node.
	b := New()
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (error, error) { return nil, nil }), ErrFuncSignature)
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (Count, error, error) { return 0, nil, nil }), ErrFuncSignature)
	// error declared in front of a value is not an error either.
	assert.ErrorIs(t, b.Use(func(ctx context.Context) (error, Count) { return nil, 0 }), ErrFuncSignature)
	// One trailing error is fine.
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, error) { return 0, nil }))
}

func TestSignatureFailedNodePublishesNothing(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (Count, Report, error) {
		return Count(1), Report("half"), errors.New("failed")
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.Error(t, err)

	node, err := prog.Node(Count(0))
	assert.NoError(t, err)
	val, err := node.Future().Get()
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestSignatureMultiOutputWithoutError(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Count, Report) {
		return Count(a.Value + 1), Report("r")
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 1}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	c, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(2), c)
}

func TestSignatureFutureResult(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) *future.Future[Count] {
		return future.Async(func() (Count, error) { return Count(a.Value + 1), nil })
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 9}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(Count(0))
	assert.NoError(t, err)
	assert.Equal(t, Count(10), got)
}

func TestSignatureFutureResultError(t *testing.T) {
	b := New()
	assert.NoError(t, b.Use(func(ctx context.Context) (*future.Future[Count], error) {
		return future.Done2[Count](0, errors.New("boom")), nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestSignatureNonComparableOutput(t *testing.T) {
	b := New()
	assert.NoError(t, b.Provide(InputA{}))
	assert.NoError(t, b.Use(func(ctx context.Context, a InputA) (Docs, error) {
		return Docs{"doc:" + strconv.Itoa(a.Value)}, nil
	}))
	assert.NoError(t, b.Freeze())

	prog, err := b.Compile([]any{InputA{Value: 1}})
	assert.NoError(t, err)
	results, err := prog.Run(context.Background())
	assert.NoError(t, err)

	// Docs is a slice and is skipped by the legacy map, which needs comparable
	// keys, but it is still readable by type.
	for k := range results {
		assert.NotEqual(t, "github.com/jizhuozhi/go-future/dagfunc.Docs", typeNameOf(k))
	}
	docs, err := prog.Get(Docs{})
	assert.NoError(t, err)
	assert.Equal(t, Docs{"doc:1"}, docs)
}
