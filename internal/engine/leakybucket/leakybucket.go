package leakybucket

import (
	"fmt"
	"sync"
	"time"

	"github.com/minhthong582000/rate-limiter/pkg/ringbuffer"
)

type leakyBucket struct {
	capacity  uint64 // Max burst
	drainRate time.Duration
	queue     *ringbuffer.RingBuffer[time.Time]
	mutex     sync.Mutex
	stopCh    <-chan struct{}
}

func NewLeakyBucket(
	capacity uint64,
	drainRate time.Duration,
	stopCh <-chan struct{},
) *leakyBucket {
	if drainRate <= 0 {
		panic("drain rate must be greater than 0")
	}

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	l := &leakyBucket{
		capacity:  capacity,
		drainRate: drainRate,
		queue:     ringbuffer.NewRingBuffer[time.Time](capacity),
		stopCh:    stopCh,
	}

	go l.leak()

	return l
}

func (l *leakyBucket) AllowAt(arriveAt time.Time) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.queue.IsFull() {
		fmt.Println("Queue is full")
		return false
	}

	err := l.queue.PushBack(arriveAt)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func (l *leakyBucket) Allow() bool {
	return l.AllowAt(time.Now())
}

func (l *leakyBucket) leak() {
	ticker := time.NewTicker(l.drainRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mutex.Lock()

			if !l.queue.IsEmpty() {
				request, err := l.queue.PopFront()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("Processed request: %v\n", request)
				}
			}

			l.mutex.Unlock()
		case <-l.stopCh:
			fmt.Println("Leaky bucket stopped")
			return
		}
	}
}
