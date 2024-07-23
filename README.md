# go-future

[![codecov](https://codecov.io/github/jizhuozhi/go-future/graph/badge.svg?token=9UZDVRZCQM)](https://codecov.io/github/jizhuozhi/go-future)

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

### Example

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

### Example

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

### Example

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
