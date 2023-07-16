package main

import (
	"os"

	"github.com/spf13/cobra"
)

var mainCmd = &cobra.Command{
	Use:          "xpdt",
	Short:        "The cross-platform dev tool.",
	SilenceUsage: true,
}

func init() {
}

func main() {
	if err := mainCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
