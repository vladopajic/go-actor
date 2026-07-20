package actortest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vladopajic/go-actor/actor/actortest"
)

type actorStub struct {
	name string
	log  *[]string
}

func (a actorStub) Start() {
	*a.log = append(*a.log, "start "+a.name)
}

func (a actorStub) Stop() {
	*a.log = append(*a.log, "stop "+a.name)
}

type tbWrapper struct {
	*testing.T
	hadFatal  bool
	fatalArgs []any
}

type fatalCalled struct{}

func (tb *tbWrapper) Fatal(args ...any) {
	tb.hadFatal = true
	tb.fatalArgs = args

	panic(fatalCalled{}) //nolint:forbidigo // test double for Fatal must not return
}

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
