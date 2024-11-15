//go:build !race

package actor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestMailboxSync_SendStopped(t *testing.T) {
	t.Parallel()

	testDoneC, senderStarted := make(chan any), make(chan any)
	m := NewMailbox[any](OptAsChan())
	m.Start()

	go func() {
		// This goroutine will notify that goroutine doing m.Send has been blocked.
		go func() {
			// this sleeps gives more chance for parent goroutine to continue executing
			time.Sleep(time.Millisecond) //nolint:forbidigo // explained above
			close(senderStarted)
		}()

		assert.ErrorIs(t, m.Send(ContextStarted(), `🌹`), ErrMailboxStopped)
		close(testDoneC)
	}()

	<-senderStarted
	m.Stop()
	<-testDoneC
}
