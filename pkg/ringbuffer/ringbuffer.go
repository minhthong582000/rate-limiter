package ringbuffer

import (
	"fmt"
)

// RingBuffer or Circular Buffer
type RingBuffer[T any] struct {
	buffer   []T
	len      uint64
	capacity uint64
	start    uint64
}

func NewRingBuffer[T any](capacity uint64) *RingBuffer[T] {
	return &RingBuffer[T]{
		capacity: capacity,
		buffer:   make([]T, capacity),
	}
}

func (r *RingBuffer[T]) IsFull() bool {
	return r.len == r.capacity
}

func (r *RingBuffer[T]) IsEmpty() bool {
	return r.len == 0
}

func (r *RingBuffer[T]) Capacity() uint64 {
	return r.capacity
}

func (r *RingBuffer[T]) Size() uint64 {
	return r.len
}

func (r *RingBuffer[T]) StartIndex() uint64 {
	return r.start
}

func (r *RingBuffer[T]) Enqueue(value T) error {
	if r.len == r.capacity {
		return fmt.Errorf("ring is full")
	}

	pos := (r.start + r.len) % r.capacity
	r.buffer[pos] = value
	r.len += 1

	return nil
}

func (r *RingBuffer[T]) Dequeue() (T, error) {
	if r.len == 0 {
		return *new(T), fmt.Errorf("ring is empty")
	}

	value := r.buffer[r.start]
	r.buffer[r.start] = *new(T) // Clear the value
	r.len -= 1
	r.start = (r.start + 1) % r.capacity

	return value, nil
}

func (r *RingBuffer[T]) Clear() {
	r.start = 0
	r.len = 0
	for i := range r.buffer {
		r.buffer[i] = *new(T)
	}
}
