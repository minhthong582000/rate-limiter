package tokenbucket

import (
	"fmt"
	"math"
	"time"
)

type tokenBucket struct {
	capacity    float64 // Max burst
	fillRate    float64 // Token fill rate per second
	consumeRate float64 // Token consume rate per request
	currToken   float64
	lastTime    time.Time
}

func NewTokenBucket(
	capacity float64,
	fillRate float64,
	consumeRate float64,
) *tokenBucket {
	if consumeRate <= 0 {
		consumeRate = 1
	}

	return &tokenBucket{
		capacity:    capacity,
		fillRate:    fillRate,
		consumeRate: consumeRate,
		currToken:   capacity,
	}
}

func (t *tokenBucket) AllowAt(arriveAt time.Time) bool {
	if t.lastTime.IsZero() {
		t.lastTime = arriveAt
	}

	elapsed := arriveAt.Sub(t.lastTime).Seconds()
	// Invalid time or possible clock skew
	if elapsed < 0 {
		fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
		return false
	}

	t.currToken = math.Min(t.capacity, t.currToken+math.Round(t.fillRate*elapsed*1000)/1000)

	// How many hours has elapsed since the last time
	fmt.Printf("Elapsed: %.3fh, Fill rate: %f, Curr token: %f\n", arriveAt.Sub(t.lastTime).Hours(), t.fillRate, t.currToken)

	t.lastTime = arriveAt

	if t.currToken >= t.consumeRate {
		t.currToken -= t.consumeRate
		return true
	}

	return false
}

func (t *tokenBucket) Allow() bool {
	return t.AllowAt(time.Now())
}
