# go-future

[![codecov](https://codecov.io/github/jizhuozhi/go-future/graph/badge.svg?token=9UZDVRZCQM)](https://codecov.io/github/jizhuozhi/go-future)
[![goreport](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)

**go-future** is a lightweight, high-performance, lock-free Future/Promise implementation for Go, built with modern concurrency in mind. It supports:

- Asynchronous task execution (`Async`, `CtxAsync`)
- ~~Lazy evaluation (`Lazy`)~~(Deprecated, will be removed in later version)
- Promise resolution (`Promise`)
- Event-driven callback registration (`Subscribe`)
- Functional chaining (`Then`, `ThenAsync`)
- Task composition (`AllOf`, `AnyOf`)
- Timeout control (`Timeout`, `Until`)
- Full support for Go generics

## 🔧 Installation

```bash
go get github.com/jizhuozhi/go-future
````

---

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	"github.com/jizhuozhi/go-future"
)

func main() {
	p := future.NewPromise[string]()
	go func() {
		p.Set("hello", nil)
	}()
	val, err := p.Future().Get()
	fmt.Println(val, err) // Output: hello <nil>
}
```

---

## 🧠 Core Concepts

### Promise and Future

* `Promise` is the **producer**, which sets the value once.
* `Future` is the **consumer**, which retrieves the result asynchronously.

Every Future is backed by a lock-free internal state. All state transitions are safe and efficient under high concurrency.

---

## 🔨 Key APIs

### `Async(func() (T, error)) *Future[T]`

Starts a new asynchronous task in a goroutine.

```go
f := future.Async(func() (string, error) {
	return "result", nil
})
val, err := f.Get()
```

---

### `Lazy(func() (T, error)) *Future[T]`

Returns a future that is lazily evaluated. The function is only executed once on the first call to `.Get()`.

```go
f := future.Lazy(func() (string, error) {
	fmt.Println("evaluated")
	return "lazy", nil
})
val, _ := f.Get() // prints "evaluated"
```

#### ⚠️ Deprecation Notice: Lazy
The Lazy API is deprecated and will be removed in future versions.

Although Lazy provides deferred execution semantics, it introduces implicit pull-based dependencies that are difficult to reason about in practice. Unlike Async, where execution is guaranteed upon construction, Lazy defers execution until .Get() is called — often far away from the site of definition.

This subtle semantic difference:
- Makes control flow harder to predict
- Breaks intuitive data dependency modeling
- Can result in unexpected bugs when chained or composed in concurrent settings

Recommendation: Prefer using Async or Promise for all use cases. These are explicit and deterministic.

---

### `Promise[T]`

Used to create and control a future manually.

```go
p := future.NewPromise[int]()
go func() {
	p.Set(42, nil)
}()
val, _ := p.Future().Get()
```

---

### `Then(f *Future[T], cb func(T, error) (R, error)) *Future[R]`

Chains computations synchronously.

```go
f := future.Async(func() (int, error) { return 1, nil })
f2 := future.Then(f, func(v int, err error) (string, error) {
	return fmt.Sprintf("num:%d", v), err
})
result, _ := f2.Get()
```

---

### `ThenAsync(f *Future[T], cb func(T, error) *Future[R]) *Future[R]`

Chains computations with asynchronous return.

```go
f := future.Async(func() (int, error) { return 1, nil })
f2 := future.ThenAsync(f, func(v int, err error) *future.Future[string] {
	return future.Async(func() (string, error) {
		return fmt.Sprintf("async:%d", v), nil
	})
})
result, _ := f2.Get()
```

---

### `AllOf(fs ...*Future[T]) *Future[[]T]`

Waits for all futures to complete successfully. Fails fast on the first error.

```go
f1 := future.Async(func() (int, error) { return 1, nil })
f2 := future.Async(func() (int, error) { return 2, nil })
fAll := future.AllOf(f1, f2)
vals, _ := fAll.Get() // [1, 2]
```

---

### `AnyOf(fs ...*Future[T]) *Future[AnyResult[T]]`

Returns the first successful result. If all fail, returns the first error.

```go
f1 := future.Async(func() (int, error) { return 0, fmt.Errorf("fail") })
f2 := future.Async(func() (int, error) { return 2, nil })
res, _ := future.AnyOf(f1, f2).Get()
// res.Index == 1, res.Val == 2
```

---

### `Timeout(f *Future[T], d time.Duration) *Future[T]`

Wraps a future and fails with `ErrTimeout` if not resolved in time.

```go
f := future.Async(func() (int, error) {
	time.Sleep(2 * time.Second)
	return 42, nil
})
val, err := future.Timeout(f, time.Second).Get()
// err == future.ErrTimeout
```

---

### `Done(val T) *Future[T]`, `Done2(val T, err error)`

Create a completed Future.

```go
f := future.Done("value")
f2 := future.Done2("value", nil)
```

---

### `Subscribe(cb func(T, error))`

Registers a callback that runs when the Future is done.

```go
f := future.Async(func() (int, error) { return 1, nil })
f.Subscribe(func(v int, err error) {
	fmt.Println("got:", v)
})
```

> ⚠️ Callbacks execute **in the same goroutine** that completes the Future. Avoid blocking operations in the callback.

---

## ✅ Advantages

* **Zero Locking:** Internals are implemented using atomic state machines, not `sync.Mutex`.
* **Type Safe:** Full support for Go generics.
* **No Goroutine Bloat:** Except `Async`, all operations are event-driven, avoiding extra goroutines.
* **Composable:** Easily chainable, supports DAG-like workflows.

---

## 📊 Benchmark

```text
goos: darwin
goarch: arm64
pkg: github.com/jizhuozhi/go-future
Benchmark/Promise           3.05M	    377 ns/op
Benchmark/WaitGroup         2.88M	    424 ns/op
Benchmark/Channel           3.00M	    399 ns/op
```

> `Promise` is competitive with `sync.WaitGroup` and `channel` in terms of performance and offers much better composition semantics.

---

## 📦 DAG Execution Engine (Experimental)

Starting from v0.1.4, `go-future` introduces a powerful **DAG (Directed Acyclic Graph) execution engine**, consisting of:

* `dagcore`: A minimal, lock-free parallel DAG scheduler
* `dagfunc`: A high-level builder that constructs DAGs using Go function signatures with type-based dependency resolution

This enables users to describe complex data flow graphs declaratively with automatic dependency wiring and parallel execution.

### Use Case

* AI model composition pipelines
* Request processing graphs
* Asynchronous task orchestration

### Example

```go
package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jizhuozhi/go-future/dagfunc"
)

func main() {
	b := dagfunc.New()

	// Provide input type
	_ = b.Provide("")

	// Register a function that depends on input
	_ = b.Use(func(ctx context.Context, s string) (int, error) {
		return len(s), nil
	})

	// Compile and execute
	prog, _ := b.Compile([]any{"hello world"})
	outputs, _ := prog.Run(context.Background())

	// Get output value by using zero-value of type as key
	fmt.Println("result:", outputs[0]) // => 11
}
```

### Avoid Type Conflicts with Type Aliases

`dagfunc` uses Go types to wire dependencies. This means if you have two different inputs or outputs with the same type (e.g., multiple `string`s), you **must disambiguate** them using **type aliases**.

#### Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/jizhuozhi/go-future/dagfunc"
)

// Define type aliases to distinguish values
type UserID string
type Greeting string

func main() {
	b := dagfunc.New()

	// Register inputs with distinct types
	_ = b.Provide(UserID(""))

	// Step 1: Get greeting from user ID
	_ = b.Use(func(ctx context.Context, uid UserID) (Greeting, error) {
		return Greeting("hello, " + string(uid)), nil
	})

	// Step 2: Convert to string for output
	_ = b.Use(func(ctx context.Context, g Greeting) (string, error) {
		return string(g), nil
	})

	prog, _ := b.Compile([]any{UserID("Alice")})
	out, _ := prog.Run(context.Background())

	// Step 3: Get output value by using zero-value of type as key 
	fmt.Println(out[""]) // => hello, Alice
}
```

By aliasing `string` to `UserID` and `Greeting`, we allow the builder to differentiate between them — enabling type-safe, unambiguous wiring.

### Advanced

You can also retrieve individual results using:

```go
res, err := prog.Get(UserID(""))
```

### Internal Architecture

* `dagcore.DAG`: Defines static DAG structure and runtime scheduling semantics
* `dagfunc.Builder`: Wraps `dagcore` and infers DAG wiring from function types
* Each node is executed in parallel once dependencies are met, using `Future`

## 🔐 License

Apache-2.0 license by [jizhuozhi](https://github.com/jizhuozhi)