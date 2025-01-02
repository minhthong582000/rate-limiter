package tokenbucket

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

type state struct {
	currToken float64
	lastTime  int64
}

type tokenBucket struct {
	capacity    float64 // Max burst
	fillRate    float64 // Token fill rate per millisecond
	consumeRate float64 // Token consume rate per request
	state       atomic.Pointer[state]
}

func NewTokenBucket(
	capacity float64,
	fillRate float64,
	consumeRate float64,
) *tokenBucket {
	if consumeRate <= 0 || consumeRate > capacity {
		panic("consume rate must be > 0 and <= capacity")
	}

	if fillRate <= 0 {
		panic("fill rate must be > 0")
	}

	t := &tokenBucket{
		capacity:    capacity,
		fillRate:    fillRate,
		consumeRate: consumeRate,
	}
	t.state.Store(&state{
		currToken: capacity,
		lastTime:  0,
	})
	return t
}

func (t *tokenBucket) AllowAt(arriveAt time.Time) bool {
	now := arriveAt.UnixMilli()

	for {
		lastState := t.state.Load()
		elapsed := now - lastState.lastTime
		newState := &state{
			lastTime: now,
		}

		if elapsed < 0 {
			fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
			return false
		}

		// Always allow the first request
		if lastState.lastTime == 0 {
			newState.currToken = t.capacity - t.consumeRate
			if t.state.CompareAndSwap(lastState, newState) {
				return true
			}
			// Retry if CAS fails
			continue
		}

		newState.currToken = math.Min(
			t.capacity,
			lastState.currToken+t.fillRate*float64(elapsed),
		)
		fmt.Println(newState.currToken)
		if newState.currToken >= t.consumeRate {
			newState.currToken -= t.consumeRate
			if t.state.CompareAndSwap(lastState, newState) {
				return true
			}
			// Retry if CAS fails
			continue
		}

		return false
	}
}

func (t *tokenBucket) Allow() bool {
	return t.AllowAt(time.Now())
}
