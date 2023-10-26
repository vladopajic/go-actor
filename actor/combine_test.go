package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Combine_TestSuite(t *testing.T) {
	t.Parallel()

	TestSuite(t, func() Actor {
		const actorsCount = 5
		actors := make([]Actor, actorsCount)
		for i := 0; i < actorsCount; i++ {
			actors[i] = New(newWorker())
		}

		return Combine(actors...).Build()
	})
}

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

// Test_Combine_StopTogether asserts that all actors will end as soon
// as first actors ends.
func Test_Combine_StopTogether(t *testing.T) {
	t.Parallel()

	const actorsCount = 5 * 2

	for i := 0; i < actorsCount/2+1; i++ {
		onStartC := make(chan any, actorsCount)
		onStopC := make(chan any, actorsCount)
		actors := make([]Actor, actorsCount/2+1)
		actorsNested := make([]Actor, actorsCount/2)

		onStart := OptOnStart(func(Context) { onStartC <- `🌞` })
		onStop := OptOnStop(func() { onStopC <- `🌚` })

		create := func(i int) Actor {
			if i%2 == 0 { // case with actorStruct wrapper
				return New(newWorker(), onStart, onStop)
			} // case with actorInterface wrapper

			return Idle(onStart, onStop)
		}

		for i := range actors {
			actors[i] = create(i)
		}

		for i := range actorsNested {
			actorsNested[i] = create(i)
		}
		// add nested actors to actors list
		actors[actorsCount/2] = Combine(actorsNested...).Build()

		a := Combine(actors...).WithOptions(OptStopTogether()).Build()

		a.Start()
		drainC(onStartC, actorsCount)

		// stop actor and assert that all actors will be stopped
		actors[i].Stop()
		drainC(onStopC, actorsCount)
	}
}

func Test_Combine_OptOnStop(t *testing.T) {
	t.Parallel()

	const actorsCount = 5

	onStopC := make(chan any, 1)
	actors := make([]Actor, actorsCount)

	for i := 0; i < actorsCount; i++ {
		actors[i] = New(newWorker())
	}

	onStopFunc := func() {
		select {
		case onStopC <- `🌚`:
		default:
			t.Fatal("onStopFunc should be called only once")
		}
	}

	a := Combine(actors...).
		WithOptions(OptOnStopFunc(onStopFunc)).
		Build()

	a.Start()

	a.Stop()
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	assert.Equal(t, `🌚`, <-onStopC)
}

func Test_Combine_OptOnStop_AfterActorStops(t *testing.T) {
	t.Parallel()

	const actorsCount = 5 * 2

	for i := 0; i < actorsCount/2+1; i++ {
		onStopC := make(chan any, 1)
		actors := make([]Actor, actorsCount/2+1)
		actorsNested := make([]Actor, actorsCount/2)

		create := func(i int) Actor {
			if i%2 == 0 {
				return New(newWorker())
			}

			return Idle()
		}

		for i := range actors {
			actors[i] = create(i)
		}

		for i := range actorsNested {
			actorsNested[i] = create(i)
		}
		// add nested actors to actors list
		actors[actorsCount/2] = Combine(actorsNested...).Build()

		onStopFunc := func() {
			select {
			case onStopC <- `🌚`:
			default:
				t.Fatal("onStopFunc should be called only once")
			}
		}

		a := Combine(actors...).
			WithOptions(OptOnStopFunc(onStopFunc), OptStopTogether()).
			Build()

		a.Start()

		actors[i].Stop()
		assert.Equal(t, `🌚`, <-onStopC)
		a.Stop() // should have no effect
	}
}
