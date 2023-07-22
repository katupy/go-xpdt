package main

import "github.com/spf13/cobra"

var servicesCmd = &cobra.Command{
	Use:          "services",
	Short:        "Manage services.",
	SilenceUsage: true,
}

var servicesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a service.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var servicesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a service.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func initServices() error {
	servicesCmd.AddCommand(servicesStartCmd)
	servicesCmd.AddCommand(servicesStopCmd)
	mainCmd.AddCommand(servicesCmd)

	return nil
}
