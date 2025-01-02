package leakybucket

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLeakyBucket_New ensures the constructor initializes the leaky bucket correctly.
func TestLeakyBucket_New(t *testing.T) {
	stopCh := make(chan struct{})
	limiter := NewLeakyBucket(5, time.Second, stopCh)

	assert.Equal(t, uint64(5), limiter.capacity, "Capacity should be 5")
	assert.Equal(t, time.Second, limiter.drainRate, "Drain rate should be 1 second")
	assert.NotNil(t, limiter.queue, "Queue should not be nil")
}

// TestLeakyBucket_Basic tests basic behavior of the leaky bucket rate limiter.
func TestLeakyBucket_Basic(t *testing.T) {
	stopCh := make(chan struct{})
	limiter := NewLeakyBucket(3, time.Second, stopCh)

	requests := []string{
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z", // This one should be denied

		// After 2 seconds, space should be available
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

	// After enough time has passed, 5th request should be allowed
	ts, _ = time.Parse(time.RFC3339, requests[4])
	time.Sleep(2 * time.Second) // Ensure the leak happens
	assert.True(t, limiter.AllowAt(ts), "Request after time window should be allowed")

	close(stopCh)
}

// TestLeakyBucket_ConcurrentAccess tests concurrent access to the leaky bucket.
func TestLeakyBucket_ConcurrentAccess(t *testing.T) {
	stopCh := make(chan struct{})
	limiter := NewLeakyBucket(10, 100*time.Millisecond, stopCh)
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

	// Only up to capacity requests should succeed instantly
	assert.LessOrEqual(t, successCount.Load(), uint64(10), "No more than 10 requests should succeed immediately")
	assert.GreaterOrEqual(t, failCount.Load(), uint64(10), "At least 10 requests should fail immediately")

	close(stopCh)
}

// TestLeakyBucket_ZeroCapacity ensures behavior when capacity is zero.
func TestLeakyBucket_ZeroCapacity(t *testing.T) {
	assert.Panics(t, func() {
		NewLeakyBucket(0, time.Second, make(chan struct{}))
	}, "Creating a leaky bucket with zero capacity should panic")
}

// TestLeakyBucket_ZeroDrainRate ensures behavior when drain rate is zero.
func TestLeakyBucket_ZeroDrainRate(t *testing.T) {
	assert.Panics(t, func() {
		NewLeakyBucket(5, 0, make(chan struct{}))
	}, "Creating a leaky bucket with zero drain rate should panic")
}
