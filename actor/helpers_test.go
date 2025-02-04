package actor_test

import (
	"fmt"
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

func tostr(v any) string { return fmt.Sprintf("%v", v) }

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

func createOnStopOption(t *testing.T, count int) (<-chan any, Option) {
	t.Helper()

	c := make(chan any, count)
	fn := func() {
		select {
		case c <- `🌚`:
		default:
			t.Fatal("onStopFunc should be called only once")
		}
	}

	return c, OptOnStop(fn)
}

func createOnStartOption(t *testing.T, count int) (<-chan any, Option) {
	t.Helper()

	c := make(chan any, count)
	fn := func(_ Context) {
		select {
		case c <- `🌞`:
		default:
			t.Fatal("onStart should be called only once")
		}
	}

	return c, OptOnStart(fn)
}

func createCombinedOnStopOption(t *testing.T, count int) (<-chan any, CombinedOption) {
	t.Helper()

	c := make(chan any, count)
	fn := func() {
		select {
		case c <- `🌚`:
		default:
			t.Fatal("onStopFunc should be called only once")
		}
	}

	return c, OptOnStopCombined(fn)
}

func createCombinedOnStartOption(t *testing.T, count int) (<-chan any, CombinedOption) {
	t.Helper()

	c := make(chan any, count)
	fn := func(_ Context) {
		select {
		case c <- `🌞`:
		default:
			t.Fatal("onStart should be called only once")
		}
	}

	return c, OptOnStartCombined(fn)
}
