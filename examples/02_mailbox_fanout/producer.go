package main

import (
	"time"

	"github.com/vladopajic/go-actor/actor"
)

// producerWorker will produce incremented number on 1 second interval
type producerWorker struct {
	outC chan<- int
	num  int
}

func (w *producerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-time.After(time.Second):
		w.num++
		w.outC <- w.num
		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}
