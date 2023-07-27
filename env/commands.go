package env

import (
	"net/http"
	"strings"

	"go.katupy.io/klib"
)

type CommandLoader interface {
	Load(cmd *Command) error
}

type defaultCommandLoader struct {
	platform       string
	commandMethods CommandMethods
}

func (l *defaultCommandLoader) Load(cmd *Command) error {
	if cmd.Platform != "" && cmd.Platform != l.platform {
		return nil
	}

	var cmdFunc func(*Command) error

	switch {
	case cmd.Add != "":
		cmdFunc = l.commandMethods.Add
	case cmd.Set != "":
		cmdFunc = l.commandMethods.Set
	case cmd.Del != "":
		cmdFunc = l.commandMethods.Del
	}

	if err := cmdFunc(cmd); err != nil {
		return klib.ForwardError("c2ebf96d-8171-4d4b-8853-eb4cf48d7c7f", err)
	}

	return nil
}

type CommandMethods interface {
	Add(cmd *Command) error
	Set(cmd *Command) error
	Del(cmd *Command) error
}

type defaultCommandMethods struct {
	container *container

	pathHandler     PathHandler
	pathLoader      PathLoader
	templateHandler klib.StringHandler
}

func (m *defaultCommandMethods) Add(cmd *Command) error {
	var values []string

	switch {
	case cmd.Value != "":
		value, err := m.templateHandler.Handle(cmd.Value)
		if err != nil {
			return klib.ForwardError("8f6fe0e7-9037-4b26-94a5-83633ea0c142", err)
		}

		values = []string{value}
	default:
		return &klib.Error{
			ID:     "ce66adb8-2e56-40b7-8268-27a8955296b5",
			Status: http.StatusBadRequest,
			Code:   klib.CodeMissingValue,
			// Path:   fmt.Sprintf("file[%d].commands[%d]", fileIndex, cmdIndex),
			Title: "Missing value",
		}
	}

	key := cmd.Add
	keyName := key

	if m.container.caseInsensitiveEnvironment {
		keyName = strings.ToUpper(keyName)
	}

	envVar, haveVar := m.container.env[keyName]
	if !haveVar {
		envVar = &environVar{
			key:     key,
			created: true,
		}
		m.container.env[keyName] = envVar
	}

	// Ensure key persists if it was deleted before.
	if envVar.delete {
		envVar.delete = false
	}

	if err := m.pathLoader.Load(envVar); err != nil {
		return klib.ForwardError("bfb999a7-55af-47ab-a8b3-bc15be757c48", err)
	}

	for i := range values {
		var index int

		if cmd.Append {
			index -= 1
		}

		if err := m.pathHandler.Add(envVar, values[i], index); err != nil {
			return klib.ForwardError("4aa49cf3-1289-403a-bbb2-b25d6ad84a4c", err)
		}
	}

	return nil
}

func (m *defaultCommandMethods) Set(cmd *Command) error {
	value, err := m.templateHandler.Handle(cmd.Value)
	if err != nil {
		return klib.ForwardError("03ba5588-7ed1-43c9-b78e-36817c63b4e0", err)
	}

	key := cmd.Set
	keyName := key

	if m.container.caseInsensitiveEnvironment {
		keyName = strings.ToUpper(keyName)
	}

	envVar, haveVar := m.container.env[keyName]
	if !haveVar {
		envVar = &environVar{
			key:     key,
			created: true,
		}
		m.container.env[keyName] = envVar
	}

	envVar.currentValue = value

	// Ensure key persists if it was deleted before.
	if envVar.delete {
		envVar.delete = false
	}

	return nil
}

func (m *defaultCommandMethods) Del(cmd *Command) error {
	key := cmd.Del
	keyName := key

	if m.container.caseInsensitiveEnvironment {
		keyName = strings.ToUpper(keyName)
	}

	if key == "*" {
		for _, envVar := range m.container.env {
			envVar.resetAndDelete()
		}
	} else if envVar, haveVar := m.container.env[keyName]; haveVar {
		envVar.resetAndDelete()
	}

	return nil
}
