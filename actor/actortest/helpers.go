package actortest

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/vladopajic/go-actor/actor"
)

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
func AssertWorkerEndSig(tb testing.TB, w actor.Worker) {
	tb.Helper()

	AssertWorkerEndSigAfterIterations(tb, w, 1)
}

// AssertWorkerEndSigAfterIterations test asserts that worker will respond
// to context.Done() signal after specified iterations count.
func AssertWorkerEndSigAfterIterations(tb testing.TB, w actor.Worker, iterations int) {
	tb.Helper()

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
