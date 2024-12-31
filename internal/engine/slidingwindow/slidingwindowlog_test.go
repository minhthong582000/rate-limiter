package slidingwindow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSlidingWindowLogs tests the behavior of the sliding window logs
func TestSlidingWindowLogs(t *testing.T) {
	limiter := NewSlidingWindowLogs(3, 1*time.Second) // 3 requests per 1 second

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
	assert.False(t, limiter.AllowAt(ts), "Request 4 should be denied")

	// 5th requests should be allowed after the window resets
	ts, _ = time.Parse(time.RFC3339, requests[4])
	assert.True(t, limiter.AllowAt(ts), "Request 5 (after window reset) should be allowed")
}

// TestRequestAtBoundary tests the rate limiter behavior at the boundary of the window.
func TestRequestAtBoundary(t *testing.T) {
	limiter := NewSlidingWindowLogs(3, 10*time.Second)

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

	// First 4 requests should be allowed
	for i := 0; i < 4; i++ {
		ts, _ := time.Parse(time.RFC3339, requests[i])
		assert.True(t, limiter.AllowAt(ts), "Request %d should be allowed", i+1)
	}

	// 5th request should be denied
	ts, _ := time.Parse(time.RFC3339, requests[4])
	assert.False(t, limiter.AllowAt(ts), "Request 5 should be denied")

	// 6th request should be denied as well
	ts, _ = time.Parse(time.RFC3339, requests[5])
	assert.False(t, limiter.AllowAt(ts), "Request 6 should be denied")
}

// TestSlidingWindowLogs_Concurrency tests thread safety under concurrent access
func TestSlidingWindowLogs_Concurrency(t *testing.T) {
	limiter := NewSlidingWindowLogs(10, 10*time.Second) // 10 requests per second

	var wg sync.WaitGroup
	var allowedRequests atomic.Int32

	// Simulate 20 concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.AllowAt(time.Now()) {
				allowedRequests.Add(1)
			}
		}()
	}

	wg.Wait()

	// Assert that no more than 10 requests were allowed
	assert.LessOrEqual(t, allowedRequests.Load(), int32(10), "No more than 10 requests should be allowed within 10 second window")
}

// TestSlidingWindowLogs_ZeroWindowSize ensures window size validation
func TestSlidingWindowLogs_ZeroWindowSize(t *testing.T) {
	assert.Panics(t, func() {
		NewSlidingWindowLogs(5, 0)
	}, "Creating a sliding window with zero window size should panic")
}
