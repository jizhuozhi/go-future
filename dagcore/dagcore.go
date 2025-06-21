package dagcore

import (
	"context"
	"errors"
	"fmt"
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
	nodes map[NodeID]*NodeSpec
}

// NodeFuncWrapper is used to wrap node execution logic (e.g. for logging, retry, metrics)
type NodeFuncWrapper func(n *NodeInstance, run NodeFunc) NodeFunc

// NewDAG creates a new DAG instance
func NewDAG() *DAG {
	return &DAG{nodes: make(map[NodeID]*NodeSpec)}
}

// AddInput adds an input node to the DAG, input value will be given in Instantiate
func (d *DAG) AddInput(id NodeID) error {
	if _, exists := d.nodes[id]; exists {
		return ErrDAGNodeExisted
	}
	d.nodes[id] = &NodeSpec{
		id:    id,
		input: true,
	}
	return nil
}

// AddNode adds a new node to the DAG
func (d *DAG) AddNode(id NodeID, deps []NodeID, fn NodeFunc) error {
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

// Instantiate builds a DAGInstance for execution with a given set of wrappers
func (d *DAG) Instantiate(inputs map[NodeID]any, wrappers ...NodeFuncWrapper) (*DAGInstance, error) {
	if err := d.checkComplete(); err != nil {
		return nil, err
	}
	if err := d.checkCycle(); err != nil {
		return nil, err
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
func (n *NodeInstance) Future() *future.Future[any] { return n.future }
func (n *NodeInstance) Duration() time.Duration     { return n.duration }

// DAGInstance is the per-execution runtime of a DAG
type DAGInstance struct {
	spec     *DAG
	nodes    map[NodeID]*NodeInstance
	wrappers []NodeFuncWrapper
}

// Execute runs the DAG instance and returns the final values or error
func (d *DAGInstance) Execute(ctx context.Context) (map[NodeID]any, error) {
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
			node.promise.Set(val, err)
		})
	}

	for _, id := range roots {
		schedule(id)
	}

	_, err := future.AllOf(futures...).Get()
	if err != nil {
		// marking not executed nodes as specially error to avoid deadlock
		for _, n := range d.nodes {
			n.promise.SetSafety(nil, ErrDAGNodeNotExecuted)
		}
		return nil, err
	}
	return result, nil
}

func (d *DAGInstance) Spec() *DAG {
	return d.spec
}

func (d *DAGInstance) Nodes() map[NodeID]*NodeInstance {
	return d.nodes
}
