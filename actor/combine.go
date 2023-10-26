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

	runningCount int
	stopping     *atomic.Bool
	onEnd        *sync.Once
}

func (a *combinedActor) onActorStopped() {
	a.runningCount--

	if a.runningCount == 0 {
		// Last actor to end should call onStopFunc
		if once, fn := a.onEnd, a.onStopFunc; once != nil && fn != nil {
			once.Do(fn)
		}
	}

	// First actor to stop should stop other actors
	if a.stopTogether && a.stopping.CompareAndSwap(false, true) {
		// run stop in goroutine because wrapped actor
		// should not block until other actors to stop
		go a.Stop()
	}
}

func (a *combinedActor) Stop() {
	for _, a := range a.actors {
		a.Stop()
	}

	if once, fn := a.onEnd, a.onStopFunc; once != nil && fn != nil {
		once.Do(fn)
	}
}

func (a *combinedActor) Start() {
	a.stopping.Store(false)
	a.onEnd = &sync.Once{}

	for _, aa := range a.actors {
		a.runningCount++
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
