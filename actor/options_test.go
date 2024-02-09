package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	// NewOptions with no options should equal to zero options
	assert.Equal(t, NewZeroOptions(), NewOptions[Option]())

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

		assert.Empty(t, opts.Mailbox)
		assert.Empty(t, opts.Combined)
	}

	{ // Assert that OnStopFunc will be set
		opts := NewOptions(OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStopFunc)
		assert.Nil(t, opts.Actor.OnStartFunc)

		assert.Empty(t, opts.Mailbox)
		assert.Empty(t, opts.Combined)
	}

	{ // Assert that OnStartFunc and OnStopFunc will be set
		opts := NewOptions(OptOnStart(func(Context) {}), OptOnStop(func() {}))
		assert.NotNil(t, opts.Actor.OnStartFunc)
		assert.NotNil(t, opts.Actor.OnStopFunc)

		assert.Empty(t, opts.Mailbox)
		assert.Empty(t, opts.Combined)
	}
}

func testMailboxOptions(t *testing.T) {
	t.Helper()

	{ // Assert that OptCapacity will be set
		opts := NewOptions(OptCapacity(16))
		assert.Equal(t, 16, opts.Mailbox.Capacity)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Combined)
	}

	{ // Assert that OptAsChan will be set
		opts := NewOptions(OptAsChan())
		assert.True(t, opts.Mailbox.AsChan)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Combined)
	}

	{ // Assert that OptStopAfterReceivingAll will be set
		opts := NewOptions(OptStopAfterReceivingAll())
		assert.True(t, opts.Mailbox.StopAfterReceivingAll)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Combined)
	}
}

func testCombinedOptions(t *testing.T) {
	t.Helper()

	{ // Assert that StopTogether will be set
		opts := NewOptions(OptStopTogether())
		assert.True(t, opts.Combined.StopTogether)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Mailbox)
	}

	{ // Assert that OnStartCombined will be set
		opts := NewOptions(OptOnStartCombined(func(Context) {}))
		assert.NotNil(t, opts.Combined.OnStartFunc)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Mailbox)
	}

	{ // Assert that OnStopCombined will be set
		opts := NewOptions(OptOnStopCombined(func() {}))
		assert.NotNil(t, opts.Combined.OnStopFunc)

		assert.Empty(t, opts.Actor)
		assert.Empty(t, opts.Mailbox)
	}
}
