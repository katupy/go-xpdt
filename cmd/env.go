package main

import "github.com/spf13/cobra"

var envCmd = &cobra.Command{
	Use:          "env",
	Short:        "Manage environment variables.",
	SilenceUsage: true,
}

var envLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load the environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	envCmd.AddCommand(envLoadCmd)
	mainCmd.AddCommand(envCmd)
}
