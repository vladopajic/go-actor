# go-actor

[![lint](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml/badge.svg)](https://github.com/vladopajic/go-actor/actions/workflows/lint.yml)
[![test](https://github.com/vladopajic/go-actor/actions/workflows/test.yml/badge.svg)](https://github.com/vladopajic/go-actor/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladopajic/go-actor?cache=v1)](https://goreportcard.com/report/github.com/vladopajic/go-actor)
[![GoDoc](https://godoc.org/github.com/vladopajic/go-actor?status.svg)](https://godoc.org/github.com/vladopajic/go-actor)
[![Release](https://img.shields.io/github/release/vladopajic/go-actor.svg?style=flat-square)](https://github.com/vladopajic/go-actor/releases/latest)


`go-actor` is tiny library for writing concurrent programs in Go using actor model.

## Motivation

This library was published with intention to bring [actor model](https://en.wikipedia.org/wiki/Actor_model) closer to Go developers and to provide easy to understand abstractions needed to build concurrent programs.

## Examples

Dive into [examples](./examples/) to see `go-actor` in action.

```go
// This program will create producer and consumer actors, where
// producer will create incremented number on every 1 second interval and
// consumer will print whaterver number it receives
func main() {
	numC := make(chan int)

	// Producer and consumer workers are created with same channel
	// so that producer worker can write directly to consumer worker
	pw := &producerWorker{outC: numC}
	cw1 := &consumerWorker{inC: numC, id: 1}
	cw2 := &consumerWorker{inC: numC, id: 2}

	// Create actors using these workers
	a := actor.Combine(
		actor.New(pw),

		// Note: We don't need two consumer actors, but we create them anyway
		// for the sake of demonstration since having one or more consumers
		// will produce the same result. Message on stdout will be written by
		// first consumer that reads from numC channel.
		actor.New(cw1),
		actor.New(cw2),
	)

	// Finally we start all actors at once
	a.Start()
	defer a.Stop()

	select {}
}

// producerWorker will produce incremented number on 1 second interval
type producerWorker struct {
	outC chan<- int
	num  int
}

func (w *producerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-time.After(time.Second):
		w.num++
		w.outC <- w.num

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}

// consumerWorker will consume numbers received on inC channel
type consumerWorker struct {
	inC <-chan int
}

func (w *consumerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case num := <-w.inC:
		fmt.Printf("consumed %d\n", num)

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}
```

## Contribution

All contributions are useful, whether it is a simple typo, a more complex change, or just pointing out an issue. We welcome any contribution so feel free to open PR or issue. 
