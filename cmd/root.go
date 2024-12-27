package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rate-limiter",
	Short: "Rate limiter is a simple rate limiting service",
	Long:  `Rate limiter is a simple rate limiting service`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
