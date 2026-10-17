from enum import StrEnum


class ContextHealthStatus(StrEnum):
    HEALTHY = "HEALTHY"
    ELEVATED = "ELEVATED"
    BLOATED = "BLOATED"
    CRITICAL = "CRITICAL"


class AgentRole(StrEnum):
    ORCHESTRATOR = "orchestrator"
    PLANNER = "planner"
    CODER = "coder"
    TESTER = "tester"
    REVIEWER = "reviewer"


class TimeGranularity(StrEnum):
    DAILY = "daily"
    WEEKLY = "weekly"
    CUMULATIVE = "cumulative"


class TimeWindowPreset(StrEnum):
    LAST_7_DAYS = "Last 7 Days"
    LAST_30_DAYS = "Last 30 Days"
    LAST_90_DAYS = "Last 90 Days"
    PROJECT_LIFETIME = "Project Lifetime"
    CUSTOM_RANGE = "Custom Range"
