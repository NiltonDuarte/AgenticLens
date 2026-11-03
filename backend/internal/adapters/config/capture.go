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
		selector := struct {
			AgenticLens *domain.EntrypointConfig `yaml:"agenticlens"`
		}{}
		if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("{")) {
			if err := yaml.Unmarshal(data, &selector); err != nil {
				return cfg, err
			}
			if selector.AgenticLens != nil {
				if _, err := parseEntrypoint(data, name != shared); err != nil {
					return cfg, fmt.Errorf("entrypoint configuration %s: %w", name, err)
				}
			}
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
	if err == nil {
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
