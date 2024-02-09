# Common hurdles

## Nothing is happening here

One of the most common hurdles is the case where actors are not started. This is usually experienced when the program is initiated, but nothing is happening in the segment of the program that we are working on and aiming to test. Whenever we experience that nothing is happening, we should double-check that all actors are started.

**Mailbox is not started**

Never forget that `actor.Mailbox` is also an actor, and it needs to be started.

**Embeded Actor interface is overriden**

When embeding `actor.Actor` interface make sure not to override methods of this interface in structre that has embeded it. Otherwise make sure to call embeded actor's Start() and Stop() methods. 

```go
type fooActor struct {
	actor.Actor
	...
}

func NewFooActor() *fooActor {
	return &fooActor{
		Actor: actor.New(&fooWorker{...})
		...
	}
}

func(f *fooActor) Start() { // <--- warning: calling fooActor.Start() will override fooActor.Actor.Start() method.
	...  					//		therfore calling this method will not execute worker that was itended
}							//		to be excuted with fooActor.
							//		if this method is necessery then make sure to call `f.Actor.Start()` manually here.

func(f *fooActor) Stop() { // <---- similar problem as described above.
	...
}

```

## Default case is undesirable

Workers should always block when there isn't anything to work on; therefore, their `select` statements shouldn't have a `default` case. If workers do not block, they will simply waste computation cycles.

Example:
```go
func (w *consumeWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case num := <-w.mailbox.ReceiveC():
		fmt.Printf("consumed %d \t(worker %d)\n", num, w.id)

		return actor.WorkerContinue
    
	default:  // <----------------------- warning: this worker will never block! default case is undesirable!
		return actor.WorkerContinue
	}
}
```

---

Your contribution is valuable; if you have encountered any challenges, please share your experiences.