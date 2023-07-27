package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.katupy.io/xpdt/conf"
	"go.katupy.io/xpdt/env"
)

var envHookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Set shell env hook.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("env.hook.list") {
			for _, item := range []string{
				"powershell",
				"zsh",
			} {
				fmt.Println(item)
			}

			return nil
		}

		if len(args) != 1 {
			return errors.New("must provide exactly one shell name")
		}

		cmdName := filepath.Base(os.Args[0])

		switch strings.ToLower(args[0]) {
		case "powershell":
			fmt.Println(env.PowerShellHook)
		case "zsh":
			fmt.Printf(env.ZshHook, cmdName)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}

		return nil
	},
}

var envCmd = &cobra.Command{
	Use:          "env",
	Short:        "Manage environment variables.",
	SilenceUsage: true,
}

var envLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load the environment.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := parseFlags(); err != nil {
			return fmt.Errorf("failed to parse flags: %w", err)
		}

		config, err := conf.Find()
		if err != nil {
			return fmt.Errorf("failed to find config: %w", err)
		}

		loader := env.NewLoader(config)

		if err := loader.Load(); err != nil {
			// TODO: improve console logging.
			return fmt.Errorf("failed to load env: %w", err)
		}

		return nil
	},
}

func initEnv() error {
	envHookCmd.PersistentFlags().Bool("list", false, "Show the list of available hooks.")
	if err := viper.BindPFlag("env.hook.list", envHookCmd.PersistentFlags().Lookup("list")); err != nil {
		return fmt.Errorf("failed to bind env.hook.list flag: %w\n", err)
	}

	envCmd.AddCommand(envHookCmd)

	envLoadCmd.PersistentFlags().StringP("dir", "C", conf.DefaultEnvLoadDir, "Change to directory before execution.")
	if err := viper.BindPFlag("env.load.dir", envLoadCmd.PersistentFlags().Lookup("dir")); err != nil {
		return fmt.Errorf("failed to bind env.load.dir flag: %w\n", err)
	}

	envLoadCmd.PersistentFlags().StringP("filename", "f", conf.DefaultEnvLoadFilename, "Config filename.")
	if err := viper.BindPFlag("env.load.filename", envLoadCmd.PersistentFlags().Lookup("filename")); err != nil {
		return fmt.Errorf("failed to bind env.load.filename flag: %w\n", err)
	}

	envLoadCmd.PersistentFlags().Bool("noLogDuration", false, "Do not log how long it took to load the environment.")
	if err := viper.BindPFlag("env.load.noLogDuration", envLoadCmd.PersistentFlags().Lookup("noLogDuration")); err != nil {
		return fmt.Errorf("failed to bind env.load.noLogDuration flag: %w\n", err)
	}

	envCmd.AddCommand(envLoadCmd)
	mainCmd.AddCommand(envCmd)

	return nil
}
