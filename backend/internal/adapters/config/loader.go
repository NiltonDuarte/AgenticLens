package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"agenticlens/internal/domain"
)

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

type rawWorkspacesConfig struct {
	Dir        *string  `json:"dir" yaml:"dir"`
	SrcFolders []string `json:"src_folders" yaml:"src_folders"`
}

type rawDatabaseConfig struct {
	Path           *string `json:"path" yaml:"path"`
	QueryTimeoutMs *int    `json:"query_timeout_ms" yaml:"query_timeout_ms"`
	BusyTimeoutMs  *int    `json:"busy_timeout_ms" yaml:"busy_timeout_ms"`
	ReadOnly       *bool   `json:"read_only" yaml:"read_only"`
}

type rawAnomalyThresholdsConfig struct {
	StallThresholdSeconds *int     `json:"stall_threshold_seconds" yaml:"stall_threshold_seconds"`
	ErrorRateThreshold    *float64 `json:"error_rate_threshold" yaml:"error_rate_threshold"`
	ContextExhaustionPct  *float64 `json:"context_exhaustion_pct" yaml:"context_exhaustion_pct"`
}

type rawLiveConfig struct {
	ActiveThresholdSeconds *int                        `json:"active_threshold_seconds" yaml:"active_threshold_seconds"`
	PollIntervalMs         *int                        `json:"poll_interval_ms" yaml:"poll_interval_ms"`
	MaxTraceBufferMessages *int                        `json:"max_trace_buffer_messages" yaml:"max_trace_buffer_messages"`
	AnomalyThresholds      *rawAnomalyThresholdsConfig `json:"anomaly_thresholds" yaml:"anomaly_thresholds"`
}

type rawCodeMetricsConfig struct {
	Include      []string `json:"include" yaml:"include"`
	Exclude      []string `json:"exclude" yaml:"exclude"`
	TestPatterns []string `json:"test_patterns" yaml:"test_patterns"`
	ProdPatterns []string `json:"prod_patterns" yaml:"prod_patterns"`
}

type rawProjectConfigDTO struct {
	ProjectName  *string               `json:"project_name" yaml:"project_name"`
	Ticket       *rawTicketConfig      `json:"ticket" yaml:"ticket"`
	Workspaces   *rawWorkspacesConfig  `json:"workspaces" yaml:"workspaces"`
	Database     *rawDatabaseConfig    `json:"database" yaml:"database"`
	Live         *rawLiveConfig        `json:"live" yaml:"live"`
	CodeMetrics  *rawCodeMetricsConfig `json:"code_metrics" yaml:"code_metrics"`
	RoleMappings map[string]string     `json:"role_mappings" yaml:"role_mappings"`
	RoleOrder    []string              `json:"role_order" yaml:"role_order"`

	// Legacy root-level fields from config.json
	LegacyTicketPrefix  *string  `json:"ticket_prefix" yaml:"ticket_prefix"`
	LegacyWorkspacesDir *string  `json:"workspaces_dir" yaml:"workspaces_dir"`
	LegacySrcFolders    []string `json:"src_folders" yaml:"src_folders"`
}

func mergeConfig(dto rawProjectConfigDTO) domain.ProjectConfig {
	cfg := domain.DefaultProjectConfig()

	if dto.ProjectName != nil && strings.TrimSpace(*dto.ProjectName) != "" {
		cfg.ProjectName = strings.TrimSpace(*dto.ProjectName)
	}

	hasModernPrefixes := false
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
			hasModernPrefixes = true
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

	if !hasModernPrefixes && dto.LegacyTicketPrefix != nil && strings.TrimSpace(*dto.LegacyTicketPrefix) != "" {
		cfg.Ticket.Prefixes = []domain.TicketPrefix{
			{
				Prefix:      strings.TrimSpace(*dto.LegacyTicketPrefix),
				Description: "",
			},
		}
	}

	if dto.Workspaces != nil {
		if dto.Workspaces.Dir != nil && strings.TrimSpace(*dto.Workspaces.Dir) != "" {
			cfg.Workspaces.Dir = strings.TrimSpace(*dto.Workspaces.Dir)
		}
		if dto.Workspaces.SrcFolders != nil {
			cfg.Workspaces.SrcFolders = dto.Workspaces.SrcFolders
		}
	}

	if dto.LegacyWorkspacesDir != nil && strings.TrimSpace(*dto.LegacyWorkspacesDir) != "" {
		if dto.Workspaces == nil || dto.Workspaces.Dir == nil || strings.TrimSpace(*dto.Workspaces.Dir) == "" {
			cfg.Workspaces.Dir = strings.TrimSpace(*dto.LegacyWorkspacesDir)
		}
	}

	if len(dto.LegacySrcFolders) > 0 {
		if dto.Workspaces == nil || dto.Workspaces.SrcFolders == nil {
			cfg.Workspaces.SrcFolders = dto.LegacySrcFolders
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
		if dto.Live.AnomalyThresholds != nil {
			if dto.Live.AnomalyThresholds.StallThresholdSeconds != nil {
				cfg.Live.AnomalyThresholds.StallThresholdSeconds = *dto.Live.AnomalyThresholds.StallThresholdSeconds
			}
			if dto.Live.AnomalyThresholds.ErrorRateThreshold != nil {
				cfg.Live.AnomalyThresholds.ErrorRateThreshold = *dto.Live.AnomalyThresholds.ErrorRateThreshold
			}
			if dto.Live.AnomalyThresholds.ContextExhaustionPct != nil {
				cfg.Live.AnomalyThresholds.ContextExhaustionPct = *dto.Live.AnomalyThresholds.ContextExhaustionPct
			}
		}
	}

	if dto.CodeMetrics != nil {
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

// FindConfigFile traverses upwards from startDir looking for .agenticlens configuration files
// in order of priority: config.yaml, config.yml, config.json.
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

	candidateNames := []string{"config.yaml", "config.yml", "config.json"}
	curr = absDir
	for {
		for _, name := range candidateNames {
			cfgPath := filepath.Join(curr, ".agenticlens", name)
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

	return mergeConfig(dto), nil
}

// Load finds the configuration file traversing upwards from repoRoot and parses it.
// If no configuration file is discovered, DefaultProjectConfig() is returned with a nil error.
func Load(repoRoot string) (domain.ProjectConfig, error) {
	cfgFile, err := FindConfigFile(repoRoot)
	if err != nil {
		return domain.DefaultProjectConfig(), err
	}
	if cfgFile == "" {
		return domain.DefaultProjectConfig(), nil
	}
	return LoadFromFile(cfgFile)
}
