package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveSource selects explicit CLI input, otherwise explicitly configured policy.
func ResolveSource(explicit, configured string) (string, error) {
	selected := strings.ToLower(strings.TrimSpace(explicit))
	if selected == "" {
		selected = strings.ToLower(strings.TrimSpace(configured))
	}
	if selected == "" {
		return "", fmt.Errorf("telemetry source is required: supply --source codex|opencode or configure telemetry.default_source in the selected local YAML")
	}
	if selected != "opencode" && selected != "codex" {
		return "", fmt.Errorf("unsupported source %q: use opencode or codex", selected)
	}
	return selected, nil
}

func ResolveOpenCodePath(explicit, configured string) (string, error) {
	return resolveSourcePath(explicit, configured, "--db-path", "telemetry.sources.opencode.database_path")
}
func ResolveCodexHome(explicit, configured string) (string, error) {
	return resolveSourcePath(explicit, configured, "--codex-home", "telemetry.sources.codex.home")
}
func resolveSourcePath(explicit, configured, flag, key string) (string, error) {
	value := strings.TrimSpace(explicit)
	if value == "" {
		value = strings.TrimSpace(configured)
	}
	if value == "" {
		return "", fmt.Errorf("source path is required: supply %s or configure %s in the selected local YAML", flag, key)
	}
	return ExpandHomePath(value)
}

// ExpandHomePath interprets explicitly supplied ~/ paths without selecting a store.
func ExpandHomePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return "", fmt.Errorf("expanding explicit path %q: user home is unavailable", path)
		}
		if path == "~" {
			return home, nil
		}
		return home + string(filepath.Separator) + path[2:], nil
	}
	if strings.IndexByte(path, 0) >= 0 {
		return "", fmt.Errorf("path contains a NUL byte")
	}
	return path, nil
}

// ResolveConfiguredPath retains physical symlink/.. traversal by concatenating
// relative components without filepath.Clean or filepath.Join.
func ResolveConfiguredPath(origin, path string) (string, error) {
	expanded, err := ExpandHomePath(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		return origin + string(filepath.Separator) + expanded, nil
	}
	return expanded, nil
}
