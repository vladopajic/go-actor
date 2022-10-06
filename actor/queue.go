package actor

import (
	queueimpl "github.com/golang-ds/queue/linkedqueue"
)

type queue[T any] interface {
	Enqueue(val T)
	Dequeue() (val T, ok bool)
	First() (val T, ok bool)
	IsEmpty() bool
	Size() int
}

func newQueue[T any]() queue[T] {
	q := queueimpl.New[T]()
	return &q
}
