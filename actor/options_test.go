package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	{ // NewOptions with no options should equal to zero options
		opts := NewOptions()
		assert.Equal(t, NewZeroOptions(), opts)
	}

	{ // Assert that OnStartFunc will be set
		opts := NewOptions(OptOnStart(func() {}))
		assert.NotNil(t, opts.Actor.OnStartFunc)
		assert.Nil(t, opts.Actor.OnStopFunc)
	}

	{ // Assert that OnStopFunc will be set
		opts := NewOptions(OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStopFunc)
		assert.Nil(t, opts.Actor.OnStartFunc)
	}

	{ // Assert that OnStartFunc and OnStopFunc will be set
		opts := NewOptions(OptOnStart(func() {}), OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStartFunc)
		assert.NotNil(t, opts.Actor.OnStopFunc)
	}

	{ // Assert that OptCapacity will be set
		opts := NewOptions(OptCapacity(16))
		assert.Equal(t, 16, opts.Mailbox.Capacity)
		assert.Equal(t, 0, opts.Mailbox.MinCapacity)
	}

	{ // Assert that OptMinCapacity will be set
		opts := NewOptions(OptMinCapacity(32))
		assert.Equal(t, 0, opts.Mailbox.Capacity)
		assert.Equal(t, 32, opts.Mailbox.MinCapacity)
	}

	{ // Assert that OptCapacity and OptMinCapacity will be set
		opts := NewOptions(OptCapacity(16), OptMinCapacity(32))
		assert.Equal(t, 16, opts.Mailbox.Capacity)
		assert.Equal(t, 32, opts.Mailbox.MinCapacity)
	}
}
