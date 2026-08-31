// Package future provides a lightweight, lock-free Future/Promise implementation
// for Go.
//
// A Promise is the producer side, a Future is the consumer side. Both are backed
// by a lock-free state machine: every transition is driven by atomic operations
// and a semaphore, so waiting never holds a mutex and completing a Future runs
// its callbacks on the completing goroutine instead of spawning new ones.
//
//	p := future.NewPromise[string]()
//	go func() { p.Set("hello", nil) }()
//	val, err := p.Future().Get()
//
// # API layering
//
// One rule decides whether an operation is a package-level function or a method
// on *Future[T]:
//
//  1. Constructors                      (Async, Done, NewPromise, ...)  -> function
//  2. Single-Future transforms          (Then, Map, Cast, Timeout, ...) -> method
//  3. Combinators over several Futures  (AllOf, AnyOf)                  -> function
//
// A transform that changes the result type has to introduce a type parameter of
// its own. That was impossible for methods until Go 1.27 added generic methods,
// which is why these transforms used to live at package scope. They are now
// methods, and the methods hold the only implementation; the old package-level
// forms remain as deprecated shims for source compatibility.
//
// A combinator over several Futures has no single receiver, so it stays a
// function. Java, Scala and friends draw the same line: "zip N futures into
// one" is a static or companion function there as well.
//
// AllOf and AnyOf cover batches of Futures that share a type. Zipping Futures
// of unrelated types needs one function per arity, which is why Tuple2..Tuple16
// and Of2..Of16 live in the tuples subpackage instead of here.
//
// # Generic methods
//
// Go 1.27 allows a method declaration to carry its own type parameters
// (https://go.dev/doc/go1.27, issue #77273):
//
//	func (f *Future[T]) Then[R any](cb func(T, error) (R, error)) *Future[R]
//
// Two restrictions apply and shape the API above:
//
//   - Interface methods may not declare type parameters, nor can an interface
//     method be implemented by a generic method. Future is therefore a concrete
//     type by design and is not exposed through an interface.
//
//   - A generic method may not return the receiver's own generic type
//     instantiated with a type built from the receiver type parameter. The
//     following looks natural but is rejected with
//     "instantiation cycle: T instantiated as Tuple2[T, R]":
//
//     func (f *Future[T]) Combine[R any](g *Future[R]) *Future[Tuple2[T, R]]
//
//     Type-checking Future[T] would require Future[Tuple2[T, R]], then
//     Future[Tuple2[Tuple2[T, R], R]] and so on forever. Every transform in this
//     package therefore returns *Future[R] with a fresh R, and combining
//     Futures is left to the package-level combinators and to the tuples
//     subpackage.
package future
