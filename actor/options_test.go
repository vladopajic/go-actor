package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	// NewOptions with no options should equal to zero options
	opts := NewOptions(nil)
	assert.Equal(t, NewZeroOptions(), opts)

	// Assert that OnStartFunc will be set
	opts = NewOptions([]Option{OptOnStart(func() {})})
	assert.NotNil(t, opts.OnStartFunc)
	assert.Nil(t, opts.OnStopFunc)

	// Assert that OnStopFunc will be set
	opts = NewOptions([]Option{OptOnStop(func() {})})
	assert.NotNil(t, opts.OnStopFunc)
	assert.Nil(t, opts.OnStartFunc)
}
