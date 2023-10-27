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
		actors:       b.actors,
		onStopFunc:   b.options.Combined.OnStopFunc,
		stopTogether: b.options.Combined.StopTogether,
		stopping:     &atomic.Bool{},
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
}

func (a *combinedActor) onActorStopped() {
	a.runningLock.Lock()

	runningCount := a.runningCount.Add(-1)
	wasRunning := a.running
	a.running = runningCount != 0

	a.runningLock.Unlock()

	// Last actor to end should call onStopFunc
	if runningCount == 0 && wasRunning && a.onStopFunc != nil {
		a.onStopFunc()
	}

	// First actor to stop should stop other actors
	if a.stopTogether && a.stopping.CompareAndSwap(false, true) {
		// Run stop in goroutine because wrapped actor
		// should not wait for other actors to stop.
		go a.Stop()
	}
}

func (a *combinedActor) Stop() {
	a.runningLock.Lock()

	if !a.running {
		a.runningLock.Unlock()
		return
	}

	a.running = false
	a.runningLock.Unlock()

	for _, actor := range a.actors {
		actor.Stop()
	}
}

func (a *combinedActor) Start() {
	a.runningLock.Lock()

	if a.running {
		a.runningLock.Unlock()
		return
	}

	a.stopping.Store(false)
	a.running = true

	a.runningLock.Unlock()

	for _, actor := range a.actors {
		a.runningCount.Add(1)
		actor.Start()
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
