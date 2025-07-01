package dagfunc

import (
	"context"
	"errors"
	"fmt"
	"github.com/jizhuozhi/go-future"
	"reflect"

	"github.com/jizhuozhi/go-future/dagcore"
)

var (
	ErrNotAFunction       = errors.New("dagfunc: not a function")
	ErrFuncSignature      = errors.New("dagfunc: function must have signature func(context.Context, ...) (R, error)")
	ErrMissingDependency  = errors.New("dagfunc: missing dependency for parameter type")
	ErrInputNotRegistered = errors.New("dagfunc: input type not registered")
	ErrTypeNotFound       = errors.New("dagfunc: requested type not found in results")
	ErrFrozen             = errors.New("dagfunc: frozen")
	ErrNotFrozen          = errors.New("dagfunc: not frozen")
)

// Builder constructs a type-driven DAG using function signatures.
// Each function must follow: func(ctx context.Context, A, B, ...) (R, error).
// Inputs and outputs are inferred by type via provided sample values.
type Builder struct {
	dag      *dagcore.DAG
	typeToID map[reflect.Type]dagcore.NodeID
	idToType map[dagcore.NodeID]reflect.Type
	counter  int
}

// Program represents an executable DAG instance with bound inputs.
type Program struct {
	builder   *Builder
	execution *dagcore.DAGInstance
}

// New creates a new DAG Builder instance.
func New() *Builder {
	return &Builder{
		dag:      dagcore.NewDAG(),
		typeToID: make(map[reflect.Type]dagcore.NodeID),
		idToType: make(map[dagcore.NodeID]reflect.Type),
	}
}

// Provide registers a sample value whose type will be used as an input node in the DAG.
// The type itself determines uniqueness; only one input per type is allowed.
func (b *Builder) Provide(sample any) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}

	t := reflect.TypeOf(sample)
	id := dagcore.NodeID(fmt.Sprintf("input:%s", fullTypeName(t)))
	if err := b.dag.AddInput(id); err != nil {
		return err
	}
	b.typeToID[t] = id
	b.idToType[id] = t
	return nil
}

// Use registers a computation node using a typed function.
// Function must have form: func(context.Context, A, B, ...) (R, error).
// Dependencies are inferred from parameter types beyond the context.
func (b *Builder) Use(fn any) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}

	v := reflect.ValueOf(fn)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return ErrNotAFunction
	}
	if t.NumOut() != 2 || !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return ErrFuncSignature
	}
	if t.NumIn() < 1 || t.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return ErrFuncSignature
	}

	retType := t.Out(0)
	id := dagcore.NodeID(fmt.Sprintf("func:%s", fullTypeName(retType)))

	var deps []dagcore.NodeID
	for i := 1; i < t.NumIn(); i++ {
		depType := t.In(i)
		depID, ok := b.typeToID[depType]
		if !ok {
			return fmt.Errorf("%w: %v", ErrMissingDependency, depType)
		}
		deps = append(deps, depID)
	}

	run := func(ctx context.Context, input map[dagcore.NodeID]any) (any, error) {
		args := make([]reflect.Value, 0, t.NumIn())
		args = append(args, reflect.ValueOf(ctx))
		for i := 1; i < t.NumIn(); i++ {
			depType := t.In(i)
			depID := b.typeToID[depType]
			args = append(args, reflect.ValueOf(input[depID]))
		}
		results := v.Call(args)
		if err := results[1]; !err.IsNil() {
			return nil, err.Interface().(error)
		}
		return results[0].Interface(), nil
	}

	if err := b.dag.AddNode(id, deps, run); err != nil {
		return err
	}
	b.typeToID[retType] = id
	b.idToType[id] = retType
	return nil
}

// Freeze finalizes the internal DAG structure and makes it immutable.
//
// After calling Freeze, the Builder's DAG topology is verified for completeness
// and cycles, and becomes read-only. Provide and Use will return an error if
// called after freezing. Compile requires the DAG to be frozen.
func (b *Builder) Freeze() error {
	return b.dag.Freeze()
}

// Compile finalizes the DAG with concrete input values.
// Each input must match the type of previously provided sample.
func (b *Builder) Compile(inputs []any) (*Program, error) {
	if !b.dag.Frozen() {
		return nil, ErrNotFrozen
	}

	dagInputs := make(map[dagcore.NodeID]any, len(inputs))
	for _, val := range inputs {
		typ := reflect.TypeOf(val)
		id, ok := b.typeToID[typ]
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrInputNotRegistered, typ)
		}
		dagInputs[id] = val
	}
	inst, err := b.dag.Instantiate(dagInputs)
	if err != nil {
		return nil, err
	}
	return &Program{builder: b, execution: inst}, nil
}

// Run executes the DAG and returns the results.
// The results are keyed by a zero-value sample of their output type for ergonomic lookup.
func (p *Program) Run(ctx context.Context) (map[any]any, error) {
	return p.RunAsync(ctx).Get()
}

// RunAsync executes the DAG and returns the results.
// The results are keyed by a zero-value sample of their output type for ergonomic lookup.
func (p *Program) RunAsync(ctx context.Context) *future.Future[map[any]any] {
	f := future.Then(p.execution.RunAsync(ctx), func(values map[dagcore.NodeID]any, err error) (map[any]any, error) {
		res := make(map[any]any)
		if err != nil {
			return nil, err
		}
		for id, val := range values {
			typ := p.builder.idToType[id]
			res[reflect.Zero(typ).Interface()] = val
		}
		return res, nil
	})
	return f
}

// Get returns the output value of a node that produces the given sample's type.
// The sample is only used to determine the type.
func (p *Program) Get(sample any) (any, error) {
	typ := reflect.TypeOf(sample)
	id, ok := p.builder.typeToID[typ]
	if !ok {
		return nil, ErrTypeNotFound
	}
	return p.execution.Nodes()[id].Future().Get()
}

func fullTypeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String() // e.g., builtin types like int, string
	}
	return t.PkgPath() + "." + t.Name()
}
