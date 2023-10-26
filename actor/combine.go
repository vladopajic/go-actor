package actor

import (
	"sync"
	"sync/atomic"
)

// Combine returns builder which is used to create single Actor that combines all
// specified actors into one.
//
// Calling Start or Stop function on combined Actor will invoke respective
// function on all underlying Actors.
func Combine(actors ...Actor) *CombineBuilder {
	return &CombineBuilder{
		actors: actors,
	}
}

type CombineBuilder struct {
	actors  []Actor
	options options
}

// Build returns combined Actor.
func (b *CombineBuilder) Build() Actor {
	a := &combinedActor{
		stopping:     &atomic.Bool{},
		actors:       b.actors,
		onStopFunc:   b.options.Combined.OnStopFunc,
		stopTogether: b.options.Combined.StopTogether,
	}

	a.actors = wrapActors(a.actors, a.onActorStopped)

	return a
}

// WithOptions adds options for combined actor.
func (b *CombineBuilder) WithOptions(opt ...CombinedOption) *CombineBuilder {
	b.options = newOptions(opt)
	return b
}

type combinedActor struct {
	actors       []Actor
	onStopFunc   func()
	stopTogether bool

	runningCount atomic.Int32
	running      bool
	runningLock  sync.Mutex
	stopping     *atomic.Bool
	onStopOnce   *sync.Once
}

func (a *combinedActor) onActorStopped() {
	a.runningLock.Lock()

	runningCount := a.runningCount.Add(-1)
	once, onStopFunc := a.onStopOnce, a.onStopFunc
	wasRunning := a.running
	a.running = runningCount != 0

	a.runningLock.Unlock()

	// Last actor to end should call onStopFunc
	if runningCount == 0 && wasRunning && onStopFunc != nil {
		once.Do(onStopFunc)
	}

	// First actor to stop should stop other actors
	if a.stopTogether && a.stopping.CompareAndSwap(false, true) {
		// run stop in goroutine because wrapped actor
		// should not block until other actors to stop
		go a.Stop()
	}
}

func (a *combinedActor) Stop() {
	a.runningLock.Lock()

	if !a.running {
		a.runningLock.Unlock()
		return
	}

	onStopOnce, onStopFunc := a.onStopOnce, a.onStopFunc
	a.running = false

	a.runningLock.Unlock()

	for _, a := range a.actors {
		a.Stop()
	}

	if onStopFunc != nil {
		onStopOnce.Do(onStopFunc)
	}
}

func (a *combinedActor) Start() {
	a.runningLock.Lock()

	if a.running {
		a.runningLock.Unlock()
		return
	}

	a.stopping.Store(false)
	a.onStopOnce = &sync.Once{}
	a.running = true

	a.runningLock.Unlock()

	for _, aa := range a.actors {
		a.runningCount.Add(1)
		aa.Start()
	}
}

func wrapActors(
	actors []Actor,
	onStopFunc func(),
) []Actor {
	wrapActorStruct := func(a *actor) *actor {
		prevOnStopFunc := a.options.Actor.OnStopFunc

		a.options.Actor.OnStopFunc = func() {
			if prevOnStopFunc != nil {
				prevOnStopFunc()
			}

			onStopFunc()
		}

		return a
	}

	wrapCombinedActorStruct := func(a *combinedActor) *combinedActor {
		prevOnStopFunc := a.onStopFunc

		a.onStopFunc = func() {
			if prevOnStopFunc != nil {
				prevOnStopFunc()
			}

			onStopFunc()
		}

		return a
	}

	wrapActorInterface := func(a Actor) Actor {
		return &wrappedActor{
			actor:      a,
			onStopFunc: onStopFunc,
		}
	}

	for i, a := range actors {
		switch v := a.(type) {
		case *actor:
			actors[i] = wrapActorStruct(v)
		case *combinedActor:
			actors[i] = wrapCombinedActorStruct(v)
		default:
			actors[i] = wrapActorInterface(v)
		}
	}

	return actors
}

type wrappedActor struct {
	actor      Actor
	onStopFunc func()
}

func (a *wrappedActor) Start() {
	a.actor.Start()
}

func (a *wrappedActor) Stop() {
	a.actor.Stop()
	a.onStopFunc()
}
