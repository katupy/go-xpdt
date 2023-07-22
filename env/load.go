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
			ID:     "df006438-8216-4fbf-a4b6-3c4f933a6c0d",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".",
			Detail: "Missing config",
		}
	}

	if c.Env == nil {
		return &klib.Error{
			ID:     "0002b99d-191f-4bb1-9120-d5853df954c9",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env",
			Detail: "Missing config.env",
		}
	}

	if c.Env.Load == nil {
		return &klib.Error{
			ID:     "13738af0-94ef-4589-bf2d-e11f6873f2cd",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env.load",
			Detail: "Missing config.env.load",
		}
	}

	files, err := GetFiles(c)
	if err != nil {
		return klib.ForwardError("b7eb276b-2aa6-4058-a711-3f09308ee200", err)
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
	loadDir := strings.TrimSpace(c.Env.Load.Dir)

	if loadDir == "" {
		loadDir = conf.DefaultEnvLoadDir
	}

	dir, err := filepath.Abs(loadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get abs path of c.Env.Load.Dir: %w", err)
	}

	loadFilename := strings.TrimSpace(c.Env.Load.Filename)

	if loadFilename == "" {
		loadFilename = conf.DefaultEnvLoadFilename
	}

	// There might be multiple overwrites for the same directory.
	globalOverwrites := make(map[string][]*conf.EnvOverwrite, len(c.Env.Overwrites))

	for i := range c.Env.Overwrites {
		overwrite := c.Env.Overwrites[i]
		overwriteDir := strings.TrimSpace(overwrite.Dir)

		if overwriteDir == "" {
			return nil, &klib.Error{
				ID:     "dc558b94-ddf9-4280-8eab-94e16cc79418",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Path:   fmt.Sprintf(".env.overwrites[%d].dir", i),
				Detail: "Overwrite dir cannot be empty.",
				Meta: map[string]any{
					"#i":  i,
					"dir": overwrite.Dir,
				},
			}
		}

		cleanDir := filepath.Clean(overwrite.Dir)

		if !filepath.IsAbs(cleanDir) {
			return nil, &klib.Error{
				ID:     "8ee66ad2-84e7-4aa1-8151-55af5c6f4b7d",
				Status: http.StatusBadRequest,
				Code:   klib.CodeInvalidValue,
				Path:   fmt.Sprintf(".env.overwrites[%d].dir", i),
				Detail: "Overwrite dir must be absolute.",
				Meta: map[string]any{
					"#i":  i,
					"dir": overwrite.Dir,
				},
			}
		}

		globalOverwrites[cleanDir] = append(globalOverwrites[overwriteDir], overwrite)
	}

	var files []*File

	// addFile adds the file to the files slice and
	// returns whether the file is a root file,
	// indicating that file discovery should stop.
	addFile := func(index int, loader *conf.EnvOverwrite) (bool, error) {
		if loader.Skip {
			return loader.Root, nil
		}

		var path string

		if loader.Dir != "" {
			path = fmt.Sprintf(".env.overwrites[%d].file", index)
		}

		b, err := os.ReadFile(loader.File)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if loader.Dir == "" {
					// Move forward if the loader is not an overwrite.
					return false, nil
				}

				return false, &klib.Error{
					ID:     "abcd4eb4-6a53-49f2-b4ce-3e6bb8562609",
					Status: http.StatusBadRequest,
					Code:   klib.CodeNotFound,
					Path:   path,
					Title:  "File not found",
					Cause:  err.Error(),
					Meta: map[string]any{
						"filepath": loader.File,
					},
				}
			}

			return false, &klib.Error{
				ID:     "5b15c330-a9a8-4426-aa1e-83120e7f9996",
				Status: http.StatusInternalServerError,
				Code:   klib.CodeFileError,
				Path:   path,
				Title:  "Failed to read file",
				Cause:  err.Error(),
				Meta: map[string]any{
					"filepath": loader.File,
				},
			}
		}

		file := new(File)
		if err := yaml.Unmarshal(b, file); err != nil {
			return false, &klib.Error{
				ID:     "f7476e4b-e592-4f2c-ab97-cae327b957a4",
				Status: http.StatusBadRequest,
				Code:   klib.CodeSerializationError,
				Path:   path,
				Title:  "Failed to unmarshal file",
				Cause:  err.Error(),
			}
		}

		file.filepath = loader.File
		file.dir = dir

		if loader.Root {
			file.Root = true
		}

		// fileIndex indicates the index of file in files after its addition.
		fileIndex := len(files)

		if index == 0 {
			// There is a single loader for this directory.
			files = append(files, file)
		} else {
			// There are multiple loaders for this directory,
			// so we need to insert the file at the correct index
			// considering that later all files are processed
			// in the reverse order.

			index := fileIndex - index - 1
			files = append(files[:index+1], files[index:]...)
			files[index] = file
			fileIndex = index
		}

		log.Debug().
			Str("_label", "envFile").
			Str("dir", dir).
			Str("file", file.filepath).
			Int("fileIndex", fileIndex).
			Int("loaderIndex", index).
			Int("filesLen", len(files)).
			Bool("root", file.Root).
			Send()

		return file.Root, nil
	}

GET_FILES_LOOP:
	for {
		// First check if there are global overwrites for this directory.
		loaders, ok := globalOverwrites[dir]

		if !ok {
			loaders = []*conf.EnvOverwrite{{
				File: filepath.Join(dir, loadFilename),
			}}
		}

		for i := range loaders {
			root, err := addFile(i, loaders[i])
			if err != nil {
				return nil, err
			}

			if root {
				break GET_FILES_LOOP
			}
		}

		parentDir := filepath.Join(dir, "../")

		if parentDir == dir {
			// Reached a root directory.
			break
		}

		dir = parentDir
	}

	return files, nil
}
