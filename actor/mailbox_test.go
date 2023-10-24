package actor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_MailboxWorker_EndSignal(t *testing.T) {
	t.Parallel()

	sendC := make(chan any)
	receiveC := make(chan any)
	q := NewQueue[any](0, 0)

	w := NewMailboxWorker(sendC, receiveC, q)
	assert.NotNil(t, w)

	// Worker should signal end with empty queue
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))

	// Worker should signal end with non-empty queue
	q.PushBack(`🌹`)
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))
}

func Test_Mailbox(t *testing.T) {
	t.Parallel()

	const sendMessagesCount = 5000

	m := NewMailbox[any]()
	assert.NotNil(t, m)

	assertSendBlocking(t, m)
	assertReceiveBlocking(t, m)

	m.Start()

	// Send values via SendC() channel, then assert that values are received
	// on ReceiveC() channel.
	// It is important to first send all data to SendC() channel to simulate excessive
	// incoming messages on this Mailbox.

	for i := 0; i < sendMessagesCount; i++ {
		assert.NoError(t, m.Send(ContextStarted(), i))
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
		assertSendReceive(t, m, `🌹`)
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
	inMbx := NewMailbox[any]()
	fanMbxx := NewMailboxes[any](fanoutCount)

	// Fan out inMbx
	FanOut(inMbx.ReceiveC(), fanMbxx)

	a := Combine(inMbx, FromMailboxes(fanMbxx)).Build()

	a.Start()
	defer a.Stop()

	wg.Add(fanoutCount)

	// Produce data on inMbx
	go func() {
		for i := 0; i < sendMessagesCount; i++ {
			assert.NoError(t, inMbx.Send(ContextStarted(), i))
		}
	}()

	// Assert that correct data is received by fanMbxx
	for _, m := range fanMbxx {
		go func(m Mailbox[any]) {
			for i := 0; i < sendMessagesCount; i++ {
				assert.Equal(t, i, <-m.ReceiveC())
			}
			wg.Done()
		}(m)
	}

	wg.Wait()

	// Closing inMbx will terminate fan out goroutine.
	// At this point we don't want to end receiving mailboxes of fan out,
	// because those Mailbox could be used for other data flows.
	inMbx.Stop()

	// Assert that Mailbox actor is still working
	for _, m := range fanMbxx {
		assertSendReceive(t, m, `🌹`)
	}
}

func Test_MailboxOptAsChan(t *testing.T) {
	t.Parallel()

	t.Run("zero cap", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan())

		m.Start()

		assertSendBlocking(t, m)
		assertReceiveBlocking(t, m)

		// Send when there is receiver
		go func() {
			assert.NoError(t, m.Send(ContextStarted(), `🌹`))
		}()
		assert.Equal(t, `🌹`, <-m.ReceiveC())

		m.Stop()

		assertMailboxChannelsClosed(t, m)
	})

	t.Run("non zero cap", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan(), OptCapacity(1))

		m.Start()

		assertSendReceive(t, m, `🌹`)

		m.Stop()

		assertMailboxChannelsClosed(t, m)
	})
}

func assertSendReceive(t *testing.T, m Mailbox[any], val any) {
	t.Helper()

	assert.NoError(t, m.Send(ContextStarted(), val))
	assert.Equal(t, val, <-m.ReceiveC())
}

func assertMailboxChannelsClosed(t *testing.T, m Mailbox[any]) {
	t.Helper()

	assert.Panics(t, func() {
		m.Send(ContextStarted(), `👹`) //nolint:errcheck // this line panics
	})

	_, ok := <-m.ReceiveC()
	assert.False(t, ok)
}

func assertReceiveBlocking(t *testing.T, m Mailbox[any]) {
	t.Helper()

	select {
	case <-m.ReceiveC():
		assert.FailNow(t, "should not be able to receive")
	default:
	}
}

func assertSendBlocking(t *testing.T, m Mailbox[any]) {
	t.Helper()

	testDoneSigC := make(chan struct{})

	ctx := NewContext()

	go func() {
		err := m.Send(ctx, `🌹`)
		if err == nil {
			assert.FailNow(t, "should not be able to send")
		}

		close(testDoneSigC)
	}()

	// This sleep is necessary to give some time goroutine from above
	// to be started and Send() method to get blocked while sending
	time.Sleep(time.Millisecond * 10) //nolint:forbidigo // relax
	ctx.End()

	<-testDoneSigC
}
