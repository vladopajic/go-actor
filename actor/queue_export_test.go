package actor

func NewQueue[T any]() queue[T] {
	return newQueue[T]()
}
