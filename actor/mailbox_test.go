package actor_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Mailbox(t *testing.T) {
	t.Parallel()

	const sendMessagesCount = 5000

	m := NewMailbox[any]()
	assert.NotNil(t, m)

	m.Start()

	// Send values via SendC() channel, then assert that values are received
	// on ReceiveC() channel.
	// It is important to first send all data to SendC() channel to simulate excessive
	// incoming messages on this Mailbox.

	for i := 0; i < sendMessagesCount; i++ {
		m.SendC() <- i
	}

	for i := 0; i < sendMessagesCount; i++ {
		assert.Equal(t, <-m.ReceiveC(), i)
	}

	// Assert that sending nil value should't cause panic
	assertSendReceive(t, m, nil)

	m.Stop()

	// After Mailbox is stopped assert that all channels are closed
	assertMailboxChannelsClosed(t, m)
}

func Test_FromMailboxes(t *testing.T) {
	t.Parallel()

	mm := []Mailbox[any]{
		NewMailbox[any](),
		NewMailbox[any](),
		NewMailbox[any](),
	}

	a := FromMailboxes(mm)
	assert.NotNil(t, a)

	a.Start()

	// After combined Agent is started all Mailboxes should be executing
	for _, m := range mm {
		assertSendReceive(t, m, "🌞")
	}

	a.Stop()

	// After combined Agent is stopped all Mailboxes should stop executing
	for _, m := range mm {
		assertMailboxChannelsClosed(t, m)
	}
}

func Test_FanOut(t *testing.T) {
	t.Parallel()

	const (
		sendMessagesCount = 100
		fanoutCount       = 5
	)

	wg := sync.WaitGroup{}
	inC := make(chan any)

	// Fan out inC channel
	fanOuts := FanOut(inC, fanoutCount)
	assert.Len(t, fanOuts, fanoutCount)

	a := FromMailboxes(fanOuts)

	a.Start()
	defer a.Stop()

	wg.Add(fanoutCount)

	// Produce data on inC channel
	go func() {
		for i := 0; i < sendMessagesCount; i++ {
			inC <- i
		}
		close(inC)
	}()

	// Assert that correct data is received by Mailbox
	for _, m := range fanOuts {
		go func(m Mailbox[any]) {
			for i := 0; i < sendMessagesCount; i++ {
				assert.Equal(t, i, <-m.ReceiveC())
			}
			wg.Done()
		}(m)
	}

	wg.Wait()

	// At this point inC is closed which will terminate fan out gorutine.
	// We don't want to end Mailbox actors at this point, because Mailbox could
	// be used for other data flows.

	// Assert that Mailbox actor is still working
	for _, m := range fanOuts {
		assertSendReceive(t, m, "🌞")
	}
}

func assertSendReceive(t *testing.T, m Mailbox[any], val any) {
	t.Helper()

	m.SendC() <- val
	assert.Equal(t, val, <-m.ReceiveC())
}

func assertMailboxChannelsClosed(t *testing.T, m Mailbox[any]) {
	t.Helper()

	assert.Panics(t, func() {
		m.SendC() <- "👹"
	})

	_, ok := <-m.ReceiveC()
	assert.False(t, ok)
}
