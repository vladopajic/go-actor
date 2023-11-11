package actor_test

import "testing"

func drainC(c <-chan any, count int) {
	for i := 0; i < count; i++ {
		<-c
	}
}

type tWrapper struct {
	*testing.T
	hadError bool
}

func (t *tWrapper) Error(args ...any) {
	t.hadError = true
}
