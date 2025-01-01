package engine

import (
	"fmt"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine/fixedsizewindow"
	"github.com/minhthong582000/rate-limiter/internal/engine/leakybucket"
	"github.com/minhthong582000/rate-limiter/internal/engine/slidingwindow"
	"github.com/minhthong582000/rate-limiter/internal/engine/tokenbucket"
)

type EngineType string

const (
	FixedWindow          EngineType = "fixed-window"
	SlidingWindowLog     EngineType = "sliding-window-log"
	SlidingWindowCounter EngineType = "sliding-window-counter"
	TokenBucket          EngineType = "token-bucket"
	LeakyBucket          EngineType = "leaky-bucket"
)

func StringToEngineType(s string) EngineType {
	switch s {
	case "fixed-window":
		return FixedWindow
	case "sliding-window-log":
		return SlidingWindowLog
	case "sliding-window-counter":
		return SlidingWindowCounter
	case "token-bucket":
		return TokenBucket
	case "leaky-bucket":
		return LeakyBucket
	default:
		return ""
	}
}

type Engine interface {
	// Allow checks if a request is allowed to be processed now
	Allow() bool
	// AllowAt checks if a request is allowed to be processed at the given time
	AllowAt(arriveAt time.Time) bool
}

func EngineFactory(opts ...Option) (Engine, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	var engine Engine
	switch config.EngineType {
	case FixedWindow:
		engine = fixedsizewindow.NewFixedSizeWindow(
			config.Capacity,
			config.windowSize,
		)
	case SlidingWindowLog:
		engine = slidingwindow.NewSlidingWindowLogs(
			config.Capacity,
			config.windowSize,
		)
	case SlidingWindowCounter:
		return nil, nil
	case TokenBucket:
		engine = tokenbucket.NewTokenBucket(
			float64(config.Capacity),
			config.FillRate,
			config.ConsumeRate,
		)
	case LeakyBucket:
		engine = leakybucket.NewLeakyBucket(
			config.Capacity,
			config.LeakRate,
			config.StopCh,
		)
	default:
		return nil, fmt.Errorf("invalid rate-limiter engine type")
	}

	return engine, nil
}
