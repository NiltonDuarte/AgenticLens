package config

import (
	"agenticlens/internal/domain"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// ResolveCapture selects one whole local policy; no shared capture state is inherited.
func ResolveCapture(project, stateOverride string) (domain.CaptureConfig, error) {
	selection, err := SelectConfig(project)
	if err != nil {
		return domain.CaptureConfig{}, err
	}
	data, err := readSelectedLocal(selection)
	if err != nil {
		return domain.CaptureConfig{}, err
	}
	cfg, err := decodeCapturePolicy(data)
	if err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	if err := cfg.ValidateState(); err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w; run setup from the base checkout", selection.LocalPath, err)
	}
	origin := filepath.Dir(selection.LocalPath)
	if stateOverride != "" {
		cfg.StateDir = stateOverride
		origin, err = os.Getwd()
		if err != nil {
			return cfg, err
		}
	}
	cfg.StateDir, err = ResolveCaptureStatePath(origin, cfg.StateDir)
	return cfg, err
}

// Disabled automatic capture still permits manual generation using explicit state.
func ResolveGenerationCapture(project, stateOverride string) (domain.CaptureConfig, error) {
	return ResolveCapture(project, stateOverride)
}

func decodeCapturePolicy(data []byte) (domain.CaptureConfig, error) {
	var overlay struct {
		Telemetry struct {
			Capture domain.CaptureConfig `yaml:"capture"`
		} `yaml:"telemetry"`
	}
	nodes, err := configurationNodes(data)
	if err != nil {
		return overlay.Telemetry.Capture, err
	}
	for _, key := range []string{"enabled", "state_dir", "refresh_ms", "debounce_ms", "idle_ms", "max_lifetime_ms", "pass_timeout_ms", "max_sessions", "max_source_bytes", "commit_warning"} {
		n := nodes["telemetry.capture."+key]
		if n == nil {
			continue
		}
		tag := "!!int"
		if key == "enabled" || key == "commit_warning" {
			tag = "!!bool"
		}
		if key == "state_dir" {
			tag = "!!str"
		}
		if err := requireNodeType(n, "telemetry.capture."+key, yaml.ScalarNode, tag); err != nil {
			return overlay.Telemetry.Capture, err
		}
	}
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return overlay.Telemetry.Capture, err
	}
	cfg := overlay.Telemetry.Capture
	cfg.Settings = make(map[string]bool)
	for key := range nodes {
		if strings.HasPrefix(key, "telemetry.capture.") {
			cfg.Settings[strings.TrimPrefix(key, "telemetry.capture.")] = true
		}
	}
	if _, present := nodes["telemetry.capture"]; present {
		if err := cfg.Validate(); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func captureStateValue(data []byte) (string, bool, error) {
	nodes, err := configurationNodes(data)
	if err != nil {
		return "", false, err
	}
	n := nodes["telemetry.capture.state_dir"]
	if n == nil {
		return "", false, nil
	}
	if err := requireNodeType(n, "telemetry.capture.state_dir", yaml.ScalarNode, "!!str"); err != nil {
		return "", true, err
	}
	return n.Value, true, nil
}

// ResolveCaptureStatePath interprets an explicitly supplied path without cleaning
// physical symlink/parent traversal or selecting another state directory.
func ResolveCaptureStatePath(project, path string) (string, error) {
	if strings.TrimSpace(path) == "" || strings.IndexByte(path, 0) >= 0 {
		return "", fmt.Errorf("telemetry.capture.state_dir must be an explicitly configured nonblank path")
	}
	return ResolveConfiguredPath(project, path)
}

func validateCaptureDocument(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		keys := make(map[string]bool)
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return errors.New("capture configuration requires string mapping keys")
			}
			if keys[key.Value] {
				return fmt.Errorf("duplicate capture configuration key %s", key.Value)
			}
			keys[key.Value] = true
		}
	}
	for _, child := range node.Content {
		if err := validateCaptureDocument(child); err != nil {
			return err
		}
	}
	return nil
}
