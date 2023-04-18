package actor_test

import (
	"testing"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Suite(t *testing.T) {
	t.Parallel()

	TestSuite(t, func() Actor { return New(newWorker()) })

	TestSuite(t, func() Actor { return Idle() })

	TestSuite(t, Noop)
}
