package fixedsizewindow

import (
	"fmt"
	"sync/atomic"
	"time"
)

type fixedSizeWindow struct {
	capacity   uint64 // Max requests allowed in the window
	windowSize time.Duration
	currCount  atomic.Uint64
	lastTime   atomic.Int64
}

func NewFixedSizeWindow(
	capacity uint64,
	windowSize time.Duration,
) *fixedSizeWindow {
	if windowSize <= 0 {
		panic("window size must be greater than 0")
	}

	return &fixedSizeWindow{
		capacity:   capacity,
		windowSize: windowSize,
	}
}

func (f *fixedSizeWindow) AllowAt(arriveAt time.Time) bool {
	now := arriveAt.UnixNano()
	lastTime := f.lastTime.Load()
	elapsed := now - lastTime

	if elapsed < 0 {
		fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
		return false
	}

	if lastTime == 0 || elapsed > f.windowSize.Nanoseconds() {
		if f.lastTime.CompareAndSwap(lastTime, now) {
			f.currCount.Store(1)
			return true
		}
	}

	if f.currCount.Load() < f.capacity {
		if f.currCount.Add(1) <= f.capacity {
			return true
		}
	}

	return false
}

func (f *fixedSizeWindow) Allow() bool {
	return f.AllowAt(time.Now())
}
