package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Context(t *testing.T) {
	t.Parallel()

	ctx := NewContext()
	assert.NotNil(t, ctx.Done())

	select {
	case <-ctx.Done():
		assert.FailNow(t, "should not be able to read from Done channel")
	default:
	}

	assert.NoError(t, ctx.Err())

	ctx.SignalEnd()

	select {
	case _, ok := <-ctx.Done():
		assert.False(t, ok)
	default:
		assert.FailNow(t, "should be able to read from Done channel")
	}

	assert.ErrorIs(t, ctx.Err(), ErrStopped)
}

func Test_Context_Ended(t *testing.T) {
	t.Parallel()

	ctx := NewContextEnded()
	assert.NotNil(t, ctx.Done())

	select {
	case _, ok := <-ctx.Done():
		assert.False(t, ok)
	default:
		assert.FailNow(t, "should be able to read from Done channel")
	}

	assert.ErrorIs(t, ctx.Err(), ErrStopped)
}
