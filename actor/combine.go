package actor

import "sync/atomic"

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
	combined := &combinedActor{
		stopping: &atomic.Bool{},
		actors:   b.actors,
	}

	if b.options.Combined.StopTogether {
		combined.actors = wrapActors(combined.actors, combined.Stop)
	}

	return combined
}

// WithOptions adds options for combined actor.
func (b *CombineBuilder) WithOptions(opt ...CombinedOption) *CombineBuilder {
	b.options = newOptions(opt)
	return b
}

type combinedActor struct {
	stopping *atomic.Bool
	actors   []Actor
}

func (a *combinedActor) Stop() {
	if !a.stopping.CompareAndSwap(false, true) {
		return
	}

	for _, a := range a.actors {
		a.Stop()
	}
}

func (a *combinedActor) Start() {
	a.stopping.Store(false)

	for _, a := range a.actors {
		a.Start()
	}
}

func wrapActors(actors []Actor, onStop func()) []Actor {
	wrapActorStruct := func(a *actor) *actor {
		prevOnStopFunc := a.options.Actor.OnStopFunc
		a.options.Actor.OnStopFunc = func() {
			if prevOnStopFunc != nil {
				prevOnStopFunc()
			}

			go onStop()
		}

		return a
	}

	wrapActorInterface := func(a Actor) Actor {
		return &wrappedActor{
			actor:    a,
			onStopFn: onStop,
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
	actor    Actor
	onStopFn func()
}

func (a *wrappedActor) Start() {
	a.actor.Start()
}

func (a *wrappedActor) Stop() {
	a.actor.Stop()
	go a.onStopFn()
}
