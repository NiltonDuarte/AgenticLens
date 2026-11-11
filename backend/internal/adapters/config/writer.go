package config

import (
	"errors"
	"fmt"
	"strings"

	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

var ErrFileExists = errors.New("configuration file already exists")

type FileExistsError struct{ Path string }

func (e *FileExistsError) Error() string {
	return fmt.Sprintf("configuration file already exists: %s", e.Path)
}
func (e *FileExistsError) Is(target error) bool { return target == ErrFileExists }

// GenerateConfigYAML emits explicit shared policy. YAML nodes quote scalar values
// according to their types, preserving metacharacters and deliberate empty lists.
func GenerateConfigYAML(cfg domain.ProjectConfig) ([]byte, error) {
	if strings.TrimSpace(cfg.Workspaces.Dir) == "" {
		return nil, fmt.Errorf("workspaces.dir must be explicitly configured before writing shared YAML")
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	if err := ValidateGenerationPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateFeedbackPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateSummaryPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateDatabasePolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateActivityPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidatePollingPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateAnomalyPolicy(cfg); err != nil {
		return nil, err
	}
	if err := ValidateReleaseVersion(cfg.AgenticLens.Version); err != nil {
		return nil, err
	}
	if cfg.LocalConfigPath != "" {
		cfg.Tags = domain.MergeTags(nil, cfg.SharedTags)
	}
	cfg.Telemetry = domain.TelemetryConfig{}
	var node yaml.Node
	if err := node.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encoding shared configuration: %w", err)
	}
	// Machine-local telemetry is selected independently and never written to shared policy.
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "telemetry" {
			node.Content = append(node.Content[:i], node.Content[i+2:]...)
			break
		}
	}
	node.HeadComment = "AgenticLens shared project policy. Values are explicit approved choices."
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		switch key.Value {
		case "generation":
			key.HeadComment = "Shared optional-output policy. A zero feedback threshold disables human stderr feedback."
		case "ticket":
			key.HeadComment = "Literal prefix mappings preserve spelling and separators; case matching is explicit."
		case "code_metrics":
			key.HeadComment = "Committed net branch contribution. Production/tests include/exclude rules are independent; overlaps fail."
		case "role_order":
			key.HeadComment = "Observed listed roles first, then unlisted observed roles alphabetically. [] orders all alphabetically."
		case "database":
			key.HeadComment = "Source database lock waiting in milliseconds. Zero means no lock waiting."
		}
	}
	applyConfigurationComments(&node, cfg.Comments)
	var out strings.Builder
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(&node); err != nil {
		return nil, fmt.Errorf("writing shared YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	data := []byte(out.String())
	if _, err := parseEntrypoint(data, false); err != nil {
		return nil, fmt.Errorf("generated bootstrap syntax: %w", err)
	}
	return data, nil
}
