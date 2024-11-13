package actor

import gocontext "context"

// SyncMailbox is used to synchronously send data, and wait for it to process before returning.
type SyncMailbox[T any] struct {
	mbx Mailbox[Callback[T]]
}

func NewSyncMailbox[T any]() *SyncMailbox[T] {
	return &SyncMailbox[T]{
		mbx: NewMailbox[Callback[T]](),
	}
}

func (sm *SyncMailbox[T]) Start() {
	sm.mbx.Start()
}

func (sm *SyncMailbox[T]) Stop() {
	sm.mbx.Stop()
}

func (sm *SyncMailbox[T]) ReceiveC() <-chan Callback[T] {
	return sm.mbx.ReceiveC()
}

func (sm *SyncMailbox[T]) Send(ctx gocontext.Context, value T) error {
	done := make(chan struct{})
	defer close(done)
	err := sm.mbx.Send(ctx, Callback[T]{
		Value: value,
		done:  done,
	})
	if err != nil {
		return err
	}
	<-done
	return nil
}

type Callback[T any] struct {
	Value T
	done  chan struct{}
}

// Notify must be called to return the synchronous call.
func (c *Callback[T]) Notify() {
	c.done <- struct{}{}
}
