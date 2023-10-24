package actor_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

// Test asserts that all Start and Stop is
// delegated to all combined actors.
func Test_Combine(t *testing.T) {
	t.Parallel()

	const actorsCount = 5

	onStartC := make(chan any, actorsCount)
	onStopC := make(chan any, actorsCount)
	actors := make([]Actor, actorsCount)

	for i := 0; i < actorsCount; i++ {
		actors[i] = New(newWorker(),
			OptOnStart(func(Context) { onStartC <- `🌞` }),
			OptOnStop(func() { onStopC <- `🌚` }),
		)
	}

	// Assert that starting and stopping combined actors
	// will start and stop all individual actors
	a := Combine(actors...).Build()

	// Start combined actor and wait for all actors to be started
	a.Start()
	drainC(onStartC, actorsCount)

	// Stopping actor and assert that onStopC has received signal from all actors.
	// Channel onStopC is not drained intentionally because after Stop() has returned
	// it is guarantied that all actors have ended.
	a.Stop()
	assert.Len(t, onStopC, actorsCount)
}

// Test_CombineAndStopTogether asserts that all actors will end as soon
// as first actors ends.
func Test_CombineAndStopTogether(t *testing.T) {
	t.Parallel()

	const actorsCount = 5

	onStartC := make(chan any, actorsCount)
	onStopC := make(chan any, actorsCount)
	actors := make([]Actor, actorsCount)

	onStart := OptOnStart(func(Context) { onStartC <- `🌞` })
	onStop := OptOnStop(func() { onStopC <- `🌚` })

	for i := 0; i < actorsCount; i++ {
		if i%2 == 0 { // case with actorStruct wrapper
			actors[i] = New(newWorker(), onStart, onStop)
		} else { // case with actorInterface wrapper
			actors[i] = Idle(onStart, onStop)
		}
	}

	a := Combine(actors...).WithOptions(OptStopTogether()).Build()

	a.Start()
	drainC(onStartC, actorsCount)

	// Stop random actor and assert that all actors will be stopped
	actors[rand.Int31n(actorsCount)].Stop() //nolint:gosec // weak random is fine
	drainC(onStopC, actorsCount)
}
