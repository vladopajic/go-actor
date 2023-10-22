package actor

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:tparallel // relax
func TestSuite(t *testing.T, fact func() Actor) {
	t.Helper()

	t.Run("start stop", func(t *testing.T) {
		t.Parallel()

		AssertStartStopAtRandom(t, fact())
	})

	t.Run("worker end sig", func(t *testing.T) {
		t.Parallel()

		AssertWorkerEndSig(t, fact())
	})
}

func AssertStartStopAtRandom(t *testing.T, a Actor) {
	t.Helper()

	assert.NotNil(t, a)

	for i := 0; i < 1000; i++ {
		if rand.Int()%2 == 0 { //nolint:gosec // weak random is fine
			a.Start()
		} else {
			a.Stop()
		}
	}

	// Make sure that actor is stopped when exiting
	a.Stop()
}

func AssertWorkerEndSig(t *testing.T, aw any) {
	t.Helper()

	assert.NotNil(t, aw)

	var w Worker

	if a, ok := aw.(*actor); ok {
		w = a.worker
	} else if ww, ok := aw.(Worker); ok {
		w = ww
	} else {
		t.Skip("couldn't test worker end sig")
	}

	assert.NotNil(t, w)

	status := w.DoWork(ContextEnded())
	assert.Equal(t, WorkerEnd, status, "worker should end when context has ended")
}
