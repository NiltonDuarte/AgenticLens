package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveSource selects one telemetry harness without probing either source.
func ResolveSource(explicit, configured string) (string, error) {
	selected := strings.ToLower(strings.TrimSpace(explicit))
	if selected == "" {
		selected = strings.ToLower(strings.TrimSpace(configured))
	}
	if selected == "" {
		selected = "opencode"
	}
	if selected != "opencode" && selected != "codex" {
		return "", fmt.Errorf("unsupported source %q: use opencode or codex", selected)
	}
	return selected, nil
}

// ResolveOpenCodePath applies explicit, local, shared-legacy, then adapter-default precedence.
func ResolveOpenCodePath(explicit, local, legacy string) string {
	for _, candidate := range []string{explicit, local, legacy} {
		if value := strings.TrimSpace(candidate); value != "" {
			return ExpandHomePath(value)
		}
	}
	return ""
}

// ResolveCodexHome applies explicit, local, environment, then built-in precedence.
func ResolveCodexHome(explicit, local, environment, builtin string) string {
	for _, candidate := range []string{explicit, local, environment, builtin} {
		if value := strings.TrimSpace(candidate); value != "" {
			return ExpandHomePath(value)
		}
	}
	return ""
}

// ExpandHomePath expands the supported ~/ notation at an imperative boundary.
func ExpandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
