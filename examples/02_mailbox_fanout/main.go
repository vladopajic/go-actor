package main

import (
	"github.com/vladopajic/go-actor/actor"
)

// This program will demonstrate how to fan-out Mailbox. Example is vary similar
// to previous except that this time we intentionally want to have single producer
// that sends messages to many consumers.
func main() {
	mailbox := actor.NewMailbox[int]()

	mm := actor.FanOut(mailbox.ReceiveC(), 3)

	pw := &producerWorker{outC: mailbox.SendC()}
	cw1 := &consumerWorker{inC: mm[0].ReceiveC(), id: 1}
	cw2 := &consumerWorker{inC: mm[1].ReceiveC(), id: 2}
	cw3 := &consumerWorker{inC: mm[2].ReceiveC(), id: 3}

	a := actor.Combine(
		mailbox,
		// Note: We need to start/stop Mailboxes created by FanOut
		actor.FromMailboxes(mm),
		actor.New(pw),
		actor.New(cw1),
		actor.New(cw2),
		actor.New(cw3),
	)

	a.Start()
	defer a.Stop()

	select {}
}
