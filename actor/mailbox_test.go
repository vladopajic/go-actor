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

	// Assert all values from receive channel
	for i := 0; i < sendCount; i++ {
		assert.Equal(t, <-receiveC, i)
	}

	// Send nil value and assert that it is received (test should't panic)
	sendC <- nil
	assert.Equal(t, <-receiveC, nil)
}
