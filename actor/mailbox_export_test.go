package actor

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
