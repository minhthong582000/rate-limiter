package slidingwindow

import (
	"sync"
	"time"

	"github.com/minhthong582000/rate-limiter/pkg/ringbuffer"
)

type slidingWindowLogs struct {
	capacity   uint64 // Max requests allowed in the window
	windowSize time.Duration
	requestLog *ringbuffer.RingBuffer[time.Time]
	mutex      sync.Mutex
}

func NewSlidingWindowLogs(
	capacity uint64,
	windowSize time.Duration,
) *slidingWindowLogs {
	if windowSize <= 0 {
		panic("window size must be greater than 0")
	}

	return &slidingWindowLogs{
		capacity:   capacity,
		windowSize: windowSize,
		requestLog: ringbuffer.NewRingBuffer[time.Time](capacity),
	}
}

func (f *slidingWindowLogs) AllowAt(arriveAt time.Time) bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for !f.requestLog.IsEmpty() {
		lastLog, _ := f.requestLog.PeekFront()
		if arriveAt.Sub(lastLog) > f.windowSize {
			f.requestLog.PopFront()
		} else {
			break
		}
	}

	if f.requestLog.Size() < f.capacity {
		if err := f.requestLog.PushBack(arriveAt); err == nil {
			return true
		}
	}

	return false
}

func (f *slidingWindowLogs) Allow() bool {
	return f.AllowAt(time.Now())
}
