package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Context(t *testing.T) {
	t.Parallel()

	ctx := NewContext()

	assertContextStarted(t, ctx)

	ctx.SignalEnd()

	assertContextEnded(t, ctx)
}

func Test_Context_Predefined(t *testing.T) {
	t.Parallel()

	assertContextStarted(t, ContextStarted())

	assertContextEnded(t, ContextEnded())
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
