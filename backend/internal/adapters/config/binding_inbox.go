package config

import (
	"errors"
	"os"
	"path/filepath"
)

// BindingStatePaths resolves the capture and publication directories that must
// remain outside the sandbox-writable inbox. Installation also supports homes
// prepared before a project has configured either state directory.
func BindingStatePaths(project string) ([]string, error) {
	selection, err := SelectConfig(project)
	if errors.Is(err, ErrConfigNotFound) || errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	data, err := readSelectedLocal(selection)
	if err != nil {
		return nil, err
	}
	capture, _, err := captureStateValue(data)
	if err != nil {
		return nil, err
	}
	entry, err := parseEntrypoint(data, true)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, raw := range []string{capture, entry.StateDir} {
		if raw != "" {
			path, err := ResolveConfiguredPath(filepath.Dir(selection.LocalPath), raw)
			if err != nil {
				return nil, err
			}
			paths = append(paths, path)
		}
	}
	return paths, nil
}
