// Package tuples provides heterogeneous tuple types and combinators that zip
// several Futures of unrelated types into one.
//
// Tuple2..Tuple16 hold values of unrelated types, and Of2..Of16 wait for the
// corresponding Futures and resolve with such a tuple, failing fast on the
// first error:
//
//	f := tuples.Of2(fetchUser(), fetchAccount())
//	t, err := f.Get()
//	if err == nil {
//	    use(t.Val0, t.Val1)
//	}
//
// There is one combinator per arity because Go cannot express a variadic
// function over heterogeneous types. That is also why they live in a
// subpackage: they are rarely needed, and they add 30 exported symbols that
// would otherwise clutter the main future package. Most code is served by
// future.AllOf for homogeneous batches, or by chaining f.Then(...) on a single
// Future.
//
// Note that the tuple combinators cannot be methods on *future.Future[T]: a
// generic method may not return its receiver's type instantiated with a type
// built from the receiver type parameter, so
//
//	func (f *Future[T]) Combine[R any](g *Future[R]) *Future[Tuple2[T, R]]
//
// is rejected by the compiler with "instantiation cycle". See the future
// package documentation for details.
package tuples
