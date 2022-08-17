package actor

import (
	queue "github.com/golang-ds/queue/linkedqueue"
)

// Mailbox is interface for channel which can receive infinite number of messages.
type Mailbox[T any] interface {
	Actor

	// SendC returns channel where data can be sent.
	SendC() chan<- T

	// ReceiveC returns channel where data can be received.
	ReceiveC() <-chan T
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

//nolint:revive // showing false error
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

//nolint:revive // showing false error
func (w *mailboxWorker[T]) onStop() {
	close(w.sendC)
	close(w.receiveC)
}
