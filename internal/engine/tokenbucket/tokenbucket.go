package tokenbucket

import (
	"fmt"
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
	// Invalid time
	if elapsed < 0 {
		return false
	}

	t.CurrToken = min(t.Capacity, t.CurrToken+t.FillRate*elapsed)
	t.LastTime = arriveAt

	fmt.Printf("Elapsed: %f, Fill rate: %f, Curr token: %f\n", elapsed, t.FillRate, t.CurrToken)

	if t.CurrToken >= t.ConsumeRate {
		t.CurrToken -= t.ConsumeRate
		return true
	}

	return false
}

func (t *tokenBucket) Allow() bool {
	return t.AllowAt(time.Now())
}
