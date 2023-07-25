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

	file  *File
	index int
}

type File struct {
	Root bool `toml:"root,omitempty" yaml:"root,omitempty"`

	Commands []*Command `toml:"commands,omitempty" yaml:"commands,omitempty"`

	dir      string
	filepath string
}

type container struct {
	caseInsensitiveEnvironment bool

	oldEnv map[string]string
	curEnv map[string]string
	delEnv map[string]bool

	pathListElementExists map[string]map[string]bool
	pathListElements      map[string][]string
	pathListExists        map[string]bool
}

func (c *container) resetPaths() {
	c.pathListElementExists = make(map[string]map[string]bool)
	c.pathListElements = make(map[string][]string)
	c.pathListExists = make(map[string]bool)
}
