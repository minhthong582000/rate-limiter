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
	limiter := NewFixedSizeWindow(5, time.Second)

	// Within the same window, allow up to 5 requests.
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(), "Request %d should be allowed", i+1)
	}

	// 6th request should be denied
	assert.False(t, limiter.Allow(), "6th request should be denied within the same window")

	// Wait for the window to reset
	time.Sleep(time.Second + 10*time.Millisecond)

	// New window, requests should be allowed again
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(), "Request %d in new window should be allowed", i+1)
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
