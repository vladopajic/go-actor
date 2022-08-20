package main

import (
	"fmt"
	"time"

	"github.com/vladopajic/go-actor/actor"
)

// NewCountdownActor creates new actor for launch pad countdowns.
func NewCountdownActor(launchReadySigC chan struct{}) actor.Actor {
	w := &countdownWorker{
		launchReadySigC: launchReadySigC,
		secondsCount:    3,
	}

	return actor.New(
		w,
		actor.OptOnStart(w.onStart),
		actor.OptOnStop(w.onStop),
	)
}

type countdownWorker struct {
	launchReadySigC chan struct{}
	secondsCount    int
}

func (w *countdownWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-time.After(time.Second):
		fmt.Printf("%d\n", w.secondsCount)

		w.secondsCount--

		if w.secondsCount == 0 {
			w.launchReadySigC <- struct{}{}
			return actor.WorkerEnd
		}

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}

func (w *countdownWorker) onStart() {
	fmt.Printf("countdown started\n")
}

func (w *countdownWorker) onStop() {
	fmt.Printf("countdown ended\n")
}
