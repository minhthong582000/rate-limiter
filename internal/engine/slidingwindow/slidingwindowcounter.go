package slidingwindow

import (
	"fmt"
	"sync/atomic"
	"time"
)

type state struct {
	currCount  float64
	prevCount  float64
	currWindow int64
}

type slidingWindowCounter struct {
	capacity   float64 // Max requests allowed in the window
	windowSize int64   // Window size in millisecond
	state      atomic.Pointer[state]
}

func NewSlidingWindowCounter(
	capacity float64,
	windowSize int64,
) *slidingWindowCounter {
	if windowSize <= 0 {
		panic("window size must be greater than 0")
	}

	s := &slidingWindowCounter{
		capacity:   capacity,
		windowSize: windowSize,
	}
	s.state.Store(&state{
		currCount:  0,
		prevCount:  0,
		currWindow: 0,
	})
	return s
}

func (s *slidingWindowCounter) AllowAt(arriveAt time.Time) bool {
	now := arriveAt.UnixMilli()

	for {
		lastState := s.state.Load()
		// Copy the last state and use it throughout the calculation
		newState := *lastState

		// Print the address of lastState and newState
		fmt.Printf("lastState: %p, newState: %p\n", lastState, &newState)

		currWindow := now / s.windowSize

		if currWindow < lastState.currWindow {
			fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
			return false
		}

		// We are in a new window
		if currWindow > lastState.currWindow {
			newState.currCount = 0
			newState.prevCount = lastState.currCount
			newState.currWindow = currWindow

			// Reset the prevCount as we don't observe any request in the previous window
			if currWindow > lastState.currWindow+1 {
				newState.prevCount = 0
			}
		}

		// Estimate the current count based on the average request count in the previous window
		prevWindowWeight := 1 - (float64(now%s.windowSize) / float64(s.windowSize))
		estimatedCurrCount := newState.prevCount*prevWindowWeight + newState.currCount

		if estimatedCurrCount < s.capacity {
			newState.currCount += 1
			if s.state.CompareAndSwap(lastState, &newState) {
				return true
			}
			// Retry if CAS fails
			continue
		}

		return false
	}
}

func (s *slidingWindowCounter) Allow() bool {
	return s.AllowAt(time.Now())
}
