# go-future

[![codecov](https://codecov.io/github/jizhuozhi/go-future/graph/badge.svg?token=9UZDVRZCQM)](https://codecov.io/github/jizhuozhi/go-future)
[![goreport](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)](https://goreportcard.com/badge/github.com/jizhuozhi/go-future)

The `Promise` provides a facility to store a value or an error that is later acquired asynchronously via a `Future` created by the `Promise` object. Note that the `Promise` object is meant to be used only once.

Except for `Async`, all functions are event-driven, which means that no additional goroutines will be created to increase runtime scheduling overhead.

## Benchmark

```
goos: darwin
goarch: arm64
pkg: github.com/jizhuozhi/go-future
Benchmark
Benchmark/Promise
Benchmark/Promise-12         	 3053698	       377.4 ns/op
Benchmark/WaitGroup
Benchmark/WaitGroup-12       	 2882623	       424.9 ns/op
Benchmark/Channel
Benchmark/Channel-12         	 3004372	       399.7 ns/op
PASS
```

## Example

```go
package main

import "github.com/jizhuozhi/go-future"

func main() {
	p := future.NewPromise[string]()
	go func() {
		p.Set("foo", nil)
	}()
	f := p.Future()
	val, err := f.Get()
	println(val, err)
}
```

## Useful API

### `Async`

Create an asynchronous task and return a `Future` to get the result.

#### Example

```go
package main

import "github.com/jizhuozhi/go-future"

func main() {
	f := future.Async(func() (string, error) {
		return "foo", nil
	})
	val, err := f.Get()
	println(val, err)
}
```

### `Lazy`

Create a lazy execution task and return a `Future` to get the result. The task will only be executed once, and the task will only be executed when `Get` is called for the first time.

#### Example

```go
package main

import "github.com/jizhuozhi/go-future"

func main() {
	f := future.Lazy(func() (string, error) {
		return "foo", nil
	})
	val, err := f.Get()
	println(val, err)
}
```

### `Then`

In reality, asynchronous tasks have dependencies. When task B depends on task A, it can be concatenated through `Then`.

#### Example

```go
package main

import (
	"strconv"

	"github.com/jizhuozhi/go-future"
)

func main() {
	f := future.Async(func() (int, error) {
		return 1, nil
	})
	ff := future.Then(f, func(val int, err error) (string, error) {
		if err != nil {
			return "", err
		}
		return strconv.Itoa(val), nil
	})
	val, err := ff.Get()
	println(val, err)
}
```

---

## DAG Execution Engine (New!)

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
	prog, _ := b.Compile(map[reflect.Type]any{
		reflect.TypeOf(""): "hello world",
	})
	outputs, _ := prog.Run(context.Background())

	fmt.Println("result:", outputs[reflect.TypeOf(0)]) // => 11
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
	"reflect"

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

	prog, _ := b.Compile(map[reflect.Type]any{
		reflect.TypeOf(UserID("")): UserID("Alice"),
	})
	out, _ := prog.Run(context.Background())

	fmt.Println(out[reflect.TypeOf("")]) // => hello, Alice
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
