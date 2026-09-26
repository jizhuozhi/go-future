package dagcore

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jizhuozhi/go-future"
)

var (
	ErrDAGNodeExisted     = errors.New("DAG node is existed")
	ErrDAGNodeNotInput    = errors.New("DAG node is not input")
	ErrDAGNodeNotRunnable = errors.New("DAG node is not runnable")
	ErrDAGNodeNotExecuted = errors.New("DAG node is not executed")

	ErrDAGFrozen     = errors.New("DAG node is frozen")
	ErrDAGNotFrozen  = errors.New("DAG node is not frozen")
	ErrDAGIncomplete = errors.New("DAG is incomplete")
	ErrDAGCyclic     = errors.New("DAG is cyclic")
)

// NodeID represents the unique identifier of a node in the DAG
type NodeID string

// NodeFunc is the signature of the business logic executed by a node
type NodeFunc func(ctx context.Context, deps map[NodeID]any) (any, error)

// NodeSpec is the immutable blueprint of a DAG node, containing metadata and logic
type NodeSpec struct {
	id    NodeID
	deps  []NodeID
	run   NodeFunc
	input bool

	// subgraph if not nil, this node presents a subgraph
	subgraph              *DAG
	subgraphInputMapping  func(map[NodeID]any) map[NodeID]any
	subgraphOutputMapping func(map[NodeID]any) any

	// skipFunc decides whether the node should be skipped, and supplies the
	// value to publish when it is. Returning false leaves the value unused.
	skipFunc func(ctx context.Context, deps map[NodeID]any) (bool, any)

	// recoverFunc decides, per failure, whether an execution error is
	// recoverable. Returning a nil error recovers by publishing the returned
	// value; returning an error propagates instead.
	recoverFunc func(ctx context.Context, deps map[NodeID]any, err error) (any, error)
}

// evalSkip evaluates the skip predicate. A node without a predicate is never
// skipped.
func (s *NodeSpec) evalSkip(ctx context.Context, deps map[NodeID]any) (bool, any) {
	if s.skipFunc == nil {
		return false, nil
	}
	return s.skipFunc(ctx, deps)
}

type NodeOpt func(*NodeSpec)

// DAG is the static structure definition holding node specs
type DAG struct {
	nodes  map[NodeID]*NodeSpec
	frozen bool
}

// NodeFuncWrapper is used to wrap node execution logic (e.g. for logging, retry, metrics)
type NodeFuncWrapper func(n *NodeInstance, run NodeFunc) NodeFunc

// NewDAG creates a new DAG instance
func NewDAG() *DAG {
	return &DAG{nodes: make(map[NodeID]*NodeSpec)}
}

// AddInput adds an input node to the DAG, input value will be given in Instantiate.
//
// Will return an error if the DAG is frozen.
func (d *DAG) AddInput(id NodeID) error {
	if d.frozen {
		return ErrDAGFrozen
	}

	if _, exists := d.nodes[id]; exists {
		return ErrDAGNodeExisted
	}
	d.nodes[id] = &NodeSpec{
		id:    id,
		input: true,
	}
	return nil
}

// AddNode adds a new node to the DAG.
//
// Will return an error if the DAG is frozen.
func (d *DAG) AddNode(id NodeID, deps []NodeID, fn NodeFunc, opts ...NodeOpt) error {
	if d.frozen {
		return ErrDAGFrozen
	}

	if _, exists := d.nodes[id]; exists {
		return ErrDAGNodeExisted
	}
	if fn == nil {
		return ErrDAGNodeNotRunnable
	}
	n := &NodeSpec{
		id:   id,
		deps: deps,
		run:  fn,
	}

	for _, opt := range opts {
		opt(n)
	}

	d.nodes[id] = n
	return nil
}

func (d *DAG) AddSubgraph(id NodeID, deps []NodeID, subgraph *DAG, inputMapping func(map[NodeID]any) map[NodeID]any, outputMapping func(map[NodeID]any) any, opts ...NodeOpt) error {
	if d.frozen {
		return ErrDAGFrozen
	}
	if _, exists := d.nodes[id]; exists {
		return ErrDAGNodeExisted
	}
	n := &NodeSpec{
		id:   id,
		deps: deps,
		run:  nil, // delayed assignment at Instantiate time

		subgraph:              subgraph,
		subgraphInputMapping:  inputMapping,
		subgraphOutputMapping: outputMapping,
	}

	for _, opt := range opts {
		opt(n)
	}

	d.nodes[id] = n
	return nil
}

// Freeze verifies that the DAG structure is complete and acyclic,
// and marks the DAG as immutable for future instantiations.
//
// After calling Freeze successfully, the DAG topology becomes read-only.
// Any attempt to modify the DAG (e.g., adding nodes) will return an error.
// Instantiate will require the DAG to be frozen before execution.
func (d *DAG) Freeze() error {
	if d.frozen {
		return ErrDAGFrozen
	}
	if err := d.checkComplete(); err != nil {
		return err
	}
	if err := d.checkCycle(); err != nil {
		return err
	}
	d.frozen = true
	return nil
}

// Frozen returns true if the DAG has been frozen.
//
// A frozen DAG is immutable and safe for repeated instantiation and execution.
func (d *DAG) Frozen() bool {
	return d.frozen
}

// checkComplete verifies that all declared dependencies exist in the DAG
func (d *DAG) checkComplete() error {
	for id, node := range d.nodes {
		for _, dep := range node.deps {
			if _, ok := d.nodes[dep]; !ok {
				return fmt.Errorf("dependency %s of node %s is not present: %w", dep, id, ErrDAGIncomplete)
			}
		}
	}
	return nil
}

// checkCycle detects any cycles in the DAG using Kahn's algorithm
func (d *DAG) checkCycle() error {
	inDegree := make(map[NodeID]int)
	queue := make([]NodeID, 0)
	visited := 0

	for id, node := range d.nodes {
		inDegree[id] = len(node.deps)
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	children := make(map[NodeID][]NodeID)
	for id, node := range d.nodes {
		for _, dep := range node.deps {
			children[dep] = append(children[dep], id)
		}
	}

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		visited++
		for _, v := range children[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if visited != len(d.nodes) {
		return ErrDAGCyclic
	}
	return nil
}

// Instantiate builds a DAGInstance for execution with the given input values
// and optional NodeFunc wrappers.
//
// The DAG must be frozen before instantiation. If the DAG is not frozen,
// Instantiate will return an error.
//
// Each instantiation produces an isolated runtime with its own promises,
// allowing the same static DAG to be executed multiple times in parallel.
func (d *DAG) Instantiate(inputs map[NodeID]any, wrappers ...NodeFuncWrapper) (*DAGInstance, error) {
	if !d.frozen {
		return nil, ErrDAGNotFrozen
	}

	nodes := make(map[NodeID]*NodeInstance)
	children := make(map[NodeID][]NodeID)
	for id, spec := range d.nodes {
		// `spec := spec` ensures that the closure inside this loop captures
		// a unique copy of spec per iteration. Without this, all closures would
		// share the same spec variable and cause incorrect behavior.
		spec := spec

		promise := future.NewPromise[any]()

		val, ok := inputs[id]
		if ok && !spec.input {
			return nil, ErrDAGNodeNotInput
		}
		if !ok && spec.input {
			return nil, ErrDAGNodeNotRunnable
		}

		if ok {
			promise.Set(val, nil)
		}

		node := &NodeInstance{
			spec:    spec,
			run:     spec.run,
			pending: int32(len(spec.deps)),
			future:  promise.Future(),
			promise: promise,
		}

		if spec.subgraph != nil {
			node.run = func(ctx context.Context, deps map[NodeID]any) (any, error) {
				subInputs := deps
				if spec.subgraphInputMapping != nil {
					subInputs = spec.subgraphInputMapping(deps)
				}
				subInstance, err := spec.subgraph.Instantiate(subInputs, wrappers...)
				if err != nil {
					return nil, err
				}
				node.subgraph = subInstance
				res, err := subInstance.Run(ctx)
				if err != nil {
					return nil, err
				}
				if spec.subgraphOutputMapping != nil {
					return spec.subgraphOutputMapping(res), nil
				}
				return res, nil
			}
		}

		nodes[id] = node
		for _, dep := range spec.deps {
			children[dep] = append(children[dep], id)
		}
	}

	for id, node := range nodes {
		node.children = children[id]
	}

	return &DAGInstance{
		spec:     d,
		nodes:    nodes,
		wrappers: wrappers,
	}, nil
}

// NodeInstance represents a runtime execution context for a single DAG node
type NodeInstance struct {
	spec     *NodeSpec
	children []NodeID
	run      NodeFunc

	subgraph *DAGInstance

	pending   int32
	future    *future.Future[any]
	start     time.Time
	duration  time.Duration
	skipped   bool
	recovered bool
	err       error

	promise *future.Promise[any]
}

func (n *NodeInstance) ID() NodeID                  { return n.spec.id }
func (n *NodeInstance) Deps() []NodeID              { return n.spec.deps }
func (n *NodeInstance) Input() bool                 { return n.spec.input }
func (n *NodeInstance) Subgraph() *DAGInstance      { return n.subgraph }
func (n *NodeInstance) Future() *future.Future[any] { return n.future }
func (n *NodeInstance) Duration() time.Duration     { return n.duration }
func (n *NodeInstance) Skipped() bool               { return n.skipped }

// Recovered reports whether the node failed and a WithRecoverFunc handler
// replaced the error with a value.
//
// It is the third state alongside Skipped and Err, and it cannot be derived
// from the other two. Err is populated both when a handler recovers and when
// it declines, and a declining handler may rewrite the error before it fails
// the run, so neither the node's error nor the run result identifies on its own
// whether a given failure was absorbed.
//
// This is the flag to count degradations with. A recovered failure does not
// appear in the error returned by DAGInstance.Run and does not appear in any
// aggregate error rate, yet it means the caller was served with a worse answer
// than a successful node would have produced.
//
// Returns false for nodes that succeeded, were skipped, or were never run.
func (n *NodeInstance) Recovered() bool { return n.recovered }

// Err returns the error the node function returned.
//
// It holds the error the node function returned, never one substituted by a
// WithRecoverFunc handler, and it is populated even when that error was
// recovered. Pair it with Recovered to tell the two cases apart: Err alone
// cannot, because a declining handler also leaves it populated.
//
// Returns nil for nodes that succeeded, were skipped, or never ran.
func (n *NodeInstance) Err() error { return n.err }

// DAGInstance is the per-execution runtime of a DAG
type DAGInstance struct {
	spec     *DAG
	nodes    map[NodeID]*NodeInstance
	wrappers []NodeFuncWrapper
}

// Run runs the DAG instance and returns the final values or error
func (d *DAGInstance) Run(ctx context.Context) (map[NodeID]any, error) {
	return d.RunAsync(ctx).Get()
}

// RunAsync runs the DAG instance async and returns future of values
func (d *DAGInstance) RunAsync(ctx context.Context) *future.Future[map[NodeID]any] {
	futures := make([]*future.Future[any], 0, len(d.nodes))

	roots := make([]NodeID, 0, len(d.nodes))
	for id, node := range d.nodes {
		if node.pending == 0 {
			roots = append(roots, id)
		}
		futures = append(futures, node.future)
	}

	for _, id := range roots {
		d.schedule(ctx, id)
	}

	f := future.Then(future.AllOf(futures...), func(_ []any, err error) (map[NodeID]any, error) {
		if err != nil {
			for _, n := range d.nodes {
				if !n.future.Done() {
					// Some node execution failed, defensively mark all unexecuted nodes with ErrDAGNodeNotExecuted.
					// This prevents Get() from blocking forever due to missing results.
					// Using SetSafety ensures concurrency safety here as Subscribe
					// may try to set these promises simultaneously.
					n.promise.SetSafety(nil, ErrDAGNodeNotExecuted)
				}
			}
			return nil, err
		}
		results := make(map[NodeID]any)
		for id, n := range d.nodes {
			v, err := n.future.Get()
			if err != nil {
				// makes linter happy, but never touched...
				return nil, err
			}
			results[id] = v
		}
		return results, nil
	})

	return f
}

func (d *DAGInstance) schedule(ctx context.Context, id NodeID) {
	node := d.nodes[id]

	if node.spec.input {
		for _, child := range node.children {
			if atomic.AddInt32(&d.nodes[child].pending, -1) == 0 {
				d.schedule(ctx, child)
			}
		}
		return
	}

	run := node.run
	for i := len(d.wrappers) - 1; i >= 0; i-- {
		run = d.wrappers[i](node, run)
	}
	node.start = time.Now()
	future.CtxAsync(ctx, func(ctx context.Context) (any, error) {
		var val any
		var err error
		deps := make(map[NodeID]any)
		for _, depid := range node.spec.deps {
			v, err := d.nodes[depid].future.Get()
			if err != nil {
				// makes linter happy, but never touched...
				return nil, fmt.Errorf("dep %s failed: %w", depid, err)
			}
			deps[depid] = v
		}
		if skip, skipValue := node.spec.evalSkip(ctx, deps); skip {
			node.skipped = true
			val = skipValue
		} else {
			val, err = run(ctx, deps)
			if err != nil {
				node.err = err
				// Recovery is a per-failure decision. Without a handler the error
				// propagates; with one, the handler decides whether this
				// particular error is recoverable and may decline by returning
				// an error of its own.
				if node.spec.recoverFunc != nil {
					if recovered, recoverErr := node.spec.recoverFunc(ctx, deps, err); recoverErr != nil {
						err = recoverErr
					} else {
						val, err = recovered, nil
						node.recovered = true
					}
				}
			}
		}
		node.duration = time.Since(node.start)
		if err != nil {
			return nil, err
		}
		for _, child := range node.children {
			if atomic.AddInt32(&d.nodes[child].pending, -1) == 0 {
				d.schedule(ctx, child)
			}
		}
		return val, err
	}).Subscribe(func(val any, err error) {
		// Use SetSafety instead of Set here because:
		// Only when future.AllOf(...).Get() returns an error (i.e., some node failed),
		// there will be concurrent attempts to mark all unfinished nodes as failed,
		// but only the first call will succeed in setting the result, avoiding panic
		node.promise.SetSafety(val, err)
	})
}

func (d *DAGInstance) Spec() *DAG {
	return d.spec
}

func (d *DAGInstance) Nodes() map[NodeID]*NodeInstance {
	return d.nodes
}

// WithSkipFunc registers a skip predicate for a node.
//
// The predicate runs after the node's dependencies have been collected and
// before the node function is invoked:
//
//	return true, value  // skip the node and publish value to downstream nodes
//	return false, nil   // run the node normally; the value is ignored
//
// The value returned alongside false is ignored, so the normal path does not
// need to produce one.
//
// Skipping and error recovery are independent. This option alone does not make
// a node tolerant of failure: a skipped node publishes a value, but if the node
// does run and fails, the error still fails the whole run. Combine it with
// WithRecoverFunc to recover failures as well.
//
// That separation matters because the two mean different things. A skip is a
// planned decision ("this model is irrelevant for this request", "the remaining
// budget is not worth spending"), whereas a recovered failure is an incident.
// Keeping them apart is what allows the two to be counted separately; folding
// them together makes either the failure rate noisy or the alerts useless.
//
// The predicate receives the execution context and the resolved dependencies,
// so it may depend on either: an experiment assignment, the remaining budget,
// or an upstream value. It cannot observe a failed dependency. If any
// dependency fails, the node is never scheduled and the predicate is never
// called; to tolerate an upstream failure, give the upstream node a
// WithRecoverFunc instead.
//
// A skipped node never runs, so its Duration carries no statistical meaning for
// latency aggregation. Filter on NodeInstance.Skipped before aggregating,
// otherwise the near-zero samples will drag percentiles down and hide the real
// distribution.
//
// The published value must match the type downstream nodes assert on. An
// untyped nil will make `deps[id].(T)` panic downstream; return a typed zero
// value instead (e.g. (*T)(nil), []T(nil)).
func WithSkipFunc(fn func(ctx context.Context, deps map[NodeID]any) (bool, any)) NodeOpt {
	return func(n *NodeSpec) {
		n.skipFunc = fn
	}
}

// WithRecoverFunc registers an error handler that decides, per failure,
// whether a node's error is recoverable.
//
// When the node function returns an error, the handler is called with it:
//
//	return value, nil   // recover: publish value, keep scheduling downstream
//	return nil, err     // do not recover: the error fails the run
//	return nil, wrap    // do not recover, adding context first
//
// Recovering has to be an explicit decision per failure. Not every error
// deserves recovery: a timeout or a transient backend error is usually worth
// degrading over, whereas a malformed request, an unknown feature field, or a
// model/feature id space mismatch is a defect that recovering would only paper
// over. Recovering unconditionally is the easy mistake here; write the
// propagate branch first and recover only the cases you have actually reasoned
// about.
//
// This option is what makes a node optional. A node whose failure must fail the
// whole run simply omits it.
//
// Recovery hides a failure in two places, the error returned by DAGInstance.Run
// and any aggregate error rate, so it has to be counted explicitly:
// NodeInstance.Recovered reports it, and NodeInstance.Err still holds the error
// the node function returned. Count them and alert on them; a recovery means
// the caller was served with a worse answer than a successful node would have
// produced.
//
// A panic does not arrive here. The node closure runs inside future.CtxAsync,
// which catches a panic with a deferred recover and turns it into an error
// wrapping ErrPanic with the stack attached — but that recover sits outside
// this closure, so unwinding skips the error branch below and the handler is
// never called. A panicking node therefore fails the run even when it is
// marked optional; NodeInstance.Recovered stays false, and NodeInstance.Err
// stays nil because the assignment to it is skipped as well.
//
// Whether that boundary is intended or an accident of where the recover sits
// deserves to be settled explicitly. If a panicking node ought to be
// tolerable, the conversion has to move inside this closure so that ErrPanic
// flows through the error branch like any other failure.
//
// The value published when recovering must match the type downstream nodes
// assert on. An untyped nil will make `deps[id].(T)` panic downstream; return a
// typed zero value instead (e.g. (*T)(nil), []T(nil)).
func WithRecoverFunc(fn func(ctx context.Context, deps map[NodeID]any, err error) (any, error)) NodeOpt {
	return func(n *NodeSpec) {
		n.recoverFunc = fn
	}
}
