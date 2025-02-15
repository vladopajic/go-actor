//go:build experimental
// +build experimental

package actor_test

// This file contains experimental tests which utilize "testing/synctest"
// package that elegantly solve race issues which have been previously
// hacked with time.Sleep()

import (
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

// Test asserts that actor should stop after worker
// has signaled that there is no more work via WorkerEnd signal.
func Test_Actor_StopAfterWorkerEnded_Experimental(t *testing.T) {
	t.Parallel()

	synctest.Run(func() {
		var ctx Context

		workIteration := 0
		doWorkC := make(chan chan int)
		workEndedC := make(chan struct{})
		workerFunc := func(c Context) WorkerStatus {
			ctx = c

			// assert that DoWork should not be called
			// after WorkerEnd signal is returned
			select {
			case <-workEndedC:
				assert.FailNow(t, "worker should be ended")
			default:
			}

			select {
			case p, ok := <-doWorkC:
				if !ok {
					close(workEndedC)
					return WorkerEnd
				}

				p <- workIteration

				workIteration++

				return WorkerContinue

			case <-c.Done():
				// Test should fail if done signal is received from Actor
				assert.FailNow(t, "worker should be ended")
				return WorkerEnd
			}
		}

		a := New(NewWorker(workerFunc))

		a.Start()

		assertDoWork(t, doWorkC)

		// Closing doWorkC will cause worker to end
		close(doWorkC)

		// Assert that context is ended after worker ends.
		// We use synctest.Wait() to ensure that actor goroutine has ended.
		<-workEndedC
		synctest.Wait()
		assertContextEnded(t, ctx)

		// Stopping actor should produce no effect (since worker has ended)
		a.Stop()

		assertContextEnded(t, ctx)
	})
}

// Test asserts that mailbox `Send()` returns error when sending data is blocked and
// Stop() is simultaneously called.
func Test_Mailbox_AsChan_SendStopped_Experimental(t *testing.T) {
	t.Parallel()

	synctest.Run(func() {
		m := NewMailbox[any](OptAsChan())
		m.Start()
		sendResultC := make(chan error, 1)

		// NOTE: must use NewContext() instead of ContextStarted() because
		// later creates channels outside of the bubble.
		ctx := NewContext()

		// Start goroutine that will send to mailbox, but since no one is waiting
		// to receive data from it should receive stopped error after mailbox is stopped.

		go func() {
			sendResultC <- m.Send(ctx, `🌹`)
		}()

		synctest.Wait()
		m.Stop() // stopping mailbox while there is some goroutines trying to send

		assert.ErrorIs(t, <-sendResultC, ErrMailboxStopped, "Send() should result with error")

		// sending again should result with stopped
		assert.ErrorIs(t, m.Send(ctx, `🌹`), ErrMailboxStopped, "Send() should result with error")
	})
}

func Test_Mailbox_AsChan_SendCanceled_Experimental(t *testing.T) {
	t.Parallel()

	synctest.Run(func() {
		m := NewMailbox[any](OptAsChan())
		m.Start()
		defer m.Stop()

		sendResultC := make(chan error, 1)

		ctx := NewContext()

		// Start goroutine that will send to mailbox, but since no one is waiting
		// to receive data from it should receive send cancelled error after context is canceled.
		go func() {
			sendResultC <- m.Send(ctx, `🌹`)
		}()

		synctest.Wait()
		ctx.End()

		sendErr := <-sendResultC
		assert.Error(t, sendErr)
		assert.ErrorIs(t, sendErr, ctx.Err())
		assert.NotErrorIs(t, sendErr, ErrMailboxStopped)
		assertReceiveBlocking(t, m) // should not have anything to receive

		// sending again with started context should succeed
		go func() {
			sendResultC <- m.Send(NewContext(), `🌹`)
		}()
		assert.Equal(t, `🌹`, <-m.ReceiveC())
		assert.NoError(t, <-sendResultC)
	})
}
