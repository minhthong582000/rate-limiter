package tokenbucket

import (
	"fmt"
	"math"
	"time"
)

type tokenBucket struct {
	Capacity    float64 // Max burst
	FillRate    float64 // Token fill rate per second
	ConsumeRate float64 // Token consume rate per request
	CurrToken   float64
	LastTime    time.Time
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
		Capacity:    capacity,
		FillRate:    fillRate,
		ConsumeRate: consumeRate,
		CurrToken:   capacity,
	}
}

func (t *tokenBucket) AllowAt(arriveAt time.Time) bool {
	if t.LastTime.IsZero() {
		t.LastTime = arriveAt
	}

	elapsed := arriveAt.Sub(t.LastTime).Seconds()
	// Invalid time or possible clock skew
	if elapsed < 0 {
		fmt.Println("Warning: Negative elapsed time detected. Possible clock skew.")
		return false
	}

	t.CurrToken = math.Min(t.Capacity, t.CurrToken+math.Round(t.FillRate*elapsed*1000)/1000)

	// How many hours has elapsed since the last time
	fmt.Printf("Elapsed: %.3fh, Fill rate: %f, Curr token: %f\n", arriveAt.Sub(t.LastTime).Hours(), t.FillRate, t.CurrToken)

	t.LastTime = arriveAt

	if t.CurrToken >= t.ConsumeRate {
		t.CurrToken -= t.ConsumeRate
		return true
	}

	return false
}

func (t *tokenBucket) Allow() bool {
	return t.AllowAt(time.Now())
}
