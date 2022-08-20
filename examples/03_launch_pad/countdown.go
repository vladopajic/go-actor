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
	// Note: it's bad practice to implement workers that are not
	// responding on c.EndWorkC() signal. See example #03

	for i := w.secondsCount; i > 0; i-- {
		fmt.Printf("%d\n", i)
		time.Sleep(time.Second)
	}

	w.launchReadySigC <- struct{}{}

	return actor.WorkerEnd
}

func (w *countdownWorker) onStart() {
	fmt.Printf("countdown started\n")
}

func (w *countdownWorker) onStop() {
	fmt.Printf("countdown ended\n")
}
