package future

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errBar = errors.New("bar")

// The tests in this file cover the generic methods introduced with Go 1.27.

func TestMethodThen(t *testing.T) {
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
		f := p.Future().Then(func(val int, err error) (string, error) {
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(val), 10), nil
		})
		p.Set(tt.val, tt.err)
		val, err := f.Get()
		assert.Equal(t, tt.rval, val)
		assert.Equal(t, tt.err, err)
	}
}

func TestMethodThenAsync(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future().ThenAsync(func(val int, err error) *Future[string] {
		return Async(func() (string, error) {
			if err != nil {
				return "", err
			}
			return strconv.FormatInt(int64(val), 10), nil
		})
	})
	p.Set(1, nil)
	val, err := f.Get()
	assert.Equal(t, "1", val)
	assert.NoError(t, err)
}

func TestMethodThenAsyncError(t *testing.T) {
	p := NewPromise[int]()
	f := p.Future().ThenAsync(func(val int, err error) *Future[string] {
		return Done2("", err)
	})
	p.Set(0, errFoo)
	_, err := f.Get()
	assert.ErrorIs(t, err, errFoo)
}

// TestMethodThenGo asserts that the callback is dispatched to the executor and
// therefore never blocks the goroutine that completes the upstream Future.
func TestMethodThenGo(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})

	f := Done(1).ThenGo(func(val int, err error) (string, error) {
		close(entered)
		<-release
		return "ok", nil
	})

	select {
	case <-entered:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("callback was not dispatched to the executor")
	}
	close(release)

	val, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, "ok", val)
}

func TestMethodThenGoPanic(t *testing.T) {
	f := Done(1).ThenGo(func(val int, err error) (string, error) {
		panic("boom")
	})
	_, err := f.Get()
	assert.ErrorIs(t, err, ErrPanic)
}

func TestMethodMap(t *testing.T) {
	val, err := Done(2).Map(func(v int) int { return v * 21 }).Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	// The error is propagated and fn is skipped.
	_, err = Done2(0, errFoo).Map(func(v int) int { return v * 21 }).Get()
	assert.ErrorIs(t, err, errFoo)
}

func TestMethodFlatMap(t *testing.T) {
	val, err := Done(2).FlatMap(func(v int) *Future[int] {
		return Async(func() (int, error) { return v * 21, nil })
	}).Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	_, err = Done2(0, errFoo).FlatMap(func(v int) *Future[int] {
		t.Fatal("fn must be skipped on error")
		return nil
	}).Get()
	assert.ErrorIs(t, err, errFoo)
}

func TestMethodCast(t *testing.T) {
	// Success.
	val, err := Done[any](42).Cast[int]().Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	// Widening to any always succeeds.
	anyVal, err := Done(42).Cast[any]().Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, anyVal)

	// The upstream error wins over the type check.
	_, err = Done2[any](nil, errFoo).Cast[int]().Get()
	assert.ErrorIs(t, err, errFoo)

	// A mismatching type is reported as ErrTypeMismatch.
	_, err = Done[any]("hello").Cast[int]().Get()
	assert.ErrorIs(t, err, ErrTypeMismatch)
}

func TestMethodRecover(t *testing.T) {
	val, err := Done2(0, errFoo).Recover(func(err error) (int, error) {
		return 42, nil
	}).Get()
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	// A successful Future is passed through.
	val, err = Done(7).Recover(func(err error) (int, error) {
		t.Fatal("fn must not be called on success")
		return 0, nil
	}).Get()
	assert.NoError(t, err)
	assert.Equal(t, 7, val)

	// Recover may rethrow.
	_, err = Done2(0, errFoo).Recover(func(err error) (int, error) {
		return 0, errBar
	}).Get()
	assert.ErrorIs(t, err, errBar)
}

func TestMethodOrElse(t *testing.T) {
	val, err := Done2(0, errFoo).OrElse(9).Get()
	assert.NoError(t, err)
	assert.Equal(t, 9, val)

	val, err = Done(7).OrElse(9).Get()
	assert.NoError(t, err)
	assert.Equal(t, 7, val)
}

func TestMethodToChan(t *testing.T) {
	res := <-Done(1).ToChan()
	assert.Equal(t, 1, res.Val)
	assert.NoError(t, res.Err)
}

func TestMethodToAny(t *testing.T) {
	val, err := Done(1).ToAny().Get()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestMethodTimeout(t *testing.T) {
	// An already resolved Future wins over the deadline.
	val, err := Done(1).Timeout(time.Hour).Get()
	assert.NoError(t, err)
	assert.Equal(t, 1, val)

	f := Async(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 1, nil
	}).Timeout(time.Millisecond)
	_, err = f.Get()
	assert.ErrorIs(t, err, ErrTimeout)
}

func TestMethodUntil(t *testing.T) {
	f := Async(func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 1, nil
	}).Until(time.Now().Add(time.Millisecond))
	_, err := f.Get()
	assert.ErrorIs(t, err, ErrTimeout)
}

// TestMethodChain exercises the fluent style that generic methods enable: the
// result type of every step is only known at the call site, so the whole chain
// has to be expressed with methods instead of package-level functions.
func TestMethodChain(t *testing.T) {
	f := Async(func() (int, error) {
		return 21, nil
	}).Map(func(v int) int {
		return v * 2
	}).FlatMap(func(v int) *Future[string] {
		return Done(strconv.FormatInt(int64(v), 10))
	}).Then(func(s string, err error) (int, error) {
		if err != nil {
			return 0, err
		}
		return len(s), nil
	})

	val, err := f.Get()
	assert.NoError(t, err)
	assert.Equal(t, 2, val)
}
