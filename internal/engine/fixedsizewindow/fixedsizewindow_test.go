package fixedsizewindow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBasicRateLimiting tests basic behavior of the fixed-size window rate limiter.
func TestBasicRateLimiting(t *testing.T) {
	limiter := NewFixedSizeWindow(3, time.Second)

	requests := []string{
		// 4 requests at the same time
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",

		// 1 request after 2 second
		"2025-01-01T00:00:02Z",
	}

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		ts, _ := time.Parse(time.RFC3339, requests[i])
		assert.True(t, limiter.AllowAt(ts), "Request %d should be allowed", i+1)
	}

	// 4th request should be denied
	ts, _ := time.Parse(time.RFC3339, requests[3])
	assert.False(t, limiter.AllowAt(ts), "Request 3 should be denied")

	// 5th requests should be allowed after the window resets
	ts, _ = time.Parse(time.RFC3339, requests[4])
	assert.True(t, limiter.AllowAt(ts), "Request after window reset should be allowed")
}

// TestRequestAtBoundary tests the rate limiter behavior at the boundary of the window.
func TestRequestAtBoundary(t *testing.T) {
	limiter := NewFixedSizeWindow(3, 10*time.Second)

	// Requests timestamps within 11 seconds window
	requests := []string{
		"2025-01-01T00:00:00Z",

		// 2 requests after 9 seconds
		"2025-01-01T00:00:09Z",
		"2025-01-01T00:00:09Z",

		// 3 new requests after first request 10 seconds
		"2025-01-01T00:00:11Z",
		"2025-01-01T00:00:11Z",
		"2025-01-01T00:00:11Z",
	}

	// All requests should be allowed
	for i := 0; i < 6; i++ {
		ts, _ := time.Parse(time.RFC3339, requests[i])
		assert.True(t, limiter.AllowAt(ts), "Request %d should be allowed", i+1)
	}
}

// TestConcurrentAccess tests concurrent access to the rate limiter.
func TestConcurrentAccess(t *testing.T) {
	limiter := NewFixedSizeWindow(10, time.Second)
	var wg sync.WaitGroup

	successCount := atomic.Uint64{}
	failCount := atomic.Uint64{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow() {
				successCount.Add(1)
			} else {
				failCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// Only 10 requests should succeed within the window
	assert.Equal(t, uint64(10), successCount.Load(), "Only 10 requests should succeed in the window")
	assert.Equal(t, uint64(10), failCount.Load(), "Remaining 10 requests should fail")
}

// TestNegativeElapsedTime ensures no behavior breaks with a negative elapsed time.
func TestNegativeElapsedTime(t *testing.T) {
	limiter := NewFixedSizeWindow(5, time.Second)

	// Simulate a scenario where arriveAt is earlier than lastTime
	limiter.Allow()                                                     // First request
	limiter.lastTime.Store(time.Now().Add(10 * time.Second).UnixNano()) // Future time

	assert.False(t, limiter.AllowAt(time.Now()), "Request with negative elapsed time should fail")
}

// TestZeroWindow ensures behavior when window time is zero (edge case).
func TestZeroWindow(t *testing.T) {
	assert.Panics(t, func() {
		NewFixedSizeWindow(5, 0)
	}, "Creating a rate limiter with zero window size should panic")
}
