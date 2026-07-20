// Package actortest provides testing helpers for actors.
package actortest

import (
	"crypto/rand"
	"io"
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

func randInt32(tb testing.TB) int32 {
	tb.Helper()
	return randInt32WithReader(tb, rand.Reader)
}

func randInt32WithReader(tb testing.TB, randReader io.Reader) int32 {
	tb.Helper()

	const byteSize = 4
	b := make([]byte, byteSize)

	if _, err := io.ReadFull(randReader, b); err != nil { // coverage-ignore
		tb.Error("failed to read random bytes")
		return 0
	}

	result := int32(0)
	for i := range byteSize {
		result <<= 8
		result += int32(b[i])
	}

	return result
}
