package dagcore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
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
}

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
func (d *DAG) AddNode(id NodeID, deps []NodeID, fn NodeFunc) error {
	if d.frozen {
		return ErrDAGFrozen
	}

	if _, exists := d.nodes[id]; exists {
		return ErrDAGNodeExisted
	}
	if fn == nil {
		return ErrDAGNodeNotRunnable
	}
	d.nodes[id] = &NodeSpec{
		id:   id,
		deps: deps,
		run:  fn,
	}
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
		nodes[id] = &NodeInstance{
			spec:    spec,
			run:     spec.run,
			pending: int32(len(spec.deps)),
			future:  promise.Future(),
			promise: promise,
		}
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

	pending  int32
	future   *future.Future[any]
	start    time.Time
	duration time.Duration

	promise *future.Promise[any]
}

func (n *NodeInstance) ID() NodeID                  { return n.spec.id }
func (n *NodeInstance) Deps() []NodeID              { return n.spec.deps }
func (n *NodeInstance) Input() bool                 { return n.spec.input }
func (n *NodeInstance) Future() *future.Future[any] { return n.future }
func (n *NodeInstance) Duration() time.Duration     { return n.duration }

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
	result := make(map[NodeID]any)
	resultMu := sync.Mutex{}
	futures := make([]*future.Future[any], 0, len(d.nodes))

	roots := make([]NodeID, 0, len(d.nodes))
	for id, node := range d.nodes {
		if node.pending == 0 {
			roots = append(roots, id)
		}
		futures = append(futures, node.future)
	}

	var schedule func(NodeID)
	schedule = func(id NodeID) {
		node := d.nodes[id]

		if node.spec.input {
			// input node never set error, ignore directly
			val, _ := node.future.Get()
			resultMu.Lock()
			result[id] = val
			resultMu.Unlock()
			for _, child := range node.children {
				if atomic.AddInt32(&d.nodes[child].pending, -1) == 0 {
					schedule(child)
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
			deps := make(map[NodeID]any)
			for _, depid := range node.spec.deps {
				v, err := d.nodes[depid].future.Get()
				if err != nil {
					return nil, fmt.Errorf("dep %s failed: %w", depid, err)
				}
				deps[depid] = v
			}
			val, err := run(ctx, deps)
			node.duration = time.Since(node.start)
			if err == nil {
				resultMu.Lock()
				result[id] = val
				resultMu.Unlock()
			}
			for _, child := range node.children {
				if atomic.AddInt32(&d.nodes[child].pending, -1) == 0 {
					schedule(child)
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

	for _, id := range roots {
		schedule(id)
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
		res := make(map[NodeID]any, len(d.nodes))
		resultMu.Lock()
		for k, v := range result {
			res[k] = v
		}
		resultMu.Unlock()
		return result, nil
	})

	return f
}

func (d *DAGInstance) Spec() *DAG {
	return d.spec
}

func (d *DAGInstance) Nodes() map[NodeID]*NodeInstance {
	return d.nodes
}

// ToMermaid converts the DAG static topology to a Mermaid.js-compatible graph string.
// This is an external utility function that reads the DAG structure for visualization.
func ToMermaid(d *DAGInstance) string {
	var b strings.Builder
	b.WriteString("graph LR\n")
	ids := make([]string, 0, len(d.nodes))
	for id := range d.nodes {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	for _, id := range ids {
		node := d.Nodes()[NodeID(id)]
		label := id
		if node.Input() {
			b.WriteString(fmt.Sprintf("\t%s[%q]\n", label, label))
		} else {
			b.WriteString(fmt.Sprintf("\t%s((%q))\n", label, label))
		}
	}
	for _, id := range ids {
		node := d.Nodes()[NodeID(id)]
		for _, dep := range node.Deps() {
			b.WriteString(fmt.Sprintf("\t%s --> %s\n", string(dep), string(id)))
		}
	}
	return b.String()
}
