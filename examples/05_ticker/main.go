package main

import (
	"fmt"
	"time"

	"github.com/vladopajic/go-actor/actor"
)

// This program demonstrates usage of actor.Ticker.
// See actor.Ticker implementation details.
func main() {
	a := actor.NewTicker(time.Millisecond * 500)
	a.Start()

	// Start new gorutine to handle ticks sent to ticker channel
	go handleTick(a.C())

	time.Sleep(time.Second * 2)
	a.Stop()
}

func handleTick(c <-chan time.Time) {
	for {
		if _, ok := <-c; !ok {
			fmt.Print("ticker stopped\n")
			return
		}

		fmt.Printf("tick\n")
	}
}
