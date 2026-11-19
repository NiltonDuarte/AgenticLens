package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

const (
	pendingDirectoryKey     = "agenticlens.pending_dir"
	defaultPendingDirectory = ".agenticlens/pending/"
)

func validatePendingDirectoryValue(value string) error {
	if strings.TrimSpace(value) == "" || len(value) > 4096 || strings.ContainsAny(value, "\x00\r\n") {
		return fmt.Errorf("%s must be an explicit nonblank single-line directory path", pendingDirectoryKey)
	}
	return nil
}

func pendingDirectoryValue(node *yaml.Node) (string, error) {
	for node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	if err := requireNodeType(node, pendingDirectoryKey, yaml.ScalarNode, "!!str"); err != nil {
		return "", err
	}
	return node.Value, validatePendingDirectoryValue(node.Value)
}

func pendingDirectoryNode(nodes map[string]*yaml.Node) *yaml.Node {
	entrypoint := nodes["agenticlens"]
	for entrypoint != nil && entrypoint.Kind == yaml.AliasNode {
		entrypoint = entrypoint.Alias
	}
	if entrypoint != nil && entrypoint.Kind == yaml.MappingNode {
		for i := 0; i < len(entrypoint.Content); i += 2 {
			if entrypoint.Content[i].Value == "pending_dir" {
				return entrypoint.Content[i+1]
			}
		}
	}
	return nil
}

// ResolvePendingDirectory validates the explicitly selected pending snapshot
// directory without creating it or interpreting missing values as defaults.
func ResolvePendingDirectory(cfg domain.ProjectConfig, project string) (string, error) {
	if err := validatePendingDirectoryValue(cfg.AgenticLens.PendingDir); err != nil {
		return "", err
	}
	root, err := ProjectRoot(project)
	if err != nil {
		return "", fmt.Errorf("%s project root: %w", pendingDirectoryKey, err)
	}
	return resolvePendingDirectory(cfg, root)
}

func resolvePendingDirectory(cfg domain.ProjectConfig, root string) (string, error) {
	selected, err := ResolveConfiguredPath(root, cfg.AgenticLens.PendingDir)
	if err != nil {
		return "", fmt.Errorf("%s: %w", pendingDirectoryKey, err)
	}
	dir, err := filesystem.PhysicalPath(selected)
	if err != nil {
		return "", fmt.Errorf("%s: %w", pendingDirectoryKey, err)
	}
	if dir == root || !insidePath(dir, root) {
		return "", fmt.Errorf("%s must select a directory inside the project", pendingDirectoryKey)
	}
	for entry := dir; entry != root; entry = filepath.Dir(entry) {
		if strings.EqualFold(filepath.Base(entry), ".git") {
			return "", fmt.Errorf("%s cannot overlap Git metadata", pendingDirectoryKey)
		}
	}
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		return "", fmt.Errorf("%s must select a directory", pendingDirectoryKey)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%s: %w", pendingDirectoryKey, err)
	}
	origin := root
	if cfg.LocalConfigPath != "" {
		origin = filepath.Dir(cfg.LocalConfigPath)
	}
	for _, private := range []struct{ key, value string }{
		{"agenticlens.state_dir", cfg.AgenticLens.StateDir},
		{"telemetry.capture.state_dir", cfg.Telemetry.Capture.StateDir},
	} {
		if private.value == "" {
			continue
		}
		path, err := ResolveConfiguredPath(origin, private.value)
		if err == nil {
			path, err = filesystem.PhysicalPath(path)
		}
		if err != nil {
			return "", fmt.Errorf("%s: checking %s: %w", pendingDirectoryKey, private.key, err)
		}
		if insidePath(dir, path) || insidePath(path, dir) {
			return "", fmt.Errorf("%s must not overlap %s", pendingDirectoryKey, private.key)
		}
	}
	return dir, nil
}

// PendingDirectoryForSetup retains the target's existing choice unless the
// approved settings supply an explicit replacement. Only setup selects a default.
func PendingDirectoryForSetup(outputPath string, cfg domain.ProjectConfig) (string, error) {
	value := cfg.AgenticLens.PendingDir
	if value == "" && !cfg.Settings[pendingDirectoryKey] {
		var err error
		value, err = readPendingDirectoryValue(outputPath)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = defaultPendingDirectory
		}
	}
	cfg.AgenticLens.PendingDir = value
	if err := validatePendingDirectoryAtOutput(outputPath, cfg); err != nil {
		return "", err
	}
	return value, nil
}

func readPendingDirectoryValue(sharedPath string) (string, error) {
	body, err := readBootstrap(sharedPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading current %s: %w", pendingDirectoryKey, err)
	}
	nodes, err := configurationNodes(body)
	if err != nil {
		return "", fmt.Errorf("reading current %s: %w", pendingDirectoryKey, err)
	}
	if node := pendingDirectoryNode(nodes); node != nil {
		return pendingDirectoryValue(node)
	}
	return "", nil
}

// ValidatePendingDirectoryForSetup checks the selected pending directory against
// the effective local YAML that setup will write, using its final path as origin.
func ValidatePendingDirectoryForSetup(sharedPath, localPath, pendingDir string, localYAML []byte) error {
	entry, err := parseEntrypoint(localYAML, true)
	if err != nil {
		return fmt.Errorf("local configuration %s: %w", localPath, err)
	}
	captureState, _, err := captureStateValue(localYAML)
	if err != nil {
		return fmt.Errorf("local configuration %s: %w", localPath, err)
	}
	entry.PendingDir = pendingDir
	return validatePendingDirectoryAtOutput(sharedPath, domain.ProjectConfig{
		AgenticLens:     entry,
		LocalConfigPath: localPath,
		Telemetry:       domain.TelemetryConfig{Capture: domain.CaptureConfig{StateDir: captureState}},
	})
}

// ValidateExistingPendingDirectoryForSetup checks an explicitly configured shared
// pending directory before a local-only edit. Legacy shared YAML can omit it.
func ValidateExistingPendingDirectoryForSetup(sharedPath, localPath string, localYAML []byte) error {
	value, err := readPendingDirectoryValue(sharedPath)
	if err != nil || value == "" {
		return err
	}
	return ValidatePendingDirectoryForSetup(sharedPath, localPath, value, localYAML)
}

func validatePendingDirectoryAtOutput(outputPath string, cfg domain.ProjectConfig) error {
	if err := validatePendingDirectoryValue(cfg.AgenticLens.PendingDir); err != nil {
		return err
	}
	root, err := filesystem.PhysicalPath(filepath.Dir(outputPath))
	if err != nil {
		return fmt.Errorf("%s project root: %w", pendingDirectoryKey, err)
	}
	_, err = resolvePendingDirectory(cfg, root)
	return err
}
