package actor

const (
	MinQueueCapacity = minQueueCapacity
)

func NewContext() *context {
	return newContext()
}

func (c *context) End() {
	c.end()
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
