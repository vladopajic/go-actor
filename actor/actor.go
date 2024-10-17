package actor

import (
	"sync"
)

// Actor represents a computational entity that manages the execution of a Worker
// in a dedicated goroutine. Each Actor is responsible for handling the lifecycle
// of a Worker, including starting and stopping its execution.
type Actor interface {
	// Start launches a new goroutine to begin the execution of the associated Worker.
	//
	// The Worker will continue running until one of the following conditions occurs:
	// 1. The Stop() method is called, signaling the Worker to terminate.
	// 2. The Worker completes its task and returns a WorkerEnd status, indicating
	//    there is no more work to process.
	//
	// This method is non-blocking and returns immediately after spawning the goroutine.
	Start()

	// Stop gracefully signals the Worker to stop execution.
	//
	// This method will block the calling goroutine until the Worker has fully terminated
	// its execution. It ensures that any ongoing operations in the Worker are completed
	// or properly interrupted before the Worker stops.
	//
	// Stop can be called at any time after Start has been invoked. If Stop is called
	// on an Actor that has not yet been started, the call has no effect.
	Stop()
}

// WorkerStatus represents the status code returned by a Worker's DoWork function,
// which indicates whether the Actor should continue executing the Worker or terminate it.
type WorkerStatus int8

const (
	WorkerContinue WorkerStatus = 1
	WorkerEnd      WorkerStatus = 2
)

// Worker represents an entity that encapsulates the executable logic within an Actor.
//
// A Worker is responsible for processing incoming messages sent via Mailboxes
// and performing actions by sending messages or spawning new Actors. It defines
// the core unit of work for an Actor, determining how tasks are processed and managed.
type Worker interface {
	// DoWork defines a single unit of executable work for the Worker.
	//
	// The method takes a Context parameter, which allows the Worker to listen for
	// stop signals initiated by the Actor. This context helps manage graceful shutdowns
	// and interruptive workloads.
	//
	// The method returns a WorkerStatus value, which indicates whether the Actor
	// should continue running this Worker or stop. If the Worker returns a status
	// signaling completion `WorkerEnd`, the Actor will terminate the Worker; otherwise,
	// the Actor will continue executing subsequent work.
	DoWork(ctx Context) WorkerStatus
}

// WorkerFunc defines the function signature for a Worker's DoWork method.
type WorkerFunc = func(ctx Context) WorkerStatus

// StartableWorker defines an optional interface that a Worker can implement
// to perform initialization tasks before its main work begins.
//
// Implementing this interface allows a Worker to execute setup or initialization
// logic exactly once, before the first call to DoWork.
type StartableWorker interface {
	// OnStart is invoked immediately before the first call to DoWork().
	//
	// This method is intended for any setup or initialization that the Worker
	// needs before starting its main execution. It is called only once during
	// the Worker’s lifecycle.
	//
	// The same Context provided to DoWork is passed to OnStart, allowing the Worker
	// to handle early termination signals in case the Actor is stopped during
	// initialization.
	OnStart(ctx Context)
}

// StoppableWorker defines an optional interface that a Worker can implement
// to perform cleanup or resource release after its execution completes.
//
// By implementing this interface, a Worker can define custom logic to release resources
// or perform any necessary finalization tasks once it has finished its work.
type StoppableWorker interface {
	// OnStop is invoked after the final call to DoWork() returns.
	//
	// This method is intended for cleaning up or releasing any resources held by
	// the Worker during its lifecycle. It is called after the Worker has completed
	// its execution and the Actor has no further tasks for it.
	//
	// A Context is not provided because the Worker’s context has already been
	// canceled or completed by the time OnStop is called.
	OnStop()
}

// NewWorker creates and returns a basic implementation of the Worker interface.
//
// This function takes a WorkerFunc as an argument, which defines the core logic
// of the Worker. The returned Worker will delegate its DoWork method to the
// provided WorkerFunc, allowing the caller to specify the work logic.
func NewWorker(fn WorkerFunc) Worker {
	return &worker{fn}
}

type worker struct {
	fn WorkerFunc
}

func (w *worker) DoWork(ctx Context) WorkerStatus {
	return w.fn(ctx)
}

// New creates and returns a new Actor with the specified Worker and
// optional configuration.
func New(w Worker, opt ...Option) Actor {
	return &actor{
		worker:  w,
		options: newOptions(opt).Actor,
	}
}

type actor struct {
	worker            Worker
	options           optionsActor
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

	// End context, then wait for worker to finish
	a.ctx.end()
	workEndedSigC := a.workEndedSigC

	a.workerRunningLock.Unlock()

	<-workEndedSigC
}

func (a *actor) Start() {
	a.workerRunningLock.Lock()
	defer a.workerRunningLock.Unlock()

	if a.workerRunning {
		return
	}

	a.workEndedSigC = make(chan struct{})
	a.ctx = newContext()
	a.workerRunning = true

	go a.doWork()
}

// doWork executes Worker of this Actor until
// Actor or Worker has signaled to stop.
func (a *actor) doWork() {
	a.onStart()

	for status := WorkerContinue; status == WorkerContinue; {
		status = a.worker.DoWork(a.ctx)
	}

	a.ctx.end()

	a.onStop()

	{ // Worker has finished
		a.workerRunningLock.Lock()
		a.workerRunning = false
		close(a.workEndedSigC)
		a.workerRunningLock.Unlock()
	}
}

func (a *actor) onStart() {
	if w, ok := a.worker.(StartableWorker); ok {
		w.OnStart(a.ctx)
	}

	if fn := a.options.OnStartFunc; fn != nil {
		fn(a.ctx)
	}
}

func (a *actor) onStop() {
	if w, ok := a.worker.(StoppableWorker); ok {
		w.OnStop()
	}

	if fn := a.options.OnStopFunc; fn != nil {
		fn()
	}
}

// Idle creates and returns a new Actor that does not have an associated Worker.
//
// This function is useful in scenarios where an Actor is needed solely for its
// OnStart and OnStop functionalities, without executing any actual work.
// The returned Actor can manage initialization and cleanup logic, allowing for
// customizable behavior based on the provided options.
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
//
// This function provides an Actor that does not perform any actions or operations.
// It is particularly useful in scenarios where an Actor interface needs to be
// implemented, but you want to avoid the overhead of defining custom logic.
// By embedding this no-op Actor in a structure, you can easily satisfy the
// Actor interface without additional implementation details.
func Noop() Actor {
	return noopActorInstance
}

//nolint:gochecknoglobals // this is singleton values
var noopActorInstance = &noopActor{}

type noopActor struct{}

func (a *noopActor) Start() {}
func (a *noopActor) Stop()  {}
