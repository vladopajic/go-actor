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

// Test asserts that Start() and Stop() is delegated to all combined actors.
func Test_Combine(t *testing.T) {
	t.Parallel()

	combineParallel(t, 0, testCombine)
	combineParallel(t, 1, testCombine)
	combineParallel(t, 2, testCombine)
	combineParallel(t, 10, testCombine)
}

func testCombine(t *testing.T, actorsCount int) {
	t.Helper()

	onStartC, onStartFn := createOnStartOption(t, actorsCount)
	onStopC, onStopFn := createOnStopOption(t, actorsCount)
	actors := createActors(actorsCount, OptOnStart(onStartFn), OptOnStop(onStopFn))

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

// Test asserts that combined actor will invoke OnStart and OnStop callbacks.
func Test_Combine_OptOnStopOptOnStart(t *testing.T) {
	t.Parallel()

	combineParallel(t, 0, testCombineOptOnStopOptOnStart)
	combineParallel(t, 1, testCombineOptOnStopOptOnStart)
	combineParallel(t, 2, testCombineOptOnStopOptOnStart)
	combineParallel(t, 10, testCombineOptOnStopOptOnStart)
}

func testCombineOptOnStopOptOnStart(t *testing.T, actorsCount int) {
	t.Helper()

	onStartC, onStartFn := createOnStartOption(t, 1)
	onStopC, onStopFn := createOnStopOption(t, 1)
	actors := createActors(actorsCount)

	a := Combine(actors...).
		WithOptions(OptOnStartCombined(onStartFn), OptOnStopCombined(onStopFn)).
		Build()

	a.Start()

	a.Stop()
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	a.Stop() // should have no effect
	assert.Equal(t, `🌚`, <-onStopC)
	assert.Equal(t, `🌞`, <-onStartC)
	assert.Empty(t, onStopC)
	assert.Empty(t, onStartC)
}

// Test asserts that combined actor will call Stop() only once on
// combined actors even when Stop() is called concurrently.
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
	stopFinishedC := make(chan any, stopCallsCount)

	for range stopCallsCount {
		go func() {
			a.Stop()
			stopFinishedC <- `🛑`
		}()
	}

	close(stopConcurrentlyFinishedC)
	drainC(stopFinishedC, stopCallsCount)

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
		onStartC, onStartFn := createOnStartOption(t, actorsCount)
		onStopC, onStopFn := createOnStopOption(t, actorsCount)
		actors := createActors(actorsCount, OptOnStart(onStartFn), OptOnStop(onStopFn))

		a := Combine(actors...).WithOptions(OptStopTogether()).Build()

		a.Start()
		drainC(onStartC, actorsCount)

		// stop actor and assert that all actors will be stopped
		actors[i].Stop()
		drainC(onStopC, actorsCount)
	}
}

// Test_Combine_OthersNotStopped asserts that if any of underlying
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
		onStartC, onStartFn := createOnStartOption(t, actorsCount)
		onStopC, onStopFn := createOnStopOption(t, actorsCount)
		actors := createActors(actorsCount, OptOnStart(onStartFn), OptOnStop(onStopFn))

		a := Combine(actors...).WithOptions().Build()

		a.Start()
		drainC(onStartC, actorsCount)

		// stop actors individually, and expect that after some time
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

//nolint:maintidx // relax
func Test_Combine_StopParallel(t *testing.T) {
	t.Parallel()

	const count = 10

	setup := func(isParallel bool) (chan struct{}, chan struct{}) {
		c1 := make(chan struct{}, count)
		c2 := make(chan struct{})
		actors := make([]Actor, count)

		for i := range count {
			actors[i] = Idle(OptOnStop(func() {
				c1 <- struct{}{}

				<-c2
			}))
		}

		a := Combine(actors...).
			WithOptions(OptStopParallelWith(isParallel)).
			Build()

		a.Start()
		go a.Stop()

		// give some time for OnStop to be called
		time.Sleep(time.Millisecond * 100) //nolint:forbidigo // relax

		return c1, c2
	}

	t.Run("sequential", func(t *testing.T) {
		t.Parallel()

		c1, c2 := setup(false)

		for range count {
			c2 <- struct{}{}

			assert.GreaterOrEqual(t, len(c1), 1)
			assert.LessOrEqual(t, len(c1), 2)
			<-c1
		}
	})

	t.Run("parallel", func(t *testing.T) {
		t.Parallel()

		c1, c2 := setup(true)

		for i := range count {
			c2 <- struct{}{}

			assert.Len(t, c1, count-i)
			<-c1
		}
	})
}

// Test asserts that wrapActors is correctly wrapping actors with onStopFunc callback.
func Test_Combine_WrapActors(t *testing.T) {
	t.Parallel()

	stopActorC := make(chan any, 3)
	actors := []Actor{
		New(newWorker(), OptOnStop(func() { stopActorC <- "new" })),
		Idle(OptOnStop(func() { stopActorC <- "idle" })),
		Combine(createActors(10)...).
			WithOptions(OptOnStopCombined(func() { stopActorC <- "combine" })).
			Build(),
	}

	// onStopFunc will write to stopWrappedC channel and result is asserted at end
	// of this test. if we assert in this callback then it might be the case that
	// function is never called so it will always pass.
	stopWrappedC := make(chan any, 3)
	onStopFunc := func() { stopWrappedC <- <-stopActorC }

	wActors := WrapActors(actors, onStopFunc)

	// Start then stop wrapped actors
	for _, a := range wActors {
		a.Start()
	}

	for _, a := range wActors {
		a.Stop()
	}

	// Expect data on stopWrappedC in order in which actors appeared in this list
	assert.Equal(t, "new", <-stopWrappedC)
	assert.Equal(t, "idle", <-stopWrappedC)
	assert.Equal(t, "combine", <-stopWrappedC)
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
