package domain

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// ProjectConfig holds the complete configuration settings for AgenticLens.
type ProjectConfig struct {
	Comments          map[string]ConfigComment `json:"-" yaml:"-"`
	Settings          map[string]bool          `json:"-" yaml:"-"`
	ConfigPath        string                   `json:"-" yaml:"-"`
	LocalConfigPath   string                   `json:"-" yaml:"-"`
	Generation        GenerationConfig         `json:"generation" yaml:"generation"`
	AgenticLens       EntrypointConfig         `json:"agenticlens,omitempty" yaml:"agenticlens,omitempty"`
	ProjectName       string                   `json:"project_name" yaml:"project_name"`
	Ticket            TicketConfig             `json:"ticket" yaml:"ticket"`
	Workspaces        WorkspacesConfig         `json:"workspaces" yaml:"workspaces"`
	Database          DatabaseConfig           `json:"database" yaml:"database"`
	Telemetry         TelemetryConfig          `json:"telemetry,omitempty" yaml:"telemetry,omitempty"`
	Live              LiveConfig               `json:"live" yaml:"live"`
	AnomalyThresholds AnomalyThresholdsConfig  `json:"anomaly_thresholds" yaml:"anomaly_thresholds"`
	CodeMetrics       CodeMetricsConfig        `json:"code_metrics" yaml:"code_metrics"`
	RoleMappings      RoleMappings             `json:"role_mappings" yaml:"role_mappings"`
	RoleOrder         []string                 `json:"role_order" yaml:"role_order"`
}

// EntrypointConfig separates shared release authority from machine-local installation paths.
type EntrypointConfig struct {
	BinaryDir   string `json:"binary_dir,omitempty" yaml:"binary_dir,omitempty"`
	Version     string `json:"version" yaml:"version"`
	InstallRoot string `json:"install_root,omitempty" yaml:"install_root,omitempty"`
}

// EffectiveHealthyCeiling returns the configured denominator. Callers must validate
// context policy before calculating; zero represents unavailable policy, never a guessed capacity.
func (c ProjectConfig) EffectiveHealthyCeiling() int64 {
	return c.AnomalyThresholds.ContextCeilingCriticalTokens
}

// TicketPrefix represents a configured ticket identifier prefix and its operational description.
type TicketPrefix struct {
	Prefix      string `json:"prefix" yaml:"prefix"`
	Description string `json:"description" yaml:"description"`
	TicketType  string `json:"ticket_type,omitempty" yaml:"ticket_type,omitempty"`
}

// TicketConfig specifies ticket identification, casing rules, and pattern parsing.
type TicketConfig struct {
	Prefixes        []TicketPrefix `json:"prefixes" yaml:"prefixes"`
	CaseInsensitive bool           `json:"case_insensitive" yaml:"case_insensitive"`
}

// PrefixNames returns a slice of non-empty prefix strings configured in TicketConfig.
func (t TicketConfig) PrefixNames() []string {
	names := make([]string, 0, len(t.Prefixes))
	for _, p := range t.Prefixes {
		clean := strings.TrimSpace(p.Prefix)
		if clean != "" {
			names = append(names, clean)
		}
	}
	return names
}

// MatchTicketPrefix recognizes a whole recorded ID against one literal prefix.
// An optional separator belongs to the ID; it never changes the configured prefix.
func MatchTicketPrefix(ticketID, prefix string, insensitive bool) bool {
	_, ok := literalTicketNumber(ticketID, prefix, insensitive)
	return ok
}
func literalTicketNumber(ticketID, prefix string, insensitive bool) (string, bool) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", false
	}
	if insensitive {
		ticketID = strings.ToLower(ticketID)
		prefix = strings.ToLower(prefix)
	}
	if !strings.HasPrefix(ticketID, prefix) {
		return "", false
	}
	number := ticketID[len(prefix):]
	if len(number) > 0 && (number[0] == '-' || number[0] == '_') {
		number = number[1:]
	}
	if number == "" {
		return "", false
	}
	for i := range number {
		if number[i] < '0' || number[i] > '9' {
			return "", false
		}
	}
	return number, true
}

// ResolveTicketTypeMapping requires all applicable configured type mappings to agree.
func (t TicketConfig) ResolveTicketTypeMapping(ticketID string) (ConventionalTicketType, bool, error) {
	var selected ConventionalTicketType
	var mappings []string
	for _, p := range t.Prefixes {
		if p.TicketType == "" || !MatchTicketPrefix(ticketID, p.Prefix, t.CaseInsensitive) {
			continue
		}
		value, ok := NormalizeTicketType(p.TicketType)
		if !ok {
			return "", false, fmt.Errorf("ticket.prefixes[%s].ticket_type is invalid: %q", p.Prefix, p.TicketType)
		}
		mappings = append(mappings, p.Prefix+"="+string(value))
		if selected != "" && selected != value {
			return "", false, fmt.Errorf("conflicting ticket type mappings %s for %s; supply --ticket-type", strings.Join(mappings, ", "), ticketID)
		}
		selected = value
	}
	return selected, selected != "", nil
}

// WorkspaceArtifactConfig defines a project-level documentation artifact inside ticket workspaces.
type WorkspaceArtifactConfig struct {
	Path        string `json:"path" yaml:"path"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// WorkspacesConfig defines filesystem paths for ticket workspaces and artifacts.
type WorkspacesConfig struct {
	Dir       string                    `json:"dir" yaml:"dir"`
	Artifacts []WorkspaceArtifactConfig `json:"artifacts" yaml:"artifacts"`
}

// DatabaseConfig specifies SQLite database connection and query execution parameters.
type DatabaseConfig struct {
	BusyTimeoutMs int `json:"busy_timeout_ms" yaml:"busy_timeout_ms"`
}

// TelemetryConfig contains optional machine-local telemetry source preferences.
type TelemetryConfig struct {
	Capture       CaptureConfig        `json:"capture" yaml:"capture"`
	DefaultSource string               `json:"default_source,omitempty" yaml:"default_source,omitempty"`
	Sources       TelemetrySourcePaths `json:"sources,omitempty" yaml:"sources,omitempty"`
}

// TelemetrySourcePaths keeps independent locations for each supported harness.
type TelemetrySourcePaths struct {
	OpenCodeDatabasePath string `json:"database_path,omitempty" yaml:"database_path,omitempty"`
	CodexHome            string `json:"home,omitempty" yaml:"home,omitempty"`
}

// AnomalyThresholdsConfig defines numerical thresholds for anomaly detection rules.
type AnomalyThresholdsConfig struct {
	StallThresholdSeconds           int     `json:"stall_threshold_seconds" yaml:"stall_threshold_seconds"`
	ErrorRateMinCalls               int     `json:"error_rate_min_calls" yaml:"error_rate_min_calls"`
	ErrorRateThreshold              float64 `json:"error_rate_threshold" yaml:"error_rate_threshold"`
	ErrorRateCriticalThreshold      float64 `json:"error_rate_critical_threshold" yaml:"error_rate_critical_threshold"`
	ConsecutiveErrorsStreakWarning  int     `json:"consecutive_errors_streak_warning" yaml:"consecutive_errors_streak_warning"`
	ConsecutiveErrorsStreakCritical int     `json:"consecutive_errors_streak_critical" yaml:"consecutive_errors_streak_critical"`
	ContextCeilingWarningTokens     int64   `json:"context_ceiling_warning_tokens" yaml:"context_ceiling_warning_tokens"`
	ContextCeilingCriticalTokens    int64   `json:"context_ceiling_critical_tokens" yaml:"context_ceiling_critical_tokens"`
	TurnUsageWarningTokens          int64   `json:"turn_usage_warning_tokens" yaml:"turn_usage_warning_tokens"`
	TurnUsageCriticalTokens         int64   `json:"turn_usage_critical_tokens" yaml:"turn_usage_critical_tokens"`
	ContextGrowthWarningTokens      int64   `json:"context_growth_warning_tokens" yaml:"context_growth_warning_tokens"`
	ContextGrowthCriticalTokens     int64   `json:"context_growth_critical_tokens" yaml:"context_growth_critical_tokens"`
	ToolThrashingMinFailures        int     `json:"tool_thrashing_min_failures" yaml:"tool_thrashing_min_failures"`
	ToolThrashingCriticalFailures   int     `json:"tool_thrashing_critical_failures" yaml:"tool_thrashing_critical_failures"`
}

// LiveConfig controls real-time session tracking and polling intervals.
type LiveConfig struct {
	ActiveThresholdSeconds int `json:"active_threshold_seconds" yaml:"active_threshold_seconds"`
	PollIntervalMs         int `json:"poll_interval_ms" yaml:"poll_interval_ms"`
}

// CodeMetricsConfig defines file classification glob patterns for git metrics.
type CodeMetricsConfig struct {
	BaseBranch string             `json:"base_branch" yaml:"base_branch"`
	Production CodeCategoryConfig `json:"production" yaml:"production"`
	Tests      CodeCategoryConfig `json:"tests" yaml:"tests"`
}

// CodeCategoryConfig independently includes and excludes repository-relative paths.
type CodeCategoryConfig struct {
	Include []string `json:"include" yaml:"include"`
	Exclude []string `json:"exclude" yaml:"exclude"`
}

// RequireSettings checks presence in decoded policy. Programmatically constructed
// values are explicit inputs; decoded files always carry their setting inventory.
func (c ProjectConfig) RequireSettings(keys ...string) error {
	for _, key := range keys {
		if c.Settings != nil && !c.Settings[key] {
			return fmt.Errorf("configuration %s: %s is required; configure it explicitly", c.ConfigPath, key)
		}
	}
	return nil
}

// RoleMappings maps agent persona names to AgentRole enums.
type RoleMappings map[string]string

// ResolveRole resolves an agent persona name to an AgentRole enum.
func (m RoleMappings) ResolveRole(rawAgent string) AgentRole {
	if strings.TrimSpace(rawAgent) == "" {
		return ""
	}
	if mapped, exists := m[rawAgent]; exists {
		return AgentRole(mapped)
	}
	return AgentRole(rawAgent)
}

// NormalizeTicketID extracts configured literal-prefix IDs without rewriting spelling.
func NormalizeTicketID(raw string, cfg TicketConfig) (string, bool) {
	input := raw
	var offsets []int
	if cfg.CaseInsensitive {
		var lower strings.Builder
		for start, r := range raw {
			encoded := string(unicode.ToLower(r))
			for range []byte(encoded) {
				offsets = append(offsets, start)
			}
			lower.WriteString(encoded)
		}
		offsets = append(offsets, len(raw))
		input = lower.String()
	}
	selected := ""
	start := -1
	for _, p := range cfg.Prefixes {
		prefix := strings.TrimSpace(p.Prefix)
		if prefix == "" {
			continue
		}
		if cfg.CaseInsensitive {
			prefix = strings.ToLower(prefix)
		}
		for offset := 0; offset < len(input); {
			relative := strings.Index(input[offset:], prefix)
			if relative < 0 {
				break
			}
			at := offset + relative
			offset = at + len(prefix)
			if at > 0 && asciiAlphaNumeric(input[at-1]) {
				continue
			}
			end := at + len(prefix)
			if end < len(input) && (input[end] == '-' || input[end] == '_') {
				end++
			}
			digits := end
			for end < len(input) && input[end] >= '0' && input[end] <= '9' {
				end++
			}
			if end == digits || (end < len(input) && asciiAlphaNumeric(input[end])) {
				continue
			}
			originalAt, originalEnd := at, end
			if cfg.CaseInsensitive {
				originalAt, originalEnd = offsets[at], offsets[end]
			}
			if start < 0 || originalAt < start || (originalAt == start && originalEnd-originalAt > len(selected)) {
				selected = raw[originalAt:originalEnd]
				start = originalAt
			}
		}

	}
	return selected, selected != ""
}

// ValidateCalculation requires every threshold consumed by fresh anomaly rules.
func (t AnomalyThresholdsConfig) ValidateCalculation() error {
	if t.ErrorRateMinCalls < 1 {
		return fmt.Errorf("anomaly_thresholds.error_rate_min_calls must be a positive integer")
	}
	if math.IsNaN(t.ErrorRateThreshold) || math.IsInf(t.ErrorRateThreshold, 0) || t.ErrorRateThreshold <= 0 || t.ErrorRateThreshold > 1 || math.IsNaN(t.ErrorRateCriticalThreshold) || math.IsInf(t.ErrorRateCriticalThreshold, 0) || t.ErrorRateCriticalThreshold < t.ErrorRateThreshold || t.ErrorRateCriticalThreshold > 1 {
		return fmt.Errorf("anomaly_thresholds failure-rate ratios must be finite, greater than 0, at most 1, warning <= critical")
	}
	for _, p := range []struct {
		key            string
		warn, critical int64
	}{{"context_ceiling", t.ContextCeilingWarningTokens, t.ContextCeilingCriticalTokens}, {"consecutive_errors_streak", int64(t.ConsecutiveErrorsStreakWarning), int64(t.ConsecutiveErrorsStreakCritical)}, {"tool_thrashing", int64(t.ToolThrashingMinFailures), int64(t.ToolThrashingCriticalFailures)}, {"turn_usage", t.TurnUsageWarningTokens, t.TurnUsageCriticalTokens}, {"context_growth", t.ContextGrowthWarningTokens, t.ContextGrowthCriticalTokens}} {
		if p.warn <= 0 || p.critical < p.warn {
			return fmt.Errorf("anomaly_thresholds.%s requires positive warning/critical values, warning <= critical", p.key)
		}
	}
	return nil
}

// ConfigComment retains supported YAML comments through explicit configuration rewrites.
type ConfigComment struct{ Head, Line, Foot, KeyHead, KeyLine, KeyFoot string }

func asciiAlphaNumeric(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
