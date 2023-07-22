package env

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.katupy.io/klib"

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
				ID:     "01H5XE3N8EG0EER1J8ZSBB341D",
				Status: http.StatusBadRequest,
				Code:   klib.ErrMissingArgument,
				Path:   ".",
			},
		},
		{
			name:   "nil-env",
			config: &conf.Config{},
			err: &klib.Error{
				ID:     "01H5XE7HV41CKBWXM2KWQE2E0V",
				Status: http.StatusBadRequest,
				Code:   klib.ErrMissingArgument,
				Path:   ".env",
			},
		},
		{
			name: "nil-env-load",
			config: &conf.Config{
				Env: &conf.Env{},
			},
			err: &klib.Error{
				ID:     "01H5XEGQ0B1003AC3YWTW74TCD",
				Status: http.StatusBadRequest,
				Code:   klib.ErrMissingArgument,
				Path:   ".env.load",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			if err := Load(tc.config); err != nil {
				if tc.err == nil {
					st.Fatalf("failed to Load: %s", err)
				}

				wantErr := tc.err
				haveErr := err.(*klib.Error)

				assert.Equal(st, wantErr.ID, haveErr.ID, "err.ID mismatch")
				assert.Equal(st, wantErr.Status, haveErr.Status, "err.Status mismatch")
				assert.Equal(st, wantErr.Code, haveErr.Code, "err.Code mismatch")
				assert.Equal(st, wantErr.Path, haveErr.Path, "err.Path mismatch")

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
		err    *klib.Error
	}{
		{
			name: "empty-env-load-dir",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{},
				},
			},
			err: &klib.Error{
				ID:     "01H5XEMMAB72AYK72CEV6JN09N",
				Status: http.StatusBadRequest,
				Code:   klib.ErrMissingArgument,
				Path:   ".env.load.dir",
			},
		},
		{
			name: "empty-env-load-filename",
			config: &conf.Config{
				Env: &conf.Env{
					Load: &conf.EnvLoad{
						Dir: ".",
					},
				},
			},
			err: &klib.Error{
				ID:     "01H5XEND4HEPG9GDKMCQ1P2VGH",
				Status: http.StatusBadRequest,
				Code:   klib.ErrMissingArgument,
				Path:   ".env.load.filename",
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			files, err := GetFiles(tc.config)
			if err != nil {
				if tc.err == nil {
					st.Fatalf("failed to GetFiles: %s", err)
				}

				wantErr := tc.err
				haveErr := err.(*klib.Error)

				assert.Equal(st, wantErr.ID, haveErr.ID, "err.ID mismatch")
				assert.Equal(st, wantErr.Status, haveErr.Status, "err.Status mismatch")
				assert.Equal(st, wantErr.Code, haveErr.Code, "err.Code mismatch")
				assert.Equal(st, wantErr.Path, haveErr.Path, "err.Path mismatch")

				return
			}

			_ = files
		})
	}
}
