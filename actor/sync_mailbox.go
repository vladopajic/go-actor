package actor

import gocontext "context"

// SyncMailbox is used to synchronously send data, and wait for it to process before returning.
type SyncMailbox[T any] struct {
	mbx Mailbox[*Callback[T]]
}

func NewSyncMailbox[T any](opts ...MailboxOption) *SyncMailbox[T] {
	return &SyncMailbox[T]{
		mbx: NewMailbox[*Callback[T]](opts...),
	}
}

func (sm *SyncMailbox[T]) Start() {
	sm.mbx.Start()
}

func (sm *SyncMailbox[T]) Stop() {
	sm.mbx.Stop()
}

func (sm *SyncMailbox[T]) ReceiveC() <-chan *Callback[T] {
	return sm.mbx.ReceiveC()
}

func (sm *SyncMailbox[T]) Send(ctx gocontext.Context, value T) error {
	done := make(chan error, 1)
	defer close(done)
	err := sm.mbx.Send(ctx, &Callback[T]{
		Value: value,
		done:  done,
	})
	if err != nil {
		return err
	}
	return <-done
}

var _ CallbackHook = (*Callback[any])(nil)

type CallbackHook interface {
	Notify(err error)
}

type Callback[T any] struct {
	Value T
	done  chan error
}

// Notify must be called to return the synchronous call.
func (c *Callback[T]) Notify(err error) {
	c.done <- err
}
