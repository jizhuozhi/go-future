# go-future

The `Promise` provides a facility to store a value or an error that is later acquired asynchronously via a `Future` created by the `Promise` object. Note that the `Promise` object is meant to be used only once.

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

func foo() *future.Promise[string] {
	p := future.NewPromise[string]()
	go func() {
		p.Set("foo", nil)
	}()
	return p
}

func main() {
	p := foo()
	f := p.Future()
	val, err := f.Get()
	println(val, err)
}
```