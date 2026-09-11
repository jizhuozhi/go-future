package dagfunc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jizhuozhi/go-future/dagcore"
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

func fnCFromA(ctx context.Context, a InputA) (ResultC, error) {
	return ResultC{Sum: a.Value + 5}, nil
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

	// Inputs are readable too.
	aVal, err := inst.Get(InputA{})
	assert.NoError(t, err)
	assert.Equal(t, InputA{Value: 10}, aVal)
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
	assert.Contains(t, err.Error(), "fault")

	_, err = prog.Get(ResultC{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fault")
}

func TestDAGFuncInvalidProvide(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.ErrorIs(t, builder.Provide(InputA{}), ErrNodeExisted)
	// A type already produced by a node cannot be declared as an input.
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))
	assert.ErrorIs(t, builder.Provide(ResultC{}), ErrNodeExisted)
}

func TestDAGFuncInvalidUse(t *testing.T) {
	builder := New()
	assert.ErrorIs(t, builder.Use(InputA{}), ErrNotAFunction)
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))

	// No result at all.
	assert.ErrorIs(t, builder.Use(func() {}), ErrFuncSignature)
	// Variadic parameters cannot be resolved by type.
	assert.ErrorIs(t, builder.Use(func(ctx context.Context, _ ...int) (int, error) { return 0, nil }), ErrFuncSignature)
	// context.Context must come first.
	assert.ErrorIs(t, builder.Use(func(_ InputA, ctx context.Context) (int, error) { return 0, nil }), ErrFuncSignature)
	// Unknown parameter type.
	err := builder.Use(func(ctx context.Context, _ string) (int, error) { return 0, nil })
	assert.ErrorIs(t, errors.Unwrap(err), ErrMissingDependency)
	// The default id is taken by fnC already.
	assert.ErrorIs(t, builder.Use(fnC), ErrNodeExisted)
}

func TestDAGFuncFreeze(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	_, err := builder.Compile(nil)
	assert.ErrorIs(t, err, ErrNotFrozen)

	assert.NoError(t, builder.Freeze())
	assert.ErrorIs(t, builder.Freeze(), ErrFrozen)
	assert.True(t, builder.Frozen())
	assert.ErrorIs(t, builder.Provide(InputB{}), ErrFrozen)
	assert.ErrorIs(t, builder.Use(fnCFromA), ErrFrozen)
}

func TestDAGFuncDefaultInput(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}, Default(InputA{Value: 7})))
	assert.NoError(t, builder.Use(fnCFromA))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile(nil)
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(ResultC{})
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 12}, got)
}

func TestDAGFuncMissingInput(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(fnCFromA))
	assert.NoError(t, builder.Freeze())

	_, err := builder.Compile(nil)
	assert.ErrorIs(t, err, ErrMissingInput)

	_, err = builder.Compile([]any{InputB{}})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
}

func TestDAGFuncInputBinding(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}, Name("first")))
	assert.NoError(t, builder.Provide(InputA{}, Name("second")))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA) (ResultC, error) {
		return ResultC{Sum: a.Value * 2}, nil
	}, From[InputA]("second")))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{
		Input("first", InputA{Value: 1}),
		Input("second", InputA{Value: 10}),
	})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	got, err := prog.Get(ResultC{})
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 20}, got)

	// An unknown node id is rejected.
	_, err = builder.Compile([]any{Input("missing", InputA{})})
	assert.ErrorIs(t, err, ErrInputNotRegistered)
}

func TestDAGFuncAmbiguousType(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(fnCFromA, Name("slow")))
	assert.NoError(t, builder.Use(func(ctx context.Context, a InputA) (ResultC, error) {
		return ResultC{Sum: a.Value}, nil
	}, Name("fast")))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 3}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	_, err = prog.Get(ResultC{})
	assert.ErrorIs(t, err, ErrAmbiguousType)

	// By id the two results are still reachable.
	slow, ok := prog.NodeByID("slow")
	assert.True(t, ok)
	v, err := slow.Future().Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 8}, v)

	fast, ok := prog.NodeByID("fast")
	assert.True(t, ok)
	v, err = fast.Future().Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 3}, v)
}

func TestDAGFuncNodeLookup(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(fnCFromA))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 1}})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	node, err := prog.Node(ResultC{})
	assert.NoError(t, err)
	assert.Equal(t, "func:github.com/jizhuozhi/go-future/dagfunc.ResultC", string(node.ID()))
	assert.Equal(t, ResultC{Sum: 6}, node.Future().GetOrDefault(ResultC{}))
	assert.GreaterOrEqual(t, node.Duration(), time.Duration(0))

	_, err = prog.Node(ResultD{})
	assert.ErrorIs(t, err, ErrTypeNotFound)

	_, ok := prog.NodeByID("missing")
	assert.False(t, ok)

	input, ok := prog.NodeByID("input:github.com/jizhuozhi/go-future/dagfunc.InputA")
	assert.True(t, ok)
	iv, err := input.Future().Get()
	assert.NoError(t, err)
	assert.Equal(t, InputA{Value: 1}, iv)
	assert.Same(t, prog.execution, prog.Instance())
}

func TestDAGFuncCompileWrappers(t *testing.T) {
	builder := New()
	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Use(fnCFromA))
	assert.NoError(t, builder.Freeze())

	seen := make(map[dagcore.NodeID]int)
	prog, err := builder.Compile([]any{InputA{Value: 1}}, func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
		return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
			seen[n.ID()]++
			return run(ctx, deps)
		}
	})
	assert.NoError(t, err)
	_, err = prog.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, seen["func:github.com/jizhuozhi/go-future/dagfunc.ResultC"])
}
