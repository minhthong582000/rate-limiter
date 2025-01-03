package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine"
	"github.com/minhthong582000/soa-404/pkg/signals"
	"github.com/spf13/cobra"
)

var (
	engineType string
	capacity   int64

	// Token bucket specific configuration
	fillDuration float64 // in milliseconds
	consumeRate  float64

	// Leaky bucket specific configuration
	drainDuration int64 // in milliseconds

	// Fixed size window and sliding window specific configuration
	windowSize int64 // in milliseconds

	// Simulation parameters
	numRequests int64
	waitTime    int64 // in milliseconds
	jitter      int64 // in milliseconds
	parallel    int8  // number of parallel workers
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the rate limiter simulator",
	Long: `A command to run the rate limiter engine based on the selected engine type.
You can choose between different rate limiting engines such as fixed-window, sliding-window, token-bucket, and leaky-bucket.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if capacity <= 0 {
			return fmt.Errorf("capacity must be greater than 0")
		}

		if fillDuration <= 0 {
			return fmt.Errorf("fill duration must be greater than 0")
		}

		if consumeRate <= 0 {
			return fmt.Errorf("consume rate must be greater than 0")
		}

		if drainDuration <= 0 {
			return fmt.Errorf("drain duration must be greater than 0")
		}

		if windowSize <= 0 {
			return fmt.Errorf("window size must be greater than 0")
		}

		if numRequests <= 0 {
			return fmt.Errorf("number of requests must be greater than 0")
		}

		if waitTime <= 0 {
			return fmt.Errorf("wait time must be greater than 0")
		}

		if jitter < 0 || jitter > waitTime {
			return fmt.Errorf("jitter must be between 0 and wait time")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stopCh := signals.SetupSignalHandler()

		ratelimiter, err := engine.EngineFactory(
			engine.WithEngineType(engine.StringToEngineType(engineType)),
			engine.WithStopCh(stopCh),
			engine.WithCapacity(uint64(capacity)),

			// Token bucket specific configuration
			engine.WithFillRate(1.0/fillDuration),
			engine.WithConsumeRate(consumeRate),

			// Leaky bucket specific configuration
			engine.WithLeakRate(time.Duration(drainDuration)*time.Millisecond),

			// Fixed size or sliding window specific configuration
			engine.WithWindowSize(windowSize),
		)
		if err != nil {
			return err
		}

		// Request simulation
		errCh := make(chan error, 1)
		var wg sync.WaitGroup
		for i := int8(0); i < parallel; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for j := int64(0); j < numRequests; j++ {
					// Random jitter between [-jitter, jitter]
					rj := int64(0)
					if jitter > 0 {
						randomJitter, err := rand.Int(rand.Reader, big.NewInt(jitter*2))
						if err != nil {
							errCh <- err
						}
						rj = randomJitter.Int64() - jitter
					}

					if ratelimiter.Allow() {
						fmt.Printf("Worker %d: Request %d ALLOWED, sleeping for %dms\n", i, j+1, waitTime+rj)
					} else {
						fmt.Printf("Worker %d: Request %d DENIED, sleeping for %dms\n", i, j+1, waitTime+rj)
					}
					time.Sleep(time.Duration(waitTime+rj) * time.Millisecond)
				}
			}()
		}

		go func() {
			wg.Wait()
			fmt.Println("All workers finished, press Ctrl+C to exit...")
		}()

		select {
		case <-stopCh:
			fmt.Println("Shutting down the rate limiter engine...")
		case err := <-errCh:
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Shared flags
	runCmd.PersistentFlags().StringVar(&engineType, "engine", "token-bucket", "Rate limiting engine (fixed-window, sliding-window-log, sliding-window-counter, token-bucket, leaky-bucket)")
	runCmd.PersistentFlags().Int64Var(&capacity, "capacity", 5, "All: Maximum number of requests allowed")

	// Token bucket specific flags
	runCmd.PersistentFlags().Float64Var(&fillDuration, "fill-duration", 500, "Token bucket: token refill duration in milliseconds. Default is 500ms (2 tokens/second)")
	runCmd.PersistentFlags().Float64Var(&consumeRate, "consume-rate", 1, "Token bucket: Token consume rate per request")

	// Leaky bucket specific flags
	runCmd.PersistentFlags().Int64Var(&drainDuration, "drain-duration", 500, "Leaky bucket: Drain duration in milliseconds")

	// Fixed size window and sliding window specific flags
	runCmd.PersistentFlags().Int64Var(&windowSize, "window-size", 1000, "Fixed/Sliding window: Window size in milliseconds")

	// Simulation parameters
	runCmd.PersistentFlags().Int64Var(&numRequests, "num-requests", 100, "Simulator: Number of requests to simulate")
	runCmd.PersistentFlags().Int64Var(&waitTime, "wait-time", 100, "Simulator: Wait time between requests in milliseconds")
	runCmd.PersistentFlags().Int64Var(&jitter, "jitter", 0, "Simulator: Random jitter in milliseconds")
	runCmd.PersistentFlags().Int8Var(&parallel, "parallel", 1, "Simulator: Number of parallel workers")
}
