package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_NewWorker(t *testing.T) {
	t.Parallel()

	ctx := NewContext()
	sigC := make(chan struct{}, 1)
	workerFunc := func(c Context) WorkerStatus {
		sigC <- struct{}{}
		return WorkerContinue
	}

	w := NewWorker(workerFunc)
	assert.NotNil(t, w)

	// We assert that worker will delgate call to supplied workerFunc
	assert.Equal(t, WorkerContinue, w.DoWork(ctx))
	assert.Len(t, sigC, 1)
}

func Test_NewActor(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

	a.Start()
	defer a.Stop()

	for i := 0; i < 20; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}
}

func Test_NewActor_StartStop(t *testing.T) {
	t.Parallel()

	w := newWorker()
	a := New(w)

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

func drain(c chan struct{}) {
	for len(c) > 0 {
		<-c
	}
}

func newWorker() *worker {
	return &worker{doWorkC: make(chan chan int, 1)}
}

type worker struct {
	workIteration int
	doWorkC       chan chan int
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
		return WorkerEnd
	}
}
