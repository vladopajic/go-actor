package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_NewActor(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := New(w)

	a.Start()
	defer a.Stop()

	for i := 0; i < count; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}
}

func Test_NewActor_StartStop(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := New(w)

	for i := 0; i < count; i++ {
		a.Start()

		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)

		a.Stop()
	}
}

func Test_NewActor_StopAfterNoWork(t *testing.T) {
	t.Parallel()

	const count = 20

	w := &worker{doWorkC: make(chan chan int, count)}
	a := New(w)

	a.Start()
	defer a.Stop()

	for i := 0; i < count; i++ {
		p := make(chan int)
		w.doWorkC <- p
		assert.Equal(t, i, <-p)
	}

	close(w.doWorkC)
}

func Test_NewActor_OptOnStartStop(t *testing.T) {
	t.Parallel()

	onStartC := make(chan struct{}, 1)
	onStopC := make(chan struct{}, 1)

	w := &worker{doWorkC: make(chan chan int, 1)}
	a := New(w,
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

	w := &worker{doWorkC: make(chan chan int, 1)}
	for i := 0; i < actorsCount; i++ {
		actors[i] = New(w,
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

func Test_Context(t *testing.T) {
	t.Parallel()

	ctx := NewContext()

	assert.NotNil(t, ctx.Done())
	assert.Len(t, ctx.Done(), 0)

	ctx.SignalEnd()
	assert.Len(t, ctx.Done(), 1)
}
