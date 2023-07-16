package conf

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
