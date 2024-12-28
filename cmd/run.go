package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/minhthong582000/rate-limiter/internal/engine"
	"github.com/minhthong582000/soa-404/pkg/signals"
	"github.com/spf13/cobra"
)

var (
	engineType     string
	trafficLogFile string
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
		)
		if err != nil {
			return err
		}

		// Read each line from the traffic log file
		file, err := os.Open(trafficLogFile)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)

		errCh := make(chan error, 1)
		go func() {
			for scanner.Scan() {
				tsStr := scanner.Text()
				if tsStr == "" {
					continue
				}

				ts, err := time.Parse(time.RFC3339, tsStr)
				if err != nil {
					errCh <- err
					return
				}

				if ratelimiter.AllowAt(ts) {
					println("ALLOWED")
				} else {
					println("BLOCKED")
				}

				time.Sleep(500 * time.Millisecond)
			}
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
	runCmd.PersistentFlags().StringVarP(&trafficLogFile, "log", "l", "", "Traffic log file path")
}
