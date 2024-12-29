package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine"
	"github.com/minhthong582000/soa-404/pkg/signals"
	"github.com/spf13/cobra"
)

var (
	engineType string
	capacity   int64

	// Token bucket specific configuration
	fillDuration float64
	consumeRate  float64

	// Leaky bucket specific configuration
	drainDuration int64 // in milliseconds

	// Fixed size window specific configuration
	windowSize int64 // in milliseconds

	// Simulation parameters
	numRequests int64
	waitTime    int64 // in milliseconds
	jitter      int64 // in milliseconds
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the rate limiter engine",
	Long: `A command to run the rate limiter engine based on the selected engine type.
You can choose between different rate limiting engines such as fixed-window, sliding-window, token-bucket, and leaky-bucket.`,
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

			// Fixed size window specific configuration
			engine.WithWindowSize(time.Duration(windowSize)*time.Millisecond),
		)
		if err != nil {
			return err
		}

		errCh := make(chan error, 1)
		go func() {
			for i := int64(0); i < numRequests; i++ {
				if !ratelimiter.Allow() {
					fmt.Printf("Request %d REJECTED\n", i)
				} else {
					fmt.Printf("Request %d ACCEPTED\n", i)
				}

				if jitter > 0 && jitter <= waitTime {
					randomJitter, err := rand.Int(rand.Reader, big.NewInt(jitter*2))
					if err != nil {
						errCh <- err
					}
					waitTime += -jitter + randomJitter.Int64()
				}
				fmt.Printf("Waiting for %d milliseconds...\n", waitTime)
				time.Sleep(time.Duration(waitTime) * time.Millisecond)
			}
			fmt.Println("Press Ctrl+C to exit...")
		}()

		select {
		case <-stopCh:
			fmt.Println("Shutting down the rate limiter engine...")
		case err := <-errCh:
			fmt.Println(err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Shared flags
	runCmd.PersistentFlags().StringVar(&engineType, "engine", "token-bucket", "Rate limiting engine (fixed-window, sliding-window, token-bucket, leaky-bucket)")
	runCmd.PersistentFlags().Int64Var(&capacity, "capacity", 5, "All: Maximum number of requests allowed")

	// Token bucket specific flags
	runCmd.PersistentFlags().Float64Var(&fillDuration, "fill-duration", 0.5, "Token bucket: token refill duration in seconds. Default is 0.5s (2 tokens/second)")
	runCmd.PersistentFlags().Float64Var(&consumeRate, "consume-rate", 1, "Token bucket: Token consume rate per request")

	// Leaky bucket specific flags
	runCmd.PersistentFlags().Int64Var(&drainDuration, "drain-duration", 500, "Leaky bucket: Drain duration in milliseconds")

	// Fixed size window specific flags
	runCmd.PersistentFlags().Int64Var(&windowSize, "window-size", 1000, "Fixed window: Window size in milliseconds")

	// Simulation parameters
	runCmd.PersistentFlags().Int64Var(&numRequests, "num-requests", 100, "Simulator: Number of requests to simulate")
	runCmd.PersistentFlags().Int64Var(&waitTime, "wait-time", 100, "Simulator: Wait time between requests in milliseconds")
	runCmd.PersistentFlags().Int64Var(&jitter, "jitter", 0, "Simulator: Random jitter in milliseconds")
}
