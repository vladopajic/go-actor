package actor

import (
	queue2 "github.com/gammazero/deque"
)

const minQueueCapacity = 64

type queue[T any] interface {
	PushBack(val T)
	Front() (val T)
	PopFront() (val T)
	Size() int
	IsEmpty() bool
}

func newQueue[T any](capacity, minimum int) queue[T] {
	if minimum < minQueueCapacity {
		minimum = minQueueCapacity
	}

	return &queueWrap[T]{
		q: queue2.New[T](capacity, minimum),
	}
}

type queueWrap[T any] struct {
	q *queue2.Deque[T]
}

func (w *queueWrap[T]) PushBack(val T) {
	w.q.PushBack(val)
}

func (w *queueWrap[T]) Front() T {
	return w.q.Front()
}

func (w *queueWrap[T]) PopFront() T {
	return w.q.PopFront()
}

func (w *queueWrap[T]) Size() int {
	return w.q.Len()
}

func (w *queueWrap[T]) IsEmpty() bool {
	return w.q.Len() == 0
}
