package actor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_NewWorker(t *testing.T) {
	t.Parallel()

	ctx := ContextStarted()
	sigC := make(chan struct{}, 1)
	workerFunc := func(c Context) WorkerStatus {
		assert.Equal(t, ctx, c)
		sigC <- struct{}{}

		return WorkerContinue
	}

	w := NewWorker(workerFunc)
	assert.NotNil(t, w)

	// We assert that worker will delegate call to supplied workerFunc
	for i := 0; i < 10; i++ {
		assert.Equal(t, WorkerContinue, w.DoWork(ctx))
		assert.Len(t, sigC, 1)
		<-sigC
	}
}

func Test_NewActor(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	a.Start()

	// Asset that worker is going to be executed by actor
	for i := 0; i < 10; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}

	a.Stop()

	// Actor was stopped; wait for done signal
	// and assert that worker is not running
	<-w.ctx.Done()

	p := make(chan int)
	w.doWorkC <- p
	select {
	case <-p:
		assert.FailNow(t, "actor should not be running worker")
	case <-time.After(time.Millisecond * 20):
	}
}

func Test_Actor_OnStartStopFnGetter(t *testing.T) {
	t.Parallel()

	onStartC := make(chan any, 1)
	onStopC := make(chan any, 1)
	onStartFn := func(c Context) { onStartC <- `🌞` }
	onStopFn := func() { onStopC <- `🌚` }

	{
		noopWorker := func(c Context) WorkerStatus { return WorkerContinue }
		a := NewActorImpl(NewWorker(noopWorker))
		assert.Nil(t, a.OnStartFunc())
		assert.Nil(t, a.OnStopFunc())
	}

	{ // Assert that getters will return functions implemented by worker
		w := newWorker()
		a := NewActorImpl(w)

		assert.NotNil(t, a.OnStartFunc())
		a.OnStartFunc()(ContextStarted())
		assert.Equal(t, `🌞`, <-w.onStartC)
		assert.Len(t, w.onStartC, 0)

		assert.NotNil(t, a.OnStopFunc())
		a.OnStopFunc()()
		assert.Equal(t, `🌚`, <-w.onStopC)
		assert.Len(t, w.onStopC, 0)
	}

	{ // Assert that functions specified as option will have more priority then worker
		w := newWorker()
		a := NewActorImpl(w, OptOnStart(onStartFn), OptOnStop(onStopFn))

		assert.NotNil(t, a.OnStartFunc())
		a.OnStartFunc()(ContextStarted())
		assert.Equal(t, `🌞`, <-onStartC)
		assert.Len(t, onStartC, 0)
		assert.Len(t, w.onStartC, 0)

		assert.NotNil(t, a.OnStopFunc())
		a.OnStopFunc()()
		assert.Equal(t, `🌚`, <-onStopC)
		assert.Len(t, onStopC, 0)
		assert.Len(t, w.onStopC, 0)
	}
}

func Test_NewActor_MultipleStartStop(t *testing.T) {
	t.Parallel()

	const count = 3

	onStartC := make(chan struct{}, count)
	onStopC := make(chan struct{}, count)

	w := newWorker()
	a := New(w,
		OptOnStart(func(Context) { onStartC <- struct{}{} }),
		OptOnStop(func() { onStopC <- struct{}{} }),
	)

	// Calling Start() multiple times should have same effect
	// as calling it once
	for i := 0; i < count; i++ {
		a.Start()
	}

	for i := 0; i < 10; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}

	// Calling Stop() multiple times should have same effect
	// as calling it once
	for i := 0; i < count; i++ {
		a.Stop()
	}

	<-w.ctx.Done()
	assert.Len(t, onStartC, 1)
	assert.Len(t, onStopC, 1)
}

func Test_NewActor_Restart(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	// Assert that restarting actor will not impact worker execution

	for i := 0; i < 20; i++ {
		a.Start()

		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)

		a.Stop()
	}
}

func Test_NewActor_StopAfterWorkerEnded(t *testing.T) {
	t.Parallel()

	doWorkC := make(chan chan int)
	workEndedC := make(chan struct{})

	workerFunc := func(c Context) WorkerStatus {
		select {
		case p, ok := <-doWorkC:
			if !ok {
				workEndedC <- struct{}{}
				return WorkerEnd
			}

			p <- 1

			return WorkerContinue

		case <-c.Done():
			// If we receive done signal from Actor we should fail test
			assert.FailNow(t, "worker should be ended")
			return WorkerEnd
		}
	}

	a := New(NewWorker(workerFunc))

	a.Start()

	for i := 0; i < 20; i++ {
		p := make(chan int)
		doWorkC <- p
		assert.Equal(t, 1, <-p)
	}

	// Closing doWorkC will cause worker to end
	close(doWorkC)
	<-workEndedC

	// Stopping actor should produce no effect (since worker has ended)
	a.Stop()
}

func Test_Actor_ContextEndedAfterWorkerEnded(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w,
		OptOnStart(func(c Context) {
			assertContextStarted(t, c)
		}),
		OptOnStop(func() {
			// When OnStop() is called assert that context has ended
			assertContextEnded(t, w.ctx)
		}),
	)

	a.Start()

	for i := 0; i < 20; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}

	a.Stop()

	assertContextEnded(t, w.ctx)
}

func Test_Combine(t *testing.T) {
	t.Parallel()

	const actorsCount = 5

	onStartC := make(chan struct{}, actorsCount)
	onStopC := make(chan struct{}, actorsCount)
	actors := make([]Actor, actorsCount)

	for i := 0; i < actorsCount; i++ {
		actors[i] = New(newWorker(),
			OptOnStart(func(Context) { onStartC <- struct{}{} }),
			OptOnStop(func() { onStopC <- struct{}{} }),
		)
	}

	// Assert that starting and stopping combined actors
	// will start and stop all individual actors
	a := Combine(actors...)
	a.Start()
	a.Stop()
	assert.Len(t, onStartC, actorsCount)
	assert.Len(t, onStopC, actorsCount)
}

func Test_Noop(t *testing.T) {
	t.Parallel()

	a := Noop()

	a.Start()
	a.Stop()

	a.Start()
	a.Start()
	a.Stop()
	a.Stop()
}

func Test_Idle_Options(t *testing.T) {
	t.Parallel()

	onStartC := make(chan struct{}, 1)
	onStopC := make(chan struct{}, 1)

	a := Idle(
		OptOnStart(func(Context) { onStartC <- struct{}{} }),
		OptOnStop(func() { onStopC <- struct{}{} }),
	)

	a.Start()
	<-onStartC
	a.Start() // Should have no effect
	assert.Len(t, onStartC, 0)

	a.Stop()
	<-onStopC
	a.Stop() // Should have no effect
	assert.Len(t, onStopC, 0)
}

func newWorker() *worker {
	return &worker{
		doWorkC:  make(chan chan int, 1),
		onStartC: make(chan any, 1),
		onStopC:  make(chan any, 1),
	}
}

type worker struct {
	workIteration int
	doWorkC       chan chan int
	ctx           Context
	onStartC      chan any
	onStopC       chan any
}

func (w *worker) DoWork(c Context) WorkerStatus {
	w.ctx = c // saving ref to context so we can assert that context

	select {
	case <-c.Done():
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

func (w *worker) OnStart(c Context) {
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
