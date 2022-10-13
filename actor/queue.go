package actor

import (
	queueImpl "github.com/gammazero/deque"
)

const minQueueCapacity = 64

func newQueue[T any](capacity, minimum int) *queue[T] {
	if minimum < minQueueCapacity {
		minimum = minQueueCapacity
	}

	return &queue[T]{
		q: queueImpl.New[T](capacity, minimum),
	}
}

type queue[T any] struct {
	q *queueImpl.Deque[T]
}

func (q *queue[T]) PushBack(val T) {
	q.q.PushBack(val)
}

func (q *queue[T]) Front() T {
	return q.q.Front()
}

func (q *queue[T]) PopFront() T {
	return q.q.PopFront()
}

func (q *queue[T]) Size() int {
	return q.q.Len()
}

func (q *queue[T]) IsEmpty() bool {
	return q.q.Len() == 0
}
