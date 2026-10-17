package domain

// TimeWindow defines standard date filtering query parameters.
type TimeWindow string

const (
	TimeWindowAll TimeWindow = "all"
	TimeWindow7D  TimeWindow = "7d"
	TimeWindow30D TimeWindow = "30d"
	TimeWindow90D TimeWindow = "90d"
)

// ParseTimeWindow parses a string into a TimeWindow enum, defaulting to TimeWindowAll.
func ParseTimeWindow(val string) TimeWindow {
	switch val {
	case "7d":
		return TimeWindow7D
	case "30d":
		return TimeWindow30D
	case "90d":
		return TimeWindow90D
	case "all":
		return TimeWindowAll
	default:
		return TimeWindowAll
	}
}

// LedgerSortField defines sorting dimensions for agentic metrics ledgers.
type LedgerSortField string

const (
	SortFieldUpdatedAt   LedgerSortField = "updated_at"
	SortFieldCost        LedgerSortField = "cost"
	SortFieldDuration    LedgerSortField = "duration"
	SortFieldInvocations LedgerSortField = "invocations"
)

// ParseLedgerSortField parses a string into a LedgerSortField enum, defaulting to SortFieldUpdatedAt.
func ParseLedgerSortField(val string) LedgerSortField {
	switch val {
	case "cost":
		return SortFieldCost
	case "duration":
		return SortFieldDuration
	case "invocations":
		return SortFieldInvocations
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
	GranularityDaily      TimeGranularity = "daily"
	GranularityWeekly     TimeGranularity = "weekly"
	GranularityCumulative TimeGranularity = "cumulative"
)

// TimeWindowPreset defines standard date filtering ranges.
type TimeWindowPreset string

const (
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
