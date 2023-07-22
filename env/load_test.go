package env

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"
	"go.katupy.io/klib/must"

	"go.katupy.io/xpdt/conf"
)

func TestLoad(t *testing.T) {
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
			err := Load(tc.config)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}
		})
	}
}

func TestGetEnviron(t *testing.T) {
	testCases := []*struct {
		name       string
		environ    []string
		ignoreCase bool
		want       map[string]string
	}{
		{
			name: "empty-environ",
			want: map[string]string{},
		},
		{
			name: "case-sensitive",
			environ: []string{
				"foo=bar1",
				"Foo=Bar2",
				"HAVE==StartingEqual",
			},
			want: map[string]string{
				"foo":  "bar1",
				"Foo":  "Bar2",
				"HAVE": "=StartingEqual",
			},
		},
		{
			name: "case-insensitive",
			environ: []string{
				"foo=bar1",
				"Foo=Bar2",
				"have=two=intermediary=equals",
			},
			ignoreCase: true,
			want: map[string]string{
				"FOO":  "Bar2",
				"HAVE": "two=intermediary=equals",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			have := GetEnviron(tc.environ, tc.ignoreCase)
			want := tc.want

			if assert.Equal(st, len(want), len(have), "Length mismatch") {
				for k := range want {
					assert.Equal(st, want[k], have[k], "Value[%s] mismatch", k)
				}
			}
		})
	}
}

func TestGetFiles(t *testing.T) {
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
						Dir:      "tests",
						Filename: "single-root-file.yaml",
					},
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			haveFiles, err := GetFiles(tc.config)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			wantFiles := tc.files

			if assert.Equal(st, len(wantFiles), len(haveFiles), "Files length mismatch") {
				for i := range wantFiles {
					want := wantFiles[i]
					have := haveFiles[i]

					assert.Equal(st, want.dir, have.dir, "File[%d].dir mismatch", i)
				}
			}
		})
	}
}
