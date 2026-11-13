package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

var ErrConfigNotFound = errors.New("configuration file not found")
var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

type ConfigValidationError struct {
	MissingFields []string
	Message       string
}

func (e *ConfigValidationError) Error() string {
	if e.Message != "" {
		return "invalid configuration: " + e.Message
	}
	return "invalid configuration: missing required fields: " + strings.Join(e.MissingFields, ", ")
}
func (e *ConfigValidationError) Is(target error) bool { return target == ErrInvalidConfig }

// ValidateConfig validates shared policy needed to interpret project records.
// Optional source/live/generation policy is required only by its actual consumer.
func ValidateConfig(cfg domain.ProjectConfig) error {
	if err := domain.ValidateTags(cfg.Tags); err != nil {
		return err
	}
	for _, entry := range []struct{ key, value string }{{"project_name", cfg.ProjectName}, {"code_metrics.base_branch", cfg.CodeMetrics.BaseBranch}} {
		if strings.TrimSpace(entry.value) == "" {
			return &ConfigValidationError{MissingFields: []string{entry.key}}
		}
	}
	if len(cfg.Ticket.Prefixes) == 0 {
		return &ConfigValidationError{MissingFields: []string{"ticket.prefixes"}}
	}
	for i, p := range cfg.Ticket.Prefixes {
		if strings.TrimSpace(p.Prefix) == "" || strings.TrimSpace(p.Description) == "" {
			return fmt.Errorf("ticket.prefixes[%d] requires nonblank prefix and description", i)
		}
		if p.TicketType != "" {
			if _, ok := domain.NormalizeTicketType(p.TicketType); !ok {
				return fmt.Errorf("ticket.prefixes[%d].ticket_type is invalid: %q", i, p.TicketType)
			}
		}
	}
	if err := cfg.RequireSettings("ticket.case_insensitive", "workspaces.artifacts", "role_mappings", "role_order", "code_metrics.production.include", "code_metrics.production.exclude", "code_metrics.tests.include", "code_metrics.tests.exclude"); err != nil {
		return err
	}
	for _, list := range []struct {
		key   string
		value []string
	}{{"code_metrics.production.include", cfg.CodeMetrics.Production.Include}, {"code_metrics.production.exclude", cfg.CodeMetrics.Production.Exclude}, {"code_metrics.tests.include", cfg.CodeMetrics.Tests.Include}, {"code_metrics.tests.exclude", cfg.CodeMetrics.Tests.Exclude}} {
		if list.value == nil {
			return fmt.Errorf("%s must be an explicit list; [] selects no patterns", list.key)
		}
		for _, pattern := range list.value {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("%s contains a blank glob", list.key)
			}
			for _, segment := range strings.Split(pattern, "/") {
				if _, err := filepath.Match(segment, ""); err != nil {
					return fmt.Errorf("%s has invalid glob %q: %w", list.key, pattern, err)
				}
			}
		}
	}
	if err := ValidateWorkspacePolicy(cfg.Workspaces); err != nil {
		return err
	}
	if cfg.Workspaces.Artifacts == nil {
		return fmt.Errorf("workspaces.artifacts must be an explicit list; [] means none")
	}
	for i, a := range cfg.Workspaces.Artifacts {
		path := a.Path
		if strings.TrimSpace(path) == "" || filepath.IsAbs(path) || strings.ContainsAny(path, "\\\x00") {
			return fmt.Errorf("workspaces.artifacts[%d].path must be a safe ticket-workspace-relative path", i)
		}
		for _, part := range strings.Split(path, "/") {
			if part == ".." || part == "" || part == "." {
				return fmt.Errorf("workspaces.artifacts[%d].path must be a safe ticket-workspace-relative path", i)
			}
		}
	}
	if cfg.RoleMappings == nil {
		return fmt.Errorf("role_mappings must be an explicit mapping; {} means none")
	}
	for k, v := range cfg.RoleMappings {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			return fmt.Errorf("role_mappings requires nonblank string names and values")
		}
	}
	if cfg.RoleOrder == nil {
		return fmt.Errorf("role_order must be an explicit list; [] means alphabetical observed roles")
	}
	for _, role := range cfg.RoleOrder {
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("role_order contains a blank role")
		}
	}
	return nil
}

func validateErrorRatePolicy(t domain.AnomalyThresholdsConfig) error {
	if t.ErrorRateMinCalls < 1 {
		return fmt.Errorf("anomaly_thresholds.error_rate_min_calls must be a positive integer")
	}
	for _, rate := range []struct {
		key   string
		value float64
	}{{"error_rate_threshold", t.ErrorRateThreshold}, {"error_rate_critical_threshold", t.ErrorRateCriticalThreshold}} {
		if math.IsNaN(rate.value) || math.IsInf(rate.value, 0) || rate.value <= 0 || rate.value > 1 {
			return fmt.Errorf("anomaly_thresholds.%s must be a finite ratio greater than 0 and at most 1", rate.key)
		}
	}
	if t.ErrorRateThreshold > t.ErrorRateCriticalThreshold {
		return fmt.Errorf("anomaly_thresholds.error_rate_threshold cannot exceed anomaly_thresholds.error_rate_critical_threshold")
	}
	return nil
}

// FindConfigFile finds the nearest shared configuration without crossing a Git root.
func FindConfigFile(startDir string) (string, error) {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	curr, err := filesystem.PhysicalPath(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving configuration origin %s: %w", startDir, err)
	}
	for {
		selected := ""
		for _, name := range []string{".agenticlens-config.yaml", ".agenticlens-config.yml"} {
			path := filepath.Join(curr, name)
			if _, err := os.Lstat(path); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return "", fmt.Errorf("configuration path %s: %w", path, err)
			}
			if selected != "" {
				return "", fmt.Errorf("ambiguous shared configuration: both %s and %s exist; keep one canonical configuration file", selected, path)
			}
			selected = path
		}
		if selected != "" {
			info, err := os.Stat(selected)
			if err != nil {
				return "", fmt.Errorf("configuration %s: %w", selected, err)
			}
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("configuration must be a regular file: %s", selected)
			}
			return selected, nil
		}
		if _, err := os.Lstat(filepath.Join(curr, ".git")); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("Git project boundary: %w", err)
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return "", nil
}

// LoadShared reads project policy without consulting machine-local source/capture state.
func LoadShared(project string) (domain.ProjectConfig, error) {
	path, err := FindConfigFile(project)
	if err != nil {
		return domain.ProjectConfig{}, err
	}
	if path == "" {
		return domain.ProjectConfig{}, ErrConfigNotFound
	}
	return LoadFromFile(path)
}

func LoadFromFile(path string) (domain.ProjectConfig, error) {
	cfg, err := loadFromFile(path)
	if err != nil {
		return cfg, err
	}
	return validateLoadedConfig(cfg, path)
}

// ReadConfigCandidate decodes explicitly supplied setup choices before CLI
// overrides complete the candidate. It never selects runtime configuration.
func ReadConfigCandidate(path string) (domain.ProjectConfig, error) {
	return decodeSharedConfig(path, false)
}
func loadFromFile(path string) (domain.ProjectConfig, error) { return decodeSharedConfig(path, true) }
func decodeSharedConfig(path string, validateBootstrap bool) (domain.ProjectConfig, error) {
	var cfg domain.ProjectConfig
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return cfg, fmt.Errorf("configuration %s must have a .yaml or .yml suffix", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("reading configuration %s: %w", path, err)
	}
	nodes, err := configurationNodes(data)
	if err != nil {
		return cfg, fmt.Errorf("configuration %s: %w", path, err)
	}
	if err = validateSharedNodes(nodes); err != nil {
		return cfg, fmt.Errorf("configuration %s: %w", path, err)
	}
	tags, tagErr := decodeTags(nodes["tags"])
	if tagErr != nil {
		return cfg, fmt.Errorf("configuration %s: %w", path, tagErr)
	}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("decoding YAML configuration %s: %w", path, err)
	}
	cfg.Tags = tags
	cfg.SharedTags = domain.MergeTags(nil, tags)
	cfg.Settings = make(map[string]bool, len(nodes))
	for key := range nodes {
		cfg.Settings[key] = true
	}
	cfg.ConfigPath = path
	cfg.Comments = readConfigurationComments(data)
	cfg.ProjectName = strings.TrimSpace(cfg.ProjectName)
	for i := range cfg.Ticket.Prefixes {
		p := &cfg.Ticket.Prefixes[i]
		p.Prefix = strings.TrimSpace(p.Prefix)
		p.Description = strings.TrimSpace(p.Description)
		p.TicketType = strings.TrimSpace(p.TicketType)
	}
	// Shared policy never provides machine source or capture state authority.
	cfg.Telemetry = domain.TelemetryConfig{}
	if _, ok := nodes["agenticlens"]; ok && validateBootstrap {
		if _, err := parseEntrypoint(data, false); err != nil {
			return cfg, fmt.Errorf("bootstrap configuration %s: %w", path, err)
		}
	}
	return cfg, nil
}

func validateLoadedConfig(cfg domain.ProjectConfig, path string) (domain.ProjectConfig, error) {
	if err := ValidateConfig(cfg); err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("validating configuration %s: %w", path, err)
	}
	return cfg, nil
}

// Load selects one whole local file while preserving shared-only persisted reads.
func Load(project string) (domain.ProjectConfig, error) {
	return loadSelected(project, true)
}

// LoadManualGeneration loads the source and project settings used by ad hoc
// generation without interpreting unused capture-only policy.
func LoadManualGeneration(project string) (domain.ProjectConfig, error) {
	return loadSelected(project, false)
}

func loadSelected(project string, capturePolicy bool) (domain.ProjectConfig, error) {
	path, err := FindConfigFile(project)
	if err != nil {
		return domain.ProjectConfig{}, err
	}
	if path == "" {
		return domain.ProjectConfig{}, ErrConfigNotFound
	}
	cfg, err := loadFromFile(path)
	if err != nil {
		return cfg, err
	}
	selection, err := SelectConfig(project)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validateLoadedConfig(cfg, path)
		}
		return cfg, err
	}
	data, err := readSelectedLocal(selection)
	if err != nil {
		return cfg, err
	}
	nodes, err := configurationNodes(data)
	if err != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
	}
	localTags, tagErr := decodeTags(nodes["tags"])
	if tagErr != nil {
		return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, tagErr)
	}
	cfg.Tags = domain.MergeTags(cfg.Tags, localTags)
	cfg.LocalConfigPath = selection.LocalPath
	if _, exists := nodes["agenticlens"]; exists {
		entry, err := parseEntrypoint(data, true)
		if err != nil {
			return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
		}
		cfg.AgenticLens.StateDir = entry.StateDir
		if entry.InstallRoot != "" {
			cfg.AgenticLens.InstallRoot = entry.InstallRoot
		}
	}
	if n := nodes["telemetry.default_source"]; n != nil {
		if err := requireNodeType(n, "telemetry.default_source", yaml.ScalarNode, "!!str"); err != nil {
			return cfg, err
		}
		value := strings.ToLower(strings.TrimSpace(n.Value))
		if value != "opencode" && value != "codex" {
			return cfg, fmt.Errorf("local configuration %s: telemetry.default_source must be opencode or codex", selection.LocalPath)
		}
		cfg.Telemetry.DefaultSource = value
	}
	for _, source := range []struct {
		key    string
		target *string
	}{{"telemetry.sources.opencode.database_path", &cfg.Telemetry.Sources.OpenCodeDatabasePath}, {"telemetry.sources.codex.home", &cfg.Telemetry.Sources.CodexHome}} {
		if n := nodes[source.key]; n != nil {
			if err := requireNodeType(n, source.key, yaml.ScalarNode, "!!str"); err != nil {
				return cfg, err
			}
			if strings.TrimSpace(n.Value) == "" {
				return cfg, fmt.Errorf("local configuration %s: %s must be nonblank", selection.LocalPath, source.key)
			}
			*source.target, err = ResolveConfiguredPath(filepath.Dir(selection.LocalPath), n.Value)
			if err != nil {
				return cfg, fmt.Errorf("local configuration %s: %s: %w", selection.LocalPath, source.key, err)
			}
		}
	}
	if capturePolicy {
		cfg.Telemetry.Capture, err = decodeCapturePolicy(data)
		if err != nil {
			return cfg, fmt.Errorf("local configuration %s: %w", selection.LocalPath, err)
		}
	}
	return validateLoadedConfig(cfg, path)
}

func ProjectRoot(start string) (string, error) {
	path, err := FindConfigFile(start)
	if err != nil {
		return "", err
	}
	if path != "" {
		return filepath.EvalSymlinks(filepath.Dir(path))
	}
	if start == "" {
		start = "."
	}
	return filesystem.PhysicalPath(start)
}
