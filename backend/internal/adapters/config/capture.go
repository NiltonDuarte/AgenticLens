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
)

// ResolveCapture overlays defaults, portable YAML and machine-local YAML in that
// order. Relative state paths always belong to the explicitly selected project.
func ResolveCapture(project, stateOverride string) (domain.CaptureConfig, error) {
	cfg := domain.DefaultCaptureConfig()
	root, err := ProjectRoot(project)
	if err != nil {
		return cfg, err
	}
	project = root
	shared, err := FindConfigFile(project)
	if err != nil {
		return cfg, err
	}
	for _, name := range []string{shared, filepath.Join(project, ".agenticlens-config.local.yaml")} {
		if name == "" {
			continue
		}
		data, err := os.ReadFile(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return cfg, err
		}
		cfg, err = decodeCapturePolicy(data, cfg)
		if err != nil {
			return cfg, fmt.Errorf("capture configuration %s: %w", name, err)
		}
	}
	if stateOverride != "" {
		cfg.StateDir = stateOverride
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	cfg.StateDir = ExpandHomePath(cfg.StateDir)
	if !filepath.IsAbs(cfg.StateDir) {
		cfg.StateDir = filepath.Join(project, cfg.StateDir)
	}
	return cfg, nil
}

func decodeCapturePolicy(data []byte, previous domain.CaptureConfig) (domain.CaptureConfig, error) {
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
	return overlay.Telemetry.Capture, err
}
