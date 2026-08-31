package dagfunc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

type InputA struct {
	Value int
}

type InputB struct {
	Text string
}

type ResultC struct {
	Sum int
}

type ResultD struct {
	Message string
}

func fnC(ctx context.Context, a InputA, b InputB) (ResultC, error) {
	return ResultC{Sum: a.Value + len(b.Text)}, nil
}

func fnD(ctx context.Context, c ResultC) (ResultD, error) {
	if c.Sum == 0 {
		return ResultD{}, errors.New("sum cannot be zero")
	}
	time.Sleep(10 * time.Millisecond) // 模拟延时
	return ResultD{Message: "Sum is " + string(rune(c.Sum))}, nil
}

func TestDagFuncFlow(t *testing.T) {
	builder := New()

	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))
	assert.NoError(t, builder.Use(fnD))

	assert.NoError(t, builder.Freeze())
	inst, err := builder.Compile([]any{InputA{Value: 10}, InputB{Text: "hello"}})
	assert.NoError(t, err)

	ctx := context.Background()
	results, err := inst.Run(ctx)
	assert.NoError(t, err)

	cVal, ok := results[ResultC{}]
	assert.True(t, ok)
	assert.IsType(t, ResultC{}, cVal)
	assert.Equal(t, cVal, ResultC{Sum: 15})

	dVal, ok := results[ResultD{}]
	assert.True(t, ok)
	assert.IsType(t, ResultD{}, dVal)
	assert.Equal(t, dVal, ResultD{Message: "Sum is \x0f"})

	// Get(sample) is the 0.1.x API and stays supported.
	cVal2, err := inst.Get(ResultC{})
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 15}, cVal2.(ResultC))

	dVal2, err := inst.Get(ResultD{})
	assert.NoError(t, err)
	assert.Equal(t, ResultD{Message: "Sum is \x0f"}, dVal2.(ResultD))

	// Value[T] is the type safe form.
	cVal3, err := inst.Value[ResultC]()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 15}, cVal3)
}

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

func TestDagFuncFlowError(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA, b InputB) (ResultC, error) {
		return ResultC{}, errors.New("fault")
	}))
	assert.NoError(t, builder.Freeze())
	prog, err := builder.Compile([]any{InputA{}, InputB{}})
	assert.NoError(t, err)
	ctx := context.Background()
	_, err = prog.Run(ctx)
	assert.Error(t, err)
}

func TestDAGFuncInvalidProvide(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.Error(t, builder.Provide(InputA{}))
}

func TestDAGFuncInvalidUse(t *testing.T) {
	builder := New()
	assert.ErrorIs(t, ErrNotAFunction, builder.Use(InputA{}))
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))
	assert.ErrorIs(t, ErrFuncSignature, builder.Use(func() {}))
	assert.ErrorIs(t, ErrFuncSignature, builder.Use(func() (int, error) { return 0, nil }))
	assert.ErrorIs(t, ErrMissingDependency, errors.Unwrap(builder.Use(func(ctx context.Context, _ string) (int, error) { return 0, nil })))
	assert.Error(t, builder.Use(fnC))
}
