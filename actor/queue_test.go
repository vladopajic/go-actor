package actor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-actor/actor"
)

func TestEnqueue(t *testing.T) {
	t.Parallel()

	queue := NewQueue[int]()

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	assert.Equal(t, 3, queue.Size())
}

func TestDequeue(t *testing.T) {
	t.Parallel()

	queue := NewQueue[int]()

	assert.Equal(t, 0, queue.Size())

	f, ok := queue.Dequeue()
	assert.False(t, ok)
	assert.Equal(t, 0, f)

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	assert.Equal(t, 3, queue.Size())

	f, ok = queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, f)

	assert.Equal(t, 2, queue.Size())
}

func TestFirst(t *testing.T) {
	t.Parallel()

	queue := NewQueue[int]()

	f, ok := queue.First()
	assert.False(t, ok)
	assert.Equal(t, 0, f)

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	f, ok = queue.First()
	assert.True(t, ok)
	assert.Equal(t, 1, f)

	queue.Dequeue()
	f, ok = queue.First()
	assert.True(t, ok)
	assert.Equal(t, 2, f)

	queue.Dequeue()
	f, ok = queue.First()
	assert.True(t, ok)
	assert.Equal(t, 3, f)

	queue.Dequeue()
	f, ok = queue.First()
	assert.False(t, ok)
	assert.Equal(t, 0, f)
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	queue := NewQueue[int]()

	assert.True(t, queue.IsEmpty())

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	assert.False(t, queue.IsEmpty())
}
