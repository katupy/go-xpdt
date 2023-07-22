package env

type Command struct {
	Declare string `toml:"declare,omitempty" yaml:"declare,omitempty"`
	Value   string `toml:"value,omitempty" yaml:"value,omitempty"`

	Add string `toml:"add,omitempty" yaml:"add,omitempty"`
	Del string `toml:"del,omitempty" yaml:"del,omitempty"`
	Set string `toml:"set,omitempty" yaml:"set,omitempty"`

	Platform string `toml:"platform,omitempty" yaml:"platform,omitempty"`
	URI      string `toml:"uri,omitempty" yaml:"uri,omitempty"`
	Append   bool   `toml:"append,omitempty" yaml:"append,omitempty"`
}

type File struct {
	Root bool `toml:"root,omitempty" yaml:"root,omitempty"`

	Commands []*Command `toml:"cmds,omitempty" yaml:"cmds,omitempty"`

	filepath string
	dir      string
}
