package actor

import (
	gocontext "context"
	"errors"
	"sync"
	"time"
)

// Context is a type alias for gocontext.Context.
//
// This Context is provided to Workers, allowing them to listen for and
// respond to stop signals sent by the Actor. It enables Workers to manage
// their execution flow, facilitating graceful termination and cancellation
// of ongoing tasks when the Actor is instructed to stop.
//
// Using Context allows Workers to check for cancellation signals and handle
// cleanup operations appropriately, ensuring that resources are released
// and tasks are completed in a controlled manner.
type Context = gocontext.Context

// ErrStopped is the error returned by Context.Err when the Actor is stopped.
var ErrStopped = errors.New("actor stopped")

//nolint:gochecknoglobals // these are singleton values
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

// ContextStarted returns a Context representing the started state of an Actor.
//
// This function is typically used in testing scenarios to provide a Context
// that indicates the Actor has begun execution. It can be passed to the
// Worker.DoWork() function to simulate the active state of an Actor during
// unit tests or integration tests.
func ContextStarted() Context {
	return contextStarted
}

// ContextEnded returns a Context representing the ended state of an Actor.
//
// This function is primarily used in testing scenarios to provide a Context
// that indicates the Actor has finished execution. It can be passed to the
// Worker.DoWork() function to simulate the inactive state of an Actor during
// unit tests or integration tests.
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

//nolint:revive // needed to satisfy context.Context interface
func (*context) Value(key any) any {
	return nil
}

func (c *context) String() string {
	return "actor.Context"
}
