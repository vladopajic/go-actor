package actor_test

import (
	"testing"

	. "github.com/vladopajic/go-actor/actor"
)

const largeCap = 1024

func BenchmarkMailbox(b *testing.B) {
	mbx := NewMailbox[any]()

	benchmarkMailbox(b, mbx)
}

func BenchmarkMailboxWithLargeCap(b *testing.B) {
	mbx := NewMailbox[any](OptCapacity(largeCap))

	benchmarkMailbox(b, mbx)
}

func BenchmarkMailboxAsChan(b *testing.B) {
	mbx := NewMailbox[any](OptAsChan())

	benchmarkMailbox(b, mbx)
}

func BenchmarkMailboxAsChanWithLargeCap(b *testing.B) {
	mbx := NewMailbox[any](OptAsChan(), OptCapacity(largeCap))

	benchmarkMailbox(b, mbx)
}

func benchmarkMailbox(b *testing.B, mbx Mailbox[any]) {
	b.Helper()

	mbx.Start()
	defer mbx.Stop()

	doneC := make(chan any)

	go func() {
		for range b.N {
			<-mbx.ReceiveC()
		}

		close(doneC)
	}()

	ctx := ContextStarted()
	for range b.N {
		mbx.Send(ctx, `🌞`) //nolint:errcheck // error should never happen
	}

	<-doneC
}
