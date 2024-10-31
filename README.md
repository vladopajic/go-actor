# go-actor

[![test](https://github.com/vladopajic/go-actor/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-actor/actions/workflows/test.yml)
[![lint](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml)
[![coverage](https://raw.githubusercontent.com/vladopajic/go-actor/badges/.badges/main/coverage.svg)](./.testcoverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-actor?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-actor)
[![GoDoc](https://godoc.org/github.com/vladopajic/go-actor?status.svg)](https://godoc.org/github.com/vladopajic/go-actor)
[![Release](https://img.shields.io/github/release/vladopajic/go-actor.svg?style=flat-square)](https://github.com/vladopajic/go-actor/releases/latest)

![goactor-cover](https://user-images.githubusercontent.com/4353513/185381081-2e2a07f3-c13a-4946-a250-b2cbe6588f60.png)

`go-actor` is a lightweight library for writing concurrent programs in Go using the Actor model.


## Motivation

The goal of go-actor is to bring the [actor model](https://en.wikipedia.org/wiki/Actor_model) closer to Go developers by providing a design pattern that facilitates the development of scalable and high-performance concurrent applications.

Without reusable design principles, maintaining a complex codebase can be challenging. Golang's unique concurrency model, based on goroutines and channels, naturally aligns with the actor model, making it an ideal fit for building actor-based programs.

## Advantage

- **Unified Design Principles**: Model the entire codebase using the same principles, where each actor is a fundamental building block.
- **Natural Fit with Go**: Leverage Go's goroutines and channels, which directly translate to actors and mailboxes.
- **Avoid Mutexes**: Design systems without the need for mutexes, reducing the potential for deadlocks and improving performance in complex components.
- **Optimal Scheduling**: Enhance performance by optimizing for Go's goroutine scheduler.
-  **Easy Transition**: Legacy codebases can transition to an actor-based design due to the simple interfaces provided by go-actor, allowing for seamless integration.
- **Zero Overhead**: Ensure optimal performance in highly concurrent environments.


## Abstractions

The core abstraction layer of go-actor consists of three primary interfaces:

- `actor.Actor`: Represents any entity that implements the `Start()` and `Stop()` methods. Actors created using the `actor.New(actor.Worker)` function spawn a dedicated goroutine to execute the supplied `actor.Worker`.
- `actor.Worker`: Encapsulates the executable logic of an actor. This is the primary interface developers need to implement to define an actor's behavior.
- `actor.Mailbox`: An interface for message transport mechanisms between actors, created using the `actor.NewMailbox(...)` function.


## Examples

Explore the [examples](https://github.com/vladopajic/go-actor-examples) repository to see `go-actor` in action. Reviewing these examples is highly recommended, as they will greatly enhance your understanding of the library.


```go
// This example will demonstrate how to create actors for producer-consumer use case.
// Producer will create incremented number on every 1 second interval and
// consumer will print whatever number it receives.
func main() {
	mbx := actor.NewMailbox[int]()

	// Produce and consume workers are created with same mailbox
	// so that produce worker can send messages directly to consume worker
	p := actor.New(&producerWorker{mailbox: mbx})
	c1 := actor.New(&consumerWorker{mailbox: mbx, id: 1})

	// Note: Example creates two consumers for the sake of demonstration
	// since having one or more consumers will produce the same result. 
	// Message on stdout will be written by first consumer that reads from mailbox.
	c2 := actor.New(&consumerWorker{mailbox: mbx, id: 2})

	// Combine all actors to singe actor so we can start and stop all at once
	a := actor.Combine(mbx, p, c1, c2).Build()
	a.Start()
	defer a.Stop()
	
	// Stdout output:
	// consumed 1      (worker 1)
	// consumed 2      (worker 2)
	// consumed 3      (worker 1)
	// consumed 4      (worker 2)
	// ...

	select {}
}

// producerWorker will produce incremented number on 1 second interval
type producerWorker struct {
	mailbox actor.MailboxSender[int]
	num  int
}

func (w *producerWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case <-time.After(time.Second):
		w.num++
		w.mailbox.Send(ctx, w.num)

		return actor.WorkerContinue
	}
}

// consumerWorker will consume numbers received on mailbox
type consumerWorker struct {
	mailbox actor.MailboxReceiver[int]
	id  int
}

func (w *consumerWorker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case num := <-w.mailbox.ReceiveC():
		fmt.Printf("consumed %d \t(worker %d)\n", num, w.id)

		return actor.WorkerContinue
	}
}
```


## Add-ons

While `go-actor` is designed to be a minimal library with lean interfaces, developers can extend its functionality with domain-specific add-ons. Some notable add-ons include:

- [super](https://github.com/vladopajic/go-super-actor): An add-on for unifying the testing of actors and workers.
- [commence](https://github.com/vladopajic/go-actor-commence): An add-on that provides a mechanism for waiting for actor execution to begin.
- [netbox](https://github.com/vladopajic/go-actor-netbox): Mailbox implemented with different network functionality.


## Pro Tips

To enhance code quality in projects that heavily rely on the actor model with `go-actor`, consider adhering to [best practices](./docs/best_practices.md) and reviewing [common hurdles](./docs/common_hurdles.md) for frequently encountered issues.

## Design Decisions

You can find detailed design decisions [here](./docs/design_decisions.md).


## Versioning

The `go-actor` library adopts a versioning scheme structured as `x.y.z`.

Initially, the library will utilize the format `0.y.z` as it undergoes refinement until it attains a level of stability where fundamental interfaces and core principles no longer necessitate significant alterations. Within this semantic, the `y` component signifies a version that is not backward-compatible. It is advisable for developers to review the release notes carefully to gain insight into these modifications. Furthermore, the final component, `z`, denotes releases incorporating changes that are backward-compatible. 


## Contribution

All contributions are useful, whether it is a simple typo, a more complex change, or just pointing out an issue. We welcome any contribution so feel free to open PR or issue. 

Continue reading [here](./docs/contributing.md).


Happy coding 🌞
