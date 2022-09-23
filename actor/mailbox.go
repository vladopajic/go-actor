package actor

import (
	queue "github.com/golang-ds/queue/linkedqueue"
)

// Mailbox is interface for message transport mechanism between Actors
// which can receive infinite number of messages.
// Mailbox is much like native go channel, except that writing to the Mailbox
// will never block, all messages are going to be queued and Actors on
// receiving end of the Mailbox will get all messages in FIFO order.
type Mailbox[T any] interface {
	Actor

	// SendC returns channel where data can be sent.
	SendC() chan<- T

	// ReceiveC returns channel where data can be received.
	ReceiveC() <-chan T
}

// FromMailboxes creates single Actor combining actors of supplied Mailboxes.
func FromMailboxes[T any](mm []Mailbox[T]) Actor {
	a := make([]Actor, len(mm))
	for i, m := range mm {
		a[i] = m
	}

	return Combine(a...)
}

// FanOut crates new Mailboxes whose receiving messages are driven by suppled
// receiveC channel. FanOut spawns new goroutine in which messages received by
// receiveC channel are forwarded to created Mailboxes. Spawned goroutine will
// be active while receiveC is open and it's up to user to start and stop Mailboxes.
func FanOut[T any](receiveC <-chan T, count int) []Mailbox[T] {
	mm := make([]Mailbox[T], count)
	for i := 0; i < count; i++ {
		mm[i] = NewMailbox[T]()
	}

	go func(receiveC <-chan T, mm []Mailbox[T]) {
		for v := range receiveC {
			for _, m := range mm {
				m.SendC() <- v
			}
		}
	}(receiveC, mm)

	return mm
}

// NewMailbox returns new Mailbox.
func NewMailbox[T any]() Mailbox[T] {
	q := queue.New[T]()

	w := &mailboxWorker[T]{
		sendC:    make(chan T),
		receiveC: make(chan T),
		queue:    &q,
	}

	return &mailboxImpl[T]{
		Actor:  New(w, OptOnStop(w.onStop)),
		worker: w,
	}
}

type mailboxImpl[T any] struct {
	Actor
	worker *mailboxWorker[T]
}

func (m *mailboxImpl[T]) SendC() chan<- T {
	return m.worker.sendC
}

func (m *mailboxImpl[T]) ReceiveC() <-chan T {
	return m.worker.receiveC
}

type mailboxWorker[T any] struct {
	receiveC chan T
	sendC    chan T
	queue    *queue.LinkedQueue[T]
}

func (w *mailboxWorker[T]) DoWork(c Context) WorkerStatus {
	if w.queue.IsEmpty() {
		select {
		case value := <-w.sendC:
			w.queue.Enqueue(value)
			return WorkerContinue

		case <-c.Done():
			return WorkerEnd
		}
	}

	select {
	case w.receiveC <- first(w.queue):
		w.queue.Dequeue()
		return WorkerContinue

	case value := <-w.sendC:
		w.queue.Enqueue(value)
		return WorkerContinue

	case <-c.Done():
		return WorkerEnd
	}
}

func first[T any](queue *queue.LinkedQueue[T]) T {
	v, _ := queue.First()
	return v
}

func (w *mailboxWorker[T]) onStop() {
	close(w.sendC)
	close(w.receiveC)
}
