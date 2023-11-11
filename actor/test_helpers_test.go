package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Suite(t *testing.T) {
	t.Parallel()

	// Test when actor is created with default constructor (New)
	TestSuite(t, func() Actor { return New(newWorker()) })

	// Test when actor is not created with default constructor
	TestSuite(t, func() Actor { return Idle() })

	// Test when actor is not created with default constructor
	TestSuite(t, Noop)
}

func Test_AssertWorkerEndSig(t *testing.T) {
	t.Parallel()

	// Test with worker
	AssertWorkerEndSig(t, newWorker())

	// Test with actor
	AssertWorkerEndSig(t, New(newWorker()))

	// Test when actor is not created with default constructor
	AssertWorkerEndSig(t, Noop)
}

func Test_RandInt32(t *testing.T) {
	t.Parallel()

	v := make(map[int32]struct{})
	for i := 0; i < 10000; i++ {
		v[RandInt32(t)] = struct{}{}
	}

	assert.GreaterOrEqual(t, len(v), 1000) // should have at least 1000 unque elements
}
