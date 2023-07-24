package env

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.katupy.io/klib"
)

type PathHandler interface {
	Add(key, value string, position int) error
}

type defaultPathHandler struct {
	caseSensitiveFilesystem bool

	container *container
}

func (h *defaultPathHandler) Add(key, value string, position int) error {
	cleanValue := filepath.Clean(strings.TrimSpace(value))
	compareValue := cleanValue

	// On case-insensitive filesystems, prevent duplicate paths with different casing.
	if !h.caseSensitiveFilesystem {
		compareValue = strings.ToUpper(cleanValue)
	}

	if h.container.pathListElementExists[key] == nil {
		h.container.pathListElementExists[key] = make(map[string]bool)
	}

	// The provided path is already in the list.
	if h.container.pathListElementExists[key][compareValue] {
		return nil
	}

	h.container.pathListElementExists[key][compareValue] = true

	if !filepath.IsAbs(cleanValue) {
		return &klib.Error{
			ID:     "0d3fb866-c0be-420d-89e7-11b0f05ff132",
			Status: http.StatusBadRequest,
			Code:   klib.CodeInvalidValue,
			Detail: fmt.Sprintf("Path for key %s is not absolute: %s (clean: %s)", key, value, cleanValue),
			Meta: map[string]any{
				"key":        key,
				"value":      value,
				"cleanValue": cleanValue,
			},
		}
	}

	var err error

	h.container.pathListElements[key], err = klib.InsertSliceElem(
		h.container.pathListElements[key],
		cleanValue,
		position,
	)
	if err != nil {
		return klib.ForwardError("570e1095-1818-45f7-aaa8-97b53fa224e3", err)
	}

	return nil
}

type PathLoader interface {
	Load(key string) error
}

type defaultPathLoader struct {
	container *container

	handler PathHandler
}

func (l *defaultPathLoader) Load(key string) error {
	if l.container.pathListExists[key] {
		return nil
	}

	l.container.pathListExists[key] = true

	value := l.container.curEnv[key]

	if value == "" {
		return nil
	}

	elements := strings.Split(value, string(os.PathListSeparator))

	for i := range elements {
		element := elements[i]

		if element == "" {
			continue
		}

		if err := l.handler.Add(key, element, -1); err != nil {
			return klib.ForwardError("b3bf3b89-656b-4882-bdb7-b773d708ea64", err)
		}
	}

	return nil
}
