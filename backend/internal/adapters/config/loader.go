package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"agenticlens/internal/domain"
)

// ErrConfigNotFound indicates that .agenticlens-config.yaml was not found in the directory hierarchy.
var ErrConfigNotFound = errors.New("configuration file not found")

// ErrInvalidConfig indicates that the configuration is missing required parameters.
var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

// ConfigValidationError specifies missing mandatory fields in the configuration.
type ConfigValidationError struct {
	MissingFields []string
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("invalid configuration: missing required fields: %s", strings.Join(e.MissingFields, ", "))
}

func (e *ConfigValidationError) Is(target error) bool {
	return target == ErrInvalidConfig
}

// ValidateConfig checks that mandatory configuration fields are present and non-empty.
func ValidateConfig(cfg domain.ProjectConfig) error {
	var missing []string
	if strings.TrimSpace(cfg.ProjectName) == "" {
		missing = append(missing, "project_name")
	}

	hasPrefix := false
	for _, p := range cfg.Ticket.Prefixes {
		if strings.TrimSpace(p.Prefix) != "" {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		missing = append(missing, "ticket.prefixes")
	}

	hasSourceRoot := false
	for _, sr := range cfg.CodeMetrics.SourceRoots {
		if strings.TrimSpace(sr) != "" {
			hasSourceRoot = true
			break
		}
	}
	if !hasSourceRoot {
		for _, sf := range cfg.Workspaces.SrcFolders {
			if strings.TrimSpace(sf) != "" {
				hasSourceRoot = true
				break
			}
		}
	}
	if !hasSourceRoot {
		missing = append(missing, "code_metrics.source_roots")
	}

	if len(missing) > 0 {
		return &ConfigValidationError{MissingFields: missing}
	}
	return nil
}

type rawPrefixItem struct {
	Prefix      string `json:"prefix" yaml:"prefix"`
	Description string `json:"description" yaml:"description"`
	TicketType  string `json:"ticket_type" yaml:"ticket_type"`
}

func (r *rawPrefixItem) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		if node.Tag != "" && node.Tag != "!!str" {
			return fmt.Errorf("invalid ticket prefix scalar type %s", node.Tag)
		}
		r.Prefix = strings.TrimSpace(node.Value)
		r.Description = ""
		r.TicketType = ""
		return nil
	}
	if node.Kind == yaml.MappingNode {
		var item struct {
			Prefix      string `yaml:"prefix"`
			Description string `yaml:"description"`
			TicketType  string `yaml:"ticket_type"`
		}
		if err := node.Decode(&item); err != nil {
			return fmt.Errorf("decoding ticket prefix mapping: %w", err)
		}
		r.Prefix = strings.TrimSpace(item.Prefix)
		r.Description = strings.TrimSpace(item.Description)
		r.TicketType = strings.TrimSpace(item.TicketType)
		return nil
	}
	return fmt.Errorf("invalid ticket prefix format: expected string or mapping, got node kind %d", node.Kind)
}

func (r *rawPrefixItem) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return fmt.Errorf("empty ticket prefix data")
	}

	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return fmt.Errorf("decoding string ticket prefix: %w", err)
		}
		r.Prefix = strings.TrimSpace(s)
		r.Description = ""
		r.TicketType = ""
		return nil
	}

	if trimmed[0] == '{' {
		var item struct {
			Prefix      string `json:"prefix"`
			Description string `json:"description"`
			TicketType  string `json:"ticket_type"`
		}
		if err := json.Unmarshal(trimmed, &item); err != nil {
			return fmt.Errorf("decoding object ticket prefix: %w", err)
		}
		r.Prefix = strings.TrimSpace(item.Prefix)
		r.Description = strings.TrimSpace(item.Description)
		r.TicketType = strings.TrimSpace(item.TicketType)
		return nil
	}

	return fmt.Errorf("invalid ticket prefix format: expected string or object, got %s", string(trimmed))
}

type rawTicketConfig struct {
	Prefixes        []rawPrefixItem `json:"prefixes" yaml:"prefixes"`
	CaseInsensitive *bool           `json:"case_insensitive" yaml:"case_insensitive"`
	CustomPatterns  []string        `json:"custom_patterns" yaml:"custom_patterns"`
}

type rawWorkspaceArtifactConfig struct {
	Path        string `json:"path" yaml:"path"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
}

type rawWorkspacesConfig struct {
	Dir        *string                      `json:"dir" yaml:"dir"`
	SrcFolders []string                     `json:"src_folders" yaml:"src_folders"`
	Artifacts  []rawWorkspaceArtifactConfig `json:"artifacts" yaml:"artifacts"`
}

type rawDatabaseConfig struct {
	Path           *string `json:"path" yaml:"path"`
	QueryTimeoutMs *int    `json:"query_timeout_ms" yaml:"query_timeout_ms"`
	BusyTimeoutMs  *int    `json:"busy_timeout_ms" yaml:"busy_timeout_ms"`
	ReadOnly       *bool   `json:"read_only" yaml:"read_only"`
}

type rawAnomalyThresholdsConfig struct {
	StallThresholdSeconds               *int     `json:"stall_threshold_seconds" yaml:"stall_threshold_seconds"`
	ErrorRateThreshold                  *float64 `json:"error_rate_threshold" yaml:"error_rate_threshold"`
	ContextExhaustionPct                *float64 `json:"context_exhaustion_pct" yaml:"context_exhaustion_pct"`
	ConsecutiveErrorsStreakWarning      *int     `json:"consecutive_errors_streak_warning" yaml:"consecutive_errors_streak_warning"`
	ConsecutiveErrorsStreakCritical     *int     `json:"consecutive_errors_streak_critical" yaml:"consecutive_errors_streak_critical"`
	ContextCeilingWarningTokens         *int64   `json:"context_ceiling_warning_tokens" yaml:"context_ceiling_warning_tokens"`
	ContextCeilingCriticalTokens        *int64   `json:"context_ceiling_critical_tokens" yaml:"context_ceiling_critical_tokens"`
	ContextCeilingApproachRatioWarning  *float64 `json:"context_ceiling_approach_ratio_warning" yaml:"context_ceiling_approach_ratio_warning"`
	ContextCeilingApproachRatioCritical *float64 `json:"context_ceiling_approach_ratio_critical" yaml:"context_ceiling_approach_ratio_critical"`
	TokenVelocityWarningTokens          *int64   `json:"token_velocity_warning_tokens" yaml:"token_velocity_warning_tokens"`
	TokenVelocityCriticalTokens         *int64   `json:"token_velocity_critical_tokens" yaml:"token_velocity_critical_tokens"`
	TokenVelocityWarningDeltaTokens     *int64   `json:"token_velocity_warning_delta_tokens" yaml:"token_velocity_warning_delta_tokens"`
	TokenVelocityCriticalDeltaTokens    *int64   `json:"token_velocity_critical_delta_tokens" yaml:"token_velocity_critical_delta_tokens"`
	ToolThrashingMinFailures            *int     `json:"tool_thrashing_min_failures" yaml:"tool_thrashing_min_failures"`
	ToolThrashingCriticalFailures       *int     `json:"tool_thrashing_critical_failures" yaml:"tool_thrashing_critical_failures"`
}

type rawLiveConfig struct {
	ActiveThresholdSeconds *int                        `json:"active_threshold_seconds" yaml:"active_threshold_seconds"`
	PollIntervalMs         *int                        `json:"poll_interval_ms" yaml:"poll_interval_ms"`
	MaxTraceBufferMessages *int                        `json:"max_trace_buffer_messages" yaml:"max_trace_buffer_messages"`
	AnomalyThresholds      *rawAnomalyThresholdsConfig `json:"anomaly_thresholds" yaml:"anomaly_thresholds"`
}

type rawMetricsConfig struct {
	HealthyCeilingTokens *int64 `json:"healthy_ceiling_tokens" yaml:"healthy_ceiling_tokens"`
}

type rawCodeMetricsConfig struct {
	SourceRoots  []string `json:"source_roots" yaml:"source_roots"`
	BaseBranch   *string  `json:"base_branch" yaml:"base_branch"`
	Include      []string `json:"include" yaml:"include"`
	Exclude      []string `json:"exclude" yaml:"exclude"`
	TestPatterns []string `json:"test_patterns" yaml:"test_patterns"`
	ProdPatterns []string `json:"prod_patterns" yaml:"prod_patterns"`
}

type rawProjectConfigDTO struct {
	ProjectName          *string                     `json:"project_name" yaml:"project_name"`
	Ticket               *rawTicketConfig            `json:"ticket" yaml:"ticket"`
	Workspaces           *rawWorkspacesConfig        `json:"workspaces" yaml:"workspaces"`
	HealthyCeilingTokens *int64                      `json:"healthy_ceiling_tokens" yaml:"healthy_ceiling_tokens"`
	Database             *rawDatabaseConfig          `json:"database" yaml:"database"`
	Live                 *rawLiveConfig              `json:"live" yaml:"live"`
	AnomalyThresholds    *rawAnomalyThresholdsConfig `json:"anomaly_thresholds" yaml:"anomaly_thresholds"`
	Metrics              *rawMetricsConfig           `json:"metrics" yaml:"metrics"`
	CodeMetrics          *rawCodeMetricsConfig       `json:"code_metrics" yaml:"code_metrics"`
	RoleMappings         map[string]string           `json:"role_mappings" yaml:"role_mappings"`
	RoleOrder            []string                    `json:"role_order" yaml:"role_order"`
}

func mergeConfig(dto rawProjectConfigDTO) domain.ProjectConfig {
	cfg := domain.DefaultProjectConfig()

	if dto.ProjectName != nil && strings.TrimSpace(*dto.ProjectName) != "" {
		cfg.ProjectName = strings.TrimSpace(*dto.ProjectName)
	}

	if dto.Ticket != nil && len(dto.Ticket.Prefixes) > 0 {
		prefixes := make([]domain.TicketPrefix, 0, len(dto.Ticket.Prefixes))
		for _, item := range dto.Ticket.Prefixes {
			cleanPrefix := strings.TrimSpace(item.Prefix)
			if cleanPrefix != "" {
				prefixes = append(prefixes, domain.TicketPrefix{
					Prefix:      cleanPrefix,
					Description: strings.TrimSpace(item.Description),
					TicketType:  strings.TrimSpace(item.TicketType),
				})
			}
		}
		if len(prefixes) > 0 {
			cfg.Ticket.Prefixes = prefixes
		}
	}

	if dto.Ticket != nil {
		if dto.Ticket.CaseInsensitive != nil {
			cfg.Ticket.CaseInsensitive = *dto.Ticket.CaseInsensitive
		}
		if dto.Ticket.CustomPatterns != nil {
			cfg.Ticket.CustomPatterns = dto.Ticket.CustomPatterns
		}
	}

	if dto.Workspaces != nil {
		if dto.Workspaces.Dir != nil {
			cfg.Workspaces.Dir = strings.TrimSpace(*dto.Workspaces.Dir)
		}
		if dto.Workspaces.SrcFolders != nil {
			var cleanFolders []string
			for _, f := range dto.Workspaces.SrcFolders {
				t := strings.TrimSpace(f)
				if t != "" {
					cleanFolders = append(cleanFolders, t)
				}
			}
			cfg.Workspaces.SrcFolders = cleanFolders
		}
		if len(dto.Workspaces.Artifacts) > 0 {
			artifacts := make([]domain.WorkspaceArtifactConfig, 0, len(dto.Workspaces.Artifacts))
			for _, item := range dto.Workspaces.Artifacts {
				cleanPath := strings.TrimSpace(item.Path)
				cleanTitle := strings.TrimSpace(item.Title)
				if cleanPath != "" {
					if cleanTitle == "" {
						cleanTitle = filepath.Base(cleanPath)
					}
					artifacts = append(artifacts, domain.WorkspaceArtifactConfig{
						Path:        cleanPath,
						Title:       cleanTitle,
						Description: strings.TrimSpace(item.Description),
					})
				}
			}
			if len(artifacts) > 0 {
				cfg.Workspaces.Artifacts = artifacts
			}
		}
	}

	if dto.Database != nil {
		if dto.Database.Path != nil && strings.TrimSpace(*dto.Database.Path) != "" {
			cfg.Database.Path = strings.TrimSpace(*dto.Database.Path)
		}
		if dto.Database.QueryTimeoutMs != nil {
			cfg.Database.QueryTimeoutMs = *dto.Database.QueryTimeoutMs
		}
		if dto.Database.BusyTimeoutMs != nil {
			cfg.Database.BusyTimeoutMs = *dto.Database.BusyTimeoutMs
		}
		if dto.Database.ReadOnly != nil {
			cfg.Database.ReadOnly = *dto.Database.ReadOnly
		}
	}

	if dto.HealthyCeilingTokens != nil && *dto.HealthyCeilingTokens > 0 {
		cfg.HealthyCeilingTokens = *dto.HealthyCeilingTokens
	} else if dto.Metrics != nil && dto.Metrics.HealthyCeilingTokens != nil && *dto.Metrics.HealthyCeilingTokens > 0 {
		cfg.HealthyCeilingTokens = *dto.Metrics.HealthyCeilingTokens
	}

	if dto.Live != nil {
		if dto.Live.ActiveThresholdSeconds != nil {
			cfg.Live.ActiveThresholdSeconds = *dto.Live.ActiveThresholdSeconds
		}
		if dto.Live.PollIntervalMs != nil {
			cfg.Live.PollIntervalMs = *dto.Live.PollIntervalMs
		}
		if dto.Live.MaxTraceBufferMessages != nil {
			cfg.Live.MaxTraceBufferMessages = *dto.Live.MaxTraceBufferMessages
		}
	}

	rawAT := dto.AnomalyThresholds
	if rawAT == nil && dto.Live != nil && dto.Live.AnomalyThresholds != nil {
		rawAT = dto.Live.AnomalyThresholds
	}
	if rawAT != nil {
		if rawAT.StallThresholdSeconds != nil {
			cfg.AnomalyThresholds.StallThresholdSeconds = *rawAT.StallThresholdSeconds
		}
		if rawAT.ErrorRateThreshold != nil {
			cfg.AnomalyThresholds.ErrorRateThreshold = *rawAT.ErrorRateThreshold
		}
		if rawAT.ContextExhaustionPct != nil {
			cfg.AnomalyThresholds.ContextExhaustionPct = *rawAT.ContextExhaustionPct
		}
		if rawAT.ConsecutiveErrorsStreakWarning != nil && *rawAT.ConsecutiveErrorsStreakWarning > 0 {
			cfg.AnomalyThresholds.ConsecutiveErrorsStreakWarning = *rawAT.ConsecutiveErrorsStreakWarning
		}
		if rawAT.ConsecutiveErrorsStreakCritical != nil && *rawAT.ConsecutiveErrorsStreakCritical > 0 {
			cfg.AnomalyThresholds.ConsecutiveErrorsStreakCritical = *rawAT.ConsecutiveErrorsStreakCritical
		}
		if rawAT.ContextCeilingWarningTokens != nil && *rawAT.ContextCeilingWarningTokens > 0 {
			cfg.AnomalyThresholds.ContextCeilingWarningTokens = *rawAT.ContextCeilingWarningTokens
		} else if rawAT.ContextCeilingApproachRatioWarning != nil && *rawAT.ContextCeilingApproachRatioWarning > 0 {
			effectiveCeiling := cfg.EffectiveHealthyCeiling()
			cfg.AnomalyThresholds.ContextCeilingWarningTokens = int64(float64(effectiveCeiling) * *rawAT.ContextCeilingApproachRatioWarning)
		}
		if rawAT.ContextCeilingCriticalTokens != nil && *rawAT.ContextCeilingCriticalTokens > 0 {
			cfg.AnomalyThresholds.ContextCeilingCriticalTokens = *rawAT.ContextCeilingCriticalTokens
		} else if rawAT.ContextCeilingApproachRatioCritical != nil && *rawAT.ContextCeilingApproachRatioCritical > 0 {
			effectiveCeiling := cfg.EffectiveHealthyCeiling()
			cfg.AnomalyThresholds.ContextCeilingCriticalTokens = int64(float64(effectiveCeiling) * *rawAT.ContextCeilingApproachRatioCritical)
		}
		if rawAT.TokenVelocityWarningTokens != nil && *rawAT.TokenVelocityWarningTokens > 0 {
			cfg.AnomalyThresholds.TokenVelocityWarningTokens = *rawAT.TokenVelocityWarningTokens
		}
		if rawAT.TokenVelocityCriticalTokens != nil && *rawAT.TokenVelocityCriticalTokens > 0 {
			cfg.AnomalyThresholds.TokenVelocityCriticalTokens = *rawAT.TokenVelocityCriticalTokens
		}
		if rawAT.TokenVelocityWarningDeltaTokens != nil && *rawAT.TokenVelocityWarningDeltaTokens > 0 {
			cfg.AnomalyThresholds.TokenVelocityWarningDeltaTokens = *rawAT.TokenVelocityWarningDeltaTokens
		}
		if rawAT.TokenVelocityCriticalDeltaTokens != nil && *rawAT.TokenVelocityCriticalDeltaTokens > 0 {
			cfg.AnomalyThresholds.TokenVelocityCriticalDeltaTokens = *rawAT.TokenVelocityCriticalDeltaTokens
		}
		if rawAT.ToolThrashingMinFailures != nil && *rawAT.ToolThrashingMinFailures > 0 {
			cfg.AnomalyThresholds.ToolThrashingMinFailures = *rawAT.ToolThrashingMinFailures
		}
		if rawAT.ToolThrashingCriticalFailures != nil && *rawAT.ToolThrashingCriticalFailures > 0 {
			cfg.AnomalyThresholds.ToolThrashingCriticalFailures = *rawAT.ToolThrashingCriticalFailures
		}
	}

	if dto.CodeMetrics != nil {
		if dto.CodeMetrics.SourceRoots != nil {
			var cleanRoots []string
			for _, r := range dto.CodeMetrics.SourceRoots {
				t := strings.TrimSpace(r)
				if t != "" {
					cleanRoots = append(cleanRoots, t)
				}
			}
			cfg.CodeMetrics.SourceRoots = cleanRoots
		}
		if dto.CodeMetrics.BaseBranch != nil && strings.TrimSpace(*dto.CodeMetrics.BaseBranch) != "" {
			cfg.CodeMetrics.BaseBranch = strings.TrimSpace(*dto.CodeMetrics.BaseBranch)
		}
		if dto.CodeMetrics.Include != nil {
			cfg.CodeMetrics.Include = dto.CodeMetrics.Include
		}
		if dto.CodeMetrics.Exclude != nil {
			cfg.CodeMetrics.Exclude = dto.CodeMetrics.Exclude
		}
		if dto.CodeMetrics.TestPatterns != nil {
			cfg.CodeMetrics.TestPatterns = dto.CodeMetrics.TestPatterns
		}
		if dto.CodeMetrics.ProdPatterns != nil {
			cfg.CodeMetrics.ProdPatterns = dto.CodeMetrics.ProdPatterns
		}
	}

	// Bidirectional fallback between CodeMetrics.SourceRoots and Workspaces.SrcFolders
	if len(cfg.CodeMetrics.SourceRoots) == 0 && len(cfg.Workspaces.SrcFolders) > 0 {
		cfg.CodeMetrics.SourceRoots = append([]string(nil), cfg.Workspaces.SrcFolders...)
	} else if len(cfg.Workspaces.SrcFolders) == 0 && len(cfg.CodeMetrics.SourceRoots) > 0 {
		cfg.Workspaces.SrcFolders = append([]string(nil), cfg.CodeMetrics.SourceRoots...)
	}

	if dto.RoleMappings != nil {
		for k, v := range dto.RoleMappings {
			cleanKey := strings.TrimSpace(k)
			cleanVal := strings.TrimSpace(v)
			if cleanKey != "" && cleanVal != "" {
				cfg.RoleMappings[cleanKey] = cleanVal
			}
		}
	}

	if dto.RoleOrder != nil {
		cleanOrder := make([]string, 0, len(dto.RoleOrder))
		for _, r := range dto.RoleOrder {
			trimmed := strings.ToLower(strings.TrimSpace(r))
			if trimmed != "" {
				cleanOrder = append(cleanOrder, trimmed)
			}
		}
		cfg.RoleOrder = cleanOrder
	}

	return cfg
}

// FindConfigFile traverses upwards from startDir looking for AgenticLens configuration files.
// At each directory level, candidates are evaluated in priority order:
// 1. .agenticlens-config.yaml (root)
// 2. .agenticlens-config.yml (root)
// 3. .agenticlens-config.json (root)
func FindConfigFile(startDir string) (string, error) {
	curr := strings.TrimSpace(startDir)
	if curr == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolving working directory: %w", err)
		}
		curr = cwd
	}

	absDir, err := filepath.Abs(curr)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %s: %w", curr, err)
	}

	curr = absDir
	for {
		candidates := []string{
			filepath.Join(curr, ".agenticlens-config.yaml"),
			filepath.Join(curr, ".agenticlens-config.yml"),
			filepath.Join(curr, ".agenticlens-config.json"),
		}
		for _, cfgPath := range candidates {
			if fi, err := os.Stat(cfgPath); err == nil && !fi.IsDir() {
				return cfgPath, nil
			}
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", nil
}

// LoadFromFile unmarshals a configuration file by path, merging values over domain defaults.
func LoadFromFile(filePath string) (domain.ProjectConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("reading configuration file %s: %w", filePath, err)
	}

	var dto rawProjectConfigDTO
	ext := strings.ToLower(filepath.Ext(filePath))

	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &dto); err != nil {
			return domain.ProjectConfig{}, fmt.Errorf("decoding yaml configuration %s: %w", filePath, err)
		}
	} else if ext == ".json" {
		if err := json.Unmarshal(data, &dto); err != nil {
			return domain.ProjectConfig{}, fmt.Errorf("decoding json configuration %s: %w", filePath, err)
		}
	} else {
		if err := yaml.Unmarshal(data, &dto); err != nil {
			if jsonErr := json.Unmarshal(data, &dto); jsonErr != nil {
				return domain.ProjectConfig{}, fmt.Errorf("decoding configuration %s: %w", filePath, err)
			}
		}
	}

	cfg := mergeConfig(dto)
	if err := ValidateConfig(cfg); err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("validating configuration %s: %w", filePath, err)
	}
	return cfg, nil
}

// Load finds the configuration file traversing upwards from repoRoot and parses it.
// If no configuration file is discovered, ErrConfigNotFound is returned.
func Load(repoRoot string) (domain.ProjectConfig, error) {
	cfgFile, err := FindConfigFile(repoRoot)
	if err != nil {
		return domain.ProjectConfig{}, err
	}
	if cfgFile == "" {
		return domain.ProjectConfig{}, ErrConfigNotFound
	}
	return LoadFromFile(cfgFile)
}
