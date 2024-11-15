# Best Practices for Using `go-actor`

To ensure optimal code quality in projects that leverage the actor model and the `go-actor` library, we recommend following these best practices.

## Avoid the `sync` Package

Projects that rely on the actor model and the `go-actor` library should avoid using synchronization primitives from the `sync` package. Instead, the actor model itself should handle concurrency. To enforce this, repositories using `go-actor` can add a linter that warns when the `sync` package is used. Here’s an example configuration:

```yml
linters-settings:
  forbidigo:
      forbid:
        - 'sync.*(# sync package is forbidden)?'
```

While the general rule is to avoid the `sync` package in actor-based code, there may be specific situations where its use is unavoidable. In such cases, the linter can be temporarily disabled to accommodate these exceptions.

## Use the `goleak` Library

The [goleak](https://github.com/uber-go/goleak) library is very helpful in detecting goroutine leaks during unit tests. This helps identify actors that remain active after a test has concluded, which can indicate issues with actor shutdown (stopping). Detecting and handling such leaks ensures graceful termination of actors and prevents potential performance issues in production.

## Begin `select` Statements in `DoWork` with Context Cancellation

Workers should always check the `Context.Done()` channel and return actor.WorkerEnd to terminate the actor. As a best practice, this case should be listed first in every select statement to ensure context cancellation is handled consistently.

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

## Handle Channel Closure in `DoWork`

Each select statement case in `DoWork` should account for scenarios where a channel might close. When this happens, the worker should gracefully end execution or perform any necessary cleanup actions.


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

## Combine Multiple Actors into a Single Actor

The `actor.Combine(...).Build()` is particularly useful for combining multiple actors into a single actor instance. This can streamline actor management and reduce the complexity of handling multiple actors individually.

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
		Actor: 	actor.Combine(mbx, a1, a2).Build()	// combines all actors into a single actor and initializes the embedded actor of fooActor
	}							// calling fooActor.Start() will start all actors at once
}

func (f *fooActor) OnMessage(ctx context.Context, msg any) error {
	return f.mbx.Send(ctx, msg)
}

```

## Avoid Unnecessary Blocking in `DoWork`

Sometimes, it’s necessary for `DoWork` to return a result, often achieved by sending a “promise-like” response channel along with data via the mailbox. For example:


```go
type request struct {
	// ... some data
	respC chan error
}


func (w *fooWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case req, ok := <-w.mbx.ReceiveC(): // Mailbox[request]
		if !ok {
			return actor.WorkerEnd
		}

		err := handleFoo(req)
		req.respC <- err  // <---- Worker can get blocked here
	}
}
```

For the best performance, avoid introducing unnecessary blocking in the worker. This can happen if `respC` is created without a buffer. In such a case, the worker will block until another goroutine reads from `respC`. This waiting time can cause inefficiencies.

To prevent this, create `respC` with a buffer of size 1:

```go
req := request{ 
	respC: make(chan error, 1) // <---- Buffer size of 1 prevents blocking
}
mbx.Send(ctx, req)
err := <- req.respC
```


---

This documentation is still under development and will be expanded with more best practices.
