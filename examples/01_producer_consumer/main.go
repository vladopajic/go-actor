package main

import (
	"github.com/vladopajic/go-actor/actor"
)

// This program will demonstrate how to create actors for producer-consumer use case, where
// producer will create incremented number on every 1 second interval and
// consumer will print whaterver number it receives
func main() {
	mailbox := actor.NewMailbox[int]()

	// Producer and consumer workers are created with same mailbox
	// so that producer worker can send messages directly to consumer worker
	pw := &producerWorker{outC: mailbox.SendC()}
	cw1 := &consumerWorker{inC: mailbox.ReceiveC(), id: 1}
	cw2 := &consumerWorker{inC: mailbox.ReceiveC(), id: 2}

	// Create actors using these workers and combine them to singe Actor
	a := actor.Combine(
		mailbox,
		actor.New(pw),

		// Note: We don't need two consumer actors, but we create them anyway
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
