// Package actortest provides testing helpers for actors.
package actortest

import (
	"testing"

	"github.com/vladopajic/go-actor/actor"
)

// Start starts actors in the provided order and registers test cleanup that
// stops them in reverse order.
func Start(tb testing.TB, actors ...actor.Actor) {
	tb.Helper()

	for _, a := range actors {
		if a == nil {
			tb.Fatal("actor should not be nil")
		}
	}

	tb.Cleanup(func() {
		for i := len(actors) - 1; i >= 0; i-- {
			actors[i].Stop()
		}
	})

	for _, a := range actors {
		a.Start()
	}
}
