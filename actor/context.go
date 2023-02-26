package actor

import (
	gocontext "context"
	"errors"
	"sync"
	"time"
)

// Context is provided to Worker so they can listen and respond
// on stop signal sent from Actor.
type Context = gocontext.Context

// ErrStopped is the error returned by Context.Err when the Actor is stopped.
var ErrStopped = errors.New("actor stopped")

//nolint:gochecknoglobals
var (
	contextStarted = newContext()

	contextEnded = func() *context {
		doneC := make(chan struct{})
		close(doneC)

		return &context{
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

type context struct {
	mu    sync.Mutex
	err   error
	doneC chan struct{}
}

func newContext() *context {
	return &context{
		doneC: make(chan struct{}),
	}
}

func (c *context) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *context) end() {
	c.mu.Lock()
	if c.err == nil {
		c.err = ErrStopped
		close(c.doneC)
	}
	c.mu.Unlock()
}

func (c *context) Done() <-chan struct{} {
	return c.doneC
}

func (c *context) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	return err
}

func (*context) Value(key any) any {
	return nil
}

func (c *context) String() string {
	return "actor.Context"
}
