package domain

import (
	"sort"
	"strings"
)

// TimeWindow defines standard date filtering query parameters.
type TimeWindow string

const (
	TimeWindowAll    TimeWindow = "all"
	TimeWindow1D     TimeWindow = "1d"
	TimeWindow7D     TimeWindow = "7d"
	TimeWindow30D    TimeWindow = "30d"
	TimeWindow90D    TimeWindow = "90d"
	TimeWindowCustom TimeWindow = "custom"
)

// ParseTimeWindow parses a string into a TimeWindow enum, defaulting to TimeWindowAll.
func ParseTimeWindow(val string) TimeWindow {
	switch val {
	case "1d":
		return TimeWindow1D
	case "7d":
		return TimeWindow7D
	case "30d":
		return TimeWindow30D
	case "90d":
		return TimeWindow90D
	case "custom":
		return TimeWindowCustom
	case "all":
		return TimeWindowAll
	default:
		return TimeWindowAll
	}
}

// LedgerSortField defines sorting dimensions for agentic metrics ledgers.
type LedgerSortField string

const (
	SortFieldCompletedAt LedgerSortField = "completed_at"
	SortFieldUpdatedAt   LedgerSortField = "updated_at"
	SortFieldCost        LedgerSortField = "cost"
	SortFieldDuration    LedgerSortField = "duration"
	SortFieldInvocations LedgerSortField = "invocations"
	SortFieldTicketID    LedgerSortField = "ticket_id"
)

// ParseLedgerSortField parses a string into a LedgerSortField enum, defaulting to SortFieldUpdatedAt.
func ParseLedgerSortField(val string) LedgerSortField {
	switch val {
	case "completed_at":
		return SortFieldCompletedAt
	case "cost":
		return SortFieldCost
	case "duration":
		return SortFieldDuration
	case "invocations":
		return SortFieldInvocations
	case "ticket_id":
		return SortFieldTicketID
	case "updated_at":
		return SortFieldUpdatedAt
	default:
		return SortFieldUpdatedAt
	}
}

// ContextHealthStatus defines the context window health classification.
type ContextHealthStatus string

const (
	ContextHealthHealthy  ContextHealthStatus = "HEALTHY"
	ContextHealthElevated ContextHealthStatus = "ELEVATED"
	ContextHealthBloated  ContextHealthStatus = "BLOATED"
	ContextHealthCritical ContextHealthStatus = "CRITICAL"
)

// AgentRole defines the role of an agent invocation.
type AgentRole string

const (
	RoleOrchestrator AgentRole = "orchestrator"
	RolePlanner      AgentRole = "planner"
	RoleCoder        AgentRole = "coder"
	RoleTester       AgentRole = "tester"
	RoleReviewer     AgentRole = "reviewer"
)

// TimeGranularity defines aggregation time windows for progression charts.
type TimeGranularity string

const (
	GranularityHourly TimeGranularity = "hourly"
	GranularityDaily  TimeGranularity = "daily"
	GranularityWeekly TimeGranularity = "weekly"
)

// ParseTimeGranularity parses a string into a TimeGranularity enum, returning false if unrecognized.
func ParseTimeGranularity(val string) (TimeGranularity, bool) {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case string(GranularityHourly):
		return GranularityHourly, true
	case string(GranularityDaily):
		return GranularityDaily, true
	case string(GranularityWeekly):
		return GranularityWeekly, true
	default:
		return "", false
	}
}

// TimeWindowPreset defines standard date filtering ranges.
type TimeWindowPreset string

const (
	TimeWindowLast24Hours     TimeWindowPreset = "Last 24 Hours"
	TimeWindowLast7Days        TimeWindowPreset = "Last 7 Days"
	TimeWindowLast30Days       TimeWindowPreset = "Last 30 Days"
	TimeWindowLast90Days       TimeWindowPreset = "Last 90 Days"
	TimeWindowProjectLifetime  TimeWindowPreset = "Project Lifetime"
	TimeWindowCustomRange      TimeWindowPreset = "Custom Range"
)

// CanonicalRoleOrder defines the display priority order for agent roles.
var CanonicalRoleOrder = []string{
	string(RoleOrchestrator),
	string(RolePlanner),
	string(RoleCoder),
	string(RoleTester),
	string(RoleReviewer),
}

// SubagentRoleOrder defines the display priority order for subagents (excluding orchestrator).
var SubagentRoleOrder = []string{
	string(RolePlanner),
	string(RoleCoder),
	string(RoleTester),
	string(RoleReviewer),
}

// IsSubagentRole returns true if the trimmed and lowercase-normalized role is non-empty
// and not equal to RoleOrchestrator.
func IsSubagentRole(role string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(role))
	return trimmed != "" && trimmed != string(RoleOrchestrator)
}

// OrderRoles deduplicates and orders roles by CanonicalRoleOrder first,
// followed by remaining custom roles sorted in alphabetical order.
func OrderRoles(roles []string) []string {
	present := make(map[string]struct{})
	for _, r := range roles {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			present[trimmed] = struct{}{}
		}
	}
	canonicalSet := make(map[string]struct{}, len(CanonicalRoleOrder))
	for _, r := range CanonicalRoleOrder {
		canonicalSet[r] = struct{}{}
	}
	var ordered []string
	for _, r := range CanonicalRoleOrder {
		if _, exists := present[r]; exists {
			ordered = append(ordered, r)
		}
	}
	var custom []string
	for r := range present {
		if _, isCanonical := canonicalSet[r]; !isCanonical {
			custom = append(custom, r)
		}
	}
	sort.Strings(custom)
	return append(ordered, custom...)
}

// OrderSubagentRoles deduplicates and orders subagent roles by SubagentRoleOrder first,
// followed by remaining custom subagent roles sorted in alphabetical order.
func OrderSubagentRoles(roles []string) []string {
	present := make(map[string]struct{})
	for _, r := range roles {
		trimmed := strings.TrimSpace(r)
		if IsSubagentRole(trimmed) {
			present[trimmed] = struct{}{}
		}
	}
	canonicalSet := make(map[string]struct{}, len(SubagentRoleOrder))
	for _, r := range SubagentRoleOrder {
		canonicalSet[r] = struct{}{}
	}
	var ordered []string
	for _, r := range SubagentRoleOrder {
		if _, exists := present[r]; exists {
			ordered = append(ordered, r)
		}
	}
	var custom []string
	for r := range present {
		if _, isCanonical := canonicalSet[r]; !isCanonical {
			custom = append(custom, r)
		}
	}
	sort.Strings(custom)
	return append(ordered, custom...)
}

// MigrationStatus defines the schema and metric migration progress state.
type MigrationStatus string

const (
	MigrationStatusPending   MigrationStatus = "pending"
	MigrationStatusPartial   MigrationStatus = "partial"
	MigrationStatusCompleted MigrationStatus = "completed"
)

// EventType defines discrete lifecycle event occurrences within an agent execution session.
type EventType string

const (
	EventTypeToolCall    EventType = "tool_call"
	EventTypeTurnMessage EventType = "turn_message"
	EventTypeCompaction  EventType = "compaction"
)

// AnomalySeverity defines the severity classification of a detected anomaly.
type AnomalySeverity string

const (
	SeverityWarning  AnomalySeverity = "WARNING"
	SeverityCritical AnomalySeverity = "CRITICAL"
)

// AnomalyRuleID defines the unique identifier for a domain anomaly rule.
type AnomalyRuleID string

const (
	RuleToolRetryStreak        AnomalyRuleID = "TOOL_RETRY_STREAK"
	RuleToolHighFailureRate    AnomalyRuleID = "TOOL_HIGH_FAILURE_RATE"
	RuleContextSaturationSpike AnomalyRuleID = "CONTEXT_SATURATION_SPIKE"
	RuleSubagentReworkLoop     AnomalyRuleID = "SUBAGENT_REWORK_LOOP"
)

// ConventionalTicketType defines canonical conventional commit classifications.
type ConventionalTicketType string

const (
	TicketTypeFeat     ConventionalTicketType = "feat"
	TicketTypeFix      ConventionalTicketType = "fix"
	TicketTypeDocs     ConventionalTicketType = "docs"
	TicketTypeChore    ConventionalTicketType = "chore"
	TicketTypeRefactor ConventionalTicketType = "refactor"
	TicketTypeTest     ConventionalTicketType = "test"
	TicketTypeCI       ConventionalTicketType = "ci"
	TicketTypePerf     ConventionalTicketType = "perf"
	TicketTypeBuild    ConventionalTicketType = "build"
)

// CanonicalTicketTypes lists all supported conventional commit types in canonical order.
var CanonicalTicketTypes = []ConventionalTicketType{
	TicketTypeFeat,
	TicketTypeFix,
	TicketTypeDocs,
	TicketTypeChore,
	TicketTypeRefactor,
	TicketTypeTest,
	TicketTypeCI,
	TicketTypePerf,
	TicketTypeBuild,
}

// NormalizeTicketType normalizes raw strings and aliases into canonical conventional commit types.
func NormalizeTicketType(raw string) (ConventionalTicketType, bool) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	switch trimmed {
	case "feat", "feature":
		return TicketTypeFeat, true
	case "fix", "bug", "bugfix":
		return TicketTypeFix, true
	case "docs", "doc", "documentation":
		return TicketTypeDocs, true
	case "chore", "task":
		return TicketTypeChore, true
	case "refactor":
		return TicketTypeRefactor, true
	case "test", "tests":
		return TicketTypeTest, true
	case "ci":
		return TicketTypeCI, true
	case "perf", "performance":
		return TicketTypePerf, true
	case "build":
		return TicketTypeBuild, true
	default:
		return "", false
	}
}

// IsValidTicketType reports whether the raw ticket type resolves to a valid conventional type.
func IsValidTicketType(raw string) bool {
	_, ok := NormalizeTicketType(raw)
	return ok
}
