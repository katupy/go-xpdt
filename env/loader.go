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

type Loader struct {
	config *conf.Config

	platform string

	files []*File

	container *container

	templateHandler klib.StringHandler
}

func NewLoader(config *conf.Config) *Loader {
	return &Loader{
		config: config,
	}
}

func (l *Loader) Load() error {
	now := time.Now()

	if l.config == nil {
		return &klib.Error{
			ID:     "df006438-8216-4fbf-a4b6-3c4f933a6c0d",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".",
			Detail: "Missing config",
		}
	}

	if l.config.Env == nil {
		return &klib.Error{
			ID:     "0002b99d-191f-4bb1-9120-d5853df954c9",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env",
			Detail: "Missing config.env",
		}
	}

	if l.config.Env.Load == nil {
		return &klib.Error{
			ID:     "13738af0-94ef-4589-bf2d-e11f6873f2cd",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   ".env.load",
			Detail: "Missing config.env.load",
		}
	}

	if err := l.FindFiles(); err != nil {
		return klib.ForwardError("b7eb276b-2aa6-4058-a711-3f09308ee200", err)
	}

	logDuration := func() {
		if !l.config.Env.Load.NoLogDuration {
			fmt.Fprintf(l.config.Logw, "xpdt: env loaded in %s\n", time.Since(now))
		}
	}

	if len(l.files) == 0 {
		// Nothing to do.
		logDuration()
		return nil
	}

	if len(l.config.Env.Load.Environ) == 0 {
		l.config.Env.Load.Environ = os.Environ()
	}

	if !l.config.CaseInsensitiveEnvironment {
		l.config.CaseInsensitiveEnvironment = runtime.GOOS == "windows"
	}

	l.platform = runtime.GOOS + "_" + runtime.GOARCH

	c := new(container)

	// The starting (static) env.
	c.oldEnv = GetEnviron(l.config.Env.Load.Environ, l.config.CaseInsensitiveEnvironment)

	// The current (modifiable) env.
	// It starts as a copy of the old env.
	c.curEnv = make(map[string]string, len(c.oldEnv))

	for k, v := range c.oldEnv {
		c.curEnv[k] = v
	}

	data := map[string]any{
		"_PLATFORM": l.platform,
	}

	l.genTemplateHandler(data)
	c.resetPaths()

	for i := len(l.files) - 1; i >= 0; i-- {
		if err := l.LoadFile(i); err != nil {
			return klib.ForwardError("2fbe24dd-bc10-403f-b777-f3dd7898c8f4", err)
		}
	}

	logDuration()

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

// FindFiles register the slice of files to be loaded, starting from the current directory
// and going up the directory tree until a root directory is reached.
func (l *Loader) FindFiles() error {
	loadDir := strings.TrimSpace(l.config.Env.Load.Dir)

	if loadDir == "" {
		loadDir = conf.DefaultEnvLoadDir
	}

	dir, err := filepath.Abs(loadDir)
	if err != nil {
		return fmt.Errorf("failed to get abs path of c.Env.Load.Dir: %w", err)
	}

	loadFilename := strings.TrimSpace(l.config.Env.Load.Filename)

	if loadFilename == "" {
		loadFilename = conf.DefaultEnvLoadFilename
	}

	// There might be multiple overwrites for the same directory.
	globalOverwrites := make(map[string][]*conf.EnvOverwrite, len(l.config.Env.Overwrites))

	for i := range l.config.Env.Overwrites {
		overwrite := l.config.Env.Overwrites[i]
		overwriteDir := strings.TrimSpace(overwrite.Dir)

		if overwriteDir == "" {
			return &klib.Error{
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
			return &klib.Error{
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

	// addFile adds the file to the files slice and
	// returns whether the file is a root file,
	// indicating that file discovery should stop.
	addFile := func(index int, overwrite *conf.EnvOverwrite) (bool, error) {
		if overwrite.Skip {
			return overwrite.Root, nil
		}

		var path string

		if overwrite.Dir != "" {
			path = fmt.Sprintf(".env.overwrites[%d].file", index)
		}

		b, err := os.ReadFile(overwrite.File)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if overwrite.Dir == "" {
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
						"filepath": overwrite.File,
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
					"filepath": overwrite.File,
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

		file.filepath = overwrite.File
		file.dir = dir

		if overwrite.Root {
			file.Root = true
		}

		// fileIndex indicates the index of file in files after its addition.
		fileIndex := len(l.files)

		if index == 0 {
			// There is a single loader for this directory.
			l.files = append(l.files, file)
		} else {
			// There are multiple loaders for this directory,
			// so we need to insert the file at the correct index
			// considering that later all files are processed
			// in the reverse order.

			fileIndex -= index
			l.files = append(l.files[:fileIndex+1], l.files[fileIndex:]...)
			l.files[fileIndex] = file
		}

		log.Debug().
			Str("_label", "envFile").
			Str("dir", dir).
			Str("file", file.filepath).
			Int("fileIndex", fileIndex).
			Int("loaderIndex", index).
			Int("filesLen", len(l.files)).
			Bool("root", file.Root).
			Send()

		return file.Root, nil
	}

GET_FILES_LOOP:
	for {
		// First check if there are global overwrites for this directory.
		overwrites, ok := globalOverwrites[dir]

		if !ok {
			overwrites = []*conf.EnvOverwrite{{
				File: filepath.Join(dir, loadFilename),
			}}
		}

		for i := range overwrites {
			root, err := addFile(i, overwrites[i])
			if err != nil {
				return err
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

	return nil
}

// genTemplateHandler generates a template handler for the given data.
func (l *Loader) genTemplateHandler(data map[string]any) {
	funcMap := klib.BaseFuncMap()

	getEnv := func(k string) string {
		if l.config.CaseInsensitiveEnvironment {
			return l.container.curEnv[strings.ToUpper(k)]
		}
		return l.container.curEnv[k]
	}

	expandEnv := func(s string) string {
		return os.Expand(s, getEnv)
	}

	// Replace sprig's env and expandenv with our own.
	funcMap["env"] = getEnv
	funcMap["expandenv"] = expandEnv

	l.templateHandler = &templateHandler{
		data:    data,
		funcMap: funcMap,
	}
}

func (l *Loader) LoadFile(index int) error {
	file := l.files[index]

	if err := os.Chdir(file.dir); err != nil {
		return &klib.Error{
			ID:     "869494ce-40c2-43f5-a5ef-a5a5d3b6201f",
			Status: http.StatusInternalServerError,
			Code:   klib.CodeFilesystemError,
			Title:  "Failed to change directory",
			Cause:  err.Error(),
			Meta: map[string]any{
				"dir": file.dir,
			},
		}
	}

	for i := range file.Commands {
		if err := l.LoadCommand(index, i); err != nil {
			return klib.ForwardError("f1b51039-fc62-4f06-b9c6-965a1b7dcc66", err)
		}
	}

	return nil
}

func (l *Loader) LoadCommand(fileIndex, cmdIndex int) error {
	file := l.files[fileIndex]
	cmd := file.Commands[cmdIndex]

	if cmd.Platform != "" && cmd.Platform != l.platform {
		// log.Debug().
		// 	Int("index", ).
		// 	Str("value", cmd.Platform).
		// 	Msg("Skipping platform.")

		return nil
	}

	var cmdFunc func(int, int) error

	switch {
	case cmd.Add != "":
		cmdFunc = l.cmdAdd
	case cmd.Set != "":
	case cmd.Del != "":
	}

	if err := cmdFunc(fileIndex, cmdIndex); err != nil {
		return klib.ForwardError("c2ebf96d-8171-4d4b-8853-eb4cf48d7c7f", err)
	}

	return nil
}

func (l *Loader) cmdAdd(fileIndex, cmdIndex int) error {
	file := l.files[fileIndex]
	cmd := file.Commands[cmdIndex]

	var values []string

	switch {
	case cmd.Value != "":
		value, err := l.templateHandler.Handle(cmd.Value)
		if err != nil {
			return &klib.Error{
				ID:     "59da8085-4cc7-4767-838f-11d8ccb6bae5",
				Status: http.StatusBadRequest,
				Code:   klib.CodeExecutionError,
				Path:   fmt.Sprintf("file[%d].commands[%d].value", fileIndex, cmdIndex),
				Title:  "Missing value",
			}
		}

		values = []string{value}
	default:
		return &klib.Error{
			ID:     "ce66adb8-2e56-40b7-8268-27a8955296b5",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			Path:   fmt.Sprintf("file[%d].commands[%d]", fileIndex, cmdIndex),
			Title:  "Missing value",
		}
	}

	_ = values

	// key := cmd.Add

	// if err := l.loadPath(key); err != nil {
	// 	return klib.ForwardError("bfb999a7-55af-47ab-a8b3-bc15be757c48", err)
	// }

	// for i := range values {
	// 	if err := l.addPath(key, values[i], cmd.Append); err != nil {
	// 		return klib.ForwardError("4aa49cf3-1289-403a-bbb2-b25d6ad84a4c", err)
	// 	}
	// }

	return nil
}
