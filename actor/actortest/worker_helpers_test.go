package actortest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladopajic/go-actor/actor"
	"github.com/vladopajic/go-actor/actor/actortest"
)

func TestAssertWorkerEndSig(t *testing.T) {
	t.Parallel()

	actortest.AssertWorkerEndSig(t, newWorker())

	tb := &tbWrapper{T: t}
	actortest.AssertWorkerEndSig(tb, nil)
	assert.True(t, tb.hadError)

	tb = &tbWrapper{T: t}
	actortest.AssertWorkerEndSig(tb, actor.NewWorker(func(actor.Context) actor.WorkerStatus {
		return actor.WorkerContinue
	}))
	assert.True(t, tb.hadError)
}

func TestAssertWorkerEndSigAfterIterations(t *testing.T) {
	t.Parallel()

	actortest.AssertWorkerEndSigAfterIterations(t, delayedEndWorker(3), 3)

	tb := &tbWrapper{T: t}
	actortest.AssertWorkerEndSigAfterIterations(tb, delayedEndWorker(3), 0)
	assert.True(t, tb.hadError)

	tb = &tbWrapper{T: t}
	actortest.AssertWorkerEndSigAfterIterations(tb, delayedEndWorker(3), -1)
	assert.True(t, tb.hadError)
}

func delayedEndWorker(iterations int) actor.Worker {
	i := 0

	return actor.NewWorker(func(actor.Context) actor.WorkerStatus {
		i++
		if i >= iterations {
			return actor.WorkerEnd
		}

		return actor.WorkerContinue
	})
}

func newWorker() actor.Worker {
	return actor.NewWorker(func(ctx actor.Context) actor.WorkerStatus {
		if ctx.Err() != nil {
			return actor.WorkerEnd
		}

		return actor.WorkerContinue
	})
}
