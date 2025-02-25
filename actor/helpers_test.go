package actor_test

import (
	"io"
	"testing"

	. "github.com/vladopajic/go-actor/actor"
)

func drainC(c <-chan any, count int) {
	for range count {
		<-c
	}
}

type tWrapper struct {
	*testing.T
	hadError bool
}

//nolint:revive // this is method override of same signature
func (t *tWrapper) Error(args ...any) {
	t.hadError = true
}

type errReader struct{}

//nolint:revive // this is method implements io.Reader
func (r errReader) Read(b []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

type delegateActor struct {
	start func()
	stop  func()
}

func (a delegateActor) Start() {
	if fn := a.start; fn != nil {
		fn()
	}
}

func (a delegateActor) Stop() {
	if fn := a.stop; fn != nil {
		fn()
	}
}

func createOnStopOption(t *testing.T, count int) (<-chan any, func()) {
	t.Helper()

	c := make(chan any, count)
	fn := func() {
		select {
		case c <- `🌚`:
		default:
			t.Fatal("onStop should be called only once")
		}
	}

	return c, fn
}

func createOnStartOption(t *testing.T, count int) (<-chan any, func(Context)) {
	t.Helper()

	c := make(chan any, count)
	fn := func(Context) {
		select {
		case c <- `🌞`:
		default:
			t.Fatal("onStart should be called only once")
		}
	}

	return c, fn
}
