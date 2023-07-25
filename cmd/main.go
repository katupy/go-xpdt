package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.katupy.io/xpdt/conf"
)

var mainCmd = &cobra.Command{
	Use:          "xpdt",
	Short:        "The cross-platform dev tool.",
	SilenceUsage: true,
}

func initMain() error {
	mainCmd.PersistentFlags().String("config", "", "Path to config file.")
	if err := viper.BindPFlag("config", mainCmd.PersistentFlags().Lookup("config")); err != nil {
		return fmt.Errorf("failed to bind config flag: %w\n", err)
	}

	mainCmd.PersistentFlags().String("logLevel", "info", "The default log level.")
	if err := viper.BindPFlag("logLevel", mainCmd.PersistentFlags().Lookup("logLevel")); err != nil {
		return fmt.Errorf("failed to bind logLevel flag: %w\n", err)
	}

	mainCmd.PersistentFlags().Bool("noLogColor", false, "Whether to log without color.")
	if err := viper.BindPFlag("noLogColor", mainCmd.PersistentFlags().Lookup("noLogColor")); err != nil {
		return fmt.Errorf("failed to bind noLogColor flag: %w\n", err)
	}

	mainCmd.PersistentFlags().Bool("caseInsensitiveEnvironment", runtime.GOOS == "windows", "Whether environment var names are case insensitive.")
	if err := viper.BindPFlag("caseInsensitiveEnvironment", mainCmd.PersistentFlags().Lookup("caseInsensitiveEnvironment")); err != nil {
		return fmt.Errorf("failed to bind caseInsensitiveEnvironment flag: %w\n", err)
	}

	mainCmd.PersistentFlags().Bool("caseSensitiveFilesystem", !(runtime.GOOS == "windows"), "Whether the filesystem object names are case sensitive.")
	if err := viper.BindPFlag("caseSensitiveFilesystem", mainCmd.PersistentFlags().Lookup("caseSensitiveFilesystem")); err != nil {
		return fmt.Errorf("failed to bind caseSensitiveFilesystem flag: %w\n", err)
	}

	return nil
}

func init() {
	for _, initFunc := range []func() error{
		initMain,
		initEnv,
		initServices,
	} {
		if err := initFunc(); err != nil {
			fmt.Printf("failed to init: %s\n", err)
			os.Exit(-2)
		}
	}
}

func main() {
	if err := mainCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func parseFlags() error {
	if configPath := strings.TrimSpace(viper.GetString("config")); configPath != "" {
		if err := os.Setenv(conf.EnvConfigPath, configPath); err != nil {
			return fmt.Errorf("failed to set %s env var: %w", conf.EnvConfigPath, err)
		}
	}

	return nil
}
