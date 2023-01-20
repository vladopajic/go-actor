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
