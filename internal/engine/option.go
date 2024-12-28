package engine

import "time"

type Config struct {
	EngineType EngineType
	Capacity   float64
	StopCh     <-chan struct{}

	// Token bucket specific configuration
	FillRate    float64
	ConsumeRate float64

	// Leaky bucket specific configuration
	LeakRate time.Duration
}

type Option func(f *Config)

func WithEngineType(engineType EngineType) Option {
	return func(f *Config) {
		f.EngineType = engineType
	}
}

func WithCapacity(capacity float64) Option {
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
