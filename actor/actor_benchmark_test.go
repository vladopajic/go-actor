package actor_test

import (
	"testing"

	. "github.com/vladopajic/go-actor/actor"
)

func BenchmarkActorProducerToConsumer(b *testing.B) {
	produceNext := int(0)
	doneC := make(chan struct{})

	// intentionally using option as chan with large capacity as it has
	// the best performance.
	mbx := NewMailbox[any](OptAsChan(), OptCapacity(largeCap))

	producer := New(NewWorker(func(c Context) WorkerStatus {
		select {
		case <-c.Done():
			return WorkerEnd

		default:
			produceNext++
			mbx.Send(c, produceNext) //nolint:errcheck // error should never happen

			if b.N == produceNext {
				return WorkerEnd
			}

			return WorkerContinue
		}
	}))

	consumer := New(NewWorker(func(c Context) WorkerStatus {
		select {
		case <-c.Done():
			return WorkerEnd

		case rcv := <-mbx.ReceiveC():
			if b.N == rcv {
				close(doneC)
			}

			return WorkerContinue
		}
	}))

	a := Combine(mbx, producer, consumer).Build()
	a.Start()
	defer a.Stop()

	<-doneC
}
