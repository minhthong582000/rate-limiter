package fixedsizewindow

import (
	"sync/atomic"
	"time"
)

type state struct {
	currCount uint64
	lastTime  time.Time
}

type fixedSizeWindow struct {
	capacity   uint64 // Max requests allowed in the window
	windowSize int64
	state      atomic.Pointer[state]
}

func NewFixedSizeWindow(
	capacity uint64,
	windowSize int64,
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
		lastTime:  time.Unix(0, 0).UTC(),
	})
	return f
}

func (f *fixedSizeWindow) AllowAt(arriveAt time.Time) bool {
	for {
		lastState := f.state.Load()
		elapsed := arriveAt.Sub(lastState.lastTime).Milliseconds()

		if elapsed < 0 {
			// A lot of contention results in lots of CAS retries.
			// This might causes the lastState.lastTime to be in the future of arriveAt.
			return false
		}

		// Reset the window if new request arrives after the window has expired
		if elapsed > f.windowSize {
			newState := &state{
				currCount: 1,
				lastTime:  arriveAt,
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
