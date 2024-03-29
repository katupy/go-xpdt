package env

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/must"

	"go.katupy.io/xpdt/conf"
)

func TestLoader_Load(t *testing.T) {
	testCases := []*struct {
		name   string
		config *conf.Config
		err    *klib.Error
	}{
		{
			name: "nil-config",
			err: &klib.Error{
				ID:     "df006438-8216-4fbf-a4b6-3c4f933a6c0d",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Path:   ".",
			},
		},
		{
			name:   "nil-env",
			config: &conf.Config{},
			err: &klib.Error{
				ID:     "0002b99d-191f-4bb1-9120-d5853df954c9",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Path:   ".env",
			},
		},
		{
			name: "nil-env-load",
			config: &conf.Config{
				Env: &conf.Env{},
			},
			err: &klib.Error{
				ID:     "13738af0-94ef-4589-bf2d-e11f6873f2cd",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Path:   ".env.load",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			loader := &Loader{
				config: tc.config,
			}

			err := loader.Load()
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}

func TestLoader_FindFiles(t *testing.T) {
	testCases := []*struct {
		name   string
		config *conf.Config
		files  []*File
		err    *klib.Error
	}{
		{
			name: "empty-env-overwrite-dir",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{},
					Overwrites: []*conf.EnvOverwrite{
						{},
					},
				},
			},
			err: &klib.Error{
				ID:     "dc558b94-ddf9-4280-8eab-94e16cc79418",
				Status: http.StatusBadRequest,
				Code:   klib.CodeMissingValue,
				Path:   ".env.overwrites[0].dir",
			},
		},
		{
			name: "non-absolute-env-overwrite-dir",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir: must.FilepathAbs("."),
						},
						{
							Dir: ".",
						},
					},
				},
			},
			err: &klib.Error{
				ID:     "8ee66ad2-84e7-4aa1-8151-55af5c6f4b7d",
				Status: http.StatusBadRequest,
				Code:   klib.CodeInvalidValue,
				Path:   ".env.overwrites[1].dir",
			},
		},
		{
			name: "file-not-found",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir:  must.FilepathAbs("."),
							File: uuid.NewString(),
						},
					},
				},
			},
			err: &klib.Error{
				ID:     "abcd4eb4-6a53-49f2-b4ce-3e6bb8562609",
				Status: http.StatusBadRequest,
				Code:   klib.CodeNotFound,
				Path:   ".env.overwrites[0].file",
			},
		},
		{
			name: "failed-to-unmarshal",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir:      "tests",
						Filename: "failed-to-unmarshal.yaml",
					},
				},
			},
			err: &klib.Error{
				ID:     "f7476e4b-e592-4f2c-ab97-cae327b957a4",
				Status: http.StatusBadRequest,
				Code:   klib.CodeSerializationError,
			},
		},
		{
			name: "no-env-files",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Filename: uuid.NewString(),
					},
				},
			},
		},
		{
			name: "single-root-file",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "single-root-file", "1", "2"),
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "single-root-file", "1", "2")),
					filepath: must.FilepathAbs(filepath.Join("tests", "single-root-file", "1", "2", conf.DefaultEnvLoadFilename)),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "single-root-file", "1")),
					filepath: must.FilepathAbs(filepath.Join("tests", "single-root-file", "1", conf.DefaultEnvLoadFilename)),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "single-root-file")),
					filepath: must.FilepathAbs(filepath.Join("tests", "single-root-file", conf.DefaultEnvLoadFilename)),
				},
			},
		},
		{
			name: "multiple-root-files",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "multiple-root-files", "1"),
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-root-files", "1")),
					filepath: must.FilepathAbs(filepath.Join("tests", "multiple-root-files", "1", conf.DefaultEnvLoadFilename)),
				},
			},
		},
		{
			name: "overwrite-root-file",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "overwrite-root-file", "1", "2"),
					},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "overwrite-root-file", "1")),
							File: filepath.Join("tests", "overwrite-root-file", "overwrite.yaml"),
							Root: true,
						},
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "overwrite-root-file", "1", "2")),
					filepath: must.FilepathAbs(filepath.Join("tests", "overwrite-root-file", "1", "2", conf.DefaultEnvLoadFilename)),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "overwrite-root-file", "1")),
					filepath: filepath.Join("tests", "overwrite-root-file", "overwrite.yaml"),
				},
			},
		},
		{
			name: "overwrite-skip",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "overwrite-skip", "1", "2"),
					},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "overwrite-skip", "1")),
							Skip: true,
						},
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "overwrite-skip", "1", "2")),
					filepath: must.FilepathAbs(filepath.Join("tests", "overwrite-skip", "1", "2", conf.DefaultEnvLoadFilename)),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "overwrite-skip")),
					filepath: must.FilepathAbs(filepath.Join("tests", "overwrite-skip", conf.DefaultEnvLoadFilename)),
				},
			},
		},
		{
			name: "multiple-overwrites-same-dir",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "multiple-overwrites-same-dir", "1", "2"),
					},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
							File: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-1.yaml"),
						},
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
							File: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-2.yaml"),
						},
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
							File: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-3.yaml"),
						},
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1", "2")),
					filepath: must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1", "2", conf.DefaultEnvLoadFilename)),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
					filepath: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-3.yaml"),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
					filepath: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-2.yaml"),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", "1")),
					filepath: filepath.Join("tests", "multiple-overwrites-same-dir", "overwrite-1.yaml"),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir")),
					filepath: must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-same-dir", conf.DefaultEnvLoadFilename)),
				},
			},
		},
		{
			name: "multiple-overwrites-intermediate-root",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: filepath.Join("tests", "multiple-overwrites-intermediate-root", "1"),
					},
					Overwrites: []*conf.EnvOverwrite{
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-intermediate-root", "1")),
							File: filepath.Join("tests", "multiple-overwrites-intermediate-root", "overwrite-1.yaml"),
						},
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-intermediate-root", "1")),
							File: filepath.Join("tests", "multiple-overwrites-intermediate-root", "overwrite-2.yaml"),
						},
						{
							Dir:  must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-intermediate-root", "1")),
							File: filepath.Join("tests", "multiple-overwrites-intermediate-root", "overwrite-3.yaml"),
						},
					},
				},
			},
			files: []*File{
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-intermediate-root", "1")),
					filepath: filepath.Join("tests", "multiple-overwrites-intermediate-root", "overwrite-2.yaml"),
				},
				{
					dir:      must.FilepathAbs(filepath.Join("tests", "multiple-overwrites-intermediate-root", "1")),
					filepath: filepath.Join("tests", "multiple-overwrites-intermediate-root", "overwrite-1.yaml"),
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			loader := &Loader{
				config: tc.config,
			}

			err := loader.FindFiles()
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			haveFiles := loader.files
			wantFiles := tc.files

			if assert.Equal(st, len(wantFiles), len(haveFiles), "Files length mismatch") {
				for i := range wantFiles {
					have := haveFiles[i]
					want := wantFiles[i]

					assert.Equal(st, want.dir, have.dir, "File[%d].dir mismatch", i)
					assert.Equal(st, want.filepath, have.filepath, "File[%d].filepath mismatch", i)
				}
			}
		})
	}
}
