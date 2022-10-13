package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	q := NewQueue[int](0, 0)

	assert.Equal(t, 0, q.Size())
	assert.True(t, q.IsEmpty())

	// Push 1
	q.PushBack(1)
	assert.Equal(t, 1, q.Size())
	assert.False(t, q.IsEmpty())

	// Push 2
	q.PushBack(2)
	assert.Equal(t, 2, q.Size())
	assert.False(t, q.IsEmpty())

	// Push 3
	q.PushBack(3)
	assert.False(t, q.IsEmpty())
	assert.Equal(t, 3, q.Size())
	assert.False(t, q.IsEmpty())

	// PopFront (expect 1)
	assert.Equal(t, 1, q.Front())
	assert.Equal(t, 1, q.PopFront())
	assert.Equal(t, 2, q.Size())

	// PopFront (expect 2)
	assert.Equal(t, 2, q.Front())
	assert.Equal(t, 2, q.PopFront())
	assert.Equal(t, 1, q.Size())

	// PopFront (expect 3)
	assert.Equal(t, 3, q.Front())
	assert.Equal(t, 3, q.PopFront())
	assert.Equal(t, 0, q.Size())
}
