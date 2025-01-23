package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Combine_TestSuite(t *testing.T) {
	t.Parallel()

	TestSuite(t, func() Actor {
		actors := createActors(0)
		return Combine(actors...).Build()
	})

	TestSuite(t, func() Actor {
		actors := createActors(1)
		return Combine(actors...).Build()
	})

	TestSuite(t, func() Actor {
		actors := createActors(10)
		return Combine(actors...).Build()
	})
}

// Test asserts that all Start and Stop is delegated to all combined actors.
func Test_Combine(t *testing.T) {
	t.Parallel()

	testCombine(t, 0)
	testCombine(t, 1)
	testCombine(t, 5)
}

func testCombine(t *testing.T, actorsCount int) {
	t.Helper()

	onStartC := make(chan any, actorsCount)
	onStopC := make(chan any, actorsCount)
	onStart := OptOnStart(func(Context) { onStartC <- `🌞` })
	onStop := OptOnStop(func() { onStopC <- `🌚` })
	actors := createActors(actorsCount, onStart, onStop)

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

func Test_Combine_OptOnStopOptOnStart(t *testing.T) {
	t.Parallel()

	testCombineOptOnStopOptOnStart(t, 0)
	testCombineOptOnStopOptOnStart(t, 1)
	testCombineOptOnStopOptOnStart(t, 5)
}

func testCombineOptOnStopOptOnStart(t *testing.T, count int) {
	t.Helper()

	onStatC, onStartOpt := createCombinedOnStartOption(t, 1)
	onStopC, onStopOpt := createCombinedOnStopOption(t, 1)
	actors := createActors(count)

	a := Combine(actors...).
		WithOptions(onStopOpt, onStartOpt).
		Build()

	a.Start()

	a.Stop()
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	assert.Equal(t, `🌚`, <-onStopC)
	assert.Equal(t, `🌞`, <-onStatC)
	assert.Empty(t, onStopC)
	assert.Empty(t, onStatC)
}

// Test_Combine_OptStopTogether asserts that all actors will end as soon
// as first actors ends.
func Test_Combine_OptStopTogether(t *testing.T) {
	t.Parallel()

	// no need to test OptStopTogether for actors count < 2
	// because single actor is always "stopped together"

	testCombineOptStopTogether(t, 1)
	testCombineOptStopTogether(t, 2)
	testCombineOptStopTogether(t, 10)
}

func testCombineOptStopTogether(t *testing.T, actorsCount int) {
	t.Helper()

	for i := range actorsCount {
		onStartC := make(chan any, actorsCount)
		onStopC := make(chan any, actorsCount)
		onStart := OptOnStart(func(Context) { onStartC <- `🌞` })
		onStop := OptOnStop(func() { onStopC <- `🌚` })
		actors := createActors(actorsCount, onStart, onStop)

		a := Combine(actors...).WithOptions(OptStopTogether()).Build()

		a.Start()
		drainC(onStartC, actorsCount)

		// stop actor and assert that all actors will be stopped
		actors[i].Stop()
		drainC(onStopC, actorsCount)
	}
}

func Test_Combine_OptOnStop_AfterActorStops(t *testing.T) {
	t.Parallel()

	const actorsCount = 5 * 2

	for i := range actorsCount/2 + 1 {
		onStopC, onStopOpt := createCombinedOnStopOption(t, 2)
		actors := createActors(actorsCount / 2)

		// append one more actor to actors list
		cmb := Combine(createActors(actorsCount / 2)...).WithOptions(onStopOpt).Build()
		actors = append(actors, cmb)

		a := Combine(actors...).
			WithOptions(onStopOpt, OptStopTogether()).
			Build()

		a.Start()

		actors[i].Stop()
		assert.Equal(t, `🌚`, <-onStopC)
		assert.Equal(t, `🌚`, <-onStopC)
		a.Stop() // should have no effect
	}
}

func createActors(count int, opts ...Option) []Actor {
	actors := make([]Actor, count)

	for i := range count {
		actors[i] = createActor(i, opts...)
	}

	return actors
}

func createActor(i int, opts ...Option) Actor {
	if i%2 == 0 {
		return New(newWorker(), opts...)
	}

	return Idle(opts...)
}

func createCombinedOnStopOption(t *testing.T, count int) (<-chan any, CombinedOption) {
	t.Helper()

	c := make(chan any, count)
	fn := func() {
		select {
		case c <- `🌚`:
		default:
			t.Fatal("onStopFunc should be called only once")
		}
	}

	return c, OptOnStopCombined(fn)
}

func createCombinedOnStartOption(t *testing.T, count int) (<-chan any, CombinedOption) {
	t.Helper()

	c := make(chan any, count)
	fn := func(_ Context) {
		select {
		case c <- `🌞`:
		default:
			t.Fatal("onStart should be called only once")
		}
	}

	return c, OptOnStartCombined(fn)
}
