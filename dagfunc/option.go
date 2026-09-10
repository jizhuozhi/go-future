package dagfunc

import (
	"reflect"
	"time"

	"github.com/jizhuozhi/go-future/dagcore"
)

// Name gives the node an explicit id instead of the type derived default.
//
// An explicit id is required to register two nodes producing the same type, and
// it is what NodeByID, Wrap and dagviz report.
func Name(id string) Option {
	return func(o *nodeOptions) { o.name = id }
}

// After declares an ordering-only dependency: the node waits for it but does
// not receive its value.
//
// Every element is either a node id (string or dagcore.NodeID) or a sample
// value whose type resolves to a node. It is how a node without results, or a
// node that must simply run after another one, is wired in.
//
//	_ = b.Use(audit, dagfunc.Name("audit"), dagfunc.After(Report{}))
func After(deps ...any) Option {
	return func(o *nodeOptions) { o.after = append(o.after, deps...) }
}

// From pins the source of one parameter to a specific node id.
//
// Use it when several nodes produce the same type and the parameter must not be
// resolved by type alone.
//
//	_ = b.Use(merge, dagfunc.From[Summary]("fast.summary"))
func From[T any](id string) Option {
	return func(o *nodeOptions) {
		t := reflect.TypeOf((*T)(nil)).Elem()
		if o.binds == nil {
			o.binds = make(map[reflect.Type]dagcore.NodeID)
		}
		o.binds[t] = dagcore.NodeID(id)
	}
}

// Timeout bounds the execution of the node.
//
// The node function has to honour context.Context for this to have an effect;
// Timeout only shortens the context handed to it.
func Timeout(d time.Duration) Option {
	return func(o *nodeOptions) { o.timeout = d }
}

// Retry reruns the node up to n times after a failure, with no delay.
func Retry(n int) Option {
	return func(o *nodeOptions) { o.retries = n }
}

// RetryWith reruns the node up to n times after a failure, waiting backoff
// between two attempts.
func RetryWith(n int, backoff time.Duration) Option {
	return func(o *nodeOptions) {
		o.retries = n
		o.backoff = backoff
	}
}

// Wrap installs dagcore node wrappers for this node only.
//
// They are applied outside of the built-in options (retry, timeout, fallback),
// so a wrapper observes the final result of the node.
func Wrap(wrappers ...dagcore.NodeFuncWrapper) Option {
	return func(o *nodeOptions) { o.wrappers = append(o.wrappers, wrappers...) }
}

// Recover turns a failure of the node into a value produced by fn.
//
// It applies to nodes with exactly one output.
func Recover[R any](fn func(err error) (R, error)) Option {
	return func(o *nodeOptions) {
		o.recover = func(err error) (any, error) { return fn(err) }
		o.recoverT = reflect.TypeOf((*R)(nil)).Elem()
	}
}

// OrElse replaces a failure of the node with fallback.
//
// It applies to nodes with exactly one output.
func OrElse[R any](fallback R) Option {
	return func(o *nodeOptions) {
		o.orElse = fallback
		o.orElseT = reflect.TypeOf((*R)(nil)).Elem()
		o.hasOrElse = true
	}
}

// Default supplies the value used by Compile when no input of that type is
// given. It only applies to Provide.
func Default(val any) Option {
	return func(o *nodeOptions) {
		o.defValue = val
		o.hasDef = true
	}
}

// Outputs declares which results of a subgraph are exposed to the parent graph,
// in order. It only applies to Subgraph and is mandatory there.
//
// With a single type the subgraph node behaves like a single-output node; with
// several it becomes a multi-output node whose results are read by type.
func Outputs(samples ...any) Option {
	return func(o *nodeOptions) {
		for _, s := range samples {
			if t := reflect.TypeOf(s); t != nil {
				o.outputs = append(o.outputs, t)
			}
		}
	}
}
