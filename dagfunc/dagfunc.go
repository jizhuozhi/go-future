package dagfunc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/jizhuozhi/go-future"
	"github.com/jizhuozhi/go-future/dagcore"
)

var (
	// ErrNotAFunction is returned by Use when the registered value is not a function.
	ErrNotAFunction = errors.New("dagfunc: not a function")

	// ErrFuncSignature is returned by Use when the function signature cannot be
	// interpreted as a node. See the package documentation for the accepted forms.
	ErrFuncSignature = errors.New("dagfunc: unsupported function signature")

	// ErrMissingDependency is returned when a parameter type has no producer.
	ErrMissingDependency = errors.New("dagfunc: missing dependency for parameter type")

	// ErrInputNotRegistered is returned by Compile for an input value whose type
	// (or id) was never declared with Provide.
	ErrInputNotRegistered = errors.New("dagfunc: input type not registered")

	// ErrTypeNotFound is returned when no node produces the requested type.
	ErrTypeNotFound = errors.New("dagfunc: requested type not found in results")

	// ErrFrozen is returned when a Builder is modified after Freeze.
	ErrFrozen = errors.New("dagfunc: frozen")

	// ErrNotFrozen is returned by Compile when the Builder was not frozen.
	ErrNotFrozen = errors.New("dagfunc: not frozen")

	// ErrNodeExisted is returned when a node id or an input type is declared twice.
	ErrNodeExisted = errors.New("dagfunc: node already exists")

	// ErrNodeNotFound is returned when a node id does not exist.
	ErrNodeNotFound = errors.New("dagfunc: node not found")

	// ErrAmbiguousType is returned when several nodes produce the requested
	// type, so it cannot be resolved by type alone.
	ErrAmbiguousType = errors.New("dagfunc: type is produced by more than one node")

	// ErrMissingInput is returned by Compile when a declared input has neither a
	// supplied value nor a default.
	ErrMissingInput = errors.New("dagfunc: missing input value")

	// ErrInvalidOption is returned when an Option does not apply to the node it
	// is passed to, e.g. Recover on a node with two outputs.
	ErrInvalidOption = errors.New("dagfunc: option does not apply to this node")

	// ErrSubgraphName is returned by Subgraph when no Name is given; a subgraph
	// node has no return type to derive an id from.
	ErrSubgraphName = errors.New("dagfunc: subgraph requires a name")

	// ErrSubgraphOutputs is returned by Freeze when a subgraph does not declare
	// which of its results are exposed to the parent graph.
	ErrSubgraphOutputs = errors.New("dagfunc: subgraph requires Outputs")
)

// Option configures a node at registration time. Options are accepted by
// Provide, Use and Subgraph; an option that does not apply to the node at hand
// is reported with ErrInvalidOption or ignored.
type Option func(*nodeOptions)

type nodeOptions struct {
	name     string
	after    []any
	binds    map[reflect.Type]dagcore.NodeID
	wrappers []dagcore.NodeFuncWrapper
	outputs  []reflect.Type

	timeout time.Duration
	retries int
	backoff time.Duration

	recover   func(err error) (any, error)
	recoverT  reflect.Type
	orElse    any
	orElseT   reflect.Type
	hasOrElse bool

	defValue any
	hasDef   bool
}

func newOptions(opts []Option) *nodeOptions {
	o := &nodeOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}

// outputRef points at one of the values produced by a node. index is -1 when
// the node produces a single value, otherwise it is the position inside the
// []any carried by a multi-output node.
type outputRef struct {
	node  dagcore.NodeID
	index int
}

// depRef is a resolved dependency: which node to wait for and where to read
// the value from. typ is nil for an ordering-only dependency.
type depRef struct {
	id    dagcore.NodeID
	index int
	typ   reflect.Type
}

type nodeDef struct {
	id    dagcore.NodeID
	deps  []depRef
	outs  []reflect.Type
	input bool
}

func (d *nodeDef) indexOf(t reflect.Type) int {
	for i, o := range d.outs {
		if o == t {
			return i
		}
	}
	return -1
}

func (d *nodeDef) refIndex(i int) int {
	if len(d.outs) <= 1 {
		return -1
	}
	return i
}

// Builder constructs a type-driven DAG from Go functions.
//
// Nodes are wired by type: Provide declares an input of some type, Use registers
// a function whose parameters are resolved to the nodes producing those types
// and whose results become available under their own types.
//
// A Builder is not safe for concurrent modification; a frozen Builder is
// immutable and can be compiled and executed from many goroutines at once.
type Builder struct {
	dag *dagcore.DAG

	// ns is prepended to the id of every node registered through this Builder.
	// Group derives a child Builder with a longer prefix, which keeps ids unique
	// without changing any resolution rule: the type registry is shared.
	ns string

	nodes    map[dagcore.NodeID]*nodeDef
	producer map[reflect.Type][]outputRef
	defaults map[reflect.Type]any
	wrappers map[dagcore.NodeID][]dagcore.NodeFuncWrapper

	// inputs indexes the input nodes by type, inputList keeps them in
	// registration order so that Compile and subgraph wiring are deterministic.
	// Several input nodes may share a type as long as they carry a Name; such a
	// type can then only be read or supplied through its node id.
	inputs    map[reflect.Type][]dagcore.NodeID
	inputList []inputRef

	subgraphs []*subgraphDef
}

type inputRef struct {
	id  dagcore.NodeID
	typ reflect.Type
}

// New creates an empty Builder.
func New() *Builder {
	return &Builder{
		dag:      dagcore.NewDAG(),
		nodes:    make(map[dagcore.NodeID]*nodeDef),
		producer: make(map[reflect.Type][]outputRef),
		inputs:   make(map[reflect.Type][]dagcore.NodeID),
		defaults: make(map[reflect.Type]any),
		wrappers: make(map[dagcore.NodeID][]dagcore.NodeFuncWrapper),
	}
}

// Group registers every node of define under the namespace ns.
//
// Group only affects node ids, it is not a subgraph: the type registry is
// shared with the parent, so a node inside the group may depend on a type
// produced outside of it and the other way round. Runtime behaviour, parallel
// scheduling and result lookup are exactly as if the nodes had been registered
// on the parent directly.
//
//	_ = b.Group("profile", func(g *dagfunc.Builder) error {
//	    _ = g.Use(loadProfile)
//	    return nil
//	})
func (b *Builder) Group(ns string, define func(*Builder) error) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}
	if define == nil {
		return nil
	}
	// Shallow copy: the DAG and every registry map are shared with the parent,
	// only the id prefix differs.
	sub := *b
	if ns != "" {
		sub.ns = b.join(ns)
	}
	return define(&sub)
}

// join prefixes a node name with the namespace of the Builder. Callers always
// pass a non-empty name: a node without an explicit one falls back to its type
// derived default first.
func (b *Builder) join(name string) string {
	if b.ns == "" {
		return name
	}
	return b.ns + "." + name
}

// Provide declares an input node of the type of sample.
//
// The value itself is supplied at Compile time; sample only carries the type.
// With Default the input becomes optional.
//
// Two inputs may share a type when at least one of them carries an explicit
// Name; that type then has to be supplied with Input(id, val) and depended on
// with From[T](id), because it no longer identifies a single node.
func (b *Builder) Provide(sample any, opts ...Option) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}
	t := reflect.TypeOf(sample)
	if t == nil {
		return fmt.Errorf("%w: nil sample", ErrInputNotRegistered)
	}
	o := newOptions(opts)

	name := o.name
	if name == "" {
		name = inputID(t)
	}
	id := dagcore.NodeID(b.join(name))
	if _, ok := b.nodes[id]; ok {
		return fmt.Errorf("%w: %s", ErrNodeExisted, id)
	}
	// An input may not shadow a value that is computed by the graph.
	for _, ref := range b.producer[t] {
		if def := b.nodes[ref.node]; def != nil && !def.input {
			return fmt.Errorf("%w: %v is already produced by %s", ErrNodeExisted, t, ref.node)
		}
	}
	if err := b.dag.AddInput(id); err != nil {
		return err
	}
	if o.hasDef {
		if dt := reflect.TypeOf(o.defValue); dt != nil && !dt.AssignableTo(t) {
			return fmt.Errorf("%w: default %v is not assignable to %v", ErrInvalidOption, dt, t)
		}
		b.defaults[t] = o.defValue
	}
	b.register(id, nil, []reflect.Type{t}, true)
	b.inputs[t] = append(b.inputs[t], id)
	b.inputList = append(b.inputList, inputRef{id: id, typ: t})
	return nil
}

// Use registers fn as a compute node.
//
// Dependencies are inferred from the parameter types, results are published
// under their own types. See the package documentation for the accepted
// signatures and for the options that control id, ordering, timeout, retry and
// fallbacks.
func (b *Builder) Use(fn any, opts ...Option) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}
	spec, err := parseFunc(fn)
	if err != nil {
		return err
	}
	o := newOptions(opts)
	if err := b.checkFallback(spec, o); err != nil {
		return err
	}

	params, extra, err := b.resolveDeps(spec, o)
	if err != nil {
		return err
	}

	name := o.name
	if name == "" {
		name = defaultFuncID(spec, fn, len(b.nodes))
	}
	id := dagcore.NodeID(b.join(name))
	if _, ok := b.nodes[id]; ok {
		return fmt.Errorf("%w: %s", ErrNodeExisted, id)
	}

	depIDs := make([]dagcore.NodeID, 0, len(params)+len(extra))
	for _, d := range params {
		depIDs = append(depIDs, d.id)
	}
	depIDs = dedupeNodeIDs(append(depIDs, extra...))

	if err := b.dag.AddNode(id, depIDs, b.buildRun(spec, params, o)); err != nil {
		return err
	}
	b.register(id, params, spec.outs, false)
	if len(o.wrappers) > 0 {
		b.wrappers[id] = o.wrappers
	}
	return nil
}

// Freeze verifies the graph and makes it immutable.
//
// It resolves the subgraphs registered with Subgraph first, so a subgraph
// Builder does not have to be frozen by hand. Compile requires a frozen
// Builder.
func (b *Builder) Freeze() error {
	if b.dag.Frozen() {
		return ErrFrozen
	}
	if err := b.materializeSubgraphs(); err != nil {
		return err
	}
	return b.dag.Freeze()
}

// Frozen reports whether the Builder has been frozen.
func (b *Builder) Frozen() bool { return b.dag.Frozen() }

// Compile binds the declared inputs and returns a runnable Program.
//
// Inputs are matched by type; Input(id, val) binds a value to a specific input
// node instead, which is needed when two nodes declare the same input type.
// Every declared input must be supplied unless it has a default.
//
// wrappers are dagcore node wrappers applied to every node of the run, in
// addition to the per-node wrappers installed through Use / Wrap. They are the
// hook for tracing, metrics and logging.
func (b *Builder) Compile(inputs []any, wrappers ...dagcore.NodeFuncWrapper) (*Program, error) {
	if !b.dag.Frozen() {
		return nil, ErrNotFrozen
	}

	byType := make(map[reflect.Type]any, len(inputs))
	byID := make(map[dagcore.NodeID]any)
	for _, in := range inputs {
		if binding, ok := in.(inputBinding); ok {
			def := b.nodes[binding.id]
			if def == nil || !def.input {
				return nil, fmt.Errorf("%w: %s", ErrInputNotRegistered, binding.id)
			}
			if _, dup := byID[binding.id]; dup {
				return nil, fmt.Errorf("%w: duplicate value for node %s", ErrInputNotRegistered, binding.id)
			}
			if t := reflect.TypeOf(binding.val); t != nil && len(def.outs) == 1 && !t.AssignableTo(def.outs[0]) {
				return nil, fmt.Errorf("%w: %v is not assignable to %v", ErrInputNotRegistered, t, def.outs[0])
			}
			byID[binding.id] = binding.val
			continue
		}
		t := reflect.TypeOf(in)
		if t == nil {
			return nil, fmt.Errorf("%w: nil input", ErrInputNotRegistered)
		}
		ids := b.inputs[t]
		if len(ids) == 0 {
			return nil, fmt.Errorf("%w: %v", ErrInputNotRegistered, t)
		}
		if len(ids) > 1 {
			return nil, fmt.Errorf("%w: %v is provided by %v, use Input(id, val) to pick one",
				ErrAmbiguousType, t, ids)
		}
		if _, dup := byType[t]; dup {
			return nil, fmt.Errorf("%w: duplicate value for %v", ErrInputNotRegistered, t)
		}
		byType[t] = in
	}

	dagInputs := make(map[dagcore.NodeID]any, len(b.inputList))
	for _, ref := range b.inputList {
		if v, ok := byID[ref.id]; ok {
			dagInputs[ref.id] = v
			continue
		}
		v, ok := byType[ref.typ]
		if !ok {
			def, hasDef := b.defaults[ref.typ]
			if !hasDef {
				return nil, fmt.Errorf("%w: %v", ErrMissingInput, ref.typ)
			}
			v = def
		}
		dagInputs[ref.id] = v
	}

	all := make([]dagcore.NodeFuncWrapper, 0, len(wrappers)+1)
	all = append(all, wrappers...)
	all = append(all, b.perNodeWrapper())

	inst, err := b.dag.Instantiate(dagInputs, all...)
	if err != nil {
		return nil, err
	}
	return &Program{builder: b, execution: inst}, nil
}

// perNodeWrapper dispatches the wrappers registered for one specific node.
//
// dagcore applies wrappers[0] as the outermost layer, so appending this
// dispatcher after the user supplied ones keeps per-node wrappers closest to
// the node body while global wrappers stay outside of them.
func (b *Builder) perNodeWrapper() dagcore.NodeFuncWrapper {
	return func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
		ws := b.wrappers[n.ID()]
		for i := len(ws) - 1; i >= 0; i-- {
			run = ws[i](n, run)
		}
		return run
	}
}

func (b *Builder) register(id dagcore.NodeID, deps []depRef, outs []reflect.Type, input bool) {
	b.nodes[id] = &nodeDef{id: id, deps: deps, outs: outs, input: input}
	for i, t := range outs {
		ref := outputRef{node: id, index: -1}
		if len(outs) > 1 {
			ref.index = i
		}
		b.producer[t] = append(b.producer[t], ref)
	}
}

// outputRef resolves a type to the single node producing it.
func (b *Builder) outputRef(t reflect.Type) (outputRef, error) {
	if t == nil {
		return outputRef{}, fmt.Errorf("%w: <nil>", ErrTypeNotFound)
	}
	refs := b.producer[t]
	switch len(refs) {
	case 0:
		return outputRef{}, fmt.Errorf("%w: %v", ErrTypeNotFound, t)
	case 1:
		return refs[0], nil
	default:
		ids := make([]string, 0, len(refs))
		for _, r := range refs {
			ids = append(ids, string(r.node))
		}
		sort.Strings(ids)
		return outputRef{}, fmt.Errorf("%w: %v, candidates are %v", ErrAmbiguousType, t, ids)
	}
}

func (b *Builder) resolveDeps(spec *funcSpec, o *nodeOptions) ([]depRef, []dagcore.NodeID, error) {
	params := make([]depRef, 0, len(spec.in))
	for _, t := range spec.in {
		ref, err := b.resolveOne(t, o)
		if err != nil {
			return nil, nil, err
		}
		params = append(params, depRef{id: ref.node, index: ref.index, typ: t})
	}
	extra, err := b.resolveAfter(o)
	if err != nil {
		return nil, nil, err
	}
	return params, extra, nil
}

// resolveAfter resolves the ordering-only dependencies declared with After.
func (b *Builder) resolveAfter(o *nodeOptions) ([]dagcore.NodeID, error) {
	if len(o.after) == 0 {
		return nil, nil
	}
	ids := make([]dagcore.NodeID, 0, len(o.after))
	for _, d := range o.after {
		var id dagcore.NodeID
		switch v := d.(type) {
		case string:
			id = dagcore.NodeID(v)
		case dagcore.NodeID:
			id = v
		default:
			t := reflect.TypeOf(d)
			if t == nil {
				return nil, fmt.Errorf("%w: nil dependency", ErrNodeNotFound)
			}
			ref, err := b.outputRef(t)
			if err != nil {
				return nil, err
			}
			id = ref.node
		}
		if b.nodes[id] == nil {
			return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, id)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (b *Builder) resolveOne(t reflect.Type, o *nodeOptions) (outputRef, error) {
	if id, ok := o.binds[t]; ok {
		def := b.nodes[id]
		if def == nil {
			return outputRef{}, fmt.Errorf("%w: %s", ErrNodeNotFound, id)
		}
		i := def.indexOf(t)
		if i < 0 {
			return outputRef{}, fmt.Errorf("%w: node %s does not produce %v", ErrTypeNotFound, id, t)
		}
		return outputRef{node: id, index: def.refIndex(i)}, nil
	}
	ref, err := b.outputRef(t)
	if err != nil {
		return outputRef{}, fmt.Errorf("%w: %v", ErrMissingDependency, t)
	}
	return ref, nil
}

func dedupeNodeIDs(ids []dagcore.NodeID) []dagcore.NodeID {
	out := make([]dagcore.NodeID, 0, len(ids))
	seen := make(map[dagcore.NodeID]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (b *Builder) checkFallback(spec *funcSpec, o *nodeOptions) error {
	if o.recover == nil && !o.hasOrElse {
		return nil
	}
	if len(spec.outs) != 1 {
		return fmt.Errorf("%w: a fallback requires a node with exactly one output", ErrInvalidOption)
	}
	want := spec.outs[0]
	if o.recoverT != nil && !o.recoverT.AssignableTo(want) {
		return fmt.Errorf("%w: fallback type %v is not assignable to %v", ErrInvalidOption, o.recoverT, want)
	}
	if o.orElseT != nil && !o.orElseT.AssignableTo(want) {
		return fmt.Errorf("%w: fallback type %v is not assignable to %v", ErrInvalidOption, o.orElseT, want)
	}
	return nil
}

// buildRun turns a parsed function into a dagcore node body and applies the
// built-in options, from the inside out: retry, timeout, fallback. The user
// supplied Wrap wrappers are applied later by perNodeWrapper, so they end up
// outside of all of them.
func (b *Builder) buildRun(spec *funcSpec, params []depRef, o *nodeOptions) dagcore.NodeFunc {
	run := spec.run(params)

	if o.retries > 0 {
		inner := run
		run = func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
			var (
				val any
				err error
			)
			for i := 0; i <= o.retries; i++ {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				if i > 0 && o.backoff > 0 {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(o.backoff):
					}
				}
				val, err = inner(ctx, deps)
				if err == nil {
					return val, nil
				}
			}
			return val, err
		}
	}

	if o.timeout > 0 {
		inner := run
		run = func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
			ctx, cancel := context.WithTimeout(ctx, o.timeout)
			defer cancel()
			return inner(ctx, deps)
		}
	}

	if o.hasOrElse {
		inner := run
		fallback := o.orElse
		run = func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
			val, err := inner(ctx, deps)
			if err != nil {
				return fallback, nil
			}
			return val, nil
		}
	}

	if o.recover != nil {
		inner := run
		recover := o.recover
		run = func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
			val, err := inner(ctx, deps)
			if err == nil {
				return val, nil
			}
			nv, nerr := recover(err)
			if nerr != nil {
				return nil, nerr
			}
			return nv, nil
		}
	}

	return run
}

// Program is a compiled DAG bound to concrete inputs.
//
// A Program is single use, like a dagcore.DAGInstance: Compile one per run and
// run several Programs of the same Builder in parallel.
type Program struct {
	builder   *Builder
	execution *dagcore.DAGInstance
}

// Run executes every node and returns all results, keyed by the zero value of
// their type.
//
// Output types that are not comparable (slices, maps, functions) cannot be map
// keys and are omitted here; read them with Value[T], Get or Node.
func (p *Program) Run(ctx context.Context) (map[any]any, error) {
	return p.RunAsync(ctx).Get()
}

// RunAsync executes every node and returns a Future of all results, keyed by
// the zero value of their type.
func (p *Program) RunAsync(ctx context.Context) *future.Future[map[any]any] {
	return future.Then(p.execution.RunAsync(ctx), func(values map[dagcore.NodeID]any, err error) (map[any]any, error) {
		if err != nil {
			return nil, err
		}
		res := make(map[any]any, len(values))
		for id, v := range values {
			def := p.builder.nodes[id]
			if def == nil {
				continue
			}
			for i, t := range def.outs {
				if !t.Comparable() {
					continue
				}
				val := v
				if len(def.outs) > 1 {
					val = unwrap(v, i)
				}
				res[reflect.Zero(t).Interface()] = val
			}
		}
		return res, nil
	})
}

// Instance exposes the underlying dagcore runtime, for node level inspection
// (duration, dependencies, per node Future) and for dagviz.
func (p *Program) Instance() *dagcore.DAGInstance { return p.execution }

// Node returns the runtime node producing the type of sample.
//
// It fails with ErrTypeNotFound when no node produces the type and with
// ErrAmbiguousType when several do; use NodeByID in the latter case. For a
// multi-output node, Future carries the []any of all of its outputs.
func (p *Program) Node(sample any) (*dagcore.NodeInstance, error) {
	ref, err := p.builder.outputRef(reflect.TypeOf(sample))
	if err != nil {
		return nil, err
	}
	return p.node(ref.node)
}

// NodeByID returns the runtime node with the given id.
func (p *Program) NodeByID(id string) (*dagcore.NodeInstance, bool) {
	n, ok := p.execution.Nodes()[dagcore.NodeID(id)]
	return n, ok
}

func (p *Program) node(id dagcore.NodeID) (*dagcore.NodeInstance, error) {
	n, ok := p.execution.Nodes()[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, id)
	}
	return n, nil
}

// Get returns the output value of the node producing the type of sample.
//
// The sample is only used to determine the type; the caller has to assert the
// result. Value[T] is the type-safe alternative on Go 1.27 and newer.
func (p *Program) Get(sample any) (any, error) {
	return p.output(reflect.TypeOf(sample))
}

// outputFuture returns the Future of the value produced for type t.
func (p *Program) outputFuture(t reflect.Type) (*future.Future[any], error) {
	ref, err := p.builder.outputRef(t)
	if err != nil {
		return nil, err
	}
	n, err := p.node(ref.node)
	if err != nil {
		return nil, err
	}
	index := ref.index
	return future.Then(n.Future(), func(v any, err error) (any, error) {
		if err != nil {
			return nil, err
		}
		return unwrap(v, index), nil
	}), nil
}

func (p *Program) output(t reflect.Type) (any, error) {
	f, err := p.outputFuture(t)
	if err != nil {
		return nil, err
	}
	return f.Get()
}

// inputBinding binds an input value to a node id instead of a type.
type inputBinding struct {
	id  dagcore.NodeID
	val any
}

// Input binds val to the input node with the given id, instead of matching it
// by type. Needed when several nodes declare inputs of the same type.
//
//	prog, err := b.Compile([]any{dagfunc.Input("profile.userID", uid)})
func Input(id string, val any) any {
	return inputBinding{id: dagcore.NodeID(id), val: val}
}

// unwrap reads the index-th value of a multi-output node.
func unwrap(v any, index int) any {
	if index < 0 {
		return v
	}
	vals, ok := v.([]any)
	if !ok || index >= len(vals) {
		return nil
	}
	return vals[index]
}

func inputID(t reflect.Type) string { return "input:" + fullTypeName(t) }

func defaultFuncID(spec *funcSpec, fn any, seq int) string {
	if len(spec.outs) > 0 {
		names := make([]string, 0, len(spec.outs))
		for _, t := range spec.outs {
			names = append(names, fullTypeName(t))
		}
		return "func:" + joinNames(names)
	}
	// A node without results has no type to derive an id from, so the runtime
	// name of the function is used instead.
	if f := runtimeFuncName(fn); f != "" {
		return "func:" + f
	}
	return fmt.Sprintf("func:#%d", seq)
}

func joinNames(names []string) string {
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ","
		}
		out += n
	}
	return out
}

func fullTypeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.String() // e.g. builtin types like int, string
	}
	return t.PkgPath() + "." + t.Name()
}
