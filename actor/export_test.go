package actor

import "testing"

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

func NewQueue[T any](capacity, minimum int) *queue[T] {
	return newQueue[T](capacity, minimum)
}

func (q *queue[T]) Cap() int {
	return q.q.Cap()
}

func RandInt32(t *testing.T) int32 { //nolint:thelper // this is just export
	return randInt32(t)
}
