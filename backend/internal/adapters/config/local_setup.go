package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckBaseSetup prevents guided local configuration writes from linked trees.
// Worktree local files remain independently user-managed.
func CheckBaseSetup(project string) error {
	root, err := filepath.Abs(project)
	if err != nil {
		return err
	}
	if output, e := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output(); e == nil {
		root = strings.TrimSpace(string(output))
	}
	marker, err := os.ReadFile(filepath.Join(root, ".git"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		if info, e := os.Stat(filepath.Join(root, ".git")); e == nil && info.IsDir() {
			return nil
		}
		return err
	}
	if strings.HasPrefix(strings.TrimSpace(string(marker)), "gitdir:") {
		output, err := exec.Command("git", "-C", root, "rev-parse", "--git-common-dir").Output()
		if err != nil {
			return err
		}
		common := strings.TrimSpace(string(output))
		if !filepath.IsAbs(common) {
			common = filepath.Join(root, common)
		}
		gitDirOutput, err := exec.Command("git", "-C", root, "rev-parse", "--git-dir").Output()
		if err != nil {
			return err
		}
		gitDir := strings.TrimSpace(string(gitDirOutput))
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(root, gitDir)
		}
		if filepath.Clean(common) != filepath.Clean(gitDir) {
			return errors.New("linked-worktree setup does not create/copy local configuration or write base settings; configure the base from its checkout, or manage this worktree local file independently")
		}
	}
	return nil
}

// InitialLocalYAML fills only missing explicitly supplied local choices. Existing
// valid values survive repeated setup, including an explicit disabled policy.
func InitialLocalYAML(project, enabled, source, codexHome, database string, statePaths ...string) ([]byte, error) {
	if err := CheckBaseSetup(project); err != nil {
		return nil, err
	}
	path := filepath.Join(project, ".agenticlens-config.local.yaml")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if errors.Is(err, os.ErrNotExist) {
		data = []byte("telemetry: {}\n")
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("local configuration must be a mapping")
	}
	if err := validateCaptureDocument(&document); err != nil {
		return nil, err
	}
	root := document.Content[0]
	changed := false
	fill := func(keys []string, value, tag string) (string, error) {
		current := root
		for i, key := range keys {
			if current.Kind != yaml.MappingNode {
				return "", fmt.Errorf("local %s must be a mapping", strings.Join(keys[:i], "."))
			}
			var next *yaml.Node
			for j := 0; j+1 < len(current.Content); j += 2 {
				if current.Content[j].Value == key {
					next = current.Content[j+1]
					break
				}
			}
			if i == len(keys)-1 {
				if next != nil {
					if next.Kind != yaml.ScalarNode || next.Tag != tag || strings.TrimSpace(next.Value) == "" {
						return "", fmt.Errorf("existing local %s is invalid; preserved", strings.Join(keys, "."))
					}
					return next.Value, nil
				}
				if strings.TrimSpace(value) == "" {
					return "", fmt.Errorf("missing required local %s; supply explicit setup choice", strings.Join(keys, "."))
				}
				next = &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
			} else if next == nil {
				next = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			found := false
			for j := 0; j+1 < len(current.Content); j += 2 {
				if current.Content[j].Value == key {
					found = true
					break
				}
			}
			if !found {
				changed = true
				current.Content = append(current.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, next)
			}
			current = next
			if i == len(keys)-1 {
				return next.Value, nil
			}
		}
		return "", nil
	}
	if enabled != "" && enabled != "true" && enabled != "false" {
		return nil, errors.New("--capture-enabled must be true or false")
	}
	if _, err := fill([]string{"telemetry", "capture", "enabled"}, enabled, "!!bool"); err != nil {
		return nil, err
	}
	state := ""
	if len(statePaths) > 0 {
		state = statePaths[0]
	}
	if _, err := fill([]string{"telemetry", "capture", "state_dir"}, state, "!!str"); err != nil {
		return nil, err
	}
	selected, err := fill([]string{"telemetry", "default_source"}, source, "!!str")
	if err != nil {
		return nil, err
	}
	if selected == "codex" {
		if _, err := fill([]string{"telemetry", "sources", "codex", "home"}, codexHome, "!!str"); err != nil {
			return nil, err
		}
	} else if selected == "opencode" {
		if _, err := fill([]string{"telemetry", "sources", "opencode", "database_path"}, database, "!!str"); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("local telemetry.default_source must be codex or opencode")
	}
	result, err := yaml.Marshal(&document)
	if err != nil {
		return nil, err
	}
	if err := localCaptureEnabled(result); err != nil {
		return nil, err
	}
	cfg, err := decodeCapturePolicy(result, captureReadPolicy())
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !changed {
		return data, nil
	}
	return result, nil
}
