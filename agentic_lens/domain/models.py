from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class InvocationTokens:
    input: int
    output: int
    cache_read: int
    reasoning: int
    peak_context: int


@dataclass(frozen=True, slots=True)
class InvocationFact:
    id: str
    parent_id: str | None
    role: str
    model: str
    user: str
    started_at: str | None
    completed_at: str | None
    active_duration_seconds: int
    turns: int
    tokens: InvocationTokens
    cost_usd: float
    compactions: int = 0


@dataclass(frozen=True, slots=True)
class ToolFact:
    name: str
    calls: int
    errors: int


@dataclass(frozen=True, slots=True)
class CodeFacts:
    lines_added: int
    lines_deleted: int
    test_lines_added: int
    test_lines_deleted: int
    prod_lines_added: int
    prod_lines_deleted: int
    files_changed: int
    sdlc_iterations: int


@dataclass(frozen=True, slots=True)
class AgenticFactLedger:
    ticket_id: str | None
    ticket_title: str | None
    ticket_type: str | None
    started_at: str | None
    completed_at: str | None
    updated_at: str
    code: CodeFacts
    invocations: tuple[InvocationFact, ...]
    tools: tuple[ToolFact, ...]


@dataclass(frozen=True, slots=True)
class GroupedMetric:
    invocations: int
    input_tokens: int
    output_tokens: int
    cache_read_tokens: int
    cache_hit_rate_pct: float
    cost_usd: float
    peak_context_tokens: int = 0
    context_health: str = "HEALTHY"


@dataclass(frozen=True, slots=True)
class TestToCodeRatio:
    ratio: str
    test_lines_added: int
    prod_lines_added: int


@dataclass(frozen=True, slots=True)
class SubagentTurnVelocity:
    avg_turns_per_invocation: float
    peak_turns: int
    by_role: dict[str, float]


@dataclass(frozen=True, slots=True)
class CodeDeliveredStats:
    lines_added: int
    lines_deleted: int
    files_changed: int


@dataclass(frozen=True, slots=True)
class ToolExecutionHealth:
    total_calls: int
    errors_and_corrections: int
    success_rate_pct: float


@dataclass(frozen=True, slots=True)
class ContextWindowHealth:
    healthy_ceiling_tokens: int
    status: str
    orchestrator_peak_tokens: int
    orchestrator_saturation_pct: float
    subagents_peak_tokens: int
    subagents_peak_role: str
    compactions: int
    peak_by_role: dict[str, int]


@dataclass(frozen=True, slots=True)
class WorkflowEfficiency:
    sdlc_iterations: int
    active_duration_seconds: int
    active_duration_formatted: str
    role_effort_distribution_pct: dict[str, float]
    subagent_turn_velocity: SubagentTurnVelocity
    test_to_production_ratio: TestToCodeRatio
    reasoning_tokens: int
    reasoning_ratio_pct: float
    code_delivered: CodeDeliveredStats
    tool_execution: ToolExecutionHealth
    context_window_health: ContextWindowHealth


@dataclass(frozen=True, slots=True)
class ModelMetric:
    model: str
    invocations: int
    input_tokens: int
    output_tokens: int
    cache_read_tokens: int
    cache_hit_rate_pct: float
    cost_usd: float
    peak_context_tokens: int = 0
    context_health: str = "HEALTHY"


type AggregatedModelMetric = ModelMetric


@dataclass(frozen=True, slots=True)
class ThreadRoleMetric:
    role: str
    models: list[ModelMetric]


@dataclass(frozen=True, slots=True)
class RoleMetric:
    role: str
    invocations: int
    input_tokens: int
    output_tokens: int
    cache_read_tokens: int
    cache_hit_rate_pct: float
    cost_usd: float
    peak_context_tokens: int = 0
    context_health: str = "HEALTHY"


@dataclass(frozen=True, slots=True)
class ThreadExecutionRecord:
    session_id: str
    user: str
    roles: list[ThreadRoleMetric]
    compactions: int = 0


@dataclass(frozen=True, slots=True)
class AgenticMetricsLedger:
    ticket_id: str | None
    ticket_title: str | None
    ticket_type: str | None
    started_at: str | None
    completed_at: str | None
    updated_at: str
    facts: AgenticFactLedger
    workflow_efficiency: WorkflowEfficiency
    threads: list[ThreadExecutionRecord]
    totals: GroupedMetric
    roles: list[RoleMetric]
    models: list[ModelMetric]


@dataclass(frozen=True, slots=True)
class TimeBucketMetric:
    bucket_key: str
    bucket_date: str
    invocations: int
    cost_usd: float
    input_tokens: int
    output_tokens: int
    cache_read_tokens: int
    reasoning_tokens: int
    cache_hit_rate_pct: float
    lines_added: int
    lines_deleted: int
    files_changed: int
    net_lines_added: int
    peak_context_tokens: int
    active_duration_seconds: int
    cumulative_cost_usd: float = 0.0
    cumulative_input_tokens: int = 0
    cumulative_output_tokens: int = 0
    cumulative_lines_added: int = 0
    cumulative_lines_deleted: int = 0
    cumulative_invocations: int = 0


@dataclass(frozen=True, slots=True)
class FleetSummary:
    total_tickets: int
    total_cost_usd: float
    total_invocations: int
    total_input_tokens: int
    total_output_tokens: int
    total_cache_read_tokens: int
    overall_cache_hit_rate_pct: float
    total_duration_seconds: int
    formatted_total_duration: str
    total_lines_added: int
    total_lines_deleted: int
    total_files_changed: int
    total_reasoning_tokens: int
    overall_reasoning_ratio_pct: float
    role_costs: dict[str, float]
    model_costs: dict[str, float]


@dataclass(frozen=True, slots=True)
class TicketComparisonRow:
    ticket_id: str
    updated_at: str
    sdlc_iterations: int
    active_duration_seconds: int
    active_duration_formatted: str
    cost_usd: float
    cache_hit_rate_pct: float
    test_to_code_ratio: str
    lines_added: int
    lines_deleted: int
    files_changed: int
    tool_success_rate_pct: float
    context_health_status: str
    orchestrator_saturation_pct: float
    peak_context_tokens: int
    ticket_title: str = "N/A"
    started_at: str | None = None
    completed_at: str | None = None
    lead_duration_seconds: int | None = None
    lead_duration_formatted: str = "N/A"


@dataclass(frozen=True, slots=True)
class ScanReport:
    ledgers: tuple[AgenticMetricsLedger, ...]
    errors: tuple[str, ...]


@dataclass(frozen=True, slots=True)
class TicketArtifacts:
    ticket_id: str
    contract_content: str | None = None
    summary_content: str | None = None
