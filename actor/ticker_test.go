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

	// Start Ticker and stop it after tickCount intervals
	ticker.Start()

	go func() {
		time.Sleep(interval * time.Duration(tickCount))
		// Sleep little bit more to be sure to receive last tick
		time.Sleep(interval / 2)
		ticker.Stop()
	}()

	// Read ticks form C() channel while it's open.
	// We should receive tickCount tick on this channel before it's closed.
	// We leave possability for flaky test because last tick may not be received
	// if Ticker finishes work before sending last tick.

	ticks := 0
	lastTick := time.Now()
	acceptableDelta := time.Nanosecond * 10

	for tickTime := range ticker.C() {
		ticks++

		assert.InDelta(t, lastTick.Add(interval).Unix(), tickTime.Unix(), float64(acceptableDelta))

		lastTick = tickTime
	}

	assert.Equal(t, tickCount, ticks)
}
