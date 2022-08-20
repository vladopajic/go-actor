package main

import (
	"time"
)

// This program shows improved example #02
func main() {
	launchReadySigC := make(chan struct{})

	a := NewCountdownActor(launchReadySigC)

	a.Start()

	// Stop before countdown has ended
	time.Sleep(time.Second * 2)
	a.Stop()

	// This will halt program execution because caundown was stopped
	// (there are no live goroutines which can write to this channel)
	<-launchReadySigC
}
