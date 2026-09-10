package dagfunc

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/jizhuozhi/go-future"
	"github.com/jizhuozhi/go-future/dagcore"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

	// errType is the error interface. A result is the error of a node only when
	// it is declared as exactly this type (is-a), never when it is some named
	// type that happens to implement it (as-a): the signature decides, not the
	// method set. A `func(ctx) (Result, MyError)` therefore produces two values
	// and can never fail, while a `func(ctx) (Result, error)` produces one
	// value and may fail.
	errType = reflect.TypeOf((*error)(nil)).Elem()

	futurePkgPath = reflect.TypeOf((*future.Future[any])(nil)).Elem().PkgPath()
)

// funcSpec is a node function reduced to a normalised form:
//
//	(ctx?) (in...) (out...|error?)
type funcSpec struct {
	fn reflect.Value

	hasCtx bool
	in     []reflect.Type // parameters without the leading context

	outs     []reflect.Type // results without the trailing error
	hasError bool

	// future is true when the single result is a *future.Future[R]; the node
	// waits for that Future instead of computing its value inline.
	future bool
}

func parseFunc(fn any) (*funcSpec, error) {
	if fn == nil {
		return nil, ErrNotAFunction
	}
	v := reflect.ValueOf(fn)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return nil, ErrNotAFunction
	}
	if t.IsVariadic() {
		return nil, fmt.Errorf("%w: variadic functions are not supported", ErrFuncSignature)
	}

	spec := &funcSpec{fn: v}

	start := 0
	if t.NumIn() > 0 && t.In(0) == ctxType {
		spec.hasCtx = true
		start = 1
	}
	for i := start; i < t.NumIn(); i++ {
		if t.In(i) == ctxType {
			return nil, fmt.Errorf("%w: context.Context must be the first parameter", ErrFuncSignature)
		}
		spec.in = append(spec.in, t.In(i))
	}

	// At most one error, and it has to come last: anything else is ambiguous
	// and is rejected instead of being guessed at.
	n := t.NumOut()
	if n > 0 && t.Out(n-1) == errType {
		spec.hasError = true
		n--
	}
	for i := 0; i < n; i++ {
		if t.Out(i) == errType {
			return nil, fmt.Errorf("%w: error must be the only and the last result", ErrFuncSignature)
		}
		spec.outs = append(spec.outs, t.Out(i))
	}
	if len(spec.outs) == 0 && !spec.hasError {
		return nil, fmt.Errorf("%w: the function returns neither a value nor an error", ErrFuncSignature)
	}

	if len(spec.outs) == 1 {
		if elem, ok := futureElem(spec.outs[0]); ok {
			spec.future = true
			spec.outs = []reflect.Type{elem}
		}
	}
	return spec, nil
}

// run builds the dagcore body of the node: it collects the resolved dependency
// values, calls the function and normalises the results.
func (spec *funcSpec) run(params []depRef) dagcore.NodeFunc {
	hasCtx := spec.hasCtx
	return func(ctx context.Context, input map[dagcore.NodeID]any) (any, error) {
		args := make([]reflect.Value, 0, len(params)+1)
		if hasCtx {
			args = append(args, reflect.ValueOf(ctx))
		}
		for _, p := range params {
			args = append(args, valueOf(unwrap(input[p.id], p.index), p.typ))
		}
		return spec.call(args)
	}
}

func (spec *funcSpec) call(args []reflect.Value) (any, error) {
	res := spec.fn.Call(args)

	var err error
	if spec.hasError {
		last := res[len(res)-1]
		if !isNilValue(last) {
			// The result is declared as error, so the assertion always holds.
			err, _ = last.Interface().(error)
		}
		res = res[:len(res)-1]
	}
	if err != nil {
		// A failed node publishes nothing. Returning the partially computed
		// values next to the error would make them readable through
		// Node(...).Future() even though the graph has already failed.
		return nil, err
	}

	if spec.future {
		f := res[0]
		if isNilValue(f) {
			return nil, nil
		}
		out := f.MethodByName("Get").Call(nil)
		if !isNilValue(out[1]) {
			return nil, out[1].Interface().(error)
		}
		return out[0].Interface(), nil
	}

	switch len(res) {
	case 0:
		return nil, nil
	case 1:
		return res[0].Interface(), nil
	default:
		vals := make([]any, 0, len(res))
		for _, r := range res {
			vals = append(vals, r.Interface())
		}
		return vals, nil
	}
}

// futureElem reports whether t is *future.Future[R] and returns R.
func futureElem(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() != reflect.Ptr {
		return nil, false
	}
	elem := t.Elem()
	if elem.Kind() != reflect.Struct || elem.PkgPath() != futurePkgPath {
		return nil, false
	}
	if name := elem.Name(); stripTypeArgs(name) != "Future" {
		return nil, false
	}
	if elem.NumField() != 1 {
		return nil, false
	}
	state := elem.Field(0).Type
	if state.Kind() != reflect.Ptr {
		return nil, false
	}
	val, ok := state.Elem().FieldByName("val")
	if !ok {
		return nil, false
	}
	return val.Type, true
}

func stripTypeArgs(name string) string {
	if i := strings.IndexByte(name, '['); i >= 0 {
		return name[:i]
	}
	return name
}

func runtimeFuncName(fn any) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return ""
	}
	f := runtime.FuncForPC(v.Pointer())
	if f == nil {
		return ""
	}
	return f.Name()
}

func valueOf(v any, t reflect.Type) reflect.Value {
	if v == nil {
		return reflect.Zero(t)
	}
	rv := reflect.ValueOf(v)
	if rv.Type().AssignableTo(t) {
		return rv
	}
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t)
	}
	return reflect.Zero(t)
}

func isNilValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}
