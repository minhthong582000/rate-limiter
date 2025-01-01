package fixedsizewindow

import (
	"fmt"
	"sync/atomic"
	"time"
)

type state struct {
	currCount uint64
	lastTime  int64
}

type fixedSizeWindow struct {
	capacity   uint64 // Max requests allowed in the window
	windowSize time.Duration
	state      atomic.Pointer[state]
}

func NewFixedSizeWindow(
	capacity uint64,
	windowSize time.Duration,
) *fixedSizeWindow {
	if windowSize <= 0 {
		panic("window size must be greater than 0")
	}

	f := &fixedSizeWindow{
		capacity:   capacity,
		windowSize: windowSize,
	}
	f.state.Store(&state{
		currCount: 0,
		lastTime:  0,
	})
	return f
}

func (f *fixedSizeWindow) AllowAt(arriveAt time.Time) bool {
	now := arriveAt.UnixMilli()

	for {
		lastState := f.state.Load()
		elapsed := now - lastState.lastTime

		if elapsed < 0 {
			fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
			return false
		}

		// Reset the window if new request arrives after the window has expired
		// or this is the first request
		if lastState.lastTime == 0 || elapsed > f.windowSize.Milliseconds() {
			newState := &state{
				currCount: 1,
				lastTime:  now,
			}
			if f.state.CompareAndSwap(lastState, newState) {
				return true
			}
			// Retry if CAS fails
			continue
		}

		if lastState.currCount < f.capacity {
			newState := &state{
				currCount: lastState.currCount + 1,
				lastTime:  lastState.lastTime,
			}
			if f.state.CompareAndSwap(lastState, newState) {
				return true
			}
			// Retry if CAS fails
			continue
		}

		return false
	}
}

func (f *fixedSizeWindow) Allow() bool {
	return f.AllowAt(time.Now())
}
