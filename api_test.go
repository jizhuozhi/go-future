package future

import (
	"context"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jizhuozhi/go-future/executors"
)

func TestAsync(t *testing.T) {
	f := Async(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestCtxAsync(t *testing.T) {
	ctx := context.WithValue(context.Background(), "foo", "bar")
	f := CtxAsync(ctx, func(ctx context.Context) (any, error) {
		return ctx.Value("foo"), nil
	})
	val, err := f.Get()
	assert.Equal(t, "bar", val)
	assert.Equal(t, nil, err)
}

func TestAsyncPanic(t *testing.T) {
	f := Async(func() (int, error) {
		panic("panic")
	})
	val, err := f.Get()
	assert.Equal(t, 0, val)
	assert.ErrorIs(t, err, ErrPanic)
}

func TestCtxAsyncPanic(t *testing.T) {
	f := CtxAsync(context.Background(), func(ctx context.Context) (int, error) {
		panic("panic")
	})
	val, err := f.Get()
	assert.Equal(t, 0, val)
	assert.ErrorIs(t, err, ErrPanic)
}

func TestSetExecutor(t *testing.T) {
	counter := 0
	SetExecutor(executors.ExecutorFunc(func(f func()) {
		counter++
		go f()
	}))

	f := Async(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, counter)

	assert.Panics(t, func() {
		SetExecutor(nil)
	})
}

func TestLazy(t *testing.T) {
	f := Lazy(func() (int, error) {
		return 1, nil
	})
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestLazyConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1

	var counter int32
	f := Lazy(func() (int, error) {
		c := atomic.AddInt32(&counter, 1)
		return int(c), nil
	})

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := f.Get()
			assert.Equal(t, val, 1)
			assert.Equal(t, err, nil)
		}()
	}
	wg.Wait()
}

func TestDone(t *testing.T) {
	f := Done(1)
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestDone2(t *testing.T) {
	f := Done2(1, errFoo)
	val, err := f.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, errFoo, err)
}

func TestAwait(t *testing.T) {
	f := Async(func() (int, error) {
		return 1, nil
	})
	val, err := Await(f)
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestThen(t *testing.T) {
	cases := []struct {
		val  int
		err  error
		rval string
		rerr error
	}{
		{1, nil, "1", nil},
		{10, errFoo, "", errFoo},
	}

	for _, tt := range cases {
		p := NewPromise[int]()
		f := p.Future()
		ff := Then(f, func(val int, err error) (string, error) {
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(val), 10), nil
		})
		p.Set(tt.val, tt.err)
		val, err := ff.Get()
		assert.Equal(t, tt.rval, val)
		assert.Equal(t, tt.err, err)
	}
}

func TestThenAfterDone(t *testing.T) {
	cases := []struct {
		val  int
		err  error
		rval string
		rerr error
	}{
		{1, nil, "1", nil},
		{10, errFoo, "", errFoo},
	}

	for _, tt := range cases {
		p := NewPromise[int]()
		p.Set(tt.val, tt.err)

		f := p.Future()
		ff := Then(f, func(val int, err error) (string, error) {
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(val), 10), nil
		})
		val, err := ff.Get()
		assert.Equal(t, tt.rval, val)
		assert.Equal(t, tt.err, err)
	}
}

func TestThenConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1
	rvals := make([]int, n)
	ffs := make([]func(int, error) (int64, error), n)
	for i := 0; i < n; i++ {
		r := rand.Intn(100)
		rvals[i] = r
		ffs[i] = func(i int, err error) (int64, error) {
			return int64(i + r), nil
		}
	}
	p := NewPromise[int]()
	f := p.Future()
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		i := i
		go func() {
			wg.Done()
			rf := Then(f, ffs[i])
			val, _ := rf.Get()
			assert.Equal(t, int64(rvals[i]+10), val)
		}()
	}
	p.Set(10, nil)
	wg.Wait()
}

func TestThenAsync(t *testing.T) {
	cases := []struct {
		val  int
		err  error
		rval string
		rerr error
	}{
		{1, nil, "1", nil},
		{10, errFoo, "", errFoo},
	}

	for _, tt := range cases {
		p := NewPromise[int]()
		f := p.Future()
		ff := ThenAsync(f, func(val int, err error) *Future[string] {
			return Async(func() (string, error) {
				if err != nil {
					return "", err
				}
				return strconv.FormatInt(int64(val), 10), nil
			})
		})
		p.Set(tt.val, tt.err)
		val, err := ff.Get()
		assert.Equal(t, tt.rval, val)
		assert.Equal(t, tt.err, err)
	}
}

func TestThenAsyncAfterDone(t *testing.T) {
	cases := []struct {
		val  int
		err  error
		rval string
		rerr error
	}{
		{1, nil, "1", nil},
		{10, errFoo, "", errFoo},
	}

	for _, tt := range cases {
		p := NewPromise[int]()
		p.Set(tt.val, tt.err)

		f := p.Future()
		ff := ThenAsync(f, func(val int, err error) *Future[string] {
			return Async(func() (string, error) {
				if err != nil {
					return "", err
				}
				return strconv.FormatInt(int64(val), 10), nil
			})
		})
		val, err := ff.Get()
		assert.Equal(t, tt.rval, val)
		assert.Equal(t, tt.err, err)
	}
}

func TestAnyOf(t *testing.T) {
	target := rand.Intn(10)
	vals := make([]int, 10)
	for i := 0; i < len(vals); i++ {
		if i == target {
			vals[i] = 1
		} else {
			vals[i] = (rand.Intn(10) + 1) * 10
		}
	}

	fs := make([]*Future[int], 10)
	for i := 0; i < 10; i++ {
		i := i
		fs[i] = Async(func() (int, error) {
			time.Sleep(time.Duration(vals[i]) * time.Millisecond)
			if i != target && rand.Intn(2) == 0 { // random error
				return 0, errFoo
			}
			return vals[i], nil
		})
	}
	f := AnyOf(fs...)
	r, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, target, r.Index, target)
	assert.Equal(t, vals[target], r.Val)
	assert.Equal(t, nil, r.Err)
}

func TestAnyOfEmpty(t *testing.T) {
	fs := make([]*Future[int], 0)
	f := AnyOf(fs...)
	r, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, -1, r.Index)
	assert.Zero(t, r.Val)
	assert.NoError(t, r.Err)
}

func TestAnyOfWhenAllErr(t *testing.T) {
	target := rand.Intn(10)
	vals := make([]int, 10)
	for i := 0; i < len(vals); i++ {
		if i == target {
			vals[i] = 1
		} else {
			vals[i] = (rand.Intn(10) + 1) * 10
		}
	}

	fs := make([]*Future[int], 10)
	for i := 0; i < 10; i++ {
		i := i
		fs[i] = Async(func() (int, error) {
			if i == target {
				return 0, errFoo
			}
			time.Sleep(time.Duration(vals[i]) * time.Millisecond)
			return 0, errFoo
		})
	}
	f := AnyOf(fs...)
	r, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, target, r.Index, target)
	assert.Equal(t, 0, r.Val)
	assert.Equal(t, errFoo, r.Err)
}

func TestToAny(t *testing.T) {
	f := Async(func() (int, error) {
		return 1, nil
	})
	ff := ToAny(f)
	val, err := ff.Get()
	assert.Equal(t, 1, val)
	assert.Equal(t, nil, err)
}

func TestAllOf(t *testing.T) {
	target := rand.Intn(10)
	vals := make([]int, 10)
	for i := 0; i < len(vals); i++ {
		if i == target {
			vals[i] = 1
		} else {
			vals[i] = (rand.Intn(10) + 1) * 10
		}
	}

	fs := make([]*Future[int], 10)
	for i := 0; i < 10; i++ {
		i := i
		fs[i] = Async(func() (int, error) {
			time.Sleep(time.Duration(vals[i]) * time.Millisecond)
			return vals[i], nil
		})
	}

	f := AllOf(fs...)
	results, err := f.Get()
	assert.Equal(t, vals, results)
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		ff := fs[i]
		val, err := ff.Get()
		assert.Equal(t, vals[i], val)
		assert.NoError(t, err)
	}
}

func TestAllOfEmpty(t *testing.T) {
	fs := make([]*Future[int], 0)
	f := AllOf(fs...)
	r, err := f.Get()
	assert.Equal(t, []int(nil), r)
	assert.NoError(t, err)
}

func TestAllOfWhenErr(t *testing.T) {
	target := rand.Intn(10)
	vals := make([]int, 10)
	for i := 0; i < len(vals); i++ {
		if i == target {
			vals[i] = 1
		} else {
			vals[i] = (rand.Intn(10) + 1) * 10
		}
	}

	fs := make([]*Future[int], 10)
	for i := 0; i < 10; i++ {
		i := i
		fs[i] = Async(func() (int, error) {
			time.Sleep(time.Duration(vals[i]) * time.Millisecond)
			if i == target {
				return 0, errFoo
			}
			return vals[i], nil
		})
	}

	f := AllOf(fs...)
	_, err := f.Get()
	assert.Equal(t, errFoo, err)

	for i := 0; i < 10; i++ {
		ff := fs[i]
		val, err := ff.Get()
		if i != target {
			assert.Equal(t, vals[i], val)
			assert.NoError(t, err)
		} else {
			assert.Equal(t, 0, val)
			assert.Equal(t, errFoo, err)
		}
	}
}

func TestTimeout(t *testing.T) {
	{
		f := Timeout(Async(func() (int, error) {
			time.Sleep(time.Millisecond)
			return 1, nil
		}), time.Nanosecond)
		val, err := f.Get()
		assert.Zero(t, 0, val)
		assert.ErrorIs(t, err, ErrTimeout)
	}
	{
		f := Timeout(Async(func() (int, error) {
			time.Sleep(time.Millisecond)
			return 1, nil
		}), 10*time.Millisecond)
		val, err := f.Get()
		assert.Equal(t, 1, val)
		assert.NoError(t, err)
	}
}

func TestUntil(t *testing.T) {
	{
		f := Until(Async(func() (int, error) {
			time.Sleep(time.Millisecond)
			return 1, nil
		}), time.Now().Add(time.Nanosecond))
		val, err := f.Get()
		assert.Zero(t, 0, val)
		assert.ErrorIs(t, err, ErrTimeout)
	}
	{
		f := Until(Async(func() (int, error) {
			time.Sleep(time.Millisecond)
			return 1, nil
		}), time.Now().Add(10*time.Millisecond))
		val, err := f.Get()
		assert.Equal(t, 1, val)
		assert.NoError(t, err)
	}
}
