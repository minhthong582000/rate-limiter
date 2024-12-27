package engine

type Config struct {
	EngineType  EngineType
	Capacity    float64
	FillRate    float64
	ConsumeRate float64
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
