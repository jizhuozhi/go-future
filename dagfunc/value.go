//go:build go1.27

package dagfunc

import (
	"fmt"
	"reflect"

	"github.com/jizhuozhi/go-future"
)

// ValueAsync returns the output of the node producing T as a Future, without
// blocking. The returned Future is resolved when that node completes, so it has
// to be triggered by Run or RunAsync first.
//
// Thanks to generic methods (Go 1.27) the result type is expressed as a method
// type parameter, which removes both the throw-away sample value and the type
// assertion needed by Get:
//
//	go prog.Run(ctx)
//	c, err := prog.ValueAsync[ResultC]().Get()
//
// If no node produces T the Future fails immediately with ErrTypeNotFound; if
// the produced value is not assignable to T it fails with future.ErrTypeMismatch.
func (p *Program) ValueAsync[T any]() *future.Future[T] {
	var zero T
	// reflect.TypeFor[T]() would be clearer but is Go 1.22+, and this file must
	// stay buildable under the module's declared language version.
	typ := reflect.TypeOf((*T)(nil)).Elem()
	id, ok := p.builder.typeToID[typ]
	if !ok {
		return future.Done2(zero, fmt.Errorf("%w: %v", ErrTypeNotFound, typ))
	}
	return p.execution.Nodes()[id].Cast[T]()
}

// Value returns the output of the node producing T, blocking until that node
// completes.
//
// The program must have been started with Run or RunAsync, otherwise Value
// blocks forever.
//
//	c, err := prog.Value[ResultC]()
func (p *Program) Value[T any]() (T, error) {
	return p.ValueAsync[T]().Get()
}
