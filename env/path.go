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
	Add(envVar *environVar, value string, position int) error
}

type defaultPathHandler struct {
	caseSensitiveFilesystem bool
}

func (h *defaultPathHandler) Add(envVar *environVar, value string, index int) error {
	cleanValue, err := filepath.Abs(strings.TrimSpace(value))
	if err != nil {
		return &klib.Error{
			ID:     "0d3fb866-c0be-420d-89e7-11b0f05ff132",
			Status: http.StatusInternalServerError,
			Code:   klib.CodeFilesystemError,
			Detail: fmt.Sprintf("Failed to calculate absolute path value for key %s: %s.", envVar.key, value),
			Cause:  err.Error(),
			Meta: map[string]any{
				"key":   envVar.key,
				"value": value,
			},
		}
	}

	compareValue := cleanValue

	// On case-insensitive filesystems, prevent duplicate paths with different casing.
	if !h.caseSensitiveFilesystem {
		compareValue = strings.ToUpper(cleanValue)
	}

	if envVar.pathListElementExists == nil {
		envVar.pathListElementExists = make(map[string]bool)
	}

	// The provided path is already in the list.
	if envVar.pathListElementExists[compareValue] {
		return nil
	}

	envVar.pathListElementExists[compareValue] = true

	envVar.pathListElements, err = klib.InsertSliceElem(
		envVar.pathListElements,
		cleanValue,
		index,
	)
	if err != nil {
		return klib.ForwardError("570e1095-1818-45f7-aaa8-97b53fa224e3", err)
	}

	return nil
}

type PathLoader interface {
	Load(envVar *environVar) error
}

type defaultPathLoader struct {
	pathHandler PathHandler
}

func (l *defaultPathLoader) Load(envVar *environVar) error {
	if envVar.pathList {
		return nil
	}

	envVar.pathList = true
	value := envVar.currentValue

	if value == "" {
		return nil
	}

	// Ensure there will be a diff for this key.
	envVar.currentValue = ""
	elements := strings.Split(value, string(os.PathListSeparator))

	for i := range elements {
		element := elements[i]

		if element == "" {
			continue
		}

		if err := l.pathHandler.Add(envVar, element, -1); err != nil {
			return klib.ForwardError("b3bf3b89-656b-4882-bdb7-b773d708ea64", err)
		}
	}

	return nil
}
