package simulator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine"
)

type Simulator struct {
	ratelimiter engine.Engine
	numWorker   int64
	numRequests int64
	waitTime    int64
	jitter      int64
	stopCh      <-chan struct{}
}

func NewSimulator(opts ...Option) *Simulator {
	s := &Simulator{}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Simulator) worker(id int64, reqCh <-chan int64) {
	for {
		select {
		case <-s.stopCh:
			return
		case req, ok := <-reqCh:
			if !ok {
				return
			}

			rj := int64(0)
			if s.jitter > 0 {
				randomJitter, _ := rand.Int(rand.Reader, big.NewInt(s.jitter*2))
				rj = randomJitter.Int64() - s.jitter
			}

			now := time.Now()
			if s.ratelimiter.AllowAt(now) {
				fmt.Printf("Request %d.%d ALLOWED, ts=\"%v\"\n", id, req, now)
			} else {
				fmt.Printf("Request %d.%d DENIED, ts=\"%v\"\n", id, req, now)
			}
			time.Sleep(time.Duration(s.waitTime+rj) * time.Millisecond)
		}
	}
}

func (s *Simulator) Run() {
	var wg sync.WaitGroup
	requestCh := make(chan int64, s.numRequests)

	// Start workers
	for i := int64(0); i < s.numWorker; i++ {
		wg.Add(1)

		go func(id int64) {
			defer func() {
				wg.Done()
				fmt.Printf("Worker %d stopped\n", id)
			}()

			fmt.Printf("Worker %d started\n", id)
			s.worker(id, requestCh)
		}(i + 1)
	}

	// Send requests to workers
	for i := int64(0); i < s.numRequests; i++ {
		select {
		// Stop signal received while sending requests
		case <-s.stopCh:
			close(requestCh)
			wg.Wait()
			return
		default:
			requestCh <- i
		}
	}

	close(requestCh)
	wg.Wait()
}
