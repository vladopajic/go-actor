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

	w := newWorkerWithDoneC()
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
	<-w.doneC

	p := make(chan int)
	w.doWorkC <- p
	select {
	case <-p:
		assert.FailNow(t, "actor should not be running worker")
	case <-time.After(time.Millisecond * 20):
	}
}

func Test_NewActor_OptOnStartStop(t *testing.T) {
	t.Parallel()

	onStartC := make(chan struct{}, 1)
	onStopC := make(chan struct{}, 1)

	a := New(newWorker(),
		OptOnStart(func() { onStartC <- struct{}{} }),
		OptOnStop(func() { onStopC <- struct{}{} }),
	)

	a.Start()
	<-onStartC

	a.Stop()
	<-onStopC
}

func Test_NewActor_MultipleStartStop(t *testing.T) {
	t.Parallel()

	const count = 3

	onStartC := make(chan struct{}, count)
	onStopC := make(chan struct{}, count)

	w := newWorkerWithDoneC()
	a := New(w,
		OptOnStart(func() { onStartC <- struct{}{} }),
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

	<-w.doneC
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

	var ctx Context

	doWorkC := make(chan chan int)
	workEndedC := make(chan struct{})

	workerFunc := func(c Context) WorkerStatus {
		ctx = c // saving ref to context so we can assert that context has ended

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

	// Stoping actor should produce no effect (since worker has ended)
	a.Stop()

	assertContextEnded(t, ctx)
}

func Test_Combine_StartAll_StopAll(t *testing.T) {
	t.Parallel()

	const actorsCount = 5

	onStartC := make(chan struct{}, actorsCount)
	onStopC := make(chan struct{}, actorsCount)
	actors := make([]Actor, actorsCount)

	for i := 0; i < actorsCount; i++ {
		actors[i] = New(newWorker(),
			OptOnStart(func() { onStartC <- struct{}{} }),
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

	// Similar test but for StartAll and StopAll methods
	drain(onStartC)
	drain(onStopC)
	assert.Len(t, onStartC, 0)
	assert.Len(t, onStopC, 0)
	StartAll(actors...)
	StopAll(actors...)
	assert.Len(t, onStartC, actorsCount)
	assert.Len(t, onStopC, actorsCount)
}

func Test_Noop(t *testing.T) {
	t.Parallel()

	a := Noop()

	a.Start()
	a.Stop()
}

func Test_Idle(t *testing.T) {
	t.Parallel()

	onStartC := make(chan struct{}, 1)
	onStopC := make(chan struct{}, 1)

	a := Idle(
		OptOnStart(func() { onStartC <- struct{}{} }),
		OptOnStop(func() { onStopC <- struct{}{} }),
	)

	a.Start()
	<-onStartC

	a.Stop()
	<-onStopC
}

func drain(c chan struct{}) {
	for len(c) > 0 {
		<-c
	}
}

func newWorker() *worker {
	return &worker{
		doWorkC: make(chan chan int, 1),
		doneC:   nil,
	}
}

func newWorkerWithDoneC() *worker {
	return &worker{
		doWorkC: make(chan chan int, 100),
		doneC:   make(chan struct{}),
	}
}

type worker struct {
	workIteration int
	doWorkC       chan chan int
	doneC         chan struct{}
}

func (w *worker) DoWork(c Context) WorkerStatus {
	select {
	case p, ok := <-w.doWorkC:
		if !ok {
			return WorkerEnd
		}

		p <- w.workIteration
		w.workIteration++

		return WorkerContinue

	case <-c.Done():
		if w.doneC != nil {
			close(w.doneC)
		}

		return WorkerEnd
	}
}
