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

		assert.Equal(t, []string{
			"start a",
			"start b",
			"start c",
		}, log)
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
