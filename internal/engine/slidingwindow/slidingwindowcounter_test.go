package slidingwindow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewSlidingWindowCounter tests the sliding window counter constructor
func TestNewSlidingWindowCounter(t *testing.T) {
	limiter := NewSlidingWindowCounter(3, 1000) // capacity=3, windowSize=1000ms

	assert.Equal(t, float64(3), limiter.capacity, "Capacity should be 3")
	assert.Equal(t, int64(1000), limiter.windowSize, "Window size should be 1000ms")
	assert.WithinDuration(t, time.Now(), limiter.startTime, time.Second, "Initial start time should be close to current time")

	state := limiter.state.Load()
	assert.Equal(t, float64(0), state.currCount, "Initial count should be 0")
	assert.Equal(t, int64(0), state.currWindow, "Initial window should be 0")
	assert.Equal(t, float64(0), state.prevCount, "Initial previous count should be 0")
}

// TestSlidingWindowCounter_Basic tests the basic behavior of the sliding window counter
func TestSlidingWindowCounter_Basic(t *testing.T) {
	limiter := NewSlidingWindowCounter(3, 10000) // capacity=10, windowSize=10s

	// Set the initial startTime to a value before the test cases below.
	// By default, startTime is set to time.Now() causing the tests to return false.
	limiter.startTime = time.Unix(0, 0).UTC()

	// Comments next to each request follow this formula: prevWeight*prevCount + currCount
	requests := []string{
		// 4 requests at the same time
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z", // Denied: cnt = 4 > capacity

		// 4 requests after 10 second
		"2025-01-01T00:00:10Z", // Denied: cnt = 1.0*3 + 0 = 3 == capacity
		"2025-01-01T00:00:11Z", // Allowed: cnt = 0.9*3 + 0 = 2.7 < capacity
		"2025-01-01T00:00:12Z", // Denied: cnt = 0.8*3 + 1 = 3.4 > capacity
		"2025-01-01T00:00:17Z", // Allowed: cnt = 0.3*3 + 1 = 1.9 < capacity
		"2025-01-01T00:00:19Z", // Allowed: cnt = 0.1*3 + 2 = 2.3 < capacity

		// 1 request in a completely new window
		"2025-01-01T00:00:30Z", // Allowed: cnt = 1.0*0 + 0 = 0 < capacity
	}

	for i := 0; i < len(requests); i++ {
		ts, err := time.Parse(time.RFC3339, requests[i])
		assert.NoError(t, err)

		if i == 3 || i == 4 || i == 6 {
			assert.False(t, limiter.AllowAt(ts), "Request %d should be denied", i+1)
		} else {
			assert.True(t, limiter.AllowAt(ts), "Request %d should be allowed", i+1)
		}
	}
}

// TestSlidingWindowCounter_RequestAtBoundary tests the rate limiter behavior at the boundary of the window.
func TestSlidingWindowCounter_RequestAtBoundary(t *testing.T) {
	limiter := NewSlidingWindowCounter(3, 10000) // capacity=3, windowSize=10s

	// Set the initial startTime to a value before the test cases below.
	// By default, startTime is set to time.Now() causing the tests to return false.
	limiter.startTime = time.Unix(0, 0).UTC()

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

// TestSlidingWindowCounter_ConcurrentAccess tests thread safety under concurrent access
func TestSlidingWindowCounter_ConcurrentAccess(t *testing.T) {
	limiter := NewSlidingWindowCounter(10, 10000) // capacity=3, windowSize=10s

	var wg sync.WaitGroup
	var allowedRequests atomic.Int32
	var deniedRequests atomic.Int32

	// Simulate 20 concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.AllowAt(time.Now()) {
				allowedRequests.Add(1)
			} else {
				deniedRequests.Add(1)
			}
		}()
	}

	wg.Wait()

	// Assert that no more than 10 requests were allowed
	assert.LessOrEqual(t, allowedRequests.Load(), int32(10), "No more than 10 requests should be allowed within 10 second window")
	assert.LessOrEqual(t, deniedRequests.Load(), int32(10), "No more than 10 requests should be denied within 10 second window")
}

// TestSlidingWindowCounter_ZeroWindowSize ensures window size validation
func TestSlidingWindowCounter_ZeroWindowSize(t *testing.T) {
	assert.Panics(t, func() {
		NewSlidingWindowCounter(5, 0)
	}, "Creating a sliding window with zero window size should panic")
}

// TestSlidingWindowCounter_NegativeElapsedTime ensures that negative elapsed time is handled safely.
func TestSlidingWindowCounter_NegativeElapsedTime(t *testing.T) {
	limiter := NewSlidingWindowCounter(5, 1000) // capacity=5, windowSize=1s

	// Set the initial startTime to a value before the test cases below.
	// By default, startTime is set to time.Now() causing the tests to return false.
	limiter.startTime = time.Unix(0, 0).UTC()

	requests := []string{
		"2025-01-02T00:00:00Z",

		// Previous day
		"2025-01-01T00:00:00Z",
	}

	ts, _ := time.Parse(time.RFC3339, requests[0])
	limiter.AllowAt(ts)

	// Simulate a scenario where arriveAt is earlier than lastTime
	ts, _ = time.Parse(time.RFC3339, requests[1])
	assert.False(t, limiter.AllowAt(ts), "Request with negative elapsed time should fail")
}
