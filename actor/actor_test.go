package actor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

// Test asserts that worker created with NewWorker utility
// will delegate call to supplied WorkerFunc.
func Test_NewWorker(t *testing.T) {
	t.Parallel()

	ctx := ContextStarted()
	workC := make(chan any, 1)
	workerFunc := func(c Context) WorkerStatus {
		assert.Equal(t, ctx, c)
		workC <- `🛠️`

		return WorkerContinue
	}

	w := NewWorker(workerFunc)
	assert.NotNil(t, w)

	for range 10 {
		assert.Equal(t, WorkerContinue, w.DoWork(ctx))
		assert.Equal(t, `🛠️`, <-workC)
	}
}

// Test asserts that Actor will execute underlying worker.
func Test_Actor_New(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	a.Start()

	// Asset that worker is going to be executed by actor
	assertDoWork(t, w.doWorkC)

	a.Stop()

	// After stopping actor assert that worker is not going to be executed
	assertNoWork(t, w.doWorkC)
}

// Test asserts that restarting actor will not impact worker's state.
func Test_Actor_Restart(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	for i := range 10 {
		a.Start()
		assertDoWorkWithStart(t, w.doWorkC, i*workIterationsPerAssert)
		a.Stop()
		assertNoWork(t, w.doWorkC)
	}
}

// Test with AssertStartStopAtRandom.
func Test_Actor_StartStopAtRandom(t *testing.T) {
	t.Parallel()

	AssertStartStopAtRandom(t, New(newWorker()))
}

// Test asserts that calling Start() and Stop() methods multiple times
// will not have effect.
func Test_Actor_MultipleStartStop(t *testing.T) {
	t.Parallel()

	const count = 3

	onStartC, onStartFn := createOnStartOption(t, count)
	onStopC, onStopFn := createOnStopOption(t, count)

	w := newWorker()
	a := New(w, OptOnStart(onStartFn), OptOnStop(onStopFn))

	// Calling Start() multiple times should have same effect as calling it once
	for range count {
		a.Start()
	}

	assertDoWork(t, w.doWorkC)

	// Calling Stop() multiple times should have same effect as calling it once
	for range count {
		a.Stop()
	}

	assert.Len(t, onStartC, 1)
	assert.Len(t, onStopC, 1)
}

// Test asserts that actor will invoke OnStart and OnStop callbacks.
//
//nolint:maintidx // long test case
func Test_Actor_OnStartOnStop(t *testing.T) {
	t.Parallel()

	{
		// Nothing should happen when calling OnStart and OnStop
		// when callbacks are not defined (no panic should occur)
		w := NewWorker(func(Context) WorkerStatus { return WorkerContinue })
		a := NewActorImpl(w)
		a.OnStart()
		a.OnStop()
	}

	{ // Assert that actor will call callbacks implemented by worker
		w := newWorker()
		a := NewActorImpl(w)

		a.OnStart()
		assert.Equal(t, `🌞`, <-w.onStartC)
		assert.Empty(t, w.onStartC)

		a.OnStop()
		assert.Equal(t, `🌚`, <-w.onStopC)
		assert.Empty(t, w.onStopC)
	}

	{ // Assert that actor will call callbacks passed by options
		onStartC, onStartFn := createOnStartOption(t, 1)
		onStopC, onStopFn := createOnStopOption(t, 1)
		w := NewWorker(func(Context) WorkerStatus { return WorkerContinue })
		a := NewActorImpl(w, OptOnStart(onStartFn), OptOnStop(onStopFn))

		a.OnStart()
		assert.Equal(t, `🌞`, <-onStartC)
		assert.Empty(t, onStartC)

		a.OnStop()
		assert.Equal(t, `🌚`, <-onStopC)
		assert.Empty(t, onStopC)
	}

	{
		// Assert that actor will call callbacks implemented by worker,
		// then callbacks passed by options.
		readySigC := make(chan any)
		onStartC, onStopC := make(chan any, 1), make(chan any, 1)
		onStartOpt := OptOnStart(func(Context) { <-readySigC; onStartC <- `🌞` })
		onStopOpt := OptOnStop(func() { <-readySigC; onStopC <- `🌚` })
		w := newWorker()
		a := NewActorImpl(w, onStartOpt, onStopOpt)

		go a.OnStart()

		assert.Equal(t, `🌞`, <-w.onStartC)
		assert.Empty(t, w.onStartC)
		assert.Empty(t, onStartC)

		readySigC <- struct{}{}

		assert.Equal(t, `🌞`, <-onStartC)
		assert.Empty(t, w.onStartC)
		assert.Empty(t, onStartC)

		go a.OnStop()

		assert.Equal(t, `🌚`, <-w.onStopC)
		assert.Empty(t, w.onStopC)
		assert.Empty(t, onStopC)

		readySigC <- struct{}{}

		assert.Equal(t, `🌚`, <-onStopC)
		assert.Empty(t, w.onStopC)
		assert.Empty(t, onStopC)
	}
}

// Test asserts that OnStart is executed in separate goroutine.
//
// Note: The same property is not checked for OnStop, because when Stop()
// is called it blocks until actor finishes - so it may be blocked due to:
//  1. executing OnStop() in the same goroutine as caller of Stop(),
//     case of Idle(...) actor, or
//  2. blocked on waiting for actor goroutine to finish, case of New(...) actor.
func Test_Actor_OnStartGoroutine(t *testing.T) {
	t.Parallel()

	assert := func(fact func(...Option) Actor) {
		c := make(chan any)
		a := fact(OptOnStart(func(Context) { c <- `🌞` }))

		a.Start()
		defer a.Stop()

		// if OnStart is execute is same goroutine, reading from channel <-c
		// will block, and test will never complete.
		assert.Equal(t, `🌞`, <-c)
	}

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		assert(func(opt ...Option) Actor {
			return New(newWorker(), opt...)
		})
	})

	t.Run("Idle", func(t *testing.T) {
		t.Parallel()
		assert(Idle)
	})
}

// Test asserts that OnStop is called even if actor is stopped
// before OnStart finishes.
func Test_Actor_OnStopCalledIfStoppedEarly(t *testing.T) {
	t.Parallel()

	assert := func(fact func(...Option) Actor) {
		startedC := make(chan any)
		blockingOnStart := func(ctx Context) {
			close(startedC)
			select {
			case <-ctx.Done():
			case <-time.After(time.Hour):
			}
		}
		onStopC := make(chan any, 1)
		a := fact(
			OptOnStart(blockingOnStart),
			OptOnStop(func() { onStopC <- `🌚` }),
		)

		a.Start()
		<-startedC
		a.Stop()
		assert.Equal(t, `🌚`, <-onStopC)
	}

	t.Run("New", func(t *testing.T) {
		t.Parallel()
		assert(func(opt ...Option) Actor {
			return New(newWorker(), opt...)
		})
	})

	t.Run("Idle", func(t *testing.T) {
		t.Parallel()
		assert(Idle)
	})
}

// Test asserts that Worker is not called if actor is stopped early.
func Test_Actor_WorkerSkippedIfStoppedEarly(t *testing.T) {
	t.Parallel()

	startedC := make(chan any)
	a := New(
		NewWorker(func(Context) WorkerStatus {
			assert.Fail(t, "DoWork should not be called")
			return WorkerEnd
		}),
		OptOnStart(func(ctx Context) {
			close(startedC)
			select {
			case <-ctx.Done():
			case <-time.After(time.Hour):
			}
		}))

	a.Start()
	<-startedC
	a.Stop()
}

// Test asserts that actor should stop after worker
// has signaled that there is no more work via WorkerEnd signal.
func Test_Actor_StopAfterWorkerEnded(t *testing.T) {
	t.Parallel()

	var ctx Context

	workIteration := 0
	doWorkC := make(chan chan int)
	workEndedC := make(chan struct{})
	workerFunc := func(c Context) WorkerStatus {
		ctx = c

		// assert that DoWork should not be called
		// after WorkerEnd signal is returned
		select {
		case <-workEndedC:
			assert.FailNow(t, "worker should be ended")
		default:
		}

		select {
		case p, ok := <-doWorkC:
			if !ok {
				close(workEndedC)
				return WorkerEnd
			}

			p <- workIteration

			workIteration++

			return WorkerContinue

		case <-c.Done():
			// Test should fail if done signal is received from Actor
			assert.FailNow(t, "worker should be ended")
			return WorkerEnd
		}
	}

	a := New(NewWorker(workerFunc))

	a.Start()

	assertDoWork(t, doWorkC)

	// Closing doWorkC will cause worker to end
	close(doWorkC)

	// Assert that context is ended after worker ends.
	// Small sleep is needed in order to fix potentially race condition
	// if actor's goroutine does not finish before this check.
	<-workEndedC
	time.Sleep(time.Millisecond * 10) //nolint:forbidigo // explained above
	assertContextEnded(t, ctx)

	// Stopping actor should produce no effect (since worker has ended)
	a.Stop()

	assertContextEnded(t, ctx)
}

// Test asserts that context supplied to worker will be ended
// after actor is stopped.
func Test_Actor_ContextEndedAfterStop(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w,
		OptOnStart(func(ctx Context) {
			assertContextStarted(t, ctx)
			assert.Equal(t, ctx, w.ctx)
		}),
		OptOnStop(func() {
			// When OnStop() is called assert that context has ended
			assertContextEnded(t, w.ctx)
		}),
	)

	a.Start()

	assertDoWork(t, w.doWorkC)

	a.Stop()

	assertContextEnded(t, w.ctx)
}

// This test could not assert much, except that test
// should not panic when Start() and Stop() are called.
func Test_Noop(t *testing.T) {
	t.Parallel()

	AssertStartStopAtRandom(t, Noop())
}

// This test could not assert much, except that test
// should not panic when Start() and Stop() are called.
func Test_Idle(t *testing.T) {
	t.Parallel()

	AssertStartStopAtRandom(t, Idle())
}

// Test asserts that OnStart and OnStop callbacks
// are being called.
func Test_Idle_Options(t *testing.T) {
	t.Parallel()

	var ctx Context

	onStartC, onStopC := make(chan any, 1), make(chan any, 1)
	a := Idle(
		OptOnStart(func(c Context) { ctx = c; onStartC <- `🌞` }),
		OptOnStop(func() { onStopC <- `🌚` }),
	)

	a.Start()
	assert.Equal(t, `🌞`, <-onStartC)
	assertContextStarted(t, ctx)
	a.Start() // Should have no effect
	assert.Empty(t, onStartC)

	a.Stop()
	assert.Equal(t, `🌚`, <-onStopC)
	assertContextEnded(t, ctx)
	a.Stop() // Should have no effect
	assert.Empty(t, onStopC)
}

func newWorker() *worker {
	return &worker{
		doWorkC:  make(chan chan int, 1),
		onStartC: make(chan any, 1),
		onStopC:  make(chan any, 1),
	}
}

var (
	_ Worker          = (*worker)(nil)
	_ StartableWorker = (*worker)(nil)
	_ StoppableWorker = (*worker)(nil)
)

type worker struct {
	workIteration int
	doWorkC       chan chan int
	ctx           Context
	onStartC      chan any
	onStopC       chan any
}

func (w *worker) DoWork(ctx Context) WorkerStatus {
	select {
	case <-ctx.Done():
		return WorkerEnd

	case p, ok := <-w.doWorkC:
		if !ok {
			return WorkerEnd
		}

		p <- w.workIteration

		w.workIteration++

		return WorkerContinue
	}
}

func (w *worker) OnStart(ctx Context) {
	w.ctx = ctx // saving ref to context so it can be asserted in tests

	select {
	case w.onStartC <- `🌞`:
	default:
	}
}

func (w *worker) OnStop() {
	select {
	case w.onStopC <- `🌚`:
	default:
	}
}

const workIterationsPerAssert = 20

func assertDoWorkWithStart(t *testing.T, doWorkC chan chan int, start int) {
	t.Helper()

	for i := start; i < workIterationsPerAssert; i++ {
		p := make(chan int)
		doWorkC <- p
		assert.Equal(t, i, <-p)
	}
}

func assertDoWork(t *testing.T, doWorkC chan chan int) {
	t.Helper()

	assertDoWorkWithStart(t, doWorkC, 0)
}

func assertNoWork(t *testing.T, doWorkC chan chan int) {
	t.Helper()

	p := make(chan int)
	doWorkC <- p

	select {
	case <-p:
		assert.FailNow(t, "actor should not be running worker")
	case <-time.After(time.Millisecond * 20):
	}

	// drain work request in order to make the same state before calling
	// this helper function
	<-doWorkC
}
