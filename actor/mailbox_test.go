package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Mailbox(t *testing.T) {
	t.Parallel()

	m := NewMailbox[interface{}]()
	sendC, receiveC := m.SendC(), m.ReceiveC()

	m.Start()
	defer m.Stop()

	const sendCount = 5

	// Send values via send channel
	for i := 0; i < sendCount; i++ {
		sendC <- i
	}
	sendC <- nil
	assert.Len(t, sendC, 0)
	assert.Len(t, receiveC, 0)

	// Assert all values from receive channel
	for i := 0; i < sendCount; i++ {
		assert.Equal(t, <-receiveC, i)
	}
	assert.Equal(t, <-receiveC, nil)
	assert.Len(t, sendC, 0)
	assert.Len(t, receiveC, 0)
}
