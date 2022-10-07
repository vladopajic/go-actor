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

	assertContextStarted(t, ctx)
	assertContextStringer(t, ctx)
	assertNoDeadline(t, ctx)

	ctx.SignalEnd()

	assertContextEnded(t, ctx)
	assertContextStringer(t, ctx)
	assertNoDeadline(t, ctx)
}

func Test_Context_NewInstance(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // intentionally checking if instances are identical
	assert.False(t, NewContext() == NewContext())
}

func Test_Context_ContextStarted(t *testing.T) {
	t.Parallel()

	assertContextStarted(t, ContextStarted())
	assertContextStringer(t, ContextStarted())
	assertNoDeadline(t, ContextStarted())
	assert.True(t, ContextStarted() == ContextStarted())
}

func Test_Context_ContextEnded(t *testing.T) {
	t.Parallel()

	assertContextEnded(t, ContextEnded())
	assertContextStringer(t, ContextEnded())
	assertNoDeadline(t, ContextStarted())
	assert.True(t, ContextEnded() == ContextEnded())
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
