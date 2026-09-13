# go-future

[![codecov](https://codecov.io/github/jizhuozhi/go-future/graph/badge.svg?token=9UZDVRZCQM)](https://codecov.io/github/jizhuozhi/go-future)
[![goreport](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)

**go-future** is a lightweight, high-performance, mutex-free Future/Promise implementation for Go, built with modern concurrency in mind. It supports:

- Asynchronous task execution (`Async`, `CtxAsync`)
- Promise resolution (`Promise`)
- Event-driven callback registration (`Subscribe`)
- Functional chaining (`Then`, `ThenAsync`)
- Task composition (`AllOf`, `AnyOf`)
- Timeout control (`Timeout`, `Until`)
- Full support for Go generics
- **Fluent combinators as generic methods** (`Then[R]`, `Map[R]`, `FlatMap[R]`, `Cast[R]`, ...) — Go 1.27+

## 🔧 Requirements

**Go 1.18 or newer.** The module declares `go 1.18` and stays compatible with every release since.

The generic methods (`Then[R]`, `Map[R]`, `Cast[R]`, …) are an optional enhancement compiled only when the toolchain is **Go 1.27 or newer**, guarded by `//go:build go1.27`. On older toolchains the package-level functions provide the same capabilities, so no upgrade is ever required. See [Go 1.27 release notes](https://go.dev/doc/go1.27).

## 🔧 Installation

```bash
go get github.com/jizhuozhi/go-future
````

---

## ⬆️ Migrating to v0.2.0

v0.2.0 adds generic methods on top of the existing API. Exactly one group of symbols moved; everything else is additive and **no Go upgrade is required**.

### 1. Tuples moved to the `tuples` subpackage (breaking)

`Tuple2`…`Tuple16` and `Of2`…`Of16` are rarely used and added 30 exported
symbols to the main package, so they now live in their own package:

| v0.1.6                       | v0.2.0                                                       |
| ---------------------------- | ------------------------------------------------------------ |
| `future.Of2(f0, f1)`         | `tuples.Of2(f0, f1)`                                         |
| `future.Tuple2[A, B]`        | `tuples.Tuple2[A, B]`                                        |

```go
import "github.com/jizhuozhi/go-future/tuples"

f := tuples.Of2(fetchUser(), fetchAccount())
t, err := f.Get()
```

There is deliberately no alias left behind in `future`: `tuples` imports
`future`, so keeping one would create an import cycle.

### 2. No Go upgrade required

The module still declares `go 1.18`. The generic methods live in files guarded by `//go:build go1.27`, which tests the **toolchain** version rather than the `go` directive, so they light up automatically without touching `go.mod`.

| Toolchain      | What compiles                                             |
| -------------- | --------------------------------------------------------- |
| Go 1.18 – 1.26 | package-level functions only: `Then(f, cb)`, `Timeout(f, d)`, … |
| Go 1.27+       | both forms: `f.Then(cb)` **and** `Then(f, cb)`            |

### 3. Package-level functions are not deprecated

They remain a fully supported API, not a compatibility shim. The two forms do **not** delegate to each other, by design: the package-level functions must not depend on a file that can be excluded from the build, so each variant carries its own (small, self-contained) body.

| Function           | Method form       |
| ------------------ | ----------------- |
| `Then(f, cb)`      | `f.Then(cb)`      |
| `ThenAsync(f, cb)` | `f.ThenAsync(cb)` |
| `ToAny(f)`         | `f.ToAny()`       |
| `ToChan(f)`        | `f.ToChan()`      |
| `Timeout(f, d)`    | `f.Timeout(d)`    |
| `Until(f, t)`      | `f.Until(t)`      |
| `Await(f)`         | `f.Get()`         |

`AllOf`, `AnyOf`, `Async`, `Done`, `NewPromise` and all `Promise` / `Future` methods are unchanged.

### New in v0.2.0 (Go 1.27+)

* Single-Future transforms as generic methods: `Then[R]`, `ThenAsync[R]`, `ThenGo[R]`, `Map[R]`, `FlatMap[R]`, `Cast[R]`, `Recover`, `OrElse`.
* `dagcore.NodeInstance.Cast[T]` and `dagfunc.Program.Value[T]` / `ValueAsync[T]` for reading `any`-typed results without assertions.
* `future.ErrTypeMismatch` for failed `Cast` conversions.

`dagfunc.Program.Get(sample any)` is **unchanged** — `Value[T]` is a type-safe alternative, not a replacement.

### New in v0.2.0: dagfunc reaches dagcore parity (additive)

`dagfunc` used to be a thin type-driven wrapper: one node per result type, one
accepted signature, no way to name a node or to wrap it. It now covers what
`dagcore` offers while keeping the type-based wiring:

* `Provide` and `Use` take options: `Name`, `After`, `From[T]`, `Timeout`,
  `Retry`, `Recover[R]`, `OrElse[R]`, `Wrap`, `Default`.
* `Use` accepts functions without a context, without an error, with several
  results, with no result at all, and functions returning a `*future.Future[R]`.
* `Subgraph` embeds a `Builder` as a node, `Group` namespaces nodes.
* `Compile` takes `dagcore` wrappers; `Program` exposes `Node`, `NodeByID` and
  `Instance`.

Everything that compiled before still compiles — `Provide(v)`, `Use(fn)`,
`Compile(inputs)`, `Run`, `Get(sample)` and `Value[T]` keep their signatures.
Two behaviours changed on purpose:

| Before                                                | Now                                                        |
| ----------------------------------------------------- | ---------------------------------------------------------- |
| `func() (int, error)` rejected with `ErrFuncSignature`  | valid: a source node with no context and no dependencies   |
| non-comparable output panicked when building `Run`'s map | skipped in the map, still readable with `Get` / `Value[T]` |

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

Every Future is backed by a single atomic word: each state transition is one compare-and-swap on that word, and waiters park on a semaphore. No mutex is taken on any path.

### Why a semaphore, not `sync.Cond`

`Get` parks a goroutine until the result is published. The textbook tool for that is a `sync.Cond`, but a `Cond` needs a `Locker`: waiting means unlocking, parking and locking again. `sync.Mutex` is adaptive — it spins, then backs off, then parks on a semaphore of its own — and that escalation pays for itself only when the critical section is short. An asynchronous task is a heavy operation, so the wait is long by nature and the mutex machinery would be overhead spent for nothing.

`go-future` therefore skips the mutex and drives the semaphore directly, keeping only the part a condition variable is actually needed for here: a queue of parked goroutines for `Set` to hand off to. There is no spin phase and no lock upgrade on the wait path.

The cost is that the semaphore is reached through `//go:linkname` rather than the public API. The Go source is candid about this — `runtime/sema.go` publishes both symbols with an explicit `//go:linkname` push and carries this note:

```text
sync_runtime_Semacquire should be an internal detail,
but widely used packages access it using linkname.
Notable members of the hall of shame include:
  - gvisor.dev/gvisor
  - github.com/sagernet/gvisor

Do not remove or change the type signature.
See go.dev/issue/67401.
```

The push makes this the handshake form rsc describes as the desired end state in [go.dev/issue/67401](https://go.dev/issue/67401), rather than an unauthorised pull, and "Do not remove or change the type signature" is a commitment the Go team has made. The note also records who else depends on it: gvisor.

---

## 🔨 Key APIs

### 🧭 Layering rule

One rule decides whether an operation is a package-level function or a method:

| Kind                       | Form        | Members                                                                                                          |
| -------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------- |
| Constructor                | function    | `Async`, `CtxAsync`, `Submit`, `CtxSubmit`, `Done`, `Done2`, `NewPromise`                                        |
| Single-Future transform    | **method**  | `Then`, `ThenAsync`, `ThenGo`, `Map`, `FlatMap`, `Cast`, `Recover`, `OrElse`, `ToAny`, `ToChan`, `Timeout`, `Until` |
| Combinator over N Futures  | function    | `AllOf`, `AnyOf`                                                                                                  |

A transform that changes the result type must introduce a type parameter of its own, so until Go 1.27 it could only live at package scope. Since Go 1.27 it can be a method, and the method form is what you want to reach for: it chains left-to-right and it is discoverable through completion.

Both forms coexist. The methods are compiled only under a Go 1.27+ toolchain (`//go:build go1.27`); the package-level functions are always available, which is what keeps Go 1.18 working.

A combinator over several Futures has no single receiver, and the language forbids a generic method from returning its receiver's type instantiated with a receiver-derived type (see the restrictions below) — so combinators stay functions. Java, Scala and friends draw the same line: "zip N futures into one" is a static/companion function there as well.

`AllOf` / `AnyOf` cover batches of Futures that share a type. Zipping Futures of
**unrelated** types needs one function per arity, so `Tuple2`…`Tuple16` and
`Of2`…`Of16` live in the [`tuples`](./tuples) subpackage — they are rarely
needed and would otherwise add 30 exported symbols to the main package:

```go
import "github.com/jizhuozhi/go-future/tuples"

f := tuples.Of2(fetchUser(), fetchAccount())
t, err := f.Get() // t.Val0 is a User, t.Val1 is an Account
```

---

### `Async(func() (T, error)) *Future[T]`

Starts a new asynchronous task in a goroutine.

```go
f := future.Async(func() (string, error) {
	return "result", nil
})
val, err := f.Get()
```

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

### Single-Future transforms — `*Future[T]` methods

> **Requires a Go 1.27+ toolchain.** These files are guarded by `//go:build go1.27`. On Go 1.18–1.26 the package-level functions below provide the same capabilities.

Go 1.27 lets a **method declare its own type parameters**, so transforms live on
`*Future[T]` and a pipeline reads left-to-right instead of inside-out.

```go
n, err := future.Async(fetchUser).
    Map(func(u User) string { return u.Email }).
    FlatMap(func(email string) *future.Future[Profile] { return future.Async(fetch(email)) }).
    Then(func(p Profile, err error) (int, error) {
        if err != nil {
            return 0, err
        }
        return len(p.Friends), nil
    }).
    Get()
```

| Method                                         | Description                                                                                                   |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `Then[R](func(T, error) (R, error))`            | Chains a synchronous step. Runs on the goroutine that completes the Future, so it must not block.              |
| `ThenAsync[R](func(T, error) *Future[R])`       | Chains an asynchronous step and flattens the nested Future.                                                    |
| `ThenGo[R](func(T, error) (R, error))`          | Like `Then`, but the callback is dispatched to the `Executor` so it may block. Panics are captured as `ErrPanic`. |
| `Map[R](func(T) R)`                             | Transforms the success value; on failure `fn` is skipped and the error is propagated.                          |
| `FlatMap[R](func(T) *Future[R])`                | `Map` + flatten; on failure `fn` is skipped and the error is propagated.                                       |
| `Cast[R]()`                                     | Reinterprets the Future as `*Future[R]` via a dynamic type assertion. Fails with `ErrTypeMismatch`.            |
| `Recover(func(error) (T, error))`               | Turns a failure into a success (or into another error).                                                        |
| `OrElse(T)`                                     | Replaces a failure with a default value.                                                                       |
| `ToAny()`, `ToChan()`, `Timeout(d)`, `Until(t)` | Interop and deadline helpers.                                                                                  |

`Cast` is what makes the `map[any]any` style results of `dagcore` / `dagfunc`
usable without type assertions:

```go
count, err := inst.Nodes()["func:TokenCount"].Cast[TokenCount]().Get()
```

#### ⚠️ Two language restrictions

1. **Interface methods may not declare type parameters, and a generic method
   cannot implement an interface method.** `Future` is therefore a concrete type
   by design; the transforms above are resolved statically and are deliberately
   not exposed through an interface.
2. **A generic method must not return its receiver's type instantiated with a
   type built from the receiver's type parameter.** The following looks natural
   but the compiler rejects it with `instantiation cycle`:

   ```go
   // func (f *Future[T]) Combine[R any](g *Future[R]) *Future[tuples.Tuple2[T, R]]
   ```

   Type-checking `Future[T]` would require `Future[Tuple2[T, R]]`, then
   `Future[Tuple2[Tuple2[T, R], R]]` and so on forever. Combining several
   Futures is exactly what the package-level combinators are for.

---

### Package-level transforms (all Go versions)

These are the original API and the only form available on Go 1.18–1.26. They are **not** deprecated, and they do not delegate to the methods above: each variant carries its own implementation so that the Go 1.18 build has no dependency on a file that may be excluded from it.

| Function           | Method form (Go 1.27+) |
| ------------------ | ---------------------- |
| `Then(f, cb)`      | `f.Then(cb)`           |
| `ThenAsync(f, cb)` | `f.ThenAsync(cb)`      |
| `ToAny(f)`         | `f.ToAny()`            |
| `ToChan(f)`        | `f.ToChan()`           |
| `Timeout(f, d)`    | `f.Timeout(d)`         |
| `Until(f, t)`      | `f.Until(t)`           |
| `Await(f)`         | `f.Get()`              |

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

### `Timeout(d time.Duration) *Future[T]` / `Until(t time.Time) *Future[T]`

Fails with `ErrTimeout` if the Future is not resolved in time.

```go
f := future.Async(func() (int, error) {
	time.Sleep(2 * time.Second)
	return 42, nil
})
val, err := f.Timeout(time.Second).Get()
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

* **Mutex-free:** Every state transition is a compare-and-swap on a single atomic word, and waiters park on a semaphore. No lock is taken on any path.
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

# 📦 DAG Execution Engine (Experimental)

Starting from v0.1.4, `go-future` introduces a powerful **DAG (Directed Acyclic Graph) execution engine**, consisting of:

* `dagcore`: A minimal parallel DAG scheduler with lock-free dependency tracking
* `dagfunc`: A high-level builder that constructs DAGs using Go function signatures with type-based dependency resolution

This enables users to describe complex data flow graphs declaratively with automatic dependency wiring and parallel execution.

## dagcore

`dagcore` is the low-level DAG execution engine powering [`go-future`](https://github.com/jizhuozhi/go-future)'s structured concurrency and dataflow execution model. It provides a dependency-driven scheduler with lock-free dependency tracking for executing static DAGs (Directed Acyclic Graphs) in parallel.

### ✨ Features

* ⚡ **Lock-free scheduling** via atomic dependency counters
* ⛓️ **Supports any static DAG with arbitrary fan-in/out structure**
* 🔁 **Exactly-once execution**: each node runs exactly once after its dependencies complete
* 🧠 **On-demand scheduling**: nodes are only triggered once all dependencies complete — goroutines are created only when the node is ready to run
* ❌ **Fast failure support**: optional early cancellation on error (fail-fast mode)
* ⏱️ **Context propagation**: full support for timeout and cancellation via `context.Context`
* 🧩 **Composable foundation**: designed for embedding in higher-level DAG builders (e.g. `dagfunc`)
* 📈 **Metrics & logging hooks**: supports per-node wrappers for observability (e.g. retry, timing, logging)

---

### 🚀 Example Usage

```go
dag := dagcore.NewDAG()

// Define DAG structure
_ = dag.AddInput("A")
_ = dag.AddNode("B", []dagcore.NodeID{"A"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
    return deps["A"].(int) + 2, nil
})
_ = dag.AddNode("C", []dagcore.NodeID{"A"}, func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
    return deps["A"].(int) * 3, nil
})

// Verifies that the graph is complete and acyclic, 
// then locks the structure to make it immutable for repeated safe instantiations.
if err := dag.Freeze(); err != nil {
	return err
}

// Execute
inst, _ := dag.Instantiate(map[dagcore.NodeID]any{"A": 10})
res, _ := inst.Execute(context.Background())
fmt.Println("B:", res["B"], "C:", res["C"])
```

---

### 🧠 Execution Model

Each node in the DAG:

* Declares its dependencies via `AddNode(id, deps, func)`
* Executes only once **after all its inputs are ready**
* Will not allocate any goroutine until scheduled — **on-demand execution**
* May run in parallel with other ready nodes
* Propagates failures down dependent nodes (fail-fast)

Internally:

* Uses atomic counters to track pending dependencies per node
* Uses `future.Future` to propagate results, cancellation, and errors
* Can be fully composed and integrated with the rest of `go-future`

---

### ⚙️ API Overview

#### `dagcore.NewDAG() *DAG`

Creates a new empty DAG instance.

#### `(*DAG).AddInput(id NodeID) error`

Adds a node that must be externally provided during execution.

#### `(*DAG).AddNode(id NodeID, deps []NodeID, fn NodeFunc) error`

Adds a computational node with declared dependencies.

#### (*DAG).Freeze() error

**Freezes the DAG topology.** Verifies that the graph is complete and acyclic, then locks the structure to make it immutable for repeated safe instantiations.

```go
dag := dagcore.NewDAG()
// Add nodes...
_ = dag.Freeze()
```

> You must call Freeze() before Instantiate or Run. Once frozen, the DAG can be instantiated and executed multiple times in parallel.

#### `(*NodeInstance).Cast[T any]() *future.Future[T]`

Returns the result of a single node as a typed Future, using the generic methods introduced in Go 1.27. Fails with `future.ErrTypeMismatch` if the produced value is not assignable to `T`.

> **Requires a Go 1.27+ toolchain** — the file is guarded by `//go:build go1.27`. Use `NodeInstance.Future()` and assert manually on older versions.

```go
count, err := inst.Nodes()["func:TokenCount"].Cast[TokenCount]().Get()
```

#### `(*DAG).Instantiate(inputs map[NodeID]any, wrappers ...NodeFuncWrapper) (*DAGInstance, error)`

Creates a runtime instance of the DAG for execution.

#### `(*DAGInstance).Run(ctx context.Context) (map[NodeID]any, error)`

Executes all nodes and returns the final results.

#### `(*DAGInstance).RunAsync(ctx context.Context) *Future[map[NodeID]any]`

Runs asynchronously and returns a future.

---

### 🔧 Advanced Features

#### NodeFunc Wrapping

Use `NodeFuncWrapper` to wrap node logic for tracing, logging, retries, etc:

```go
dag.Instantiate(inputs, func(n *dagcore.NodeInstance, fn dagcore.NodeFunc) dagcore.NodeFunc {
    return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
        start := time.Now()
        out, err := fn(ctx, deps)
        log.Printf("node %s took %s", n.ID(), time.Since(start))
        return out, err
    }
})
```

#### Mermaid Graph Output

Convert the DAG to a [Mermaid.js](https://mermaid-js.github.io/) compatible graph string:

```go
fmt.Println(dagcore.ToMermaid(instance))
```

---

### 🧱 Designed for Composition

`dagcore` is intended to be embedded in high-level tools:

* [`dagfunc`](../dagfunc): type-safe DAG builder with Go function signature inference
* Custom domain-specific orchestrators, AI pipelines, CI/CD workflows
* Any static dependency graph evaluation with result propagation

## dagfunc

`dagfunc` is a high-level, type-safe DAG (Directed Acyclic Graph) builder built on top of [`dagcore`](../dagcore). It allows you to define and execute dependency graphs by simply wiring Go functions based on their parameter and return types.

### 🚀 What It Solves

`dagfunc` abstracts away manual DAG construction by:

- Automatically inferring node dependencies from function signatures
- Resolving dependency order at compile-time
- Mapping types to results without needing manual wiring

Ideal for:

- AI agent pipelines
- Asynchronous service orchestration
- Task graph composition with clear dependency semantics

---

### ✅ Features

Wiring

- ✅ Type-based dependency inference
- ✅ Every Go function shape that reduces to "dependencies in, results out"
- ✅ Nodes with several results, and nodes with no result at all
- ✅ Nodes returning a `*future.Future[R]`, for already asynchronous APIs
- ✅ Explicit node ids, plus namespaces through `Group`
- ✅ Several nodes of the same type, selected with `From[T](id)`

Composition

- ✅ Subgraphs: a whole `Builder` embedded as a single node, nested arbitrarily
- ✅ Per-node and per-run `dagcore` wrappers, for tracing, metrics and logging
- ✅ Per-node `Timeout`, `Retry`, `Recover` and `OrElse`
- ✅ Optional inputs through `Default`, and input binding by node id

Execution

- ✅ Fully integrated with `go-future` for parallel execution
- ✅ Built-in support for context propagation and error handling
- ✅ Typed result access without sample values or assertions (`Value[T]`)

---

### 🔧 Installation

```bash
go get github.com/jizhuozhi/go-future/dagfunc
````

---

### ✨ Example

```go
package main

import (
	"context"
	"fmt"
	"github.com/jizhuozhi/go-future/dagfunc"
)

func main() {
	type Input string
	type TokenCount int

	b := dagfunc.New()

	// Step 1: Declare input
	_ = b.Provide(Input(""))

	// Step 2: Register function
	_ = b.Use(func(ctx context.Context, text Input) (TokenCount, error) {
		return TokenCount(len(text)), nil
	})

	// Step 3: Verifies that the graph is complete and acyclic, 
	// then locks the structure to make it immutable for repeated safe instantiations.
	if err := b.Freeze(); err != nil {
		panic(err)
    }
	
	// Step 4: Run
	prog, _ := b.Compile([]any{Input("hello world")})
	_, _ = prog.Run(context.Background())

	// Step 5: Read a typed result (generic method, Go 1.27+)
	count, _ := prog.Value[TokenCount]()
	fmt.Println(count) // Output: 11
}
```

---

### 🧠 Type-Based Wiring

`dagfunc` determines node dependencies using **parameter types** and result types:

* Inputs and outputs must use unique Go types or **aliases**
* The DAG will automatically determine execution order

> ⚠️ If two inputs/outputs are of the same type (e.g., multiple `string` values), use `type alias` to disambiguate — or give the nodes a `Name` and select one with `From[T](id)`.

#### With type alias

```go
type UserID string
type Greeting string

b.Provide(UserID(""))
b.Use(func(ctx context.Context, uid UserID) (Greeting, error) {
	return Greeting("Hello, " + string(uid)), nil
})
b.Use(func(ctx context.Context, g Greeting) (string, error) {
	return string(g), nil
})
```

#### Accepted function signatures

`Use` accepts every shape that reduces to "take dependencies, produce results":

| Signature                                        | Meaning                                     |
| ------------------------------------------------ | ------------------------------------------- |
| `func(ctx, A, B) (R, error)`                     | canonical form                              |
| `func(ctx, A, B) R`                              | cannot fail                                 |
| `func(ctx, A, B) error`                          | sink, produces no value                     |
| `func(ctx, A, B) (R, S, error)`                  | several results                             |
| `func(ctx, A, B) (R, S)`                         | several results, cannot fail                |
| `func(A, B) (R, error)`                          | no context                                  |
| `func(ctx) (R, error)`                           | source, no dependencies                     |
| `func(ctx, A) *future.Future[R]`                 | asynchronous node                           |

The trailing `error` is optional and always last; every other result becomes an output of the node and can be depended on by its type. A node returning `*future.Future[R]` waits for that Future instead of computing `R` inline.

Three rules decide what counts as an error:

* **is-a, not as-a**: only a result declared as exactly `error` is an error. A named type that merely implements it is an ordinary value, so `func(ctx) (Result, MyError)` produces two values and can never fail, while `func(ctx) (Result, error)` produces one value and may fail. What the signature says is what the graph does.
* **at most one error, and it has to be the last result**: `(R, error, error)` and `(error, R)` are rejected with `ErrFuncSignature` instead of being guessed at.
* **a failed node publishes nothing**, not even the values it computed before failing.

#### Node ids

Every node has a string id, used by `NodeByID`, `Wrap` and `dagviz`:

| Node                | Default id                        |
| ------------------- | --------------------------------- |
| `Provide(T{})`      | `input:<type>`                    |
| `Use` with results  | `func:<type>`, or `func:<t1>,<t2>` |
| `Use` without results | `func:<symbol>`                 |
| `Name("x")`         | `x`                               |
| inside `Group("g")` | `g.<default>`                     |

---

### 🧰 API Overview

#### `dagfunc.New() *Builder`

Creates a new DAG builder.

#### `(*Builder).Provide(sample any, opts ...Option) error`

Declares an input node of the type of `sample`; the value itself is supplied at
`Compile`. `Default(val)` makes it optional, `Name(id)` gives it an explicit id.

#### `(*Builder).Use(fn any, opts ...Option) error`

Registers a function as a DAG node. See
[accepted function signatures](#accepted-function-signatures).

#### `(*Builder).Subgraph(sub *Builder, opts ...Option) error`

Embeds a whole builder as a single node. The inputs of `sub` are fed from the
nodes of the parent producing those types (or from their `Default`), and
`Outputs(...)` declares which results the parent can read:

```go
qa := dagfunc.New()
_ = qa.Provide(Question(""))
_ = qa.Use(retrieve)   // func(ctx, Question) (Candidate, error)
_ = qa.Use(rerank)     // func(ctx, Question, Candidate) (Ranked, error)

root := dagfunc.New()
_ = root.Provide(Question(""))
_ = root.Subgraph(qa, dagfunc.Name("qa"), dagfunc.Outputs(Ranked{}))
_ = root.Use(answer)   // func(ctx, Question, Ranked) (Answer, error)
```

The subgraph keeps its own scheduler, so its nodes still run in parallel, and
subgraphs nest. The parent freezes it, so `qa.Freeze()` is optional.

#### `(*Builder).Group(ns string, define func(*Builder) error) error`

Registers every node of `define` under the namespace `ns`. It is only a naming
device: the type registry is shared, so the nodes stay part of the same graph.

#### Options

| Option                       | Applies to                    | Effect                                              |
| ---------------------------- | ----------------------------- | --------------------------------------------------- |
| `Name(id)`                   | all                           | explicit node id                                    |
| `After(deps...)`             | all                           | ordering-only dependency, by id or by sample type   |
| `From[T](id)`                | `Use`                         | pins one parameter to a specific node               |
| `Timeout(d)`                 | `Use`                         | bounds the context handed to the node               |
| `Retry(n)` / `RetryWith(n, d)` | `Use`                       | reruns the node after a failure                     |
| `Recover[R](fn)`             | `Use` (one result)            | turns a failure into a value                        |
| `OrElse[R](val)`             | `Use` (one result)            | replaces a failure with a value                     |
| `Wrap(w...)`                 | all                           | `dagcore` wrappers for this node only               |
| `Default(val)`               | `Provide`                     | value used when `Compile` gets none                 |
| `Outputs(samples...)`        | `Subgraph`                    | results exposed to the parent                       |

The built-ins are layered from the inside out as retry, timeout, fallback, and
`Wrap` wrappers always observe the final result:

```go
_ = b.Use(search,
    dagfunc.Name("search"),
    dagfunc.Timeout(200*time.Millisecond),
    dagfunc.RetryWith(2, 10*time.Millisecond),
    dagfunc.OrElse[Result](Result{}),
)
```

#### `(*Builder).Freeze() error` / `(*Builder).Compile(inputs []any, wrappers ...dagcore.NodeFuncWrapper) (*Program, error)`

`Freeze` verifies the graph and locks it. `Compile` binds the inputs — matched
by type, or by node id with `dagfunc.Input(id, val)` — and returns a `Program`.
The variadic wrappers apply to every node of that run and are the hook for
tracing, metrics and logging, exactly like `dagcore.Instantiate`:

```go
prog, err := b.Compile(inputs, func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
    return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
        start := time.Now()
        out, err := run(ctx, deps)
        log.Printf("%s took %s", n.ID(), time.Since(start))
        return out, err
    }
})
```

#### `(*Program).Run(ctx context.Context) (map[any]any, error)`

Executes the DAG. Outputs are keyed by result types with typed zero. Output
types that are not comparable (slices, maps) cannot be map keys and are omitted;
read them with `Value[T]`, `Get` or `Node`.

#### `(*Program).RunAsync(ctx context.Context) *future.Future[map[any]any]`

Executes the DAG. Return a future with outputs are keyed by result types with typed zero.

#### `(*Program).Get(sample any) (any, error)`

Gets the result value for a specific type. The sample is only used to determine
the type; the caller has to assert the result.

```go
count, err := prog.Get(TokenCount(0)) // count is an any, needs an assertion
```

#### `(*Program).Value[T any]() (T, error)`

> **Requires a Go 1.27+ toolchain**, as do `ValueAsync[T]`. `Get(sample any)` is the version-independent alternative.

Gets the typed result value produced by the node whose result type is `T`.
Blocked until that node completes, so `Run` / `RunAsync` must have been called.

```go
count, err := prog.Value[TokenCount]() // no sample value, no type assertion
```

#### `(*Program).ValueAsync[T any]() *future.Future[T]`

Same as `Value[T]` but non-blocking: it can be subscribed to before the DAG is
started. Fails with `ErrTypeNotFound` when no node produces `T`, with
`ErrAmbiguousType` when several do, and with `future.ErrTypeMismatch` when the
produced value is not assignable to `T`.

#### `(*Program).Node(sample any) (*dagcore.NodeInstance, error)` / `(*Program).NodeByID(id string) (*dagcore.NodeInstance, bool)`

Returns the runtime node behind a type or an id, which carries its `Future`, its
`Duration`, its dependencies and, on Go 1.27, `Cast[T]`. `NodeByID` is how the
results of two same-typed nodes are told apart.

#### `(*Program).Instance() *dagcore.DAGInstance`

Hands over the whole runtime, which is what `dagviz` renders:

```go
fmt.Println(dagviz.ToMermaid(prog.Instance()))
```

#### Error propagation

* DAG execution will **fail fast** by default
* Downstream nodes will not be executed if inputs fail
* A node can opt out with `Recover` / `OrElse`, or with a `dagcore` wrapper

---

### 🧩 Relationship to dagcore

| Layer     | Role                                      |
| --------- | ----------------------------------------- |
| dagfunc   | High-level: Build DAGs from Go functions  |
| dagcore   | Low-level: Execute DAGs with scheduling   |
| go-future | Runtime: Power async execution via Future |

---

### 💡 Use Cases

* LLM / Agent planning pipelines
* Microservice DAG invocation
* Declarative orchestration of business logic
* Build systems / task runners

---

### 📌 Notes

* Outputs are retrieved by Go types, either through the generic methods
  `Value[T]()` / `ValueAsync[T]()` or, for the raw map, by typed zero value
* `Run` returns `map[any]any`, so every entry still needs an assertion; prefer
  `Value[T]()` which does the assertion for you and reports `ErrTypeMismatch` on
  mismatch
* Type aliasing, `Name` or `From[T](id)` is required for disambiguation
* All dependencies must be resolvable before `Freeze`
* `dagfunc` is a layer over `dagcore`, not a subset of it: naming, per-node
  wrappers, subgraphs and defaults are reachable without dropping down, and
  `Program.Instance()` hands over the `dagcore` runtime when needed

## 🔐 License

Apache-2.0 license by [jizhuozhi](https://github.com/jizhuozhi)