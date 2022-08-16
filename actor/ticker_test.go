package actor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func Test_Ticker(t *testing.T) {
	t.Parallel()

	const (
		tickCount = 20
		interval  = time.Millisecond * 10
	)

	ticker := NewTicker(interval)
	assert.NotNil(t, ticker)

	ticker.Start()
	defer ticker.Stop()

	go func() {
		time.Sleep(interval * time.Duration(tickCount))
		ticker.Stop()
	}()

	ticks := 0

	for tickTime := range ticker.C() {
		assert.NotEmpty(t, tickTime)
		ticks++
	}

	assert.GreaterOrEqual(t, tickCount, ticks)
}
