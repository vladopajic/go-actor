package actor

import "sync"

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

// Context is provided so Worker can listen and respond on stop signal sent from Actor.
type Context interface {
	// Done returns channel which will receive signal that Worker should end execution.
	Done() <-chan struct{}
}

type contextImpl struct {
	endWorkC chan struct{}
}

func newContext() *contextImpl {
	return &contextImpl{
		endWorkC: make(chan struct{}, 1),
	}
}

func (c *contextImpl) Done() <-chan struct{} {
	return c.endWorkC
}

func (c *contextImpl) signalEnd() {
	c.endWorkC <- struct{}{}
}

// WorkerStatus is returned by Worker's DoWork function indicating if Actor should continue
// executing Worker.
type WorkerStatus int8

const (
	WorkerContinue WorkerStatus = 1
	WorkerEnd      WorkerStatus = 2
)

// Worker is entity which encapsulates Actor's single unit/iteration of it's executable logic.
//
// Worker's implementation should listen on messages sent via go channels and preform actions by
// sending new messages or creating new actors.
type Worker interface {
	// DoWork function is Worker's single executable unit.
	//
	// Context is provided so Worker can listen and respond on stop signal sent from Actor.
	//
	// WorkerStatus is returned indicating if Actor should continue executing this worker.
	// Actor will check this status and stop execution if Worker has no more work, or otherwise
	// proceed execution.
	DoWork(c Context) WorkerStatus
}

// WorkerFunc is signature of Worker's DoWork function.
type WorkerFunc = func(c Context) WorkerStatus

// NewWorker returns basic Worker implementation which delegates
// DoWork to supplied WorkerFunc.
func NewWorker(fn WorkerFunc) Worker {
	return &workerImpl{fn}
}

type workerImpl struct {
	fn WorkerFunc
}

func (w *workerImpl) DoWork(c Context) WorkerStatus {
	return w.fn(c)
}

// New returns new Actor with specified Worker and Options.
func New(w Worker, opt ...Option) Actor {
	return &actorImpl{
		worker:  w,
		options: newOptions(opt),
		ctx:     newContext(),
	}
}

type actorImpl struct {
	worker            Worker
	options           options
	ctx               *contextImpl
	workEndedSigC     chan struct{}
	workerRunning     bool
	workerRunningLock sync.Mutex
}

func (a *actorImpl) Stop() {
	a.workerRunningLock.Lock()
	if !a.workerRunning {
		defer a.workerRunningLock.Unlock()
		return
	}

	a.workEndedSigC = make(chan struct{})
	a.workerRunningLock.Unlock()

	a.ctx.signalEnd()

	<-a.workEndedSigC
}

func (a *actorImpl) Start() {
	a.workerRunningLock.Lock()
	if a.workerRunning {
		a.workerRunningLock.Unlock()
		return
	}

	a.workerRunning = true
	a.workerRunningLock.Unlock()

	go a.doWork()
}

// doWork executes Worker of this Actor until
// Actor or Worker has signaled to stop.
func (a *actorImpl) doWork() {
	executeFunc(a.options.OnStartFunc)
	defer executeFunc(a.options.OnStopFunc)

	for wStatus := WorkerContinue; wStatus == WorkerContinue; {
		wStatus = a.worker.DoWork(a.ctx)
	}

	a.workerRunningLock.Lock()

	a.workerRunning = false
	if c := a.workEndedSigC; c != nil {
		c <- struct{}{}
	}

	a.workerRunningLock.Unlock()
}

func executeFunc(fn func()) {
	if fn != nil {
		fn()
	}
}

// StartAll starts all specified actors.
func StartAll(actors ...Actor) {
	for _, a := range actors {
		a.Start()
	}
}

// StopAll starts all specified actors.
func StopAll(actors ...Actor) {
	for _, a := range actors {
		a.Stop()
	}
}

// Combine returns single Actor which combines all specified actors into one.
// Calling Start or Stop function on this Actor will invoke respective function
// on all Actors provided to this function.
func Combine(actors ...Actor) Actor {
	return &combinedActorImpl{actors}
}

type combinedActorImpl struct {
	actors []Actor
}

func (a *combinedActorImpl) Stop() {
	StopAll(a.actors...)
}

func (a *combinedActorImpl) Start() {
	StartAll(a.actors...)
}
