package env

import (
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.katupy.io/klib"
)

func Test_templateHandler_Handle(t *testing.T) {
	testCases := []*struct {
		name            string
		templateHandler *templateHandler
		input           string
		err             *klib.Error
		output          string
	}{
		{
			name:            "parse-error",
			templateHandler: &templateHandler{},
			input:           "{{ unknownFunc }}",
			err: &klib.Error{
				ID:     "f99a56d8-bbd5-4c53-a9c7-b5cf3ea5c0e9",
				Status: http.StatusBadRequest,
				Code:   klib.CodeParseError,
			},
		},
		{
			name: "execution-error",
			templateHandler: &templateHandler{
				funcMap: klib.BaseFuncMap(),
			},
			input: "{{ .UNKNOWN_VAR | env }}",
			err: &klib.Error{
				ID:     "6c9d0b27-3026-423e-93a5-10697a252fd8",
				Status: http.StatusInternalServerError,
				Code:   klib.CodeExecutionError,
			},
		},
		{
			name:            "empty",
			templateHandler: &templateHandler{},
		},
		{
			name: "with-data",
			templateHandler: &templateHandler{
				data: map[string]any{
					"_GOOS":  runtime.GOOS,
					"goarch": runtime.GOARCH,
				},
			},
			input:  `{{ ._GOOS }}_{{ .goarch }}`,
			output: fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d:%s", i, tc.name), func(st *testing.T) {
			have, err := tc.templateHandler.Handle(tc.input)
			if klib.CheckTestError(st, err, tc.err) {
				return
			}

			want := tc.output

			assert.Equal(st, want, have, "Template output mismatch")
		})
	}
}
