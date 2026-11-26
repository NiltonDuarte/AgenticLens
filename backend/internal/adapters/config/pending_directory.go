package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	if err := validatePendingPlacement(root, dir); err != nil {
		return "", err
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

func validatePendingPlacement(root, directory string) error {
	if directory == root || !insidePath(directory, root) {
		return fmt.Errorf("%s must select a directory inside the project", pendingDirectoryKey)
	}
	for entry := directory; entry != root; entry = filepath.Dir(entry) {
		if strings.EqualFold(filepath.Base(entry), ".git") {
			return fmt.Errorf("%s cannot overlap Git metadata", pendingDirectoryKey)
		}
	}
	return nil
}

// ResolveStagedPendingDirectories maps the shared selection to literal Git
// paths. The configured path stays usable for index reads after its working
// copy is removed or replaced. A valid physical alias can also hold snapshots.
func ResolveStagedPendingDirectories(project string) ([]string, error) {
	shared, err := FindConfigFile(project)
	if err != nil {
		return nil, err
	}
	if shared == "" {
		return nil, ErrConfigNotFound
	}
	value, err := readPendingDirectoryValue(shared)
	if err != nil {
		return nil, err
	}
	if err := validatePendingDirectoryValue(value); err != nil {
		return nil, err
	}
	selected, err := ResolveConfiguredPath(project, value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", pendingDirectoryKey, err)
	}
	logical := filepath.Clean(selected)
	candidates := []string{logical}
	physical, physicalErr := filesystem.PhysicalPath(selected)
	if traversesPendingParent(selected) {
		// Cleaning before resolving changes parent traversal after a symlink.
		// Such a lexical path cannot identify a second snapshot directory.
		if physicalErr != nil {
			return nil, fmt.Errorf("resolve staged %s: %w", pendingDirectoryKey, physicalErr)
		}
		candidates = []string{physical}
	} else if physicalErr == nil && physical != logical {
		if info, err := os.Stat(physical); errors.Is(err, os.ErrNotExist) || err == nil && info.IsDir() {
			candidates = append(candidates, physical)
		}
	}
	private := stagedPendingPrivatePaths(project)
	var paths []string
	var rejected error
	for _, candidate := range candidates {
		if err := validatePendingPlacement(project, candidate); err != nil {
			rejected = errors.Join(rejected, err)
			continue
		}
		if err := rejectPendingPrivateOverlap(candidate, private); err != nil {
			rejected = errors.Join(rejected, err)
			continue
		}
		relative, err := filepath.Rel(project, candidate)
		if err != nil {
			return nil, fmt.Errorf("resolve staged %s: %w", pendingDirectoryKey, err)
		}
		paths = append(paths, filepath.ToSlash(relative))
	}
	if len(paths) == 0 {
		return nil, rejected
	}
	return paths, nil
}

// Only known private placements can disqualify an index path. A missing or
// inaccessible local path leaves the shared staged selection usable.
func stagedPendingPrivatePaths(project string) []string {
	selection, err := SelectConfig(project)
	if err != nil {
		return nil
	}
	data, err := readSelectedLocal(selection)
	if err != nil {
		return nil
	}
	origin := filepath.Dir(selection.LocalPath)
	var values []string
	if entrypoint, err := parseEntrypoint(data, true); err == nil {
		values = append(values, entrypoint.StateDir)
	}
	if captureState, present, err := captureStateValue(data); err == nil && present {
		values = append(values, captureState)
	}
	var paths []string
	for _, value := range values {
		if value == "" {
			continue
		}
		path, err := ResolveConfiguredPath(origin, value)
		if err != nil {
			continue
		}
		physical, physicalErr := filesystem.PhysicalPath(path)
		if traversesPendingParent(path) {
			if physicalErr == nil {
				paths = append(paths, physical)
			}
			continue
		}
		logical := filepath.Clean(path)
		paths = append(paths, logical)
		if physicalErr == nil && physical != logical {
			paths = append(paths, physical)
		}
	}
	return paths
}

func traversesPendingParent(path string) bool {
	return slices.Contains(strings.Split(filepath.ToSlash(path), "/"), "..")
}

func rejectPendingPrivateOverlap(directory string, private []string) error {
	for _, path := range private {
		if insidePath(directory, path) || insidePath(path, directory) {
			return fmt.Errorf("%s must not overlap private state", pendingDirectoryKey)
		}
	}
	return nil
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
