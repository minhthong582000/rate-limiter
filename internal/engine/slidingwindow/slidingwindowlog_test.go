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

	// First 3 requests should be allowed
	assert.True(t, limiter.AllowAt(time.Now()), "First request should be allowed")
	assert.True(t, limiter.AllowAt(time.Now()), "Second request should be allowed")
	assert.True(t, limiter.AllowAt(time.Now()), "Third request should be allowed")

	// 4th request should be denied
	assert.False(t, limiter.AllowAt(time.Now()), "Fourth request should be denied")

	// Wait for the window to expire
	time.Sleep(1500 * time.Millisecond)

	// New requests should be allowed after the window resets
	assert.True(t, limiter.AllowAt(time.Now()), "Request after window reset should be allowed")
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
