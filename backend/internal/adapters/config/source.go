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
			return home + string(filepath.Separator) + path[2:]
		}
	}
	return path
}

// ProjectSourcePath anchors configured relative source locations to their owning
// project. Explicit CLI paths retain the caller's normal path semantics.
func ProjectSourcePath(project, configured string) string {
	if strings.TrimSpace(configured) == "" {
		return ""
	}
	path := ExpandHomePath(configured)
	if !filepath.IsAbs(path) {
		// Preserve raw symlink/.. traversal until the physical source boundary.
		return project + string(filepath.Separator) + path
	}
	return path
}
