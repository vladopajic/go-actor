# go-actor

[![lint](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml)
[![test](https://github.com/vladopajic/go-actor/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/vladopajic/go-actor/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-actor?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-actor)
[![codecov](https://codecov.io/gh/vladopajic/go-actor/branch/main/graph/badge.svg?token=WYCKb1MLgl)](https://codecov.io/gh/vladopajic/go-actor)
[![GoDoc](https://godoc.org/github.com/vladopajic/go-actor?status.svg)](https://godoc.org/github.com/vladopajic/go-actor)
[![Release](https://img.shields.io/github/release/vladopajic/go-actor.svg?style=flat-square)](https://github.com/vladopajic/go-actor/releases/latest)

![goactor-cover](https://user-images.githubusercontent.com/4353513/185381081-2e2a07f3-c13a-4946-a250-b2cbe6588f60.png)

`go-actor` is tiny library for writing concurrent programs in Go using actor model.

## Motivation

Intention of go-actor is to bring [actor model](https://en.wikipedia.org/wiki/Actor_model) closer to Go developers and to provide design pattern needed to build scalable and high performing concurrent programs.

Without re-usable design principles codebase of complex system can become hard to maintain. Codebase written using Golang can highly benefit from design principles based on actor model as gorutines and channels naturally translate to actors and mailboxes.

## `go-actor` benefits

 - Entire codebase can be modeld with same design principles where _actor_ is the universal primitive. 
Example: in microservice architectured system each service is actor which reacts and sends messages to other services (actors). Services themself could be made of multiple components (actors) which interact with other components by responding and sending messages. 
- Golang’s gorutines and channels naturally translate to actor and mailboxes.
- System can be designed without use of mutex. This can give performance gains as overlocking is not rare in complex components. 
- Optimal for Golang's gorutine scheduler
- Legacy codebase can transition to actor based design because components modeld with `go-actor` have simple interface which could be integrated in anywhere.


## `go-actor` abstractions

`go-actor`'s base abstraction layer only have three interfaces:


- `actor.Actor` is anything that implements `Start()` and `Stop()` methods. Actors created using `actor.New(actor.Worker)` function will create prefered actor implementation which will on start spawn dedicated goroutine to perform work of supplied `actor.Worker`.
  - `actor.Worker` encapsulates actor's executable logic. This is only interface which developers need to wirte in order to describe behavior of actor.
- `actor.Mailbox` is interface for message transport mechanism between actors. Mailboxes are created using `actor.NewMailbox(...)` function.


## Examples

Dive into [examples](https://github.com/vladopajic/go-actor-examples) to see `go-actor` in action.

```go
// This program will demonstrate how to create actors for producer-consumer use case, where
// producer will create incremented number on every 1 second interval and
// consumer will print whaterver number it receives
func main() {
	mailbox := actor.NewMailbox[int]()

	// Produce and consume workers are created with same mailbox
	// so that produce worker can send messages directly to consume worker
	pw := &produceWorker{outC: mailbox.SendC()}
	cw1 := &consumeWorker{inC: mailbox.ReceiveC(), id: 1}
	cw2 := &consumeWorker{inC: mailbox.ReceiveC(), id: 2}

	// Create actors using these workers and combine them to singe Actor
	a := actor.Combine(
		mailbox,
		actor.New(pw),

		// Note: We don't need two consume actors, but we create them anyway
		// for the sake of demonstration since having one or more consumers
		// will produce the same result. Message on stdout will be written by
		// first consumer that reads from mailbox.
		actor.New(cw1),
		actor.New(cw2),
	)

	// Finally we start all actors at once
	a.Start()
	defer a.Stop()

	select {}
}

// produceWorker will produce incremented number on 1 second interval
type produceWorker struct {
	outC chan<- int
	num  int
}

func (w *produceWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-time.After(time.Second):
		w.num++
		w.outC <- w.num

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}

// consumeWorker will consume numbers received on inC channel
type consumeWorker struct {
	inC <-chan int
	id  int
}

func (w *consumeWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case num := <-w.inC:
		fmt.Printf("consumed %d \t(worker %d)\n", num, w.id)

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}
```

## Contribution

All contributions are useful, whether it is a simple typo, a more complex change, or just pointing out an issue. We welcome any contribution so feel free to open PR or issue. 
