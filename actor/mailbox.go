package actor

import (
	"fmt"
	"sync"
	"sync/atomic"
)

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

	return Combine(a...).Build()
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
func NewMailboxes[T any](count int, opt ...MailboxOption) []Mailbox[T] {
	mm := make([]Mailbox[T], count)
	for i := 0; i < count; i++ {
		mm[i] = NewMailbox[T](opt...)
	}

	return mm
}

const mbxChanBufferCap = 64

// NewMailbox returns new local Mailbox implementation.
// Mailbox is much like native go channel, except that writing to the Mailbox
// will never block, all messages are going to be queued and Actors on
// receiving end of the Mailbox will get all messages in FIFO order.
func NewMailbox[T any](opt ...MailboxOption) Mailbox[T] {
	options := newOptions(opt).Mailbox

	if options.AsChan {
		c := make(chan T, options.Capacity)

		return &mailboxSync[T]{
			Actor:    Idle(OptOnStop(func() { close(c) })),
			sendC:    c,
			receiveC: c,
		}
	}

	var (
		sendC       = make(chan T, mbxChanBufferCap)
		receiveC    = make(chan T, mbxChanBufferCap)
		w           = newMailboxWorker(sendC, receiveC, options)
		sendHandler = &atomic.Value{}
	)

	sendHandler.Store(creatPanicHandler[T]("could not send to Mailbox that was not started"))

	return &mailbox[T]{
		actor:       New(w),
		sendC:       sendC,
		receiveC:    receiveC,
		sendHandler: sendHandler,
	}
}

type mailboxSync[T any] struct {
	Actor
	sendC    chan<- T
	receiveC <-chan T
}

func (m *mailboxSync[T]) Send(ctx Context, msg T) error {
	select {
	case m.sendC <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("Mailbox.Send canceled: %w", ctx.Err())
	}
}

func (m *mailboxSync[T]) ReceiveC() <-chan T {
	return m.receiveC
}

type sendHandler[T any] func(ctx Context, msg T) error

type mailbox[T any] struct {
	actor    Actor
	sendC    chan<- T
	receiveC <-chan T

	running     bool
	lock        sync.Mutex
	sendHandler *atomic.Value
}

func (m *mailbox[T]) Start() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.sendHandler.Store(createSendHandler(m.sendC))
	m.actor.Start()
}

func (m *mailbox[T]) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	if !m.running {
		return
	}

	m.running = false
	m.sendHandler.Store(
		creatPanicHandler[T]("could not send to Mailbox after it has been stopped"),
	)
	m.actor.Stop()
}

func (m *mailbox[T]) Send(ctx Context, msg T) error {
	h := m.sendHandler.Load().(sendHandler[T]) //nolint:forcetypeassert // relax
	return h(ctx, msg)
}

func creatPanicHandler[T any](msg string) sendHandler[T] {
	return func(_ Context, _ T) error {
		panic(msg) //nolint:forbidigo // panic is intentional
	}
}

func createSendHandler[T any](sendC chan<- T) sendHandler[T] {
	return func(ctx Context, msg T) error {
		select {
		case sendC <- msg:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("Mailbox.Send canceled: %w", ctx.Err())
		}
	}
}

func (m *mailbox[T]) ReceiveC() <-chan T {
	return m.receiveC
}

type mailboxWorker[T any] struct {
	receiveC chan T
	sendC    chan T
	queue    *queue[T]
	options  optionsMailbox
}

func newMailboxWorker[T any](
	sendC,
	receiveC chan T,
	options optionsMailbox,
) *mailboxWorker[T] {
	queue := newQueue[T](options.Capacity, options.MinCapacity)

	return &mailboxWorker[T]{
		sendC:    sendC,
		receiveC: receiveC,
		queue:    queue,
		options:  options,
	}
}

func (w *mailboxWorker[T]) DoWork(c Context) WorkerStatus {
	if w.queue.IsEmpty() {
		select {
		case <-c.Done():
			return WorkerEnd

		case value := <-w.sendC:
			w.queue.PushBack(value)
			return WorkerContinue
		}
	}

	select {
	case <-c.Done():
		return WorkerEnd

	case w.receiveC <- w.queue.Front():
		w.queue.PopFront()
		return WorkerContinue

	case value := <-w.sendC:
		w.queue.PushBack(value)
		return WorkerContinue
	}
}

func (w *mailboxWorker[T]) OnStop() {
	// close sendC to prevent anyone from writing to this mailbox
	close(w.sendC)

	// close receiveC channel after all data from queue is received
	if w.options.StopAfterReceivingAll {
		for !w.queue.IsEmpty() {
			w.receiveC <- w.queue.PopFront()
		}

		for msg := range w.sendC {
			w.receiveC <- msg
		}
	}

	close(w.receiveC)
}
