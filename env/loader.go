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

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
	"go.katupy.io/klib"
	"gopkg.in/yaml.v3"

	"go.katupy.io/xpdt/conf"
)

type Loader struct {
	config    *conf.Config
	data      map[string]any
	files     []*File
	container *container
	platform  string

	templateHandler klib.StringHandler
	fileLoader      FileLoader
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

	if len(l.config.Env.Load.Environ) == 0 {
		l.config.Env.Load.Environ = os.Environ()
	}

	if !l.config.CaseInsensitiveEnvironment {
		l.config.CaseInsensitiveEnvironment = runtime.GOOS == "windows"
	}

	c := &container{
		caseInsensitiveEnvironment: l.config.CaseInsensitiveEnvironment,
	}

	l.container = c

	if err := c.loadEnviron(l.config.Env.Load.Environ); err != nil {
		return klib.ForwardError("860fd303-8d01-45ac-9cce-9c901ec7d05d", err)
	}

	if err := c.applyReverse(); err != nil {
		return klib.ForwardError("e19c02fa-1e38-45b5-b520-ece8edcb621b", err)
	}

	pathHandler := &defaultPathHandler{
		caseSensitiveFilesystem: l.config.CaseSensitiveFilesystem,
	}

	pathLoader := &defaultPathLoader{
		pathHandler: pathHandler,
	}

	l.genTemplateHandler(l.data)

	l.fileLoader = &defaultFileLoader{
		commandLoader: &defaultCommandLoader{
			platform: l.platform,
			commandMethods: &defaultCommandMethods{
				container:       c,
				pathHandler:     pathHandler,
				pathLoader:      pathLoader,
				templateHandler: l.templateHandler,
			},
		},
	}

	for i := len(l.files) - 1; i >= 0; i-- {
		file := l.files[i]

		if err := l.fileLoader.Load(file); err != nil {
			return klib.ForwardError("2fbe24dd-bc10-403f-b777-f3dd7898c8f4", err)
		}
	}

	c.makeDiff()

	if err := c.writeDiff(l.config.Outw); err != nil {
		return klib.ForwardError("35c11746-07ad-4bf0-86f9-a811a7e57aff", err)
	}

	logDuration()

	return nil
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

	l.platform = runtime.GOOS + "_" + runtime.GOARCH
	l.data = map[string]any{
		"_PLATFORM": l.platform,
	}

	for k := range l.config.Env.Data {
		dataKey := k
		dataFile := l.config.Env.Data[k]

		b, err := os.ReadFile(dataFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &klib.Error{
					ID:     "145a329c-9fe2-493a-ac6f-e70603dae4ec",
					Status: http.StatusBadRequest,
					Code:   klib.CodeNotFound,
					Path:   fmt.Sprintf(".env.data[%s]", dataKey),
					Title:  "File not found",
					Cause:  err.Error(),
					Meta: map[string]any{
						"filepath": dataFile,
					},
				}
			}

			return &klib.Error{
				ID:     "96dbb73e-d7e4-4c65-aa2e-81ab2d8bf588",
				Status: http.StatusInternalServerError,
				Code:   klib.CodeFileError,
				Path:   fmt.Sprintf(".env.data[%s]", dataKey),
				Title:  "Failed to read file",
				Cause:  err.Error(),
				Meta: map[string]any{
					"filepath": dataFile,
				},
			}
		}

		ext := strings.ToLower(filepath.Ext(dataFile))
		values := make(map[string]any)

		switch ext {
		case ".toml":
			if err := toml.Unmarshal(b, &values); err != nil {
				return &klib.Error{
					ID:     "5705aeef-f9bb-45c8-815c-b8ca07bcaeda",
					Status: http.StatusBadRequest,
					Code:   klib.CodeSerializationError,
					Path:   fmt.Sprintf(".env.data[%s]", dataKey),
					Title:  "Failed to unmarshal toml file",
					Cause:  err.Error(),
				}
			}
		case ".yaml":
			if err := yaml.Unmarshal(b, &values); err != nil {
				return &klib.Error{
					ID:     "c86e9b6f-a41c-4656-855e-7c6f6723efdd",
					Status: http.StatusBadRequest,
					Code:   klib.CodeSerializationError,
					Path:   fmt.Sprintf(".env.data[%s]", dataKey),
					Title:  "Failed to unmarshal yaml file",
					Cause:  err.Error(),
				}
			}
		}

		l.data[k] = values
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

		ext := strings.ToLower(filepath.Ext(overwrite.File))
		file := new(File)

		switch ext {
		case ".toml":
			if err := toml.Unmarshal(b, file); err != nil {
				return false, &klib.Error{
					ID:     "eaac6f84-4e1c-44de-a059-6821006d976a",
					Status: http.StatusBadRequest,
					Code:   klib.CodeSerializationError,
					Path:   path,
					Title:  "Failed to unmarshal toml file",
					Cause:  err.Error(),
				}
			}
		case ".yaml":
			if err := yaml.Unmarshal(b, file); err != nil {
				return false, &klib.Error{
					ID:     "f7476e4b-e592-4f2c-ab97-cae327b957a4",
					Status: http.StatusBadRequest,
					Code:   klib.CodeSerializationError,
					Path:   path,
					Title:  "Failed to unmarshal yaml file",
					Cause:  err.Error(),
				}
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
		keyName := k

		if l.container.caseInsensitiveEnvironment {
			keyName = strings.ToUpper(keyName)
		}

		envVar, haveVar := l.container.env[keyName]
		if !haveVar || envVar.delete {
			return ""
		}

		return envVar.currentValue
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
