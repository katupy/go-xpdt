package env

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.katupy.io/klib"

	"go.katupy.io/xpdt/conf"
)

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

type environVar struct {
	// Original key name.
	key string

	// Original value when the environment was loaded.
	originalValue string

	// Current value.
	currentValue string

	// Whether this key is a path list,
	// and relevant path list data.
	pathList              bool
	pathListElements      []string
	pathListElementExists map[string]bool

	// Whether this key was created by commands.
	created bool

	// Whether this key should be deleted.
	delete bool

	// Whether this key was set for reversal.
	// If true, and the value is unchanged after
	// all operations, this key won't be added
	// to the returned reversal list, only the diff.
	reversal bool

	// Whether this key was set to be deleted on reversal.
	// This is necessary because if a key that was set for deletion
	// is restored, we would lose this information.
	reversalDelete bool
}

func (ev *environVar) resetAndDelete() {
	ev.currentValue = ""
	ev.delete = true

	if ev.pathList {
		ev.pathList = false
		ev.pathListElementExists = nil
		ev.pathListElements = nil
	}
}

type container struct {
	caseInsensitiveEnvironment bool

	env map[string]*environVar

	diff    []string
	reverse []string
}

func (c *container) loadEnviron(environ []string) error {
	c.env = make(map[string]*environVar, len(environ))

	for i := range environ {
		p := strings.SplitN(environ[i], "=", 2)
		envVar := &environVar{
			key:           p[0],
			originalValue: p[1],
			currentValue:  p[1],
		}

		if c.caseInsensitiveEnvironment {
			c.env[strings.ToUpper(p[0])] = envVar
		} else {
			c.env[p[0]] = envVar
		}
	}

	return nil
}

func (c *container) applyReverse() error {
	reverseVar, ok := c.env[conf.EnvReverseVar]
	if !ok {
		return nil
	}

	reverse := []string{}

	if err := json.Unmarshal([]byte(reverseVar.originalValue), &reverse); err != nil {
		return &klib.Error{
			ID:     "57221f4e-bf90-4998-81b9-2544447fcf33",
			Status: http.StatusBadRequest,
			Code:   klib.CodeSerializationError,
			Detail: fmt.Sprintf("Env var %q has an invalid format.", conf.EnvReverseVar),
			Cause:  err.Error(),
		}
	}

	// Ensure this key will be deleted since it has been consumed.
	// Later we will check if it should be recreated or not.
	c.env[conf.EnvReverseVar].delete = true

	for i := 0; i < len(reverse); i++ {
		if len(reverse) == i+1 {
			return &klib.Error{
				ID:     "d8fd6be3-9555-464d-b6c7-2189545ba9b0",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Detail: fmt.Sprintf("Cmd %q is missing a key.", conf.EnvReverseVar),
			}
		}

		cmd := reverse[i]
		key := reverse[i+1]
		keyName := key

		if c.caseInsensitiveEnvironment {
			keyName = strings.ToUpper(key)
		}

		envVar, haveVar := c.env[keyName]
		if !haveVar {
			envVar = &environVar{key: key}
			c.env[keyName] = envVar
		}

		envVar.reversal = true

		switch cmd {
		case "SET":
			// Set the originalValue to ensure this reversal
			// is propagated if the key changes again.

			envVar.originalValue = reverse[i+2]
			envVar.currentValue = envVar.originalValue
			i += 2
		case "DEL":
			envVar.delete = true
			envVar.reversalDelete = true
			envVar.originalValue = ""
			envVar.currentValue = ""
			i += 1
		default:
			return &klib.Error{
				ID:     "52b39360-02a8-4646-9b1e-2cfc75f7b025",
				Status: http.StatusBadRequest,
				Code:   klib.CodeInvalidValue,
				Detail: fmt.Sprintf("Unsupported command %q on key %q", cmd, key),
			}
		}
	}

	return nil
}

func (c *container) makeDiff() {
	c.diff = []string{}
	c.reverse = []string{}

	var reverseEnvVar *environVar

	for _, envVar := range c.env {
		key := envVar.key

		if envVar.delete {
			if key == conf.EnvReverseVar {
				// We'll check if this should be deleted later,
				// after all possible reversals have been calculated.
				reverseEnvVar = envVar
				continue
			}

			c.diff = append(c.diff, "DEL", key)

			// Only create a reversal for this deletion if it was not propagated from the previous env.
			if !envVar.reversalDelete {
				c.reverse = append(c.reverse, "SET", key, envVar.originalValue)
			}

			continue
		}

		// If envVar was created during this run, current and original value are likely the same.
		// If envVar was created because of a reversalDelete, but later was updated with an empty value,
		// it would fall into this case as well, so we must explicitly ignore it.
		if !(envVar.created || envVar.reversalDelete) && envVar.currentValue == envVar.originalValue {
			if envVar.reversal {
				// If this was originally a reversal, we must apply it.
				c.diff = append(c.diff, "SET", key, envVar.originalValue)
			}

			continue
		}

		value := envVar.currentValue

		if envVar.pathList {
			value = strings.Join(envVar.pathListElements, string(os.PathListSeparator))
		}

		c.diff = append(c.diff, "SET", key, value)

		// If this was originally a reversal, we must propagate it.
		if envVar.reversalDelete {
			c.reverse = append(c.reverse, "DEL", key)
		} else if envVar.reversal {
			c.reverse = append(c.reverse, "SET", key, envVar.originalValue)
		} else if envVar.originalValue == "" {
			// Key was created.
			c.reverse = append(c.reverse, "DEL", key)
		} else {
			// Key was updated.
			c.reverse = append(c.reverse, "SET", key, envVar.originalValue)
		}
	}

	if reverseEnvVar == nil {
		return
	}

	if len(c.reverse) == 0 {
		c.diff = append(c.diff, "DEL", conf.EnvReverseVar)
	}
}

func (c *container) writeDiff(w io.Writer) error {
	if len(c.reverse) > 0 {
		b, err := json.Marshal(c.reverse)
		if err != nil {
			return &klib.Error{
				ID:     "9924802b-ef25-4864-97e0-3f3d7ce9f907",
				Status: http.StatusInternalServerError,
				Code:   klib.CodeSerializationError,
				Title:  "Failed to serialize reverse env var",
				Cause:  err.Error(),
			}
		}

		c.diff = append(c.diff, "SET", conf.EnvReverseVar, string(b))
	}

	if _, err := fmt.Fprintln(w, strings.Join(c.diff, "\n")); err != nil {
		return &klib.Error{
			ID:     "1e6591ae-229c-4655-a57d-f3fe13f53ebf",
			Status: http.StatusInternalServerError,
			Code:   klib.CodeBufferError,
			Title:  "Failed to write diff",
			Cause:  err.Error(),
		}
	}

	return nil
}
