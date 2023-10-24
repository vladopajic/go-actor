package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	// NewOptions with no options should equal to zero options
	assert.Equal(t, NewZeroOptions(), NewOptions[ActorOption]())

	testActorOptions(t)
	testMailboxOptions(t)
	testCombinedOptions(t)
}

func testActorOptions(t *testing.T) {
	t.Helper()

	{ // Assert that OnStartFunc will be set
		opts := NewOptions(OptOnStart(func(Context) {}))
		assert.NotNil(t, opts.Actor.OnStartFunc)
		assert.Nil(t, opts.Actor.OnStopFunc)
	}

	{ // Assert that OnStopFunc will be set
		opts := NewOptions(OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStopFunc)
		assert.Nil(t, opts.Actor.OnStartFunc)
	}

	{ // Assert that OnStartFunc and OnStopFunc will be set
		opts := NewOptions(OptOnStart(func(Context) {}), OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStartFunc)
		assert.NotNil(t, opts.Actor.OnStopFunc)
	}
}

func testMailboxOptions(t *testing.T) {
	t.Helper()

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

	{ // Assert that OptCapacity and OptMinCapacity will be set
		opts := NewOptions(OptMailbox(16, 32))
		assert.Equal(t, 16, opts.Mailbox.Capacity)
		assert.Equal(t, 32, opts.Mailbox.MinCapacity)
	}

	{ // Assert that OptAsChan will be set
		opts := NewOptions(OptAsChan())
		assert.True(t, opts.Mailbox.AsChan)
	}
}

func testCombinedOptions(t *testing.T) {
	t.Helper()

	{ // Assert that StopTogether will be set
		opts := NewOptions(OptStopTogether())
		assert.True(t, opts.Combined.StopTogether)
	}
}
