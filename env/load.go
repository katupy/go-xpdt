package env

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.katupy.io/klib"
	"gopkg.in/yaml.v3"

	"go.katupy.io/xpdt/conf"
)

func Load(c *conf.Config) error {
	now := time.Now()

	if c == nil {
		return &klib.Error{
			ID:     "01H5XE3N8EG0EER1J8ZSBB341D",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".",
			Detail: "Missing config",
		}
	}

	if c.Env == nil {
		return &klib.Error{
			ID:     "01H5XE7HV41CKBWXM2KWQE2E0V",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env",
			Detail: "Missing config.env",
		}
	}

	if c.Env.Load == nil {
		return &klib.Error{
			ID:     "01H5XEGQ0B1003AC3YWTW74TCD",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env.load",
			Detail: "Missing config.env.load",
		}
	}

	files, err := GetFiles(c)
	if err != nil {
		return klib.ForwardError("01H5XEPKYK78YNAX4SMFD3F98M", err)
	}

	if len(c.Env.Load.Environ) == 0 {
		c.Env.Load.Environ = os.Environ()
	}

	// The starting (static) env.
	oldEnv := GetEnviron(c.Env.Load.Environ, runtime.GOOS == "windows")

	// The current (modifiable) env.
	// It starts as a copy of the old env.
	curEnv := make(map[string]string, len(oldEnv))

	for k, v := range oldEnv {
		curEnv[k] = v
	}

	platform := runtime.GOOS + "_" + runtime.GOARCH

	_ = files
	_ = platform
	_ = curEnv

	// End.

	if !c.Env.Load.NoLogDuration {
		fmt.Fprintf(c.Logw, "xpdt: env loaded in %s\n", time.Since(now))
	}

	return nil
}

// GetEnviron returns a map of environment variables from the given environ slice.
// It also accounts for case-insensitive keys for environments like Windows.
func GetEnviron(environ []string, ignoreCase bool) map[string]string {
	env := make(map[string]string, len(environ))

	for i := range environ {
		p := strings.SplitN(environ[i], "=", 2)

		if ignoreCase {
			env[strings.ToUpper(p[0])] = p[1]
		} else {
			env[p[0]] = p[1]
		}
	}

	return env
}

// GetFiles returns a slice of files to be loaded, starting from the current directory
// and going up the directory tree until a root directory is reached.
func GetFiles(c *conf.Config) ([]*File, error) {
	c.Env.Load.Dir = strings.TrimSpace(c.Env.Load.Dir)

	if c.Env.Load.Dir == "" {
		return nil, &klib.Error{
			ID:     "01H5XEMMAB72AYK72CEV6JN09N",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env.load.dir",
			Detail: "Missing config.env.load.dir",
		}
	}

	dir, err := filepath.Abs(c.Env.Load.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get abs path of c.Env.Load.Dir: %w", err)
	}

	filename := strings.TrimSpace(c.Env.Load.Filename)

	if filename == "" {
		return nil, &klib.Error{
			ID:     "01H5XEND4HEPG9GDKMCQ1P2VGH",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env.load.filename",
			Detail: "Missing config.env.load.filename",
		}
	}

	// There might be multiple overwrites for the same directory.
	globalOverwrites := make(map[string][]*conf.EnvOverwrite, len(c.Env.Overwrites))

	for i := range c.Env.Overwrites {
		overwrite := c.Env.Overwrites[i]
		overwriteDir := filepath.Clean(overwrite.Dir)

		globalOverwrites[overwriteDir] = append(globalOverwrites[overwriteDir], overwrite)
	}

	var files []*File

	stop := func() bool {
		parentDir := filepath.Join(dir, "../")

		if parentDir == dir {
			return true
		}

		dir = parentDir

		return false
	}

	for {
		var b []byte
		var f string
		var overwriteLoop bool
		var overwriteIndex int

	OVERWRITES:

		// First check if there are global overwrites for this directory.
		overwrites, ok := globalOverwrites[dir]

		var root, skip bool

		if ok {
			if !overwriteLoop {
				overwriteLoop = true
			}

			overwrite := overwrites[overwriteIndex]

			skip = overwrite.Skip

			if !skip {
				b, err = os.ReadFile(overwrite.File)
				if err != nil {
					return nil, &klib.Error{
						ID:     "01H5XEW98E1T0EFFW6T2VEY3W9",
						Status: http.StatusInternalServerError,
						Code:   klib.CodeFileError,
						Path:   fmt.Sprintf(".overwrites[%d].file", overwriteIndex),
						Title:  "Failed to read file",
						Cause:  err.Error(),
						Meta: map[string]any{
							"filepath": overwrite.File,
						},
					}
				}
			}

			f = overwrite.File
			root = overwrite.Root

			overwriteIndex++
		} else {
			f = filepath.Join(dir, filename)

			b, err = os.ReadFile(f)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if stop() {
						break
					}

					continue
				}

				return nil, &klib.Error{
					ID:     "01H5XF4Y6HPNG0D68558A0JSG5",
					Status: http.StatusInternalServerError,
					Code:   klib.CodeFileError,
					Title:  "Failed to read file",
					Cause:  err.Error(),
					Meta: map[string]any{
						"filepath": f,
					},
				}
			}
		}

		if !skip {
			file := new(File)
			if err := yaml.Unmarshal(b, file); err != nil {
				return nil, &klib.Error{
					ID:     "01H5XFA1K3HW36GG5721H9GB2Q",
					Status: http.StatusBadRequest,
					Code:   klib.CodeSerializationError,
					Title:  "Failed to unmarshal file",
					Cause:  err.Error(),
				}
			}

			file.filepath = f
			file.dir = dir

			switch {
			case root:
				file.Root = true
			}

			log.Debug().
				Str("_label", "envFile").
				Str("dir", dir).
				Str("file", file.filepath).
				Int("index", len(files)).
				Bool("overwriteLoop", overwriteLoop).
				Int("overwriteIndex", overwriteIndex).
				Bool("root", file.Root).
				Send()

			if overwriteLoop && overwriteIndex > 1 {
				// Insert the current overwrite before others so they are
				// processed in the correct order later.
				index := len(files) - overwriteIndex + 1
				files = append(files[:index+1], files[index:]...)
				files[index] = file
			} else {
				files = append(files, file)
			}

			if file.Root {
				break
			}
		}

		overwriteLoop = overwriteLoop && overwriteIndex != len(overwrites)

		if overwriteLoop {
			goto OVERWRITES
		}

		if stop() {
			break
		}
	}

	return files, nil
}
