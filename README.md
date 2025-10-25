# go-future

[![codecov](https://codecov.io/github/jizhuozhi/go-future/graph/badge.svg?token=9UZDVRZCQM)](https://codecov.io/github/jizhuozhi/go-future)
[![goreport](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)

**go-future** is a lightweight, high-performance, lock-free Future/Promise implementation for Go, built with modern concurrency in mind. It supports:

- Asynchronous task execution (`Async`, `CtxAsync`)
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

# 📦 DAG Execution Engine (Experimental)

Starting from v0.1.4, `go-future` introduces a powerful **DAG (Directed Acyclic Graph) execution engine**, consisting of:

* `dagcore`: A minimal, lock-free parallel DAG scheduler
* `dagfunc`: A high-level builder that constructs DAGs using Go function signatures with type-based dependency resolution

This enables users to describe complex data flow graphs declaratively with automatic dependency wiring and parallel execution.

## dagcore

`dagcore` is the low-level DAG execution engine powering [`go-future`](https://github.com/jizhuozhi/go-future)'s structured concurrency and dataflow execution model. It provides a lock-free, dependency-driven scheduler for executing static DAGs (Directed Acyclic Graphs) in parallel.

### ✨ Features

* ⚡ **Lock-free execution** via atomic dependency counters
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

- ✅ Type-based dependency inference
- ✅ Fully integrated with `go-future` for parallel execution
- ✅ Reusable functions with Go-style declarations
- ✅ Built-in support for context propagation and error handling
- ✅ Alias support to distinguish same-type dependencies

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
	out, _ := prog.Run(context.Background())
	fmt.Println(out[TokenCount(0)]) // Output: 11
}
```

---

### 🧠 Type-Based Wiring

`dagfunc` determines node dependencies using **parameter types** and result types:

* Each function must accept `context.Context` as the first argument
* Inputs and outputs must use unique Go types or **aliases**
* The DAG will automatically determine execution order

> ⚠️ If two inputs/outputs are of the same type (e.g., multiple `string` values), use `type alias` to disambiguate.

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

---

### 🧰 API Overview

#### `dagfunc.New() *Builder`

Creates a new DAG builder.

#### `(*Builder).Provide(val any) error`

Declares a root node with known value.

#### `(*Builder).Use(fn any) error`

Registers a function as a DAG node. Must match:

```go
func(ctx context.Context, A, B, ...) (X, Y, ..., error)
```

#### `(*Builder).Compile(inputs []any) (*Program, error)`

Builds a DAG using the provided inputs.

#### `(*Program).Run(ctx context.Context) (map[any]any, error)`

Executes the DAG. Outputs are keyed by result types with typed zero.

#### `(*Program).RunAsync(ctx context.Context) *future.Future[map[any]any]`

Executes the DAG. Return a future with outputs are keyed by result types with typed zero.

#### `(*Program).Get(any) (any, error)`

Gets the result value for a specific type.

#### Error propagation

* DAG execution will **fail fast** by default
* Downstream nodes will not be executed if inputs fail
* You can customize error behavior using `dagcore`

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

* Outputs are retrieved by Go types (typed zero), not labels
* Type aliasing is required for disambiguation
* All dependencies must be resolvable at compile-time

## 🔐 License

Apache-2.0 license by [jizhuozhi](https://github.com/jizhuozhi)