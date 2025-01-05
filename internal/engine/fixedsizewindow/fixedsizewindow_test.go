package fixedsizewindow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewFixedSizeWindow tests the fixed-size window rate limiter constructor.
func TestNewFixedSizeWindow(t *testing.T) {
	limiter := NewFixedSizeWindow(3, 1000) // capacity=3, windowSize=1s

	assert.Equal(t, uint64(3), limiter.capacity, "Capacity should be 3")
	assert.Equal(t, int64(1000), limiter.windowSize, "Window size should be 1000ms")

	state := limiter.state.Load()
	assert.Equal(t, uint64(0), state.currCount, "Initial count should be 0")
	assert.WithinDuration(t, time.Now(), state.lastTime, time.Second, "Initial last time should be close to current time")
}

// TestFixedSizeWindow_Basic tests basic behavior of the fixed-size window rate limiter.
func TestFixedSizeWindow_Basic(t *testing.T) {
	limiter := NewFixedSizeWindow(3, 1000) // capacity=3, windowSize=1s

	// Set the initial lastTime to a value before the test cases below.
	// By default, lastTime is set to time.Now() causing the tests to return false.
	limiter.state.Store(&state{currCount: 0, lastTime: time.Unix(0, 0).UTC()})

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

// TestFixedSizeWindow_RequestAtBoundary tests the rate limiter behavior at the boundary of the window.
func TestFixedSizeWindow_RequestAtBoundary(t *testing.T) {
	limiter := NewFixedSizeWindow(3, 10000) // capacity=3, windowSize=10s

	// Set the initial lastTime to a value before the test cases below.
	// By default, lastTime is set to time.Now() causing the tests to return false.
	limiter.state.Store(&state{currCount: 0, lastTime: time.Unix(0, 0).UTC()})

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

// TestFixedSizeWindow_ConcurrentAccess tests concurrent access to the rate limiter.
func TestFixedSizeWindow_ConcurrentAccess(t *testing.T) {
	limiter := NewFixedSizeWindow(10, 1000)
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

// TestFixedSizeWindow_NegativeElapsedTime ensures that we handle negative elapsed time correctly.
func TestFixedSizeWindow_NegativeElapsedTime(t *testing.T) {
	limiter := NewFixedSizeWindow(5, 1000)

	// Set the initial lastTime to a value before the test cases below.
	// By default, lastTime is set to time.Now() causing the tests to return false.
	limiter.state.Store(&state{currCount: 0, lastTime: time.Unix(0, 0).UTC()})

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

// TestFixedSizeWindow_ZeroWindow ensures behavior when window time is zero (edge case).
func TestFixedSizeWindow_ZeroWindow(t *testing.T) {
	assert.Panics(t, func() {
		NewFixedSizeWindow(5, 0)
	}, "Creating a rate limiter with zero window size should panic")
}
