package actor_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func combineParallel(
	t *testing.T,
	actorsCount int,
	testFn func(*testing.T, int),
) {
	t.Helper()

	t.Run(fmt.Sprintf("actors count %v", actorsCount), func(t *testing.T) {
		t.Parallel()
		testFn(t, actorsCount)
	})
}

func Test_Combine_TestSuite(t *testing.T) {
	t.Parallel()

	combineParallel(t, 0, testCombineTestSuite)
	combineParallel(t, 1, testCombineTestSuite)
	combineParallel(t, 2, testCombineTestSuite)
	combineParallel(t, 10, testCombineTestSuite)
}

func testCombineTestSuite(t *testing.T, actorsCount int) {
	t.Helper()

	TestSuite(t, func() Actor {
		actors := createActors(actorsCount)
		return Combine(actors...).Build()
	})
}

// Test asserts that all Start and Stop is delegated to all combined actors.
func Test_Combine(t *testing.T) {
	t.Parallel()

	combineParallel(t, 0, testCombine)
	combineParallel(t, 1, testCombine)
	combineParallel(t, 2, testCombine)
	combineParallel(t, 10, testCombine)
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

	combineParallel(t, 0, testCombineOptOnStopOptOnStart)
	combineParallel(t, 1, testCombineOptOnStopOptOnStart)
	combineParallel(t, 2, testCombineOptOnStopOptOnStart)
	combineParallel(t, 10, testCombineOptOnStopOptOnStart)
}

func testCombineOptOnStopOptOnStart(t *testing.T, actorsCount int) {
	t.Helper()

	onStatC, onStartOpt := createCombinedOnStartOption(t, 1)
	onStopC, onStopOpt := createCombinedOnStopOption(t, 1)
	actors := createActors(actorsCount)

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

func Test_Combine_StoppingOnce(t *testing.T) {
	t.Parallel()

	combineParallel(t, 0, testCombineStoppingOnce)
	combineParallel(t, 1, testCombineStoppingOnce)
	combineParallel(t, 2, testCombineStoppingOnce)
	combineParallel(t, 10, testCombineStoppingOnce)
	combineParallel(t, 20, testCombineStoppingOnce)
}

func testCombineStoppingOnce(t *testing.T, actorsCount int) {
	t.Helper()

	c := atomic.Int32{}
	stopConcurrentlyFinishedC := make(chan any)
	actors := make([]Actor, actorsCount)

	for i := range actorsCount {
		actors[i] = delegateActor{stop: func() {
			<-stopConcurrentlyFinishedC
			c.Add(1)
		}}
	}

	a := Combine(actors...).Build()
	a.Start()

	// Call Stop() multiple times in separate goroutine to force concurrency
	const stopCallsCount = 100
	stopFinsihedC := make(chan any, stopCallsCount)

	for range stopCallsCount {
		go func() {
			a.Stop()
			stopFinsihedC <- `🛑`
		}()
	}

	close(stopConcurrentlyFinishedC)
	drainC(stopFinsihedC, stopCallsCount)

	assert.Equal(t, actorsCount, int(c.Load()))
}

// Test_Combine_OptStopTogether asserts that all actors will end as soon
// as first actors ends.
func Test_Combine_OptStopTogether(t *testing.T) {
	t.Parallel()

	// no need to test OptStopTogether for actors count < 2
	// because single actor is always "stopped together"

	combineParallel(t, 1, testCombineOptStopTogether)
	combineParallel(t, 2, testCombineOptStopTogether)
	combineParallel(t, 10, testCombineOptStopTogether)
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

// Test_Combine_OthersNotStopped asserts that if any of underlaying
// actors end it wont affect other actors. This is default behavior when
// `OptStopTogether` is not provided.
func Test_Combine_OthersNotStopped(t *testing.T) {
	t.Parallel()

	combineParallel(t, 1, testCombineOthersNotStopped)
	combineParallel(t, 2, testCombineOthersNotStopped)
	combineParallel(t, 10, testCombineOthersNotStopped)
}

func testCombineOthersNotStopped(t *testing.T, actorsCount int) {
	t.Helper()

	for i := range actorsCount {
		onStartC := make(chan any, actorsCount)
		onStopC := make(chan any, actorsCount)
		onStart := OptOnStart(func(Context) { onStartC <- `🌞` })
		onStop := OptOnStop(func() { onStopC <- `🌚` })
		actors := createActors(actorsCount, onStart, onStop)

		a := Combine(actors...).WithOptions().Build()

		a.Start()
		drainC(onStartC, actorsCount)

		// stop actors indiviually, and expect that after some time
		// there wont be any other actors stopping.
		stopCount := i + 1
		for j := range stopCount {
			actors[j].Stop()
		}

		drainC(onStopC, stopCount)

		select {
		case <-onStartC:
			t.Fatal("should not have any more stopped actors")
		case <-time.After(time.Millisecond * 20):
		}

		a.Stop()
		drainC(onStopC, actorsCount-stopCount)
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
