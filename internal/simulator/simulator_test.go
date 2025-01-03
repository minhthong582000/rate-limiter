package simulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/minhthong582000/rate-limiter/internal/engine/mocks"
)

// TestSimulator_Run tests the normal operation of the simulator
func TestSimulator_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name          string
		numWorker     int64
		numRequests   int64
		setupMock     func() *mocks.MockEngine
		expectedPanic bool
	}{
		{
			name:        "Normal operation with engine allowing all requests",
			numWorker:   2,
			numRequests: 5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Return(true).Times(5)
				return mockEngine
			},
		},
		{
			name:        "Normal operation with engine denying all requests",
			numWorker:   2,
			numRequests: 5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Return(false).Times(5)
				return mockEngine
			},
		},
		{
			name:        "Single Worker",
			numWorker:   1,
			numRequests: 5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Return(true).Times(5)
				return mockEngine
			},
		},
		{
			name:        "Single Request with Single Worker",
			numWorker:   1,
			numRequests: 1,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Return(true).Times(1)
				return mockEngine
			},
		},
		{
			name:        "Single Request with Multiple Workers",
			numWorker:   5,
			numRequests: 1,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Return(true).Times(1)
				return mockEngine
			},
		},
		{
			name:        "Empty Request",
			numWorker:   2,
			numRequests: 0,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Times(0)
				return mockEngine
			},
		},
		{
			name:        "Empty Worker",
			numWorker:   0,
			numRequests: 5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Times(0)
				return mockEngine
			},
		},
		{
			name:        "Negative Worker",
			numWorker:   -1,
			numRequests: 5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Times(0)
				return mockEngine
			},
		},
		{
			name:        "Negative Request",
			numWorker:   2,
			numRequests: -5,
			setupMock: func() *mocks.MockEngine {
				mockEngine := mocks.NewMockEngine(ctrl)
				mockEngine.EXPECT().AllowAt(gomock.Any()).Times(0)
				return mockEngine
			},
			expectedPanic: true, // Channel with negative capacity
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEngine := tc.setupMock()

			stopCh := make(chan struct{})
			defer close(stopCh)

			sim := NewSimulator(
				WithRateLimiter(mockEngine),
				WithNumWorker(tc.numWorker),
				WithNumRequests(tc.numRequests),
				WithWaitTime(10),
				WithJitter(5),
				WithStopChannel(stopCh),
			)

			// Run simulator
			if tc.expectedPanic {
				assert.Panics(t, func() {
					sim.Run()
				}, "Expected panic")
				return
			}

			sim.Run()
		})
	}
}

// TestSimulator_Shutdown tests the behavior when the stopCh is closed before the simulator finishes
func TestSimulator_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEngine := mocks.NewMockEngine(ctrl)
	mockEngine.EXPECT().AllowAt(gomock.Any()).Return(true).AnyTimes()

	stopCh := make(chan struct{})

	sim := NewSimulator(
		WithRateLimiter(mockEngine),
		WithNumWorker(2),
		WithNumRequests(5),
		WithWaitTime(10),
		WithJitter(5),
		WithStopChannel(stopCh),
	)
	// Run the simulator in a separate goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		sim.Run()
	}()

	// Trigger stop
	close(stopCh)

	select {
	case <-done:
		// Simulator has stopped
	case <-time.After(1 * time.Second):
		t.Fatal("Test timeout")
	}
}

// TestSimulator_EarlyStop tests if the simulator handles an early stop correctly
func TestSimulator_EarlyStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEngine := mocks.NewMockEngine(ctrl)
	mockEngine.EXPECT().AllowAt(gomock.Any()).Times(0)

	stopCh := make(chan struct{})
	close(stopCh) // Stop immediately

	sim := NewSimulator(
		WithRateLimiter(mockEngine),
		WithNumWorker(2),
		WithNumRequests(5),
		WithWaitTime(10),
		WithJitter(5),
		WithStopChannel(stopCh),
	)

	// Run the simulator
	sim.Run()
}
