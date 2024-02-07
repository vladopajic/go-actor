package actor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Context_Stopping(t *testing.T) {
	t.Parallel()

	ctx := NewContext()

	// Assert that new context is in started state
	assertContextStarted(t, ctx)
	assertContextStringer(t, ctx)
	assertNoDeadline(t, ctx)

	// End should end context
	ctx.End()
	assertContextEnded(t, ctx)
	assertContextStringer(t, ctx)
	assertNoDeadline(t, ctx)

	// Ending context again should have no effect
	ctx.End()
	assertContextEnded(t, ctx)
	assertContextStringer(t, ctx)
	assertNoDeadline(t, ctx)
}

func Test_Context_NewInstance(t *testing.T) {
	t.Parallel()

	assert.NotSame(t, NewContext(), NewContext()) //nolint:testifylint // relax
}

func Test_Context_ContextStarted(t *testing.T) {
	t.Parallel()

	assertContextStarted(t, ContextStarted())
	assertContextStringer(t, ContextStarted())
	assertNoDeadline(t, ContextStarted())
	assert.Same(t, ContextStarted(), ContextStarted()) //nolint:testifylint // relax
}

func Test_Context_ContextEnded(t *testing.T) {
	t.Parallel()

	assertContextEnded(t, ContextEnded())
	assertContextStringer(t, ContextEnded())
	assertNoDeadline(t, ContextStarted())
	assert.Same(t, ContextEnded(), ContextEnded()) //nolint:testifylint // relax
}

func Test_Context_Value(t *testing.T) {
	t.Parallel()

	assertNoValue(t, func() Context {
		endedCtx := NewContext()
		endedCtx.End()

		return endedCtx
	}())
	assertNoValue(t, NewContext())
	assertNoValue(t, ContextEnded())
	assertNoValue(t, ContextStarted())
}

func assertContextStarted(t *testing.T, ctx Context) {
	t.Helper()

	assert.NotNil(t, ctx)

	select {
	case <-ctx.Done():
		assert.FailNow(t, "should not be able to read from Done channel")
	default:
	}

	assert.NoError(t, ctx.Err())
}

func assertContextEnded(t *testing.T, ctx Context) {
	t.Helper()

	assert.NotNil(t, ctx)

	select {
	case _, ok := <-ctx.Done():
		assert.False(t, ok)
	default:
		assert.FailNow(t, "should be able to read from Done channel")
	}

	assert.ErrorIs(t, ctx.Err(), ErrStopped)
}

func assertContextStringer(t *testing.T, ctx Context) {
	t.Helper()

	assert.Equal(t, "actor.Context", fmt.Sprintf("%s", ctx))
}

func assertNoDeadline(t *testing.T, ctx Context) {
	t.Helper()

	d, ok := ctx.Deadline()
	assert.Equal(t, time.Time{}, d)
	assert.False(t, ok)
}

func assertNoValue(t *testing.T, ctx Context) {
	t.Helper()

	var (
		keyStr string
		keyInt int
	)

	assert.Empty(t, ctx.Value(nil))
	assert.Empty(t, ctx.Value(&keyInt))
	assert.Empty(t, ctx.Value(&keyStr))
}
