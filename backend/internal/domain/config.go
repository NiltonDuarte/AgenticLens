package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ProjectConfig holds the complete configuration settings for AgenticLens.
type ProjectConfig struct {
	ProjectName  string                  `json:"project_name" yaml:"project_name"`
	Ticket       TicketConfig            `json:"ticket" yaml:"ticket"`
	Workspaces   WorkspacesConfig        `json:"workspaces" yaml:"workspaces"`
	Database     DatabaseConfig          `json:"database" yaml:"database"`
	Live         LiveConfig              `json:"live" yaml:"live"`
	CodeMetrics  CodeMetricsConfig       `json:"code_metrics" yaml:"code_metrics"`
	RoleMappings RoleMappings            `json:"role_mappings" yaml:"role_mappings"`
}

// TicketPrefix represents a configured ticket identifier prefix and its operational description.
type TicketPrefix struct {
	Prefix      string `json:"prefix" yaml:"prefix"`
	Description string `json:"description" yaml:"description"`
}

// TicketConfig specifies ticket identification, casing rules, and pattern parsing.
type TicketConfig struct {
	Prefixes        []TicketPrefix `json:"prefixes" yaml:"prefixes"`
	CaseInsensitive bool           `json:"case_insensitive" yaml:"case_insensitive"`
	CustomPatterns  []string       `json:"custom_patterns" yaml:"custom_patterns"`
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

// WorkspacesConfig defines filesystem paths for ticket workspaces and source repositories.
type WorkspacesConfig struct {
	Dir        string   `json:"dir" yaml:"dir"`
	SrcFolders []string `json:"src_folders" yaml:"src_folders"`
}

// DatabaseConfig specifies SQLite database connection and query execution parameters.
type DatabaseConfig struct {
	Path           string `json:"path" yaml:"path"`
	QueryTimeoutMs int    `json:"query_timeout_ms" yaml:"query_timeout_ms"`
	BusyTimeoutMs  int    `json:"busy_timeout_ms" yaml:"busy_timeout_ms"`
	ReadOnly       bool   `json:"read_only" yaml:"read_only"`
}

// AnomalyThresholdsConfig defines numerical thresholds for anomaly detection rules.
type AnomalyThresholdsConfig struct {
	StallThresholdSeconds int     `json:"stall_threshold_seconds" yaml:"stall_threshold_seconds"`
	ErrorRateThreshold    float64 `json:"error_rate_threshold" yaml:"error_rate_threshold"`
	ContextExhaustionPct  float64 `json:"context_exhaustion_pct" yaml:"context_exhaustion_pct"`
}

// LiveConfig controls real-time session tracking, polling intervals, and anomaly detection.
type LiveConfig struct {
	ActiveThresholdSeconds int                     `json:"active_threshold_seconds" yaml:"active_threshold_seconds"`
	PollIntervalMs         int                     `json:"poll_interval_ms" yaml:"poll_interval_ms"`
	MaxTraceBufferMessages int                     `json:"max_trace_buffer_messages" yaml:"max_trace_buffer_messages"`
	AnomalyThresholds      AnomalyThresholdsConfig `json:"anomaly_thresholds" yaml:"anomaly_thresholds"`
}

// CodeMetricsConfig defines file classification glob patterns for git metrics.
type CodeMetricsConfig struct {
	Include      []string `json:"include" yaml:"include"`
	Exclude      []string `json:"exclude" yaml:"exclude"`
	TestPatterns []string `json:"test_patterns" yaml:"test_patterns"`
	ProdPatterns []string `json:"prod_patterns" yaml:"prod_patterns"`
}

// RoleMappings maps agent persona names to canonical AgentRole enums.
type RoleMappings map[string]string

// ResolveRole resolves an agent persona name to a canonical AgentRole enum.
func (m RoleMappings) ResolveRole(rawAgent string) AgentRole {
	cleaned := strings.ToLower(strings.TrimSpace(rawAgent))
	if cleaned == "" {
		return RoleOrchestrator
	}

	if m != nil {
		if mapped, ok := m[cleaned]; ok && strings.TrimSpace(mapped) != "" {
			return AgentRole(strings.ToLower(strings.TrimSpace(mapped)))
		}
		if mapped, ok := m[rawAgent]; ok && strings.TrimSpace(mapped) != "" {
			return AgentRole(strings.ToLower(strings.TrimSpace(mapped)))
		}
		for k, v := range m {
			if strings.EqualFold(strings.TrimSpace(k), cleaned) && strings.TrimSpace(v) != "" {
				return AgentRole(strings.ToLower(strings.TrimSpace(v)))
			}
		}
	}

	switch cleaned {
	case "orchestrator", "build", "root":
		return RoleOrchestrator
	case "planner":
		return RolePlanner
	case "coder":
		return RoleCoder
	case "tester":
		return RoleTester
	case "reviewer":
		return RoleReviewer
	default:
		return AgentRole(cleaned)
	}
}

// DefaultProjectConfig returns the standard baseline configuration with domain defaults.
func DefaultProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "AgenticLens",
		Ticket: TicketConfig{
			Prefixes: []TicketPrefix{
				{
					Prefix:      "AL",
					Description: "Core AgenticLens development features, bugfixes, and technical initiatives",
				},
				{
					Prefix:      "GEN",
					Description: "General multi-agent engineering tasks and ad-hoc untracked work",
				},
				{
					Prefix:      "PO",
					Description: "Product Owner specifications, requirements definitions, and roadmap breakdowns",
				},
			},
			CaseInsensitive: true,
			CustomPatterns:  []string{},
		},
		Workspaces: WorkspacesConfig{
			Dir:        "workspaces",
			SrcFolders: []string{"web", "backend"},
		},
		Database: DatabaseConfig{
			Path:           "~/.local/share/opencode/opencode.db",
			QueryTimeoutMs: 5000,
			BusyTimeoutMs:  5000,
			ReadOnly:       true,
		},
		Live: LiveConfig{
			ActiveThresholdSeconds: 300,
			PollIntervalMs:         2000,
			MaxTraceBufferMessages: 500,
			AnomalyThresholds: AnomalyThresholdsConfig{
				StallThresholdSeconds: 120,
				ErrorRateThreshold:    0.25,
				ContextExhaustionPct:  0.85,
			},
		},
		CodeMetrics: CodeMetricsConfig{
			Include:      []string{},
			Exclude:      []string{},
			TestPatterns: []string{},
			ProdPatterns: []string{},
		},
		RoleMappings: RoleMappings{
			"orchestrator": "orchestrator",
			"build":        "orchestrator",
			"planner":      "planner",
			"coder":        "coder",
			"tester":       "tester",
			"reviewer":     "reviewer",
		},
	}
}

// NormalizeTicketID extracts and normalizes a ticket identifier according to configuration rules.
func NormalizeTicketID(raw string, cfg TicketConfig) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	prefixes := cfg.PrefixNames()
	if len(prefixes) == 0 {
		prefixes = []string{"AL", "GEN", "PO"}
	}

	sortedPrefixes := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		clean := strings.TrimSpace(p)
		clean = strings.TrimSuffix(clean, "-")
		clean = strings.TrimSuffix(clean, "_")
		if clean != "" {
			sortedPrefixes = append(sortedPrefixes, clean)
		}
	}
	if len(sortedPrefixes) == 0 {
		sortedPrefixes = []string{"AL", "GEN", "PO"}
	}
	sort.Slice(sortedPrefixes, func(i, j int) bool {
		return len(sortedPrefixes[i]) > len(sortedPrefixes[j])
	})

	for _, prefix := range sortedPrefixes {
		var pattern string
		escaped := regexp.QuoteMeta(prefix)
		if cfg.CaseInsensitive {
			pattern = `(?i)(?:^|[^a-zA-Z0-9])` + escaped + `-(\d+)(?:[^a-zA-Z0-9]|$)`
		} else {
			pattern = `(?:^|[^a-zA-Z0-9])` + escaped + `-(\d+)(?:[^a-zA-Z0-9]|$)`
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			digits := matches[1]
			canonical := fmt.Sprintf("%s-%s", strings.ToUpper(prefix), digits)
			return canonical, true
		}
	}

	for _, pattern := range cfg.CustomPatterns {
		if strings.TrimSpace(pattern) == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) > 1 && matches[1] != "" {
			return strings.ToUpper(strings.TrimSpace(matches[1])), true
		} else if len(matches) > 0 && matches[0] != "" {
			return strings.ToUpper(strings.TrimSpace(matches[0])), true
		}
	}

	return "", false
}
