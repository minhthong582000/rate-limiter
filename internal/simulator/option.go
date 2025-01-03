package simulator

import "github.com/minhthong582000/rate-limiter/internal/engine"

type Option func(*Simulator)

func WithRateLimiter(rateLimiter engine.Engine) Option {
	return func(s *Simulator) {
		s.ratelimiter = rateLimiter
	}
}

func WithNumWorker(numWorker int64) Option {
	return func(s *Simulator) {
		s.numWorker = numWorker
	}
}

func WithNumRequests(numRequests int64) Option {
	return func(s *Simulator) {
		s.numRequests = numRequests
	}
}

func WithWaitTime(waitTime int64) Option {
	return func(s *Simulator) {
		s.waitTime = waitTime
	}
}

func WithJitter(jitter int64) Option {
	return func(s *Simulator) {
		s.jitter = jitter
	}
}

func WithStopChannel(stopCh <-chan struct{}) Option {
	return func(s *Simulator) {
		s.stopCh = stopCh
	}
}
