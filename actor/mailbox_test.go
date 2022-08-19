package actor_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Mailbox(t *testing.T) {
	t.Parallel()

	m := NewMailbox[any]()
	sendC, receiveC := m.SendC(), m.ReceiveC()

	m.Start()
	defer m.Stop()

	const sendCount = 5000

	// Send values via sendC channel, then assert values from receive channel.
	// It is important to first write to sendC channel to simulate excessive
	// incoming messages.

	for i := 0; i < sendCount; i++ {
		sendC <- i
	}

	for i := 0; i < sendCount; i++ {
		assert.Equal(t, <-receiveC, i)
	}

	// Send nil value and assert that it is received (test should't panic)
	assertSendReceive(t, m, nil)
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
	defer a.Stop()

	for _, m := range mm {
		assertSendReceive(t, m, "🌞")
	}
}

func Test_FanOut(t *testing.T) {
	t.Parallel()

	const (
		testDataCount = 100
		fanoutCount   = 5
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
		for i := 0; i < testDataCount; i++ {
			inC <- i
		}
		close(inC)
	}()

	// Assert that correct data is received by Mailbox
	for _, m := range fanOuts {
		go func(m Mailbox[any]) {
			for i := 0; i < testDataCount; i++ {
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
