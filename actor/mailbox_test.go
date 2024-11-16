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
	options := OptionsMailbox{}

	w := NewMailboxWorker(sendC, receiveC, options)
	assert.NotNil(t, w)

	// Worker should signal end with empty queue
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))

	// Worker should signal end with non-empty queue
	w.Queue().PushBack(`🌹`)
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))
}

func Test_Mailbox(t *testing.T) {
	t.Parallel()

	const sendMessagesCount = 5000

	m := NewMailbox[any]()
	assert.NotNil(t, m)

	assert.Error(t, m.Send(ContextStarted(), `👹`))
	assertReceiveBlocking(t, m)

	m.Start()

	// Send values via Send() method, then assert that values are received
	// on ReceiveC() channel.
	// It is important to first send all data to Send() method to simulate excessive
	// incoming messages on this Mailbox.

	for i := range sendMessagesCount {
		assert.NoError(t, m.Send(ContextStarted(), i))
	}

	for i := range sendMessagesCount {
		assert.Equal(t, <-m.ReceiveC(), i)
	}

	// Assert that sending nil value should't cause panic
	assertSendReceive(t, m, nil)

	m.Stop()

	// After Mailbox is stopped assert that all channels are closed
	assertMailboxStopped(t, m)
}

func Test_Mailbox_StartStop(t *testing.T) {
	t.Parallel()

	m := NewMailbox[any]()

	m.Start()
	m.Start()

	assertSendReceive(t, m, `🌹`)

	m.Stop()
	m.Stop()

	assertMailboxStopped(t, m)
}

func Test_Mailbox_SendWithEndedCtx(t *testing.T) {
	t.Parallel()

	m := NewMailbox[any]()

	m.Start()
	defer m.Stop()

	// Assert that sending with canceled context will end with error.
	// Since sendC has some buffer it is possible that some attempts will succeed.
	for {
		err := m.Send(ContextEnded(), `🌹`)
		if err != nil {
			assert.ErrorIs(t, err, ContextEnded().Err())
			break
		}
	}
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
		assertMailboxStopped(t, m)
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
		for i := range sendMessagesCount {
			assert.NoError(t, inMbx.Send(ContextStarted(), i))
		}
	}()

	// Assert that correct data is received by fanMbxx
	for _, m := range fanMbxx {
		go func(m Mailbox[any]) {
			for i := range sendMessagesCount {
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

func Test_Mailbox_AsChan(t *testing.T) {
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

		assertMailboxStopped(t, m)
	})

	t.Run("non zero cap", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan(), OptCapacity(1))
		m.Start()

		assertSendReceive(t, m, `🌹`)

		m.Stop()

		assertMailboxStopped(t, m)
	})

	t.Run("send with canceled context", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan())
		m.Start()

		err := m.Send(ContextEnded(), `🌹`)
		assert.ErrorIs(t, err, ContextEnded().Err())
	})

	t.Run("send not started", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan())

		err := m.Send(ContextStarted(), `🌹`)
		assert.ErrorIs(t, err, ErrMailboxNotStarted)
	})
}

func Test_Mailbox_AsChan_SendStopped(t *testing.T) {
	t.Parallel()

	testDoneC, senderBlockedC := make(chan any), make(chan any)
	m := NewMailbox[any](OptAsChan())
	m.Start()

	// Start goroutine that will send to mailbox, but since no one is waiting
	// to receive data from it should receive sopped error after mailbox is stopped.
	go func() {
		// This goroutine will notify that goroutine doing m.Send has been blocked.
		go func() {
			// sleeps gives more chance for parent goroutine to continue executing
			time.Sleep(time.Millisecond) //nolint:forbidigo // explained above
			close(senderBlockedC)
		}()

		assert.ErrorIs(t, m.Send(ContextStarted(), `🌹`), ErrMailboxStopped)
		close(testDoneC)
	}()

	<-senderBlockedC
	m.Stop() // stopping mailbox wile there is some goroutines trying to send
	<-testDoneC
}

// This test asserts that Mailbox will end only after all messages have been received.
func Test_Mailbox_OptEndAferReceivingAll(t *testing.T) {
	t.Parallel()

	const messagesCount = 1000

	sendMessages := func(m Mailbox[any]) {
		t.Helper()

		for i := range messagesCount {
			assert.NoError(t, m.Send(ContextStarted(), `🥥`+tostr(i)))
		}
	}
	assertGotAllMessages := func(m Mailbox[any]) {
		t.Helper()

		gotMessages := 0

		for msg := range m.ReceiveC() {
			assert.Equal(t, `🥥`+tostr(gotMessages), msg)

			gotMessages++
		}

		assert.Equal(t, messagesCount, gotMessages)
	}

	t.Run("the-best-way", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptStopAfterReceivingAll())
		m.Start()
		sendMessages(m)

		// Stop has to be called in goroutine because Stop is blocking until
		// actor (mailbox) has fully ended. And current thread of execution is needed
		// to read data from mailbox.
		go m.Stop()

		assertGotAllMessages(m)
	})

	t.Run("suboptimal-way", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptStopAfterReceivingAll())
		m.Start()
		sendMessages(m)

		// This time we start goroutine which will read all messages from mailbox instead of
		// stopping in separate goroutine.
		// There are no guaranies that this goroutine will finish after Stop is called, so
		// it could be the case that this goroutine has received all messages from mailbox,
		// even before mailbox was stopped. Which wouldn't correctly assert this feature.
		go assertGotAllMessages(m)

		m.Stop()
	})
}

func assertSendReceive(t *testing.T, m Mailbox[any], val any) {
	t.Helper()

	assert.NoError(t, m.Send(ContextStarted(), val))
	assert.Equal(t, val, <-m.ReceiveC())
}

func assertMailboxStopped(t *testing.T, m Mailbox[any]) {
	t.Helper()

	assert.ErrorIs(t, m.Send(ContextStarted(), `👹`), ErrMailboxStopped)

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
			assert.FailNow(t, "should not be able to send") //nolint:testifylint // relax
		}

		close(testDoneSigC)
	}()

	// This sleep is necessary to give some time goroutine from above
	// to be started and Send() method to get blocked while sending
	time.Sleep(time.Millisecond * 10) //nolint:forbidigo // relax
	ctx.End()

	<-testDoneSigC
}
