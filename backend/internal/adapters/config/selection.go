package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters/filesystem"
	"gopkg.in/yaml.v3"
)

// SelectedConfig records the whole-file machine policy and its path origin.
// Selection is read-only and never creates configuration or capture state.
type SelectedConfig struct {
	Project     string `json:"project"`
	SharedPath  string `json:"shared_path"`
	LocalPath   string `json:"local_path"`
	LocalOrigin string `json:"local_origin"`
}

// SelectConfig prefers one complete worktree local file, otherwise the base local
// file. In particular, a present invalid local file cannot trigger fallback.
func SelectConfig(start string) (SelectedConfig, error) {
	var result SelectedConfig
	shared, err := FindConfigFile(start)
	if err != nil {
		return result, err
	}
	if shared == "" {
		return result, ErrConfigNotFound
	}
	result.Project = filepath.Dir(shared)
	result.SharedPath = shared
	result.LocalPath = filepath.Join(result.Project, ".agenticlens-config.local.yaml")
	result.LocalOrigin = "worktree"
	if _, err := os.Lstat(result.LocalPath); err == nil {
		return result, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("local configuration %s: %w", result.LocalPath, err)
	}
	// Only Git metadata establishes the base; directory naming is not authority.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := filesystem.RepositoryGitCommand(ctx, result.Project, "worktree", "list", "--porcelain", "-z")
	output, gitErr := command.Output()
	if gitErr == nil {
		for _, line := range strings.Split(string(output), "\x00") {
			if strings.HasPrefix(line, "worktree ") {
				base := strings.TrimPrefix(line, "worktree ")
				result.LocalPath = filepath.Join(base, ".agenticlens-config.local.yaml")
				result.LocalOrigin = "base"
				break
			}
		}
	}
	if _, err := os.Lstat(result.LocalPath); err != nil {
		return result, fmt.Errorf("required local configuration %s (%s origin) unavailable: %w", result.LocalPath, result.LocalOrigin, err)
	}
	return result, nil
}

func readSelectedLocal(selection SelectedConfig) ([]byte, error) {
	data, err := readBootstrap(selection.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("local configuration %s (%s origin): %w", selection.LocalPath, selection.LocalOrigin, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("local configuration %s must be a mapping", selection.LocalPath)
	}
	if err := validateCaptureDocument(&document); err != nil {
		return nil, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	return data, nil
}

func localCaptureEnabled(data []byte) error {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return err
	}
	node := &document
	for _, key := range []string{"telemetry", "capture", "enabled"} {
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			node = node.Content[0]
		}
		var found *yaml.Node
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				if node.Content[i].Value == key {
					found = node.Content[i+1]
					break
				}
			}
		}
		if found == nil {
			return errors.New("telemetry.capture.enabled must be explicitly configured true or false")
		}
		node = found
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" || (node.Value != "true" && node.Value != "false") {
		return errors.New("telemetry.capture.enabled must be explicitly configured true or false")
	}
	return nil
}
