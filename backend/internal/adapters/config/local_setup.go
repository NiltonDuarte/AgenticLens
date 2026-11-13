package config

import (
	"agenticlens/internal/adapters/filesystem"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckBaseSetup prevents guided local configuration writes from linked trees.
// Worktree local files remain independently user-managed.
func CheckBaseSetup(project string) error {
	root, err := filesystem.PhysicalPath(project)
	if err != nil {
		return err
	}
	if output, e := filesystem.RepositoryGitCommand(context.Background(), root, "rev-parse", "--show-toplevel").Output(); e == nil {
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
		output, err := filesystem.RepositoryGitCommand(context.Background(), root, "rev-parse", "--git-common-dir").Output()
		if err != nil {
			return err
		}
		common := strings.TrimSpace(string(output))
		if !filepath.IsAbs(common) {
			common = filepath.Join(root, common)
		}
		gitDirOutput, err := filesystem.RepositoryGitCommand(context.Background(), root, "rev-parse", "--git-dir").Output()
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
	state := ""
	if len(statePaths) > 0 {
		state = statePaths[0]
	}
	return InitialLocalYAMLWithSettings(project, "", enabled, source, codexHome, database, state)
}

// InitialLocalYAMLWithSettings fills missing approved choices only; the settings
// candidate is explicit input, never a discovered local policy or installation request.
func InitialLocalYAMLWithSettings(project, candidatePath, enabled, source, codexHome, database, state string, publicationStates ...string) ([]byte, error) {
	physical, err := filesystem.PhysicalPath(project)
	if err != nil {
		return nil, err
	}
	project = physical
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
	if candidatePath != "" {
		ext := strings.ToLower(filepath.Ext(candidatePath))
		if ext != ".yaml" && ext != ".yml" {
			return nil, fmt.Errorf("--local-settings must have a .yaml or .yml suffix")
		}
		candidate, err := os.ReadFile(candidatePath)
		if err != nil {
			return nil, err
		}
		nodes, err := configurationNodes(candidate)
		if err != nil {
			return nil, err
		}
		allowed := map[string]bool{"telemetry.default_source": true, "telemetry.sources.opencode.database_path": true, "telemetry.sources.codex.home": true, "agenticlens.install_root": true, "agenticlens.state_dir": true}
		for _, key := range []string{"enabled", "state_dir", "refresh_ms", "debounce_ms", "idle_ms", "max_lifetime_ms", "pass_timeout_ms", "max_sessions", "max_source_bytes", "commit_warning"} {
			allowed["telemetry.capture."+key] = true
		}
		for key, node := range nodes {
			if node.Kind == yaml.MappingNode {
				continue
			}
			if !allowed[key] {
				return nil, fmt.Errorf("unsupported local setup setting %s", key)
			}
		}
		var prepared yaml.Node
		if err := yaml.Unmarshal(candidate, &prepared); err != nil {
			return nil, err
		}
		var fillMissing func(*yaml.Node, *yaml.Node) error
		fillMissing = func(target, input *yaml.Node) error {
			if target.Kind != yaml.MappingNode || input.Kind != yaml.MappingNode {
				return errors.New("local settings must use mappings")
			}
			for i := 0; i < len(input.Content); i += 2 {
				key, value := input.Content[i], input.Content[i+1]
				var existing *yaml.Node
				for j := 0; j < len(target.Content); j += 2 {
					if target.Content[j].Value == key.Value {
						existing = target.Content[j+1]
						break
					}
				}
				if existing == nil {
					target.Content = append(target.Content, key, value)
					changed = true
				} else if value.Kind == yaml.MappingNode {
					if err := fillMissing(existing, value); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if err := fillMissing(root, prepared.Content[0]); err != nil {
			return nil, err
		}
	}
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
	if _, err := fill([]string{"telemetry", "capture", "state_dir"}, state, "!!str"); err != nil {
		return nil, err
	}
	publicationState := "~/.agenticlens/state"
	if len(publicationStates) > 0 {
		publicationState = publicationStates[0]
	}
	if _, err := fill([]string{"agenticlens", "state_dir"}, publicationState, "!!str"); err != nil {
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
	cfg, err := decodeCapturePolicy(result)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Enabled {
		if err := cfg.ValidateWorker(); err != nil {
			return nil, err
		}
	}
	if _, present := findMappingValue(root, "agenticlens"); present {
		if _, err := parseEntrypoint(result, true); err != nil {
			return nil, err
		}
	}
	if !changed {
		return data, nil
	}
	return result, nil
}

func findMappingValue(root *yaml.Node, key string) (*yaml.Node, bool) {
	if root.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			return root.Content[i+1], true
		}
	}
	return nil, false
}
