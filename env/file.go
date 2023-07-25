package env

import (
	"net/http"
	"os"

	"go.katupy.io/klib"
)

type FileLoader interface {
	Load(file *File) error
}

type defaultFileLoader struct {
	commandLoader CommandLoader
}

func (l *defaultFileLoader) Load(file *File) error {
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
		cmd := file.Commands[i]
		cmd.file = file
		cmd.index = i

		if err := l.commandLoader.Load(cmd); err != nil {
			return klib.ForwardError("f1b51039-fc62-4f06-b9c6-965a1b7dcc66", err)
		}
	}

	return nil
}
