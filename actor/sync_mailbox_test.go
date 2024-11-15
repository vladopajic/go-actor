package actor_test

import (
	gocontext "context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/vladopajic/go-actor/actor"
)

func Test_SyncMailbox_Basic(t *testing.T) {
	t.Parallel()

	m := NewSyncMailbox[string]()
	assert.NotNil(t, m)

	m.Start()
	defer m.Stop()

	// Set up receiver
	var received []string
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for cb := range m.ReceiveC() {
			received = append(received, cb.Value)
			cb.Notify(nil)
		}
	}()

	// Test sending multiple messages
	ctx := gocontext.Background()
	assert.NoError(t, m.Send(ctx, "hello"))
	assert.NoError(t, m.Send(ctx, "world"))

	m.Stop()
	wg.Wait()

	assert.Equal(t, []string{"hello", "world"}, received)
}

func Test_SyncMailbox_OptionsDontPanicOnStop(t *testing.T) {
	t.Parallel()

	m := NewSyncMailbox[string](OptDontPanicOnStop())
	assert.NotNil(t, m)

	m.Start()
	m.Stop()

	// Test sending multiple messages
	ctx := gocontext.Background()
	assert.Error(t, m.Send(ctx, "hello"))
}

func Test_SyncMailbox_StopBeforeSend(t *testing.T) {
	t.Parallel()

	m := NewSyncMailbox[string]()
	assert.NotNil(t, m)

	m.Stop()
	// Don't start the mailbox at all
	assert.Panics(t, func() {
		m.Send(gocontext.Background(), "should fail")
	})
}

func Test_SyncMailbox_StopBeforeSendWithOptDontPanicOnStop(t *testing.T) {
	t.Parallel()

	m := NewSyncMailbox[string](OptDontPanicOnStop())
	assert.NotNil(t, m)

	m.Start()
	m.Stop()
	err := m.Send(gocontext.Background(), "should fail")
	assert.Error(t, err)
}

func Test_SyncMailbox_Concurrent(t *testing.T) {
	t.Parallel()

	m := NewSyncMailbox[int]()
	assert.NotNil(t, m)

	m.Start()
	defer m.Stop()

	// Set up receiver
	var received []int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for cb := range m.ReceiveC() {
			mu.Lock()
			received = append(received, cb.Value)
			mu.Unlock()
			cb.Notify(nil)
		}
	}()

	// Send concurrently
	const numSenders = 10
	const messagesPerSender = 100
	var sendWg sync.WaitGroup
	sendWg.Add(numSenders)

	for i := 0; i < numSenders; i++ {
		go func(senderID int) {
			defer sendWg.Done()
			for j := 0; j < messagesPerSender; j++ {
				value := senderID*messagesPerSender + j
				assert.NoError(t, m.Send(gocontext.Background(), value))
			}
		}(i)
	}

	sendWg.Wait()
	m.Stop()
	wg.Wait()

	assert.Equal(t, numSenders*messagesPerSender, len(received))
}
