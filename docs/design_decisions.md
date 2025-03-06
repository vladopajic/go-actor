# Design decisions

This page will document reasoning behind some design decisions made in this library.

## Light library

`go-actor` is intended to be a tiny library with lean interfaces and basic mechanisms providing core building blocks needed to build concurrent programs using the actor model. Having lean interfaces gives this library advantage to be easily integrated in any codebase. Bloating `go-actor` with additional functionalities may result in framework-like library which could turn down developers that need just basic building blocks.
 
Having lean interfaces does not prevent this library from to evolve. Furthermore, developers may build on top of it and extend its functionality with user or domain specific addon abstractions. These addon libraries would enrich _actorverse_  and give developers opportunity to stack up different libraries that fit their development goals/style the most.


## Errorless interface

`actor.Actor` interface defines two methods `Start()` and `Stop()`, both of which do not return any error to the caller.

```go
type Actor interface {
	Start()
	Stop()
}
```

It's not hard to find examples where developers have defined a similar interface, but with methods which could return error to the caller. As an example we can consider following interface:

```go
// example of interface similar to Actor's, but whose methods return error
type Service interface {
    // Start will establish database connection and start new goroutine to process
    // all incoming requests.
    Start() error

    // Stop releases all resources occupied by this service. 
    Stop() error
    
    ...
}
```

Interface whose methods do not return errors (errorless interface) could be easily converted to interface whose methods return error (error-returning interface), because returned value can be `nil` in later scenarios. Having this in mind it seems that `Actor` interface would be compatible with more interfaces if it had error-returning interface. This certainly would make sense, but let's also consider following analogy.

If we make analogy that `Actor` is just a goroutine that is started by calling `Actor`'s `Start()` method then why would `Actor` return error on start if starting goroutine is errorless operation? Now, this reasoning makes completely different, but very logical statement from `Actor`'s point of view.

In a system entirely made of actors it makes sense for `Service` interface to implement `Actor` interface. But how would that be possible if `Service` is error-returning interface? In the case of `Service` it appears to be that some piece of code which requires error handling was added at the place where service's goroutine is about to be started. This code could be dealt with in two ways:
- It could be moved before service is created. For example database connection would need to be created at the top of the bootstrap process of a program and passed on where needed.
- More robust approach would be to move this code inside goroutine which would repeat this operation (establishing database connection) until it succeeds.
 
Based on previous statements we can see that errorless interfaces are more robust and more natural in the world of actors.


## Consecutive Start() or Stop() calls have no effect

Calling `Start()` multiple times on the same actor has no effect after the first call. If an actor is already started, additional calls to `Start()` do nothing.

```go
a := actor.New(...)
a.Start() // this call starts the actor
a.Start() // this call has no effect (actor is already started)

```

Since `Start()` does not return an error, it cannot indicate repeated calls. Panicking in this scenario would be excessive, as the intent of the second call (to start the actor) is already fulfilled.

Similarly, consecutive calls to `Stop()` behave the same way. If an actor is already stopped, additional calls to `Stop()` have no effect.


## Stop() blocks the caller
The `Stop()` function blocks the caller until the goroutine created by `Start()` has fully stopped. For some actors, depending on the logic defined in their worker functions, `Stop()` may take an extended period to complete.

This blocking behavior ensures that all actor-related resources are properly released, allowing programs to gracefully terminate.

```go
a := actor.New(...)
a.Start()
...

a.Stop()
// at this point all resources occupied by actor have been released
```

To prevent potential delays, `Stop()` can be called in a separate goroutine and use a timeout to limit the blocking behavior:

```go
stoppedC := make(chan struct{})
go func() {
    a.Stop()
    close(stoppedC)
}()

select {
case <-stoppedC:
case <-time.After(timeout):
}
```


## Actor restartability
Actors created using `New(...)`, `Idle(...)`, and `Noop()` can be restarted, with the only limitations being the logic defined within their `OnStart()`, `OnStop()`, and `DoWork()` functions.

The restart behavior depends on how `OnStart()`, `OnStop()`, and `DoWork()` functions manage the Worker’s state and lifecycle. If these functions properly handle transitions, an actor can stop and start again without issues. However, if these functions introduce constraints, such as preventing reinitialization or maintaining state inconsistencies, restarting may be affected.

Example: The following sunshine actor can be restarted multiple times:
```go
a := actor.New(NewWorker(func(c Context) WorkerStatus {
    select {
    case <-c.Done():
        return WorkerEnd
    default:
        fmt.Print(`🌞`)
        return WorkerContinue
    }
}))
a.Start()
a.Stop()

// Restart the actor after stopping it
a.Start() 
a.Stop()

// Restart again
a.Start() 
a.Stop()
```

### NewMailbox(...) is not restartable
Unlike other actors, an actor created using `NewMailbox(...)` cannot be restarted. This is because its `ReceiveC()` channel is closed when the mailbox is stopped, signaling that no further messages will be received. Restarting the mailbox would require creating a new `ReceiveC()`, which is not supported (more on this below).

## Mailbox.ReceiveC() channel is never changed
When a mailbox is created using `NewMailbox(...)`, it initializes a `ReceiveC()` channel that remains unchanged throughout the mailbox's lifecycle. This design provides the following advantages:
- A reference to the `ReceiveC()` channel can be passed even before the mailbox is started.
- Since `ReceiveC()` never changes, users can safely pass it as `<-chan T` without worrying about potential issues caused by channel reassignment. If the channel were mutable, users would have to access it exclusively through the mailbox `mbx.ReceiveC()` to avoid inconsistencies.