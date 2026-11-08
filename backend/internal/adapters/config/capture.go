package config

import (
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// ResolveCapture reads numeric defaults and exactly one selected local policy.
// It validates before applying disablement; no shared capture values are inherited.
func ResolveCapture(project, stateOverride string) (domain.CaptureConfig, error) {
	cfg := captureReadPolicy()
	selection, err := SelectConfig(project)
	if err != nil {
		return cfg, err
	}
	data, err := readSelectedLocal(selection)
	if err != nil {
		return cfg, err
	}
	if err := localCaptureEnabled(data); err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	cfg, err = decodeCapturePolicy(data, cfg)
	if err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	if stateOverride != "" {
		cfg.StateDir = stateOverride
	}
	cfg.StateDir, err = ResolveCaptureStatePath(filepath.Dir(selection.LocalPath), cfg.StateDir)
	return cfg, err
}

// ResolveGenerationCapture retains the legacy manual operational policy when no
// local file exists. It does not silently authorize automatic capture.
func ResolveGenerationCapture(project, stateOverride string) (domain.CaptureConfig, error) {
	selection, selectErr := SelectConfig(project)
	if selectErr != nil && !errors.Is(selectErr, os.ErrNotExist) {
		return domain.CaptureConfig{}, selectErr
	}
	data, err := os.ReadFile(selection.SharedPath)
	if err != nil {
		return domain.CaptureConfig{}, err
	}
	cfg, err := decodeCapturePolicy(data, captureReadPolicy())
	if err != nil {
		return cfg, err
	}
	origin := selection.Project
	if selectErr == nil {
		local, err := readSelectedLocal(selection)
		if err != nil {
			return cfg, err
		}
		cfg, err = decodeCapturePolicy(local, cfg)
		if err != nil {
			return cfg, err
		}
		if _, present, err := captureStateValue(local); err != nil {
			return cfg, err
		} else if present {
			origin = filepath.Dir(selection.LocalPath)
		}
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	if stateOverride != "" {
		cfg.StateDir = stateOverride
	}
	cfg.StateDir, err = ResolveCaptureStatePath(origin, cfg.StateDir)
	return cfg, err
}

func decodeCapturePolicy(data []byte, previous domain.CaptureConfig) (domain.CaptureConfig, error) {
	state, present, stateErr := captureStateValue(data)
	if stateErr != nil {
		return domain.CaptureConfig{}, stateErr
	}
	overlay := struct {
		Telemetry struct {
			Capture domain.CaptureConfig `yaml:"capture"`
		} `yaml:"telemetry"`
	}{}
	overlay.Telemetry.Capture = previous
	var err error
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
		err = json.Unmarshal(data, &overlay)
	} else {
		err = yaml.Unmarshal(data, &overlay)
	}
	if err == nil {
		if present {
			overlay.Telemetry.Capture.StateDir = state
		}
		deprecated := struct {
			Telemetry struct {
				Capture struct {
					DetailedReport *bool `yaml:"detailed_report" json:"detailed_report"`
				} `yaml:"capture" json:"capture"`
			} `yaml:"telemetry" json:"telemetry"`
		}{}
		if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
			err = json.Unmarshal(data, &deprecated)
		} else {
			err = yaml.Unmarshal(data, &deprecated)
		}
		if deprecated.Telemetry.Capture.DetailedReport != nil {
			overlay.Telemetry.Capture.DeprecatedDetailedReport = true
		}
		overlay.Telemetry.Capture.DetailedReport = false
	}
	return overlay.Telemetry.Capture, err
}

// captureReadPolicy intentionally has no state-path default. Construction and
// explicit configuration updates own insertion of the home-state value.
func captureReadPolicy() domain.CaptureConfig {
	cfg := domain.DefaultCaptureConfig()
	cfg.StateDir = ""
	return cfg
}

func expandCaptureHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expanding telemetry.capture.state_dir: user home is unavailable: %w", err)
	}
	if strings.TrimSpace(home) == "" {
		return "", errors.New("expanding telemetry.capture.state_dir: user home is unavailable")
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

// captureStateValue checks the syntax tree because scalar coercion and null
// overlays must not turn an invalid configured path into an inherited value.
func captureStateValue(data []byte) (string, bool, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return "", false, err
	}
	if len(document.Content) == 0 {
		return "", false, nil
	}
	node := document.Content[0]
	for _, key := range []string{"telemetry", "capture", "state_dir"} {
		if node.Kind != yaml.MappingNode {
			return "", true, fmt.Errorf("telemetry.capture.state_dir must be configured in mapping fields")
		}
		var found *yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				if found != nil {
					return "", true, fmt.Errorf("duplicate capture configuration key %s", key)
				}
				found = node.Content[i+1]
			}
		}
		if found == nil {
			return "", false, nil
		}
		node = found
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return "", true, fmt.Errorf("telemetry.capture.state_dir must be a nonblank string")
	}
	return node.Value, true, nil
}

// ResolveCaptureStatePath expands an explicitly provided state path; it never
// selects a default or validates an otherwise incomplete operational config.
func ResolveCaptureStatePath(project, path string) (string, error) {
	policy := domain.DefaultCaptureConfig()
	policy.StateDir = path
	if err := policy.Validate(); err != nil {
		return "", err
	}
	expanded, err := expandCaptureHome(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(expanded) {
		expanded = filepath.Join(project, expanded)
	}
	return filepath.Clean(expanded), nil
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
