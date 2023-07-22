package conf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const EnvPrefix = "XPDT"
const EnvConfigPath = EnvPrefix + "_CONFIG_PATH"

type Config struct {
	Env        *Env   `toml:"env,omitempty" yaml:"env,omitempty"`
	LogLevel   string `toml:"logLevel,omitempty" yaml:"logLevel,omitempty"`
	NoLogColor bool   `toml:"noLogColor,omitempty" yaml:"noLogColor,omitempty"`

	// Where to write common outputs and logs to.
	Outw io.Writer `toml:"-" yaml:"-"`
	Logw io.Writer `toml:"-" yaml:"-"`
}

type Env struct {
	Load       *EnvLoad        `toml:"load,omitempty" yaml:"load,omitempty"`
	Overwrites []*EnvOverwrite `toml:"overwrites,omitempty" yaml:"overwrites,omitempty"`
}

type EnvLoad struct {
	Dir           string `toml:"dir,omitempty" yaml:"dir,omitempty"`
	Filename      string `toml:"filename,omitempty" yaml:"filename,omitempty"`
	NoLogDuration bool   `toml:"noLogDuration,omitempty" yaml:"noLogDuration,omitempty"`

	// The original environment, before loading changes.
	Environ []string `toml:"environ,omitempty" yaml:"environ,omitempty"`
}

type EnvOverwrite struct {
	Dir  string `toml:"dir,omitempty" yaml:"dir,omitempty"`
	File string `toml:"file,omitempty" yaml:"file,omitempty"`
	Root bool   `toml:"root,omitempty" yaml:"root,omitempty"`
	Skip bool   `toml:"skip,omitempty" yaml:"skip,omitempty"`
}

type Service struct {
	Name string `toml:"name,omitempty" yaml:"name,omitempty"`

	Dependencies []string `toml:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	HealthCheck *HealthCheck `toml:"health_check,omitempty" yaml:"health_check,omitempty"`
}

type HealthCheck struct {
	Address string `toml:"address,omitempty" yaml:"address,omitempty"`
	Port    int    `toml:"port,omitempty" yaml:"port,omitempty"`

	Interval string `toml:"interval,omitempty" yaml:"interval,omitempty"`
	Retries  int    `toml:"retries,omitempty" yaml:"retries,omitempty"`
}

func Find() (*Config, error) {
	configPath := os.Getenv(EnvConfigPath)

	var body, ext string

	if configPath == "" {
		ext = "yaml"
	} else {
		b, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		body = os.ExpandEnv(string(b))
		ext = strings.ToLower(filepath.Ext(configPath))

		if len(ext) > 1 {
			ext = ext[1:]
		}
	}

	viper.SetConfigType(ext)

	if err := viper.MergeConfig(bytes.NewBufferString(body)); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := &Config{}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	lvl, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(config.LogLevel)))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if !config.NoLogColor {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339Nano,
		})
	}

	config.Outw = os.Stdout
	config.Logw = os.Stderr

	return config, nil
}
