package actor_test

import (
	"testing"

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
