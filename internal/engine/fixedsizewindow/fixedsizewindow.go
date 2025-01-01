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
	done := false

	for !done {
		lastState := f.state.Load()
		elapsed := now - lastState.lastTime
		fmt.Println(lastState.currCount, elapsed)

		if elapsed < 0 {
			fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
			return false
		}

		if lastState.lastTime == 0 || elapsed > f.windowSize.Milliseconds() {
			done = f.state.CompareAndSwap(lastState, &state{
				currCount: 1,
				lastTime:  now,
			})
			continue
		}

		if lastState.currCount < f.capacity {
			done = f.state.CompareAndSwap(lastState, &state{
				currCount: lastState.currCount + 1,
				lastTime:  lastState.lastTime,
			})
			continue
		}

		return false
	}

	return true
}

func (f *fixedSizeWindow) Allow() bool {
	return f.AllowAt(time.Now())
}
