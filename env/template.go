package env

import (
	"bytes"
	"net/http"
	"text/template"

	"go.katupy.io/klib"
)

type templateHandler struct {
	buf     *bytes.Buffer
	data    map[string]any
	funcMap template.FuncMap
}

func (h *templateHandler) Handle(input string) (string, error) {
	t, err := template.New("").Funcs(h.funcMap).Parse(input)
	if err != nil {
		return "", &klib.Error{
			ID:     "f99a56d8-bbd5-4c53-a9c7-b5cf3ea5c0e9",
			Status: http.StatusBadRequest,
			Code:   klib.CodeParseError,
			Title:  "Failed to parse template",
			Cause:  err.Error(),
		}
	}

	if h.buf == nil {
		h.buf = new(bytes.Buffer)
	} else {
		h.buf.Reset()
	}

	if err := t.Execute(h.buf, h.data); err != nil {
		return "", &klib.Error{
			ID:     "6c9d0b27-3026-423e-93a5-10697a252fd8",
			Status: http.StatusInternalServerError,
			Code:   klib.CodeExecutionError,
			Title:  "Failed to execute template",
			Cause:  err.Error(),
		}
	}

	return h.buf.String(), nil
}
