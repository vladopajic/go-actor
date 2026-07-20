package actortest

import (
	"testing"

	"github.com/vladopajic/go-actor/actor"
)

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

	if iterations < 1 {
		tb.Error("iterations should be >= 1")
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
