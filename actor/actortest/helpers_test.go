package actortest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladopajic/go-actor/actor"
	"github.com/vladopajic/go-actor/actor/actortest"
)

func TestSuite(t *testing.T) {
	t.Parallel()

	actortest.TestSuite(t, actor.Noop)
	actortest.TestSuite(t, func() actor.Actor {
		return actor.New(newWorker())
	})
}

func TestAssertStartStopAtRandom(t *testing.T) {
	t.Parallel()

	actortest.AssertStartStopAtRandom(t, actor.New(newWorker()))
	actortest.AssertStartStopAtRandom(t, actor.Noop())

	tb := &tbWrapper{T: t}
	actortest.AssertStartStopAtRandom(tb, nil)
	assert.True(t, tb.hadError)
}

func TestAssertWorkerEndSig(t *testing.T) {
	t.Parallel()

	actortest.AssertWorkerEndSig(t, newWorker())
	actortest.AssertWorkerEndSig(t, actor.New(newWorker()))

	tb := &tbWrapper{T: t}
	actortest.AssertWorkerEndSig(tb, nil)
	assert.True(t, tb.hadError)

	tb = &tbWrapper{T: t}
	actortest.AssertWorkerEndSig(tb, actor.NewWorker(func(actor.Context) actor.WorkerStatus {
		return actor.WorkerContinue
	}))
	assert.True(t, tb.hadError)
}

func newWorker() actor.Worker {
	return actor.NewWorker(func(ctx actor.Context) actor.WorkerStatus {
		if ctx.Err() != nil {
			return actor.WorkerEnd
		}

		return actor.WorkerContinue
	})
}
