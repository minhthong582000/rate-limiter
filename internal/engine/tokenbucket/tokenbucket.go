package tokenbucket

import (
	"math"
	"sync/atomic"
	"time"
)

type state struct {
	currToken float64
	lastTime  time.Time
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
		lastTime:  time.Now(),
	})
	return t
}

func (t *tokenBucket) AllowAt(arriveAt time.Time) bool {
	for {
		lastState := t.state.Load()
		elapsed := arriveAt.Sub(lastState.lastTime).Milliseconds()
		newState := &state{
			lastTime: arriveAt,
		}

		if elapsed < 0 {
			// A lot of contention results in lots of CAS retries.
			// This might causes the lastState.lastTime to be in the future of arriveAt.
			return false
		}

		newState.currToken = math.Min(
			t.capacity,
			lastState.currToken+t.fillRate*float64(elapsed),
		)
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
