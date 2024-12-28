package engine

import (
	"fmt"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine/leakybucket"
	"github.com/minhthong582000/rate-limiter/internal/engine/tokenbucket"
)

type EngineType string

const (
	FixedWindow   EngineType = "fixed-window"
	SlidingWindow EngineType = "sliding-window"
	TokenBucket   EngineType = "token-bucket"
	LeakyBucket   EngineType = "leaky-bucket"
)

func StringToEngineType(s string) EngineType {
	switch s {
	case "fixed-window":
		return FixedWindow
	case "sliding-window":
		return SlidingWindow
	case "token-bucket":
		return TokenBucket
	case "leaky-bucket":
		return LeakyBucket
	default:
		return ""
	}
}

type Engine interface {
	Allow() bool
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
		return nil, nil
	case SlidingWindow:
		return nil, nil
	case TokenBucket:
		engine = tokenbucket.NewTokenBucket(
			config.Capacity,
			config.FillRate,
			config.ConsumeRate,
		)
	case LeakyBucket:
		engine = leakybucket.NewLeakyBucket(
			uint64(config.Capacity),
			config.LeakRate,
			config.StopCh,
		)
	default:
		return nil, fmt.Errorf("invalid rate-limiter engine type")
	}

	return engine, nil
}
