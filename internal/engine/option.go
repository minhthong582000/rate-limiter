package engine

import "time"

type Config struct {
	EngineType EngineType
	Capacity   uint64
	StopCh     <-chan struct{}

	// Token bucket specific configuration
	FillRate    float64
	ConsumeRate float64

	// Leaky bucket specific configuration
	LeakRate time.Duration

	// Fixed size window and sliding window specific configuration
	windowSize time.Duration
}

type Option func(f *Config)

func WithEngineType(engineType EngineType) Option {
	return func(f *Config) {
		f.EngineType = engineType
	}
}

func WithCapacity(capacity uint64) Option {
	return func(f *Config) {
		f.Capacity = capacity
	}
}

func WithFillRate(fillRate float64) Option {
	return func(f *Config) {
		f.FillRate = fillRate
	}
}

func WithConsumeRate(consumeRate float64) Option {
	return func(f *Config) {
		f.ConsumeRate = consumeRate
	}
}

func WithLeakRate(leakRate time.Duration) Option {
	return func(f *Config) {
		f.LeakRate = leakRate
	}
}

func WithStopCh(stopCh <-chan struct{}) Option {
	return func(f *Config) {
		f.StopCh = stopCh
	}
}

func WithWindowSize(windowSize time.Duration) Option {
	return func(f *Config) {
		f.windowSize = windowSize
	}
}
