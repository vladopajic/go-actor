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

// Test asserts basic Actor functions
func Test_Actor_New(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	a.Start()

	// Asset that worker is going to be executed by actor
	assertDoWork(t, w.doWorkC, 0)

	a.Stop()

	// After stopping actor assert that worker is not going to be executed
	assertNoWork(t, w.doWorkC)
}

// Test asserts that restarting actor will no impact on worker execution.
func Test_Actor_Restart(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	for i := range 20 {
		a.Start()

		assertDoWork(t, w.doWorkC, i*workIterationsPerAssert)

		a.Stop()
	}

	AssertStartStopAtRandom(t, a)
}

// Test asserts that nothing should happen if
// Start() or Stop() methods are called multiple times.
func Test_Actor_MultipleStartStop(t *testing.T) {
	t.Parallel()

	const count = 3

	onStartC := make(chan any, count)
	onStopC := make(chan any, count)

	w := newWorker()
	a := New(w,
		OptOnStart(func(Context) { onStartC <- `🌞` }),
		OptOnStop(func() { onStopC <- `🌚` }),
	)

	// Calling Start() multiple times should have same effect as calling it once
	for range count {
		a.Start()
	}

	assertDoWork(t, w.doWorkC, 0)

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

	readySigC := make(chan any)
	onStartC, onStopC := make(chan any, 1), make(chan any, 1)
	onStartFn := func(_ Context) { <-readySigC; onStartC <- `🌞` }
	onStopFn := func() { <-readySigC; onStopC <- `🌚` }

	{
		// Nothing should happen when calling OnStart and OnStop
		// when callbacks are not defined (no panic should occur)
		w := NewWorker(func(_ Context) WorkerStatus { return WorkerContinue })
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
		w := NewWorker(func(_ Context) WorkerStatus { return WorkerContinue })
		a := NewActorImpl(w, OptOnStart(onStartFn), OptOnStop(onStopFn))

		go a.OnStart()
		readySigC <- struct{}{}

		assert.Equal(t, `🌞`, <-onStartC)
		assert.Empty(t, onStartC)

		go a.OnStop()
		readySigC <- struct{}{}

		assert.Equal(t, `🌚`, <-onStopC)
		assert.Empty(t, onStopC)
	}

	{
		// Assert that actor will call callbacks implemented by worker,
		// then callbacks passed by options
		w := newWorker()
		a := NewActorImpl(w, OptOnStart(onStartFn), OptOnStop(onStopFn))

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
				defer close(workEndedC)
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

	assertDoWork(t, doWorkC, 0)

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

	assertDoWork(t, w.doWorkC, 0)

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
	ctx           Context //nolint:containedctx // ctx needed for test case
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

func assertDoWork(t *testing.T, doWorkC chan chan int, start int) {
	t.Helper()

	for i := start; i < workIterationsPerAssert; i++ {
		p := make(chan int)
		doWorkC <- p
		assert.Equal(t, i, <-p)
	}
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
}
