package actor

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Context is provided so Worker can listen and respond on stop signal sent from Actor.
type Context = context.Context

// ErrStopped is the error returned by Context.Err when the Actor is stopped.
var ErrStopped = errors.New("actor stopped")

//nolint:gochecknoglobals
var (
	contextStarted = newContext()

	contextEnded = func() *contextImpl {
		doneC := make(chan struct{})
		close(doneC)

		return &contextImpl{
			doneC: doneC,
			err:   ErrStopped,
		}
	}()
)

// ContextStarted returns Context representing started state for Actor.
// It is typically used in tests for passing to Worker.DoWork() function.
func ContextStarted() Context {
	return contextStarted
}

// ContextStarted returns Context representing ended state for Actor.
// It is typically used in tests for passing to Worker.DoWork() function.
func ContextEnded() Context {
	return contextEnded
}

type contextImpl struct {
	mu    sync.Mutex
	err   error
	doneC chan struct{}
}

func newContext() *contextImpl {
	return &contextImpl{
		doneC: make(chan struct{}),
	}
}

func (c *contextImpl) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *contextImpl) signalEnd() {
	c.mu.Lock()
	c.err = ErrStopped
	close(c.doneC)
	c.mu.Unlock()
}

func (c *contextImpl) Done() <-chan struct{} {
	return c.doneC
}

func (c *contextImpl) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	return err
}

func (*contextImpl) Value(key any) any {
	return nil
}

func (c *contextImpl) String() string {
	return "actor.Context"
}
