package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/filesystem"
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

// CodexInboxName is the directory explicitly granted by the user installer.
const CodexInboxName = "agenticlens-binding-inbox"

// RuntimeCodexHome identifies the native client's user-controlled settings home.
func RuntimeCodexHome() string {
	if home := os.Getenv("CODEX_HOME"); home != "" {
		return home
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex")
}

// OutsideCodexInbox returns the physical executable path checked by the hook.
// Dispatch uses this same path, including when the selected artifact is a symlink.
func OutsideCodexInbox(path string, homes ...string) (string, error) {
	physical, err := filesystem.PhysicalPath(path)
	if err != nil {
		return "", err
	}
	for _, home := range homes {
		if home == "" {
			continue
		}
		parent, err := filesystem.PhysicalPath(home)
		if err != nil {
			return "", err
		}
		inbox := filepath.Join(parent, CodexInboxName)
		// A previously granted directory may have been relocated via a
		// symlink since installation. Exclude the current physical destination.
		inbox, err = filesystem.PhysicalPath(inbox)
		if err != nil {
			return "", err
		}
		if physical == inbox || strings.HasPrefix(physical, strings.TrimRight(inbox, string(filepath.Separator))+string(filepath.Separator)) {
			return "", fmt.Errorf("native executable %s is inside the Codex binding inbox", physical)
		}
	}
	return physical, nil
}
