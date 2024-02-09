# Best practices

To enhance the code quality of projects that heavily rely on the actor model and utilize the `go-actor` library, it's recommended to adhere following best practices.

## Forget about `sync` package

Projects that fully relay on actor model and `go-actor` library shouldn't use any synchronization primitives from `sync` package. Therefore repositories based on `go-actor` could add linter that will warn them if `sync` package is used, eg:

```yml
linters-settings:
  forbidigo:
      forbid:
        - 'sync.*(# sync package is forbidden)?'
```

While the general rule is to avoid `sync` package usage in actor-based code, there might be specific situations where its use becomes necessary. In such cases, the linter can be temporarily disabled to accommodate these exceptional needs.

## Use `goleak` library

[goleak](https://github.com/uber-go/goleak) is vary helpful library that detects goroutine leaks in unit tests. This can be helpful because it can identify actors that are still running after unit test have ended. Gracefully ending actors is important to test because it can identify various problems in implementation.

## Start select statements in `DoWork` with context

Workers should always respond to `Context.Done()` channel and return `actor.WorkerEnd` status in order to end it's actor. As a rule of thumb it's advised to always list this case first since it should be included in every `select` statement.

```go
func (w *fooWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done(): // <----------------------- handle ctx.Done() first
		return actor.WorkerEnd

	case msg := <-w.mbx.ReceiveC():
		handleFoo(msg)
	}
}
```

## Check channel closed indicator in `DoWork`

Every case statement in `DoWork` should handle case when channel is closed. In these cases worker should end execution; or it can perform any other logic that is necessery.

```go
func (w *fooWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case msg, ok := <-w.mbx.ReceiveC():
		if !ok { // <----------------------- handle channel close (mailbox stop) case
			return actor.WorkerEnd
		}

		handleFoo(msg)
	}
}
```

## Combine multiple actors to singe actor

`actor.Combine(...)` is vary handy to combine multiple actors to single actor.

```go
type fooActor struct {
	actor.Actor
	mbx actor.Mailbox[any]
	...
}

func NewFooActor() *fooActor {
	mbx := actor.NewMailbox[any]()
	
	a1 := actor.New(&fooWorker{mbx: mbx})
	a2 := actor.New(&fooWorker{mbx: mbx})

	return &fooActor{
		mbx: 	mbx,
		Actor: 	actor.Combine(mbx, a1, a2).Build()	// combine all actors to single actor and initialize embeded actor of fooActor struct.
	}							// when calling fooActor.Start() it will start all actors at once.
}

func (f *fooActor) OnMessage(ctx context.Context, msg any) error {
	return f.mbx.Send(ctx, msg)
}

```

---

This page is not yet complete.
