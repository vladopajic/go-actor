package actortest_test

import "testing"

type actorStub struct {
	name string
	log  *[]string
}

func (a actorStub) Start() {
	*a.log = append(*a.log, "start "+a.name)
}

func (a actorStub) Stop() {
	*a.log = append(*a.log, "stop "+a.name)
}

type tbWrapper struct {
	*testing.T
	hadError  bool
	hadFatal  bool
	fatalArgs []any
}

func (tb *tbWrapper) Error(...any) {
	tb.hadError = true
}

type fatalCalled struct{}

func (tb *tbWrapper) Fatal(args ...any) {
	tb.hadFatal = true
	tb.fatalArgs = args

	panic(fatalCalled{}) //nolint:forbidigo // test double for Fatal must not return
}
