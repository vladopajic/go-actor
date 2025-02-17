package actor_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
//
//nolint:maintidx // relax
func Test_Mailbox_OptEndAfterReceivingAll(t *testing.T) {
	t.Parallel()

	const initialMessagesCount = 1000

	sentMessages := atomic.Int64{}
	sendMessages := func(m Mailbox[int64]) {
		t.Helper()

		for i := range initialMessagesCount {
			assert.NoError(t, m.Send(ContextStarted(), int64(i)))
		}

		sentMessages.Add(initialMessagesCount)
	}
	sendMessagesConcurrentlyWithStop := func(m Mailbox[int64]) {
		t.Helper()

		for {
			id := sentMessages.Add(1) - 1 // -1 because we need old value

			err := m.Send(ContextStarted(), id)
			if err != nil {
				sentMessages.Add(-1) // because msg was not sent
				return
			}
		}
	}
	assertGotAllMessages := func(m Mailbox[int64]) <-chan int {
		t.Helper()

		resultC := make(chan int, 1)
		allMsgs := make(map[int64]struct{})

		go func() {
			for msg := range m.ReceiveC() {
				allMsgs[msg] = struct{}{}
			}

			resultC <- len(allMsgs)
		}()

		return resultC
	}

	m := NewMailbox[int64](OptStopAfterReceivingAll())
	m.Start()

	// send some messages to fill queue
	sendMessages(m)

	// before Stop is called, we are going to send messages (concurrently with Stop),
	// because we want to ensure that all those messages will be received as well.
	for range 20 {
		go sendMessagesConcurrentlyWithStop(m)
	}

	go func() {
		// Small sleep is needed because we want to give goroutine from above
		// greater chances to be running and sending messages when Stop() is called
		time.Sleep(time.Millisecond * 100) //nolint:forbidigo // explained
		m.Stop()
	}()

	gotMessages := <-assertGotAllMessages(m)

	assert.Equal(t, int(sentMessages.Load()), gotMessages)
	// must ensure to get more messages then initially sent in order to
	// be sure that messages have been send concurrently with Stop
	assert.Greater(t, gotMessages, initialMessagesCount)
}

// Test asserts mailbox invariants when `OptAsChan()` option is used.
func Test_Mailbox_AsChan(t *testing.T) {
	t.Parallel()

	t.Run("zero cap", func(t *testing.T) {
		t.Parallel()

		m := NewMailbox[any](OptAsChan())
		assertMailboxNotStarted(t, m)

		m.Start()

		assertSendBlocking(t, m)
		assertReceiveBlocking(t, m)
		assertSendWithCanceledCtx(t, m, true)
		assertSendReceiveSync(t, m, `🌹`)
		assertSendReceiveSync(t, m, nil)

		m.Stop()
		assertMailboxStopped(t, m)
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

	sendResultC, senderBlockedC := make(chan error, 1), make(chan any)
	m := NewMailbox[any](OptAsChan())
	m.Start()

	// Start goroutine that will send to mailbox, but since no one is waiting
	// to receive data from it should receive ErrMailboxStopped after mailbox is stopped.
	go func() {
		// This goroutine will notify that goroutine doing m.Send has been blocked.
		go func() {
			// sleeps gives more chance for parent goroutine to continue executing
			time.Sleep(time.Millisecond) //nolint:forbidigo // explained above
			close(senderBlockedC)
		}()

		sendResultC <- m.Send(ContextStarted(), `🌹`)
	}()

	<-senderBlockedC
	m.Stop() // stopping mailbox wile there is some goroutines trying to send

	assert.ErrorIs(t, <-sendResultC, ErrMailboxStopped, "Send() should result with error")
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

func assertSendBlocking(t *testing.T, m Mailbox[any]) {
	t.Helper()

	sendResultC := make(chan error, 1)
	ctx := NewContext()

	go func() {
		sendResultC <- m.Send(ctx, `🌹`)
	}()

	// This sleep is necessary to give some time goroutine from above
	// to be started and Send() method to get blocked while sending
	time.Sleep(time.Millisecond * 10) //nolint:forbidigo // relax
	ctx.End()

	sendErr := <-sendResultC
	assert.Error(t, sendErr)
	assert.ErrorIs(t, sendErr, ctx.Err())
	assert.NotErrorIs(t, sendErr, ErrMailboxStopped)
}
