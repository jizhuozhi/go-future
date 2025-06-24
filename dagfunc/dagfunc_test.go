package dagfunc

import (
	"context"
	"errors"
	"testing"
	"time"

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

	cVal2, err := inst.Get(ResultC{})
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 15}, cVal2.(ResultC))

	dVal2, err := inst.Get(ResultD{})
	assert.NoError(t, err)
	assert.Equal(t, ResultD{Message: "Sum is \x0f"}, dVal2.(ResultD))
}

func TestDagFuncFlowError(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA, b InputB) (ResultC, error) {
		return ResultC{}, errors.New("fault")
	}))
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
