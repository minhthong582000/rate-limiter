package ringbuffer

import (
	"fmt"
)

// RingBuffer
type RingBuffer[T any] struct {
	capacity   uint64
	start, end uint64
	buffer     []T
}

func NewRingBuffer[T any](capacity uint64) *RingBuffer[T] {
	return &RingBuffer[T]{
		capacity: capacity + 1,
		buffer:   make([]T, capacity+1),
	}
}

func (r *RingBuffer[T]) IsFull() bool {
	return r.incrementIndex(r.end) == r.start
}

func (r *RingBuffer[T]) IsEmpty() bool {
	return r.start == r.end
}

func (r *RingBuffer[T]) Capacity() uint64 {
	return r.capacity - 1
}

func (r *RingBuffer[T]) Size() uint64 {
	return r.end - r.start
}

func (r *RingBuffer[T]) StartIndex() uint64 {
	return r.start
}

// incrementIndex use modulus to calculate when the index should wrap to the beginning
// in a circular way
func (r *RingBuffer[T]) incrementIndex(index uint64) uint64 {
	return (index + 1) % r.capacity
}

func (r *RingBuffer[T]) PushBack(value T) error {
	if r.IsFull() {
		return fmt.Errorf("ring is full")
	}

	pos := r.end
	r.buffer[r.end] = value
	r.end = r.incrementIndex(pos)

	return nil
}

func (r *RingBuffer[T]) PopFront() (T, error) {
	if r.IsEmpty() {
		return *new(T), fmt.Errorf("ring is empty")
	}

	value := r.buffer[r.start]
	r.buffer[r.start] = *new(T) // Clear the value
	r.start = r.incrementIndex(r.start)

	return value, nil
}

func (r *RingBuffer[T]) PeekFront() (T, error) {
	if r.IsEmpty() {
		return *new(T), fmt.Errorf("ring is empty")
	}

	return r.buffer[r.start], nil
}

func (r *RingBuffer[T]) Clear() {
	for i := range r.capacity {
		r.buffer[i] = *new(T)
	}
	r.start = 0
	r.end = 0
}
