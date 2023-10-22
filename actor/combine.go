package actor

import "sync/atomic"

// Combine returns single Actor which combines all specified actors into one.
// Calling Start or Stop function on this Actor will invoke respective function
// on all Actors provided to this function.
func Combine(actors ...Actor) Actor {
	return &combinedActor{
		stopping: &atomic.Bool{},
		actors:   actors,
	}
}

// CombineAndStopTogether returns single Actor which combines all specified actors into one,
// just like Combine function, and additionally actors combined with this function will end
// together as soon as any of supplied actors ends.
func CombineAndStopTogether(actors ...Actor) Actor {
	combined := &combinedActor{
		stopping: &atomic.Bool{},
	}

	combined.actors = wrapActors(actors, combined.Stop)

	return combined
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
