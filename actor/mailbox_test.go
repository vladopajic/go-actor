package actor_test

import (
	"sync"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

// Test asserts that FromMailboxes creates single actor using multiple mailboxes.
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
		assertSendReceiveAsync(t, m, `🌹`)
	}

	a.Stop()

	// After combined Agent is stopped all Mailboxes should stop executing
	for _, m := range mm {
		assertMailboxStopped(t, m)
	}
}

// Test asserts correct behavior of FanOut utility.
func Test_FanOut(t *testing.T) {
	t.Parallel()

	const (
		sendMessagesCount = 100
		fanOutCount       = 5
	)

	wg := sync.WaitGroup{}
	inMbx := NewMailbox[any]()
	fanMbxx := NewMailboxes[any](fanOutCount)

	// Fan out inMbx
	FanOut(inMbx.ReceiveC(), fanMbxx)

	a := Combine(inMbx, FromMailboxes(fanMbxx)).Build()

	a.Start()
	defer a.Stop()

	wg.Add(fanOutCount)

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
		assertSendReceiveAsync(t, m, `🌹`)
	}
}

// Test asserts that MailboxWorker returns `WorkerEnd` when context is canceled.
func Test_MailboxWorker_EndSignal(t *testing.T) {
	t.Parallel()

	w := NewMailboxWorker(make(chan any), make(chan any), OptionsMailbox{})
	assert.NotNil(t, w)

	// Worker should signal end with empty queue
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))

	// Worker should signal end with non-empty queue
	w.Queue().PushBack(`🌹`)
	assert.Equal(t, WorkerEnd, w.DoWork(ContextEnded()))
}

// Test which runs tests in AssertMailboxInvariantsAsync helper function.
func Test_Mailbox_Invariants(t *testing.T) {
	t.Parallel()

	AssertMailboxInvariantsAsync(t, func() Mailbox[any] {
		return NewMailbox[any]()
	})
}

// Test asserts that mailbox will receive (enqueue)
// larger number of messages without blocking.
func Test_Mailbox_MessageQueue(t *testing.T) {
	t.Parallel()

	const sendMessagesCount = 5000

	m := NewMailbox[any]()
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

	m.Stop()
	assertMailboxStopped(t, m)
}

// Test asserts that Mailbox will end only after all messages have been received.
func Test_Mailbox_OptEndAfterReceivingAll(t *testing.T) {
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

// Test asserts mailbox invariants when `OptAsChan()` option is used.
func Test_Mailbox_AsChan(t *testing.T) {
	t.Parallel()

	t.Run("zero cap", func(t *testing.T) {
		t.Parallel()

		synctest.Run(func() {
			m := NewMailbox[any](OptAsChan())
			assertMailboxNotStarted(t, m)

			m.Start()

			assertSendBlocking(t, m, synctest.Wait)
			assertReceiveBlocking(t, m)
			assertSendWithCanceledCtx(t, m, true)
			assertSendReceiveSync(t, m, `🌹`)
			assertSendReceiveSync(t, m, nil)

			m.Stop()
			assertMailboxStopped(t, m)
		})
	})

	t.Run("non zero cap", func(t *testing.T) {
		t.Parallel()

		AssertMailboxInvariantsAsync(t, func() Mailbox[any] {
			return NewMailbox[any](OptAsChan(), OptCapacity(1))
		})
	})
}

// Test asserts that mailbox `Send()` returns error when sending data is blocked and
// Stop() is simultaneously called.
func Test_Mailbox_AsChan_SendStopped(t *testing.T) {
	t.Parallel()

	synctest.Run(func() {
		m := NewMailbox[any](OptAsChan())
		m.Start()
		sendResultC := make(chan error, 1)

		// Start goroutine that will send to mailbox, but since no one is waiting
		// to receive data from it should receive stopped error after mailbox is stopped.
		go func() {
			// NOTE: must use NewContext() instead of ContextStarted() because
			// later creates channels outside of the bubble.
			sendResultC <- m.Send(NewContext(), `🌹`)
		}()

		synctest.Wait()
		m.Stop() // stopping mailbox while there is some goroutines trying to send

		sendErr := <-sendResultC
		assert.Error(t, sendErr)
		assert.ErrorIs(t, sendErr, ErrMailboxStopped)
	})
}

// AssertMailboxInvariantsAsync is helper functions that asserts mailbox invariants.
func AssertMailboxInvariantsAsync(t *testing.T, mFact func() Mailbox[any]) {
	t.Helper()

	t.Run("basic invariants", func(t *testing.T) {
		t.Parallel()

		m := mFact()
		assertMailboxNotStarted(t, m)

		m.Start()
		assertSendReceiveAsync(t, m, `🌹`)
		assertSendReceiveAsync(t, m, nil)
		assertSendWithCanceledCtx(t, m, false)

		m.Stop()
		assertMailboxStopped(t, m)
	})

	t.Run("start-stop", func(t *testing.T) {
		t.Parallel()

		m := mFact()

		m.Start()
		m.Start()
		assertSendReceiveAsync(t, m, `🌹`)

		m.Stop()
		m.Stop()
		assertMailboxStopped(t, m)

		m.Start() // Should have no effect
		assertMailboxStopped(t, m)
	})

	//nolint:testifylint // intentionally using ==
	t.Run("receive channel is not changed", func(t *testing.T) {
		t.Parallel()

		m := mFact()
		initialC := m.ReceiveC()

		// when mailbox is started assert that ReceiveC is the same as initial
		m.Start()
		// sending some data to ensure that mbx has fully started
		assertSendReceiveAsync(t, m, `🌹`)
		assert.True(t, initialC == m.ReceiveC(), "expecting the same reference for ReceiveC")

		// when mailbox is stopped assert ReceiveC should be the same as initial
		m.Stop()
		assert.True(t, initialC == m.ReceiveC(), "expecting the same reference for ReceiveC")
	})
}

// Asserts that sending with canceled context will end with error.
func assertSendWithCanceledCtx(t *testing.T, m Mailbox[any], immediate bool) {
	t.Helper()

	// Because send is blocking, it must return error immediately
	if immediate {
		assert.ErrorIs(t, m.Send(ContextEnded(), `🌹`), ContextEnded().Err())
		return
	}

	// Since sendC has some buffer it is possible that some attempts will succeed
	for {
		err := m.Send(ContextEnded(), `🌹`)
		if err != nil {
			assert.ErrorIs(t, err, ContextEnded().Err())
			return
		}

		<-m.ReceiveC() // Drain message since it went through
	}
}

func assertSendReceiveAsync(t *testing.T, m Mailbox[any], val any) {
	t.Helper()

	assert.NoError(t, m.Send(ContextStarted(), val))
	assert.Equal(t, val, <-m.ReceiveC())
}

func assertSendReceiveSync(t *testing.T, m Mailbox[any], val any) {
	t.Helper()

	go func() {
		assert.NoError(t, m.Send(ContextStarted(), val))
	}()
	assert.Equal(t, val, <-m.ReceiveC())
}

func assertMailboxStopped(t *testing.T, m Mailbox[any]) {
	t.Helper()

	assert.ErrorIs(t, m.Send(ContextStarted(), `👹`), ErrMailboxStopped)

	_, ok := <-m.ReceiveC()
	assert.False(t, ok)
}

func assertMailboxNotStarted(t *testing.T, m Mailbox[any]) {
	t.Helper()

	assert.ErrorIs(t, m.Send(ContextStarted(), `👹`), ErrMailboxNotStarted,
		"should not be able to send as Mailbox is not started")

	assertReceiveBlocking(t, m)
}

func assertReceiveBlocking(t *testing.T, m Mailbox[any]) {
	t.Helper()

	select {
	case <-m.ReceiveC():
		assert.FailNow(t, "should not be able to receive")
	default:
	}
}

func assertSendBlocking(t *testing.T, m Mailbox[any], synctestWait func()) {
	t.Helper()

	sendResultC := make(chan error, 1)
	ctx := NewContext()

	// Start goroutine that will send to mailbox, but since no one is waiting
	// to receive data from it should receive send cancelled error after context is canceled.
	go func() {
		sendResultC <- m.Send(ctx, `🌹`)
	}()

	synctestWait()
	ctx.End()

	sendErr := <-sendResultC
	assert.Error(t, sendErr)
	assert.ErrorIs(t, sendErr, ctx.Err())
	assert.NotErrorIs(t, sendErr, ErrMailboxStopped)
}
