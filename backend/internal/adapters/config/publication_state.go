package config

import (
	"fmt"
	"path/filepath"
)

// ResolvePublicationState reads only the operational state choice, never capture registry state.
// It does not create the directory; only a writer initializes operational state.
func ResolvePublicationState(project string) (string, error) {
	selection, err := SelectConfig(project)
	if err != nil {
		return "", err
	}
	body, err := readSelectedLocal(selection)
	if err != nil {
		return "", err
	}
	entry, err := parseEntrypoint(body, true)
	if err != nil {
		return "", err
	}
	if entry.StateDir == "" {
		return "", fmt.Errorf("local agenticlens.state_dir must be explicitly configured; run setup before publication")
	}
	path, err := ResolveConfiguredPath(filepath.Dir(selection.LocalPath), entry.StateDir)
	if err != nil {
		return "", fmt.Errorf("agenticlens.state_dir: %w", err)
	}
	return path, nil
}
