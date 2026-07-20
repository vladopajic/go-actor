package actortest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladopajic/go-actor/actor"
	"github.com/vladopajic/go-actor/actor/actortest"
)

//nolint:tparallel // subtest is used in order to see how actors are closed
func TestStart(t *testing.T) {
	t.Parallel()

	var log []string

	t.Run("start order", func(t *testing.T) {
		actortest.Start(t,
			actorStub{name: "a", log: &log},
			actorStub{name: "b", log: &log},
			actorStub{name: "c", log: &log},
		)
	})

	assert.Equal(t, []string{
		"start a",
		"start b",
		"start c",
		"stop c",
		"stop b",
		"stop a",
	}, log)
}

func TestStart_NilActor(t *testing.T) {
	t.Parallel()

	tb := &tbWrapper{T: t}

	func() {
		defer func() {
			assert.IsType(t, fatalCalled{}, recover())
		}()
		actortest.Start(tb, nil)
	}()

	assert.True(t, tb.hadFatal)
	assert.Equal(t, []any{"actor should not be nil"}, tb.fatalArgs)
}

func TestAssertStartStopAtRandom(t *testing.T) {
	t.Parallel()

	actortest.AssertStartStopAtRandom(t, actor.New(newWorker()))
	actortest.AssertStartStopAtRandom(t, actor.Noop())

	tb := &tbWrapper{T: t}
	actortest.AssertStartStopAtRandom(tb, nil)
	assert.True(t, tb.hadError)
}
