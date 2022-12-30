package actor

import (
	"sync"
)

// Actor is computational entity that executes Worker in individual goroutine.
type Actor interface {
	// Start spawns new goroutine and begins Worker execution.
	//
	// Execution will last until Stop() method is called or Worker returned
	// status indicating that Worker has ended (there is no more work).
	Start()

	// Stop sends signal to Worker to stop execution. Method will block
	// until Worker finishes.
	Stop()
}

// WorkerStatus is returned by Worker's DoWork function indicating if Actor should
// continue executing Worker.
type WorkerStatus int8

const (
	WorkerContinue WorkerStatus = 1
	WorkerEnd      WorkerStatus = 2
)

// Worker is entity which encapsulates Actor's executable logic.
//
// Worker's implementation should listen on messages sent via Mailboxes and preform
// actions by sending new messages or creating new actors.
type Worker interface {
	// DoWork function is encapsulating single executable unit of work for this Worker.
	//
	// Context is provided so Worker can listen and respond on stop signal sent from Actor.
	//
	// WorkerStatus is returned indicating if Actor should continue executing this Worker.
	// Actor will check this status and stop execution if Worker has no more work, otherwise
	// proceed execution.
	DoWork(c Context) WorkerStatus
}

// WorkerFunc is signature of Worker's DoWork function.
type WorkerFunc = func(c Context) WorkerStatus

// startableWorker defines optional interface which Worker can implement
type startableWorker interface {
	// OnStart is called right before DoWork() is called for first time. It can be used to
	// initialize Worker as it will be called only once.
	// Context is provided in case when Actor is stopped early and OnStop should terminated
	// with initialization. This is same Context as one which will be provided to DoWork method
	// in later stages of Worker lifecycle.
	OnStart(Context)
}

// stoppableWorker defines optional interface which Worker can implement
type stoppableWorker interface {
	// OnStop is called after last DoWork() returns. It can be used to release all
	// resources occupied by Worker.
	// Context is not proved as at this point as it was already ended.
	OnStop()
}

// NewWorker returns basic Worker implementation which delegates
// DoWork to supplied WorkerFunc.
func NewWorker(fn WorkerFunc) Worker {
	return &worker{fn}
}

type worker struct {
	fn WorkerFunc
}

func (w *worker) DoWork(c Context) WorkerStatus {
	return w.fn(c)
}

// New returns new Actor with specified Worker and Options.
func New(w Worker, opt ...Option) Actor {
	return &actor{
		worker:  w,
		options: newOptions(opt),
	}
}

type actor struct {
	worker            Worker
	options           options
	ctx               *context
	workEndedSigC     chan struct{}
	workerRunning     bool
	workerRunningLock sync.Mutex
}

func (a *actor) Stop() {
	a.workerRunningLock.Lock()
	if !a.workerRunning {
		a.workerRunningLock.Unlock()
		return
	}

	a.workEndedSigC = make(chan struct{})
	a.workerRunningLock.Unlock()

	a.ctx.end()

	<-a.workEndedSigC
}

func (a *actor) Start() {
	a.workerRunningLock.Lock()
	defer a.workerRunningLock.Unlock()

	if a.workerRunning {
		return
	}

	a.ctx = newContext()
	a.workerRunning = true

	go a.doWork()
}

// doWork executes Worker of this Actor until
// Actor or Worker has signaled to stop.
func (a *actor) doWork() {
	if fn := a.onStartFunc(); fn != nil {
		fn(a.ctx)
	}

	for status := WorkerContinue; status == WorkerContinue; {
		status = a.worker.DoWork(a.ctx)
	}

	a.ctx.end()

	if fn := a.onStopFunc(); fn != nil {
		fn()
	}

	{ // Worker has finished
		a.workerRunningLock.Lock()
		a.workerRunning = false
		if c := a.workEndedSigC; c != nil {
			c <- struct{}{}
		}
		a.workerRunningLock.Unlock()
	}
}

func (a *actor) onStartFunc() func(Context) {
	if fn := a.options.Actor.OnStartFunc; fn != nil {
		return fn
	}

	if w, ok := a.worker.(startableWorker); ok {
		return w.OnStart
	}

	return nil
}

func (a *actor) onStopFunc() func() {
	if fn := a.options.Actor.OnStopFunc; fn != nil {
		return fn
	}

	if w, ok := a.worker.(stoppableWorker); ok {
		return w.OnStop
	}

	return nil
}

// Combine returns single Actor which combines all specified actors into one.
// Calling Start or Stop function on this Actor will invoke respective function
// on all Actors provided to this function.
func Combine(actors ...Actor) Actor {
	return &combinedActor{actors}
}

type combinedActor struct {
	actors []Actor
}

func (a *combinedActor) Stop() {
	for _, a := range a.actors {
		a.Stop()
	}
}

func (a *combinedActor) Start() {
	for _, a := range a.actors {
		a.Start()
	}
}

// Idle returns new Actor without Worker.
func Idle(opt ...Option) Actor {
	return &idleActor{
		options: newOptions(opt),
	}
}

type idleActor struct {
	options options
	ctx     *context
	lock    sync.Mutex
}

func (a *idleActor) Start() {
	a.lock.Lock()

	// early return if this actor is already running
	if a.ctx != nil {
		a.lock.Unlock()
		return
	}

	a.ctx = newContext()
	a.lock.Unlock()

	if fn := a.options.Actor.OnStartFunc; fn != nil {
		fn(a.ctx)
	}
}

func (a *idleActor) Stop() {
	a.lock.Lock()

	// early return if this actor is already stopped
	if a.ctx == nil {
		a.lock.Unlock()
		return
	}

	a.ctx.end()
	a.ctx = nil
	a.lock.Unlock()

	if fn := a.options.Actor.OnStopFunc; fn != nil {
		fn()
	}
}

// Noop returns no-op Actor.
func Noop() Actor {
	return noopActorInstance
}

//nolint:gochecknoglobals
var noopActorInstance = &noopActor{}

type noopActor struct{}

func (a *noopActor) Start() {}
func (a *noopActor) Stop()  {}
