package actor

import "time"

// Ticker interface wraps time.Ticker in actor like fashion.
type Ticker interface {
	Actor

	// C is the channel on which the ticks are delivered.
	C() <-chan time.Time
}

// NewTicker returns new ticker actor with period of the
// ticks specified by the duration argument.
func NewTicker(duration time.Duration) Ticker {
	w := &tickerWorker{
		duration: duration,
		outC:     make(chan time.Time, 1),
	}

	return &tickerImpl{
		Actor:  New(w, OptOnStart(w.onStart), OptOnStop(w.onStop)),
		worker: w,
	}
}

type tickerImpl struct {
	Actor
	worker *tickerWorker
}

func (t *tickerImpl) C() <-chan time.Time {
	return t.worker.outC
}

type tickerWorker struct {
	duration time.Duration
	outC     chan time.Time
	ticker   *time.Ticker
}

func (w *tickerWorker) onStart() {
	w.ticker = time.NewTicker(w.duration)
}

func (w *tickerWorker) onStop() {
	w.ticker.Stop()
	close(w.outC)
}

func (w *tickerWorker) DoWork(c Context) WorkerStatus {
	select {
	case t, ok := <-w.ticker.C:
		if !ok {
			return WorkerEnd
		}

		w.outC <- t

		return WorkerContinue

	case <-c.Done():
		return WorkerEnd
	}
}
