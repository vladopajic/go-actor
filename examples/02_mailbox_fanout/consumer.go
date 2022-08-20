package main

import (
	"fmt"

	"github.com/vladopajic/go-actor/actor"
)

// consumerWorker will consume numbers received on inC channel
type consumerWorker struct {
	inC <-chan int
	id  int
}

func (w *consumerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case num := <-w.inC:
		fmt.Printf("consumed %d \t(worker %d)\n", num, w.id)

		return actor.WorkerContinue

	case <-c.Done():
		return actor.WorkerEnd
	}
}
