package actor

import "fmt"

// Mailbox is interface for message transport mechanism between Actors.
type Mailbox[T any] interface {
	Actor
	MailboxSender[T]
	MailboxReceiver[T]
}

// MailboxSender is interface for sender bits of Mailbox.
type MailboxSender[T any] interface {
	// Send message via mailbox.
	Send(ctx Context, msg T) error
}

// MailboxReceiver is interface for receiver bits of Mailbox.
type MailboxReceiver[T any] interface {
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

// FanOut spawns new goroutine in which messages received by receiveC are forwarded
// to senders. Spawned goroutine will be active while receiveC is open.
func FanOut[T any, MS MailboxSender[T]](receiveC <-chan T, senders []MS) {
	ctx := ContextStarted()

	go func(receiveC <-chan T, senders []MS) {
		for v := range receiveC {
			for _, m := range senders {
				m.Send(ctx, v) //nolint:errcheck // errors are swallowed
			}
		}
	}(receiveC, senders)
}

// NewMailboxes returns slice of new Mailbox instances with specified count.
func NewMailboxes[T any](count int, opt ...Option) []Mailbox[T] {
	mm := make([]Mailbox[T], count)
	for i := 0; i < count; i++ {
		mm[i] = NewMailbox[T](opt...)
	}

	return mm
}

// NewMailbox returns new Mailbox.
func NewMailbox[T any](opt ...Option) Mailbox[T] {
	var (
		opts  = newOptions(opt)
		mOpts = opts.Mailbox
	)

	if mOpts.UsingChan {
		c := make(chan T, mOpts.Capacity)

		return &mailbox[T]{
			Actor:    Idle(OptOnStop(func() { close(c) })),
			sendC:    c,
			receiveC: c,
		}
	}

	var (
		sendC    = make(chan T)
		receiveC = make(chan T)
		queue    = newQueue[T](mOpts.Capacity, mOpts.MinCapacity)
		w        = newMailboxWorker(sendC, receiveC, queue)
	)

	return &mailbox[T]{
		Actor:    New(w),
		sendC:    sendC,
		receiveC: receiveC,
	}
}

type mailbox[T any] struct {
	Actor
	sendC    chan<- T
	receiveC <-chan T
}

func (m *mailbox[T]) Send(ctx Context, msg T) error {
	select {
	case m.sendC <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Mailbox.Send failed: %w", ctx.Err())
	}
}

func (m *mailbox[T]) ReceiveC() <-chan T {
	return m.receiveC
}

type mailboxWorker[T any] struct {
	receiveC chan T
	sendC    chan T
	queue    *queue[T]
}

func newMailboxWorker[T any](
	sendC,
	receiveC chan T,
	queue *queue[T],
) *mailboxWorker[T] {
	return &mailboxWorker[T]{
		sendC:    sendC,
		receiveC: receiveC,
		queue:    queue,
	}
}

func (w *mailboxWorker[T]) DoWork(c Context) WorkerStatus {
	if w.queue.IsEmpty() {
		select {
		case value := <-w.sendC:
			w.queue.PushBack(value)
			return WorkerContinue

		case <-c.Done():
			return WorkerEnd
		}
	}

	select {
	case w.receiveC <- w.queue.Front():
		w.queue.PopFront()
		return WorkerContinue

	case value := <-w.sendC:
		w.queue.PushBack(value)
		return WorkerContinue

	case <-c.Done():
		return WorkerEnd
	}
}

func (w *mailboxWorker[T]) OnStop() {
	close(w.sendC)
	close(w.receiveC)
}
