package actor

const (
	MinQueueCapacity = minQueueCapacity
)

func NewContext() *contextImpl {
	return newContext()
}

func (c *contextImpl) SignalEnd() {
	c.signalEnd()
}

func NewOptions(opts ...Option) options {
	return newOptions(opts)
}

func NewZeroOptions() options {
	return options{}
}

func NewMailboxWorker[T any](
	sendC,
	receiveC chan T,
	queue *queue[T],
) *mailboxWorker[T] {
	return newMailboxWorker(sendC, receiveC, queue)
}

func NewQueue[T any](capacity, minimum int) *queue[T] {
	return newQueue[T](capacity, minimum)
}

func (q *queue[T]) Cap() int {
	return q.q.Cap()
}
