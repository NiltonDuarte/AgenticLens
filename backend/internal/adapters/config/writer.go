package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"agenticlens/internal/domain"
)

// ErrFileExists is returned when target configuration file already exists and force is false.
var ErrFileExists = errors.New("configuration file already exists")

// DefaultTestPatterns defines standard fallback glob patterns for identifying test code.
var DefaultTestPatterns = []string{
	"**/*_test.go",
	"**/*.test.ts",
	"**/*.test.tsx",
	"**/*.spec.ts",
	"**/*.spec.tsx",
	"**/*.test.js",
	"**/*.test.jsx",
	"**/*.spec.js",
	"**/*.spec.jsx",
	"**/test_*.py",
	"**/*_test.py",
	"**/testdata/**",
	"**/__tests__/**",
}

// DefaultProdPatterns defines standard fallback glob patterns for identifying production runtime code.
var DefaultProdPatterns = []string{
	"**/*.go",
	"**/*.ts",
	"**/*.tsx",
	"**/*.py",
	"**/*.js",
	"**/*.jsx",
	"**/*.html",
	"**/*.css",
}

// FileExistsError is an explicit typed error containing the conflicting file path.
type FileExistsError struct {
	Path string
}

func (e *FileExistsError) Error() string {
	return fmt.Sprintf("configuration file already exists: %s", e.Path)
}

func (e *FileExistsError) Is(target error) bool {
	return target == ErrFileExists
}

// GenerateConfigYAML produces cleanly formatted, richly commented YAML bytes for a ProjectConfig.
func GenerateConfigYAML(cfg domain.ProjectConfig) ([]byte, error) {
	var b strings.Builder

	b.WriteString("# AgenticLens Global Configuration\n")
	b.WriteString("# Defines project parameters, ticket identification patterns, workspace layout,\n")
	b.WriteString("# database connectivity, live observability thresholds, code metrics classification,\n")
	b.WriteString("# and agent-backend bridge role mappings.\n\n")

	// Project Name
	projName := strings.TrimSpace(cfg.ProjectName)
	if projName == "" {
		return nil, fmt.Errorf("project_name cannot be empty")
	}
	b.WriteString("# Human-readable name of the project.\n")
	b.WriteString(fmt.Sprintf("project_name: %s\n\n", projName))

	// Ticket Settings
	var cleanPrefixes []domain.TicketPrefix
	for _, p := range cfg.Ticket.Prefixes {
		cleanPrefix := strings.TrimSpace(p.Prefix)
		if cleanPrefix != "" {
			cleanPrefixes = append(cleanPrefixes, domain.TicketPrefix{
				Prefix:      cleanPrefix,
				Description: strings.TrimSpace(p.Description),
				TicketType:  strings.TrimSpace(p.TicketType),
			})
		}
	}
	if len(cleanPrefixes) == 0 {
		return nil, fmt.Errorf("ticket.prefixes cannot be empty")
	}

	b.WriteString("# Ticket identification and normalization settings.\n")
	b.WriteString("ticket:\n")
	b.WriteString("  # Supported ticket prefix identifiers for this repository.\n")
	b.WriteString("  # Accepts structured mappings (prefix and description) or plain string scalars.\n")
	b.WriteString("  prefixes:\n")

	for _, p := range cleanPrefixes {
		b.WriteString(fmt.Sprintf("    - prefix: %s\n", p.Prefix))
		if strings.TrimSpace(p.Description) != "" {
			b.WriteString(fmt.Sprintf("      description: %s\n", strings.TrimSpace(p.Description)))
		}
		if strings.TrimSpace(p.TicketType) != "" {
			b.WriteString(fmt.Sprintf("      ticket_type: %s\n", strings.TrimSpace(p.TicketType)))
		}
	}

	b.WriteString("  # Enable case-insensitive prefix parsing (e.g. al-21 -> AL-21).\n")
	b.WriteString(fmt.Sprintf("  case_insensitive: %t\n", cfg.Ticket.CaseInsensitive))
	b.WriteString("  # Additional regex patterns to match unconventional ticket naming formats.\n")
	b.WriteString("  custom_patterns:")
	formatStringList(&b, cfg.Ticket.CustomPatterns, "    ")
	b.WriteString("\n")

	// Workspaces Settings
	wsDir := strings.TrimSpace(cfg.Workspaces.Dir)
	var cleanSourceRoots []string
	for _, sr := range cfg.CodeMetrics.SourceRoots {
		clean := strings.TrimSpace(sr)
		if clean != "" {
			cleanSourceRoots = append(cleanSourceRoots, clean)
		}
	}
	var cleanSrcFolders []string
	for _, sf := range cfg.Workspaces.SrcFolders {
		clean := strings.TrimSpace(sf)
		if clean != "" {
			cleanSrcFolders = append(cleanSrcFolders, clean)
		}
	}
	if len(cleanSourceRoots) == 0 && len(cleanSrcFolders) > 0 {
		cleanSourceRoots = append([]string(nil), cleanSrcFolders...)
	} else if len(cleanSrcFolders) == 0 && len(cleanSourceRoots) > 0 {
		cleanSrcFolders = append([]string(nil), cleanSourceRoots...)
	}
	if len(cleanSourceRoots) == 0 {
		return nil, fmt.Errorf("workspaces.src_folders cannot be empty")
	}

	b.WriteString("# Workspace storage and source folder definitions.\n")
	b.WriteString("workspaces:\n")
	b.WriteString("  # Directory where ticket artifacts and session telemetry ledgers are stored.\n")
	if wsDir != "" {
		b.WriteString(fmt.Sprintf("  dir: %s\n", wsDir))
	}
	if len(cleanSrcFolders) > 0 {
		b.WriteString("  # Monitored source code folders for project telemetry and change tracking.\n")
		b.WriteString("  src_folders:")
		formatStringList(&b, cleanSrcFolders, "    ")
	}
	b.WriteString("  # Configured workspace documentation artifacts to read and render in dashboard.\n")
	artifacts := cfg.Workspaces.Artifacts
	if len(artifacts) == 0 {
		b.WriteString("  artifacts: []\n\n")
	} else {
		b.WriteString("  artifacts:\n")
		for _, art := range artifacts {
			cleanPath := strings.TrimSpace(art.Path)
			cleanTitle := strings.TrimSpace(art.Title)
			if cleanPath == "" {
				continue
			}
			if cleanTitle == "" {
				cleanTitle = filepath.Base(cleanPath)
			}
			b.WriteString(fmt.Sprintf("    - path: %s\n", cleanPath))
			b.WriteString(fmt.Sprintf("      title: %s\n", cleanTitle))
			if strings.TrimSpace(art.Description) != "" {
				b.WriteString(fmt.Sprintf("      description: %s\n", strings.TrimSpace(art.Description)))
			}
		}
		b.WriteString("\n")
	}

	// Database Settings
	dbPath := strings.TrimSpace(cfg.Database.Path)
	if dbPath == "" {
		dbPath = "~/.local/share/opencode/opencode.db"
	}
	queryTimeout := cfg.Database.QueryTimeoutMs
	if queryTimeout <= 0 {
		queryTimeout = 5000
	}
	busyTimeout := cfg.Database.BusyTimeoutMs
	if busyTimeout <= 0 {
		busyTimeout = 5000
	}
	b.WriteString("# OpenCode SQLite database connectivity and timeout parameters.\n")
	b.WriteString("database:\n")
	b.WriteString("  # Path to local OpenCode SQLite database file.\n")
	b.WriteString(fmt.Sprintf("  path: %s\n", dbPath))
	b.WriteString("  # Query execution timeout in milliseconds.\n")
	b.WriteString(fmt.Sprintf("  query_timeout_ms: %d\n", queryTimeout))
	b.WriteString("  # SQLite busy handler wait timeout in milliseconds.\n")
	b.WriteString(fmt.Sprintf("  busy_timeout_ms: %d\n", busyTimeout))
	b.WriteString("  # Enforce read-only connection to prevent unintended database mutations.\n")
	b.WriteString(fmt.Sprintf("  read_only: %t\n\n", cfg.Database.ReadOnly))

	// Live Settings
	activeSec := cfg.Live.ActiveThresholdSeconds
	if activeSec <= 0 {
		activeSec = 300
	}
	pollMs := cfg.Live.PollIntervalMs
	if pollMs <= 0 {
		pollMs = 2000
	}
	maxTrace := cfg.Live.MaxTraceBufferMessages
	if maxTrace <= 0 {
		maxTrace = 500
	}

	b.WriteString("# Live session observability and real-time streaming engine configuration.\n")
	b.WriteString("live:\n")
	b.WriteString("  # Window in seconds within which a session is considered actively running.\n")
	b.WriteString(fmt.Sprintf("  active_threshold_seconds: %d\n", activeSec))
	b.WriteString("  # SQLite polling interval in milliseconds for live session updates.\n")
	b.WriteString(fmt.Sprintf("  poll_interval_ms: %d\n", pollMs))
	b.WriteString("  # Maximum number of recent trace messages retained in memory buffer per session.\n")
	b.WriteString(fmt.Sprintf("  max_trace_buffer_messages: %d\n\n", maxTrace))

	// Anomaly Thresholds Settings
	stallSec := cfg.AnomalyThresholds.StallThresholdSeconds
	if stallSec <= 0 {
		stallSec = 120
	}
	errorRate := cfg.AnomalyThresholds.ErrorRateThreshold
	if errorRate <= 0 {
		errorRate = 0.25
	}
	contextPct := cfg.AnomalyThresholds.ContextExhaustionPct
	if contextPct <= 0 {
		contextPct = 0.85
	}

	b.WriteString("# Automated anomaly detection rules governing offline generation and live stream monitoring.\n")
	b.WriteString("anomaly_thresholds:\n")
	b.WriteString("  # Time in seconds without session turn activity before flagging a stall anomaly.\n")
	b.WriteString(fmt.Sprintf("  stall_threshold_seconds: %d\n", stallSec))
	b.WriteString("  # Ratio of tool errors to total invocations triggering a high failure anomaly.\n")
	b.WriteString(fmt.Sprintf("  error_rate_threshold: %0.2f\n", errorRate))
	b.WriteString("  # Token context window utilization percentage triggering saturation warnings.\n")
	b.WriteString(fmt.Sprintf("  context_exhaustion_pct: %0.2f\n", contextPct))
	at := cfg.AnomalyThresholds
	if at.ConsecutiveErrorsStreakWarning > 0 {
		b.WriteString(fmt.Sprintf("  consecutive_errors_streak_warning: %d\n", at.ConsecutiveErrorsStreakWarning))
	}
	if at.ConsecutiveErrorsStreakCritical > 0 {
		b.WriteString(fmt.Sprintf("  consecutive_errors_streak_critical: %d\n", at.ConsecutiveErrorsStreakCritical))
	}
	b.WriteString("  # Fixed context ceiling token thresholds triggering warning and critical alerts.\n")
	b.WriteString("  # context_ceiling_critical_tokens governs context saturation percentage and health status grading.\n")
	ceilingWarn := at.ContextCeilingWarningTokens
	if ceilingWarn <= 0 {
		ceilingWarn = domain.ContextCeilingWarningTokens
	}
	ceilingCrit := at.ContextCeilingCriticalTokens
	if ceilingCrit <= 0 {
		ceilingCrit = domain.ContextCeilingCriticalTokens
	}
	b.WriteString(fmt.Sprintf("  context_ceiling_warning_tokens: %d\n", ceilingWarn))
	b.WriteString(fmt.Sprintf("  context_ceiling_critical_tokens: %d\n", ceilingCrit))
	if at.TokenVelocityWarningTokens > 0 {
		b.WriteString(fmt.Sprintf("  token_velocity_warning_tokens: %d\n", at.TokenVelocityWarningTokens))
	}
	if at.TokenVelocityCriticalTokens > 0 {
		b.WriteString(fmt.Sprintf("  token_velocity_critical_tokens: %d\n", at.TokenVelocityCriticalTokens))
	}
	if at.TokenVelocityWarningDeltaTokens > 0 {
		b.WriteString(fmt.Sprintf("  token_velocity_warning_delta_tokens: %d\n", at.TokenVelocityWarningDeltaTokens))
	}
	if at.TokenVelocityCriticalDeltaTokens > 0 {
		b.WriteString(fmt.Sprintf("  token_velocity_critical_delta_tokens: %d\n", at.TokenVelocityCriticalDeltaTokens))
	}
	if at.ToolThrashingMinFailures > 0 {
		b.WriteString(fmt.Sprintf("  tool_thrashing_min_failures: %d\n", at.ToolThrashingMinFailures))
	}
	if at.ToolThrashingCriticalFailures > 0 {
		b.WriteString(fmt.Sprintf("  tool_thrashing_critical_failures: %d\n", at.ToolThrashingCriticalFailures))
	}
	b.WriteString("\n")

	// Code Metrics Settings
	b.WriteString("# Git commit diff analysis and code classification glob patterns.\n")
	b.WriteString("code_metrics:\n")
	b.WriteString("  # Monitored source code root directories for project telemetry and change tracking.\n")
	b.WriteString("  source_roots:")
	formatStringList(&b, cleanSourceRoots, "    ")
	b.WriteString("  # Specific files or glob patterns explicitly included as production code.\n")
	b.WriteString("  include:")
	formatStringList(&b, cfg.CodeMetrics.Include, "    ")
	b.WriteString("  # Files or directories excluded from code volume and diff calculations.\n")
	b.WriteString("  exclude:")
	formatStringList(&b, cfg.CodeMetrics.Exclude, "    ")
	b.WriteString("  # Glob patterns identifying test files, test fixtures, and test utilities.\n")
	b.WriteString("  test_patterns:")
	formatStringList(&b, cfg.CodeMetrics.TestPatterns, "    ")
	b.WriteString("  # Glob patterns identifying production runtime code.\n")
	b.WriteString("  prod_patterns:")
	formatStringList(&b, cfg.CodeMetrics.ProdPatterns, "    ")
	b.WriteString("\n")

	// Role Mappings
	b.WriteString("# Bridge definitions mapping agent persona names to canonical AgentRole enums.\n")
	b.WriteString("# Canonical domain roles: orchestrator, planner, coder, tester, reviewer.\n")
	mappings := cfg.RoleMappings
	if len(mappings) == 0 {
		b.WriteString("role_mappings: {}\n\n")
	} else {
		b.WriteString("role_mappings:\n")
		roleKeys := make([]string, 0, len(mappings))
		for k := range mappings {
			roleKeys = append(roleKeys, k)
		}
		sort.Strings(roleKeys)
		for _, k := range roleKeys {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, mappings[k]))
		}
		b.WriteString("\n")
	}

	// Role Order
	b.WriteString("# Configurable sequence for sorting agent roles in CLI tables and metric cards.\n")
	b.WriteString("# Roles appearing here are sorted first in this exact order; unlisted roles follow alphabetically.\n")
	roleOrder := cfg.RoleOrder
	if len(roleOrder) == 0 {
		b.WriteString("role_order: []\n")
	} else {
		b.WriteString("role_order:\n")
		for _, r := range roleOrder {
			b.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	result := []byte(b.String())

	// Validate output syntax with yaml parser
	var validationNode yaml.Node
	if err := yaml.Unmarshal(result, &validationNode); err != nil {
		return nil, fmt.Errorf("validating generated yaml syntax: %w", err)
	}

	return result, nil
}

// GenerateDefaultYAML creates configuration YAML bytes initialized from default settings with overrides.
func GenerateDefaultYAML(projectName string, workspacesDir string, dbPath string, srcFolders []string, prefixes []domain.TicketPrefix, customArtifacts ...[]domain.WorkspaceArtifactConfig) ([]byte, error) {
	projName := strings.TrimSpace(projectName)
	if projName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	wsDir := strings.TrimSpace(workspacesDir)
	if wsDir == "" {
		wsDir = "workspaces"
	}
	databasePath := strings.TrimSpace(dbPath)
	if databasePath == "" {
		databasePath = "~/.local/share/opencode/opencode.db"
	}
	var cleanFolders []string
	for _, f := range srcFolders {
		trimmed := strings.TrimSpace(f)
		if trimmed != "" {
			cleanFolders = append(cleanFolders, trimmed)
		}
	}
	if len(cleanFolders) == 0 {
		return nil, fmt.Errorf("source folders cannot be empty")
	}
	var cleanPrefixes []domain.TicketPrefix
	for _, p := range prefixes {
		trimmed := strings.TrimSpace(p.Prefix)
		if trimmed != "" {
			cleanPrefixes = append(cleanPrefixes, domain.TicketPrefix{
				Prefix:      trimmed,
				Description: strings.TrimSpace(p.Description),
				TicketType:  strings.TrimSpace(p.TicketType),
			})
		}
	}
	if len(cleanPrefixes) == 0 {
		return nil, fmt.Errorf("ticket prefixes cannot be empty")
	}

	var artifacts []domain.WorkspaceArtifactConfig
	if len(customArtifacts) > 0 && len(customArtifacts[0]) > 0 {
		artifacts = customArtifacts[0]
	}

	cfg := domain.ProjectConfig{
		ProjectName: projName,
		Ticket: domain.TicketConfig{
			Prefixes:        cleanPrefixes,
			CaseInsensitive: true,
			CustomPatterns:  []string{},
		},
		Workspaces: domain.WorkspacesConfig{
			Dir:        wsDir,
			SrcFolders: cleanFolders,
			Artifacts:  artifacts,
		},
		Database: domain.DatabaseConfig{
			Path:           databasePath,
			QueryTimeoutMs: 5000,
			BusyTimeoutMs:  5000,
			ReadOnly:       true,
		},
		Live: domain.LiveConfig{
			ActiveThresholdSeconds: 300,
			PollIntervalMs:         2000,
			MaxTraceBufferMessages: 500,
		},
		AnomalyThresholds: domain.DefaultAnomalyThresholdsConfig(),
		CodeMetrics: domain.CodeMetricsConfig{
			SourceRoots:  cleanFolders,
			Include:      []string{},
			Exclude:      []string{},
			TestPatterns: append([]string(nil), DefaultTestPatterns...),
			ProdPatterns: append([]string(nil), DefaultProdPatterns...),
		},
		RoleMappings: domain.RoleMappings{},
		RoleOrder:    []string{},
	}

	return GenerateConfigYAML(cfg)
}

// GenerateDefaultConfigYAML creates configuration YAML bytes initialized from default settings with overrides.
func GenerateDefaultConfigYAML(projectName string, workspacesDir string, dbPath string, srcFolders []string, prefixes []domain.TicketPrefix, customArtifacts ...[]domain.WorkspaceArtifactConfig) ([]byte, error) {
	return GenerateDefaultYAML(projectName, workspacesDir, dbPath, srcFolders, prefixes, customArtifacts...)
}

// WriteConfigFile writes configuration bytes to targetPath atomically with collision detection.
// If the target file already exists and force is false, it returns a *FileExistsError.
func WriteConfigFile(targetPath string, data []byte, force bool) error {
	cleanPath := filepath.Clean(targetPath)

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating parent directory %s: %w", dir, err)
	}

	if !force {
		if _, err := os.Stat(cleanPath); err == nil {
			return &FileExistsError{Path: cleanPath}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("checking file existence %s: %w", cleanPath, err)
		}
	}

	tempFile, err := os.CreateTemp(dir, ".agenticlens-config-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary config file: %w", err)
	}
	tmpName := tempFile.Name()
	cleaned := false
	defer func() {
		if !cleaned {
			if removeErr := os.Remove(tmpName); removeErr != nil && !os.IsNotExist(removeErr) {
				// Non-fatal cleanup
			}
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			return fmt.Errorf("closing temp file after write error: %w", closeErr)
		}
		return fmt.Errorf("writing temporary config file %s: %w", tmpName, err)
	}

	if err := tempFile.Sync(); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			return fmt.Errorf("closing temp file after sync error: %w", closeErr)
		}
		return fmt.Errorf("syncing temporary config file %s: %w", tmpName, err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temporary config file %s: %w", tmpName, err)
	}

	if err := os.Rename(tmpName, cleanPath); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmpName, cleanPath, err)
	}

	cleaned = true
	return nil
}

// WriteConfig converts cfg to YAML and writes it atomically to targetPath.
func WriteConfig(targetPath string, cfg domain.ProjectConfig, force bool) error {
	data, err := GenerateConfigYAML(cfg)
	if err != nil {
		return fmt.Errorf("generating config yaml: %w", err)
	}
	return WriteConfigFile(targetPath, data, force)
}

func formatStringList(b *strings.Builder, items []string, indent string) {
	if len(items) == 0 {
		b.WriteString(" []\n")
		return
	}
	b.WriteString("\n")
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean != "" {
			b.WriteString(fmt.Sprintf("%s- %s\n", indent, formatStringItem(clean)))
		}
	}
}

func formatStringItem(item string) string {
	if strings.HasPrefix(item, "\"") && strings.HasSuffix(item, "\"") && len(item) >= 2 {
		return item
	}
	if strings.HasPrefix(item, "'") && strings.HasSuffix(item, "'") && len(item) >= 2 {
		return item
	}
	if strings.ContainsAny(item, "*?[]{}#:&@!`") || strings.HasPrefix(item, "-") {
		return fmt.Sprintf("%q", item)
	}
	return item
}
