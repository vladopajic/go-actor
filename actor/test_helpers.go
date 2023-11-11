package actor

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSuite is test helper function that tests all basic actor functionality.
//
//nolint:tparallel // this is helper to test case (lint fake positive)
func TestSuite(t *testing.T, fact func() Actor) {
	t.Helper()

	t.Run("start stop", func(t *testing.T) {
		t.Parallel()

		AssertStartStopAtRandom(t, fact())
	})

	t.Run("worker end signal", func(t *testing.T) {
		t.Parallel()

		AssertWorkerEndSig(t, fact())
	})
}

// AssertStartStopAtRandom is test helper that starts and stops actor repeatedly, which
// will catch potential panic, race conditions, or some other issues.
func AssertStartStopAtRandom(t *testing.T, a Actor) {
	t.Helper()

	assert.NotNil(t, a)

	for i := 0; i < 1000; i++ {
		if randInt32(t)%2 == 0 {
			a.Start()
		} else {
			a.Stop()
		}
	}

	// Make sure that actor is stopped when exiting
	a.Stop()
}

// AssertWorkerEndSig test asserts that worker will respond to context.Done() signal.
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

func randInt32(t *testing.T) int32 {
	t.Helper()

	b := make([]byte, 4) //nolint:gomnd // 4 bytes = int32
	_, err := rand.Read(b)
	assert.NoError(t, err)

	result := int32(0)
	for i := 0; i < 4; i++ {
		result <<= 8
		result += int32(b[i])
	}

	return result
}
