package actor_test

import (
	"testing"

	. "github.com/vladopajic/go-actor/actor"
)

func BenchmarkActorProducerToConsumerMailbox(b *testing.B) {
	produceNext := int(0)
	doneC := make(chan struct{})

	// intentionally using option as chan with large capacity as it has
	// the best performance.
	mbx := NewMailbox[any](OptAsChan(), OptCapacity(largeCap))

	producer := New(NewWorker(func(ctx Context) WorkerStatus {
		produceNext++
		mbx.Send(ctx, produceNext) //nolint:errcheck // error should never happen

		if b.N == produceNext {
			return WorkerEnd
		}

		return WorkerContinue
	}))

	consumer := New(NewWorker(func(Context) WorkerStatus {
		rcv := <-mbx.ReceiveC()
		if b.N == rcv {
			close(doneC)
			return WorkerEnd
		}

		return WorkerContinue
	}))

	a := Combine(mbx, producer, consumer).Build()
	a.Start()
	defer a.Stop()

	<-doneC
}

func BenchmarkActorProducerToConsumerChannel(b *testing.B) {
	produceNext := int(0)
	doneC := make(chan struct{})

	c := make(chan any, largeCap)

	producer := New(NewWorker(func(Context) WorkerStatus {
		produceNext++
		c <- produceNext

		if b.N == produceNext {
			return WorkerEnd
		}

		return WorkerContinue
	}))

	consumer := New(NewWorker(func(Context) WorkerStatus {
		rcv := <-c
		if b.N == rcv {
			close(doneC)
			return WorkerEnd
		}

		return WorkerContinue
	}))

	a := Combine(producer, consumer).Build()
	a.Start()
	defer a.Stop()

	<-doneC
}

// This benchmark implements logic of ActorProducerToConsumer benchmark
// using two goroutines and a channel.
func BenchmarkNativeProducerToConsumer(b *testing.B) {
	b.Helper()

	produceNext := int(0)
	doneC := make(chan struct{})
	c := make(chan any, largeCap)

	go func() { // producer
		for {
			produceNext++
			c <- produceNext

			if b.N == produceNext {
				return
			}
		}
	}()

	go func() { // consumer
		for {
			rcv := <-c
			if b.N == rcv {
				close(doneC)
				return
			}
		}
	}()

	<-doneC
}
