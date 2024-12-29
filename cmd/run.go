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
			engine.WithCapacity(5),

			// Token bucket specific configuration
			engine.WithFillRate(1.0/(60*60)), // 1 token per hour
			engine.WithConsumeRate(1),

			// Leaky bucket specific configuration
			engine.WithLeakRate(500*time.Millisecond),

			// Fixed size window specific configuration
			engine.WithWindowSize(1000*time.Millisecond),
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

				randomJitter, err := rand.Int(rand.Reader, big.NewInt(jitter))
				if err != nil {
					errCh <- err
				}
				time.Sleep(time.Duration(waitTime+randomJitter.Int64()) * time.Millisecond)
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

	runCmd.PersistentFlags().StringVarP(&engineType, "engine", "e", "token-bucket", "Rate limiting engine (fixed-window, sliding-window, token-bucket, leaky-bucket)")
	runCmd.PersistentFlags().Int64VarP(&numRequests, "num-requests", "n", 100, "Number of requests to simulate")
	runCmd.PersistentFlags().Int64VarP(&waitTime, "wait-time", "w", 0, "Wait time between requests in milliseconds")
	runCmd.PersistentFlags().Int64VarP(&jitter, "jitter", "j", 1, "Random jitter in milliseconds")
}
