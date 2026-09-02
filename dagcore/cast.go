//go:build go1.27

package dagcore

import (
	"github.com/jizhuozhi/go-future"
)

// Cast returns the result of the node as a *future.Future[T].
//
// Node results are carried as `any` because a DAG is built dynamically; Cast is
// the typed bridge back to static types. It relies on the generic methods
// introduced in Go 1.27, which let a method declare its own type parameters.
//
//	count, err := inst.Nodes()["func:TokenCount"].Cast[TokenCount]().Get()
//
// If the node fails, the error is propagated. If the produced value is not
// assignable to T, the returned Future fails with future.ErrTypeMismatch.
func (n *NodeInstance) Cast[T any]() *future.Future[T] {
	return n.future.Cast[T]()
}
