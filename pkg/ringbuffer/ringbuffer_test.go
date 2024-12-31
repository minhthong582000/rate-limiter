package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRingBuffer tests buffer initialization
func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int64](5)

	assert.NotNil(t, rb)
	assert.Equal(t, uint64(5), rb.Capacity())
	assert.Equal(t, uint64(0), rb.Size())
	assert.Equal(t, uint64(0), rb.StartIndex())
}

// TestEnqueue tests enqueuing values into the buffer
func TestEnqueue(t *testing.T) {
	rb := NewRingBuffer[int64](3)

	// Enqueue 3 items
	assert.NoError(t, rb.PushBack(1))
	assert.NoError(t, rb.PushBack(2))
	assert.NoError(t, rb.PushBack(3))

	// Buffer should now be full
	assert.True(t, rb.IsFull())
	assert.Equal(t, uint64(3), rb.Size())

	// Attempt to enqueue beyond capacity
	assert.Error(t, rb.PushBack(4))
}

// TestDequeue tests dequeuing values from the buffer
func TestDequeue(t *testing.T) {
	rb := NewRingBuffer[string](3)

	// Dequeue from an empty buffer
	_, err := rb.PopFront()
	assert.Error(t, err)

	// Enqueue 2 items
	assert.NoError(t, rb.PushBack("10"))
	assert.NoError(t, rb.PushBack("20"))

	// Dequeue and validate
	value, err := rb.PopFront()
	assert.NoError(t, err)
	assert.Equal(t, "10", value)

	value, err = rb.PopFront()
	assert.NoError(t, err)
	assert.Equal(t, "20", value)

	// Buffer should now be empty
	assert.True(t, rb.IsEmpty())
}

// TestEnqueueDequeueCycle tests multiple enqueue and dequeue operations
func TestEnqueueDequeueCycle(t *testing.T) {
	rb := NewRingBuffer[int64](2)

	// Fill buffer
	assert.NoError(t, rb.PushBack(5))
	assert.NoError(t, rb.PushBack(10))
	assert.True(t, rb.IsFull())

	// Dequeue one item
	value, err := rb.PopFront()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value)

	// Add another item
	assert.NoError(t, rb.PushBack(15))

	// Dequeue remaining items
	value, err = rb.PopFront()
	assert.NoError(t, err)
	assert.Equal(t, int64(10), value)

	value, err = rb.PeekFront()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), value)
	value, err = rb.PopFront()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), value)

	assert.True(t, rb.IsEmpty())
}

// TestIsEmpty tests the IsEmpty method
func TestIsEmpty(t *testing.T) {
	rb := NewRingBuffer[int64](2)

	assert.True(t, rb.IsEmpty())

	assert.NoError(t, rb.PushBack(1))
	assert.False(t, rb.IsEmpty())

	_, _ = rb.PopFront()
	assert.True(t, rb.IsEmpty())
}

// TestIsFull tests the IsFull method
func TestIsFull(t *testing.T) {
	rb := NewRingBuffer[int64](2)

	assert.False(t, rb.IsFull())

	assert.NoError(t, rb.PushBack(1))
	assert.NoError(t, rb.PushBack(2))
	assert.True(t, rb.IsFull())
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	rb := NewRingBuffer[int64](3)

	assert.NoError(t, rb.PushBack(1))
	assert.NoError(t, rb.PushBack(2))

	assert.False(t, rb.IsEmpty())

	rb.Clear()
	assert.True(t, rb.IsEmpty())
	assert.Equal(t, uint64(0), rb.Size())
	assert.Equal(t, uint64(0), rb.StartIndex())
}

// TestPeekFront tests the PeekFront method
func TestPeekFront(t *testing.T) {
	rb := NewRingBuffer[int64](3)

	_, err := rb.PeekFront()
	assert.Error(t, err, "PeekFront on empty buffer should return an error")

	assert.NoError(t, rb.PushBack(1))
	assert.NoError(t, rb.PushBack(2))

	value, err := rb.PeekFront()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), value)
}
