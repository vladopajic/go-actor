package actor

import (
	"io"
	"testing"
)

const (
	MinQueueCapacity = minQueueCapacity
)

type (
	ActorImpl      = actor
	OptionsMailbox = optionsMailbox
)

func NewActorImpl(w Worker, opt ...Option) *ActorImpl {
	a := New(w, opt...)
	return a.(*actor) //nolint:forcetypeassert //relax
}

func (a *ActorImpl) OnStart() {
	a.onStart()
}

func (a *ActorImpl) OnStop() {
	a.onStop()
}

func NewContext() *context {
	return newContext()
}

func (c *context) End() {
	c.end()
}

func NewOptions[T ~func(o *options)](opts ...T) options {
	return newOptions(opts)
}

func NewZeroOptions() options {
	return options{}
}

func NewMailboxWorker[T any](
	sendC,
	receiveC chan T,
	mOpts optionsMailbox,
) *mailboxWorker[T] {
	return newMailboxWorker(sendC, receiveC, mOpts)
}

func (w *mailboxWorker[T]) Queue() *queue[T] {
	return w.queue
}

func NewQueue[T any](capacity int) *queue[T] {
	return newQueue[T](capacity)
}

func (q *queue[T]) Cap() int {
	return q.q.Cap()
}

//nolint:thelper // this is just export
func RandInt32WithReader(tb testing.TB, reader io.Reader) int32 {
	return randInt32WithReader(tb, reader)
}
