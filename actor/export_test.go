package actor

func NewContext() *contextImpl {
	return newContext()
}

func (c *contextImpl) SignalEnd() {
	c.signalEnd()
}

func NewOptions(opts []Option) options {
	return newOptions(opts)
}

func NewZeroOptions() options {
	return options{}
}

func NewMailboxWorker[T any](
	sendC,
	receiveC chan T,
	queue queue[T],
) *mailboxWorker[T] {
	return &mailboxWorker[T]{
		sendC:    sendC,
		receiveC: receiveC,
		queue:    queue,
	}
}

func NewQueue[T any]() queue[T] {
	return newQueue[T]()
}
