package tokenbucket

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewTokenBucket tests token bucket constructor.
func TestNewTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(5, 1, 1) // capacity=5, fillRate=1/ms, consumeRate=1

	assert.NotNil(t, bucket)
	assert.Equal(t, float64(5), bucket.capacity)
	assert.Equal(t, float64(1), bucket.fillRate)
	assert.Equal(t, float64(1), bucket.consumeRate)

	state := bucket.state.Load()
	assert.Equal(t, float64(5), state.currToken)
	assert.Equal(t, int64(0), state.lastTime)
}

// TestTokenBucket_Basic validates basic token bucket behavior.
func TestTokenBucket_Basic(t *testing.T) {
	bucket := NewTokenBucket(3, 1, 1) // capacity=3, fillRate=1/ms, consumeRate=1

	requests := []string{
		// 4 requests at the same time
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",

		// 1 request after 2 milliseconds
		"2025-01-01T00:00:00.002Z",
	}

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		ts, _ := time.Parse(time.RFC3339Nano, requests[i])
		assert.True(t, bucket.AllowAt(ts), "Request %d should be allowed", i+1)
	}

	// 4th request should be denied (tokens exhausted)
	ts, _ := time.Parse(time.RFC3339Nano, requests[3])
	assert.False(t, bucket.AllowAt(ts), "Request 4 should be denied due to empty tokens")

	// 5th request should be allowed after refill (2ms later, 2 tokens refilled)
	ts, _ = time.Parse(time.RFC3339Nano, requests[4])
	assert.True(t, bucket.AllowAt(ts), "Request 5 should be allowed after token refill")
}

// TestTokenBucket_ConcurrentAccess verifies thread safety of the token bucket.
func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	bucket := NewTokenBucket(10, 1.0/1000, 1) // capacity=10, fillRate=1/s, consumeRate=1

	var wg sync.WaitGroup
	successCount := atomic.Uint64{}
	failCount := atomic.Uint64{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.Allow() {
				successCount.Add(1)
			} else {
				failCount.Add(1)
			}
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, successCount.Load(), uint64(10), "Allowed requests should not exceed capacity")
	assert.GreaterOrEqual(t, failCount.Load(), uint64(10), "Remaining requests should fail")
	t.Logf("Allowed: %d, Rejected: %d", successCount.Load(), failCount.Load())
}

// TestTokenBucket_NegativeElapsedTime ensures that negative elapsed time is handled safely.
func TestTokenBucket_NegativeElapsedTime(t *testing.T) {
	bucket := NewTokenBucket(5, 1, 1) // capacity=5, fillRate=1/ms, consumeRate=1

	requests := []string{
		"2025-01-02T00:00:00Z",
		"2025-01-01T23:59:59", // Earlier timestamp (clock skew)
	}

	// First request should be allowed
	ts, _ := time.Parse(time.RFC3339, requests[0])
	assert.True(t, bucket.AllowAt(ts), "First request should be allowed")

	// Second request with earlier timestamp should fail
	ts, _ = time.Parse(time.RFC3339, requests[1])
	assert.False(t, bucket.AllowAt(ts), "Request with negative elapsed time should fail")
}

// TestTokenBucket_ZeroFillRate checks edge case with zero fill rate.
func TestTokenBucket_ZeroFillRate(t *testing.T) {
	assert.Panics(t, func() {
		NewTokenBucket(5, 0, 1)
	}, "Creating a token bucket with zero fill rate should panic")
}

// TestTokenBucket_ZeroConsumeRate checks edge case with zero consume rate.
func TestTokenBucket_ZeroConsumeRate(t *testing.T) {
	assert.Panics(t, func() {
		NewTokenBucket(5, 1, 0)
	}, "Creating a token bucket with zero consume rate should panic")
}

// TestTokenBucket_ConsumeRateExceedsCapacity checks edge case with invalid consume rate.
func TestTokenBucket_ConsumeRateExceedsCapacity(t *testing.T) {
	assert.Panics(t, func() {
		NewTokenBucket(5, 1, 10)
	}, "Creating a token bucket with consume rate exceeding capacity should panic")
}
