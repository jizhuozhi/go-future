//go:build go1.27

package dagfunc

import (
	"context"
	"testing"

	"github.com/jizhuozhi/go-future"
	"github.com/stretchr/testify/assert"
)

type Unregistered struct{}

// TestDagFuncGenericGet exercises the generic methods introduced with Go 1.27.
func TestDagFuncGenericGet(t *testing.T) {
	builder := New()

	assert.NoError(t, builder.Provide(InputA{}))
	assert.NoError(t, builder.Provide(InputB{}))
	assert.NoError(t, builder.Use(fnC))
	assert.NoError(t, builder.Use(fnD))
	assert.NoError(t, builder.Freeze())

	prog, err := builder.Compile([]any{InputA{Value: 10}, InputB{Text: "hello"}})
	assert.NoError(t, err)

	// ValueAsync does not block, it can be subscribed before the DAG is started.
	fc := prog.ValueAsync[ResultC]()
	fd := prog.ValueAsync[ResultD]()

	_, err = prog.Run(context.Background())
	assert.NoError(t, err)

	c, err := fc.Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultC{Sum: 15}, c)

	d, err := fd.Get()
	assert.NoError(t, err)
	assert.Equal(t, ResultD{Message: "Sum is \x0f"}, d)

	// After the DAG completed, Value returns the very same values.
	c2, err := prog.Value[ResultC]()
	assert.NoError(t, err)
	assert.Equal(t, c, c2)

	// An unregistered type is reported eagerly, without executing anything.
	_, err = prog.ValueAsync[Unregistered]().Get()
	assert.ErrorIs(t, err, ErrTypeNotFound)

	// Casting a node result to an unrelated type fails with ErrTypeMismatch.
	_, err = prog.ValueAsync[ResultC]().Cast[ResultD]().Get()
	assert.ErrorIs(t, err, future.ErrTypeMismatch)
}
