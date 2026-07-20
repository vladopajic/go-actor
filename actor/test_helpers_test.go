package actor_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
	"github.com/vladopajic/go-actor/actor/actortest"
)

func Test_Suite(t *testing.T) {
	t.Parallel()

	// Test when actor is created with default constructor (New)
	actortest.TestSuite(t, func() Actor { return New(newWorker()) })

	// Test when actor is not created with default constructor
	actortest.TestSuite(t, func() Actor { return Idle() })

	// Test when actor is not created with default constructor
	actortest.TestSuite(t, Noop)
}

func Test_AssertWorkerEndSig(t *testing.T) {
	t.Parallel()

	// Test with worker
	actortest.AssertWorkerEndSig(t, newWorker())

	// Test with actor
	actortest.AssertWorkerEndSig(t, New(newWorker()))

	// Test expected to fail because argument nil
	tw := &tWrapper{T: t}
	actortest.AssertWorkerEndSig(tw, nil)
	assert.True(t, tw.hadError)

	// Test expected to fail because worker is nil
	tw = &tWrapper{T: t}
	actortest.AssertWorkerEndSig(tw, New(nil))
	assert.True(t, tw.hadError)

	// Test expected to fail because worker didn't return end signal
	tw = &tWrapper{T: t}
	actortest.AssertWorkerEndSig(tw, NewWorker(func(Context) WorkerStatus {
		return WorkerContinue
	}))
	assert.True(t, tw.hadError)
}

func Test_AssertStartStopAtRandom(t *testing.T) {
	t.Parallel()

	actortest.AssertStartStopAtRandom(t, New(newWorker()))
	actortest.AssertStartStopAtRandom(t, Noop())

	// Test expected to fail because actor is nil
	tw := &tWrapper{T: t}
	actortest.AssertStartStopAtRandom(tw, nil)
	assert.True(t, tw.hadError)
}

func Test_RandInt32WithReader(t *testing.T) {
	t.Parallel()

	v := make(map[int32]struct{})
	for range 10000 {
		v[RandInt32WithReader(t, rand.Reader)] = struct{}{}
	}

	assert.GreaterOrEqual(t, len(v), 1000) // should have at least 1000 unique elements

	// Test expected to fail because bytes could not be read
	tw := &tWrapper{T: t}
	RandInt32WithReader(tw, errReader{})
	assert.True(t, tw.hadError)
}
