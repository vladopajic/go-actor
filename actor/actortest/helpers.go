package actortest

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/vladopajic/go-actor/actor"
)

// TestSuite is test helper function that tests all basic actor functionality.
//
//nolint:tparallel // this is helper to test case (lint fake positive)
func TestSuite(t *testing.T, fact func() actor.Actor) {
	t.Helper()

	t.Run("start stop", func(t *testing.T) {
		t.Parallel()

		AssertStartStopAtRandom(t, fact())
	})

	t.Run("worker end signal", func(t *testing.T) {
		t.Parallel()

		w, ok := workerFrom(fact())
		if !ok {
			t.Skip("worker end signal skipped - could not get Worker")
			return
		}

		AssertWorkerEndSig(t, w)
	})
}

// AssertStartStopAtRandom is test helper that starts and stops actor repeatedly, which
// will catch potential panic, race conditions, or some other issues.
func AssertStartStopAtRandom(tb testing.TB, a actor.Actor) {
	tb.Helper()

	if a == nil {
		tb.Error("actor should not be nil")
		return
	}

	for range 1000 {
		if randInt32(tb)%2 == 0 {
			a.Start()
		} else {
			a.Stop()
		}
	}

	// Make sure that actor is stopped when exiting
	a.Stop()
}

// AssertWorkerEndSig test asserts that worker will respond to context.Done() signal.
func AssertWorkerEndSig(tb testing.TB, aw any) {
	tb.Helper()

	AssertWorkerEndSigAfterIterations(tb, aw, 1)
}

// AssertWorkerEndSigAfterIterations test asserts that worker will respond
// to context.Done() signal after specified iterations count.
func AssertWorkerEndSigAfterIterations(tb testing.TB, aw any, iterations int) {
	tb.Helper()

	if aw == nil {
		tb.Error("actor or worker should not be nil")
		return
	}

	w, ok := workerFrom(aw)
	if !ok {
		tb.Error("couldn't test worker end sig")
		return
	}

	if w == nil {
		tb.Error("worker should be initialized")
		return
	}

	for range iterations {
		status := w.DoWork(actor.ContextEnded())
		if status == actor.WorkerEnd {
			return
		}
	}

	tb.Error("worker should end when context has ended")
}

type workerGetter interface {
	Worker() actor.Worker
}

func workerFrom(aw any) (actor.Worker, bool) {
	if w, ok := aw.(actor.Worker); ok {
		return w, true
	}

	if wg, ok := aw.(workerGetter); ok {
		return wg.Worker(), true
	}

	return nil, false
}

func randInt32(tb testing.TB) int32 {
	tb.Helper()
	return randInt32WithReader(tb, rand.Reader)
}

func randInt32WithReader(tb testing.TB, randReader io.Reader) int32 {
	tb.Helper()

	const byteSize = 4
	b := make([]byte, byteSize)

	_, err := randReader.Read(b)
	if err != nil {
		tb.Error("failed to read random bytes")
	}

	result := int32(0)
	for i := range byteSize {
		result <<= 8
		result += int32(b[i])
	}

	return result
}
