// Package future provides a lightweight, mutex-free Future/Promise
// implementation for Go.
//
// A Promise is the producer side, a Future is the consumer side. Both are backed
// by a single atomic word: every state transition is one compare-and-swap on
// that word, and waiters park on a semaphore. No mutex is taken on any path, and
// completing a Future runs its callbacks on the completing goroutine instead of
// spawning new ones.
//
//	p := future.NewPromise[string]()
//	go func() { p.Set("hello", nil) }()
//	val, err := p.Future().Get()
//
// # Waiting: a semaphore, not sync.Cond
//
// Get parks a goroutine until the result is published. The textbook tool for
// that is a sync.Cond, but a Cond needs a Locker: waiting means unlocking,
// parking and locking again. A sync.Mutex is adaptive — it spins, then backs
// off, then parks on a semaphore of its own — and that escalation pays for
// itself only when the critical section is short. An asynchronous task is a
// heavy operation, so the wait here is long by nature and the mutex machinery
// would be overhead spent for nothing.
//
// This package therefore skips the mutex and drives the semaphore directly,
// keeping only the part a condition variable is actually needed for here: a
// queue of parked goroutines for Set to hand off to. There is no spin phase and
// no lock upgrade on the wait path.
//
// The cost is that the semaphore is reached through //go:linkname rather than
// through the public API. The Go source is candid about this: runtime/sema.go
// publishes both symbols with an explicit //go:linkname push and carries this
// note —
//
//	sync_runtime_Semacquire should be an internal detail,
//	but widely used packages access it using linkname.
//	Notable members of the hall of shame include:
//	  - gvisor.dev/gvisor
//	  - github.com/sagernet/gvisor
//
//	Do not remove or change the type signature.
//	See go.dev/issue/67401.
//
// The push makes this the handshake form rsc describes as the desired end state
// in go.dev/issue/67401, rather than an unauthorised pull, and "Do not remove or
// change the type signature" is a commitment the Go team has made. The note also
// records who else depends on it: gvisor. See linkname.go.
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
// which is why the earlier transforms were written as package-level
// functions taking the Future as their first argument.
//
// Both forms are available for those transforms: the methods are the
// expressive, chainable form, the package-level functions are the original API
// and remain fully supported, not deprecated. The six transforms added in
// v0.2.0 (ThenGo, Map, FlatMap, Cast, Recover, OrElse) were written against
// generic methods and exist in method form only, so they need Go 1.27.
// They are independent implementations rather than shims, see
// "Build tags" below.
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
//
// # Build tags
//
// The module declares go 1.18 in go.mod, so it keeps building with Go 1.18
// through Go 1.26. The generic methods live in files guarded by
//
//	//go:build go1.27
//
// which is satisfied by the toolchain version, not by the go directive. On
// Go 1.27 and newer both the methods and the package-level functions exist; on
// older toolchains only the package-level functions are compiled.
//
// This is why the two forms do not delegate to each other: the package-level
// functions must not depend on a file that may be excluded from the build. The
// duplicated bodies are small and self-contained.
//
// The same guard applies to dagcore.NodeInstance.Cast and to
// dagfunc.Program.Value / ValueAsync.
package future
