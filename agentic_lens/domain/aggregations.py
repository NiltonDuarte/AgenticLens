from __future__ import annotations

from collections import defaultdict
from collections.abc import Sequence
from datetime import UTC, date, datetime, timedelta

from agentic_lens.domain.constants import (
    HEALTHY_CEILING_TOKENS,
    get_base_health_status,
    get_health_label,
)
from agentic_lens.domain.enums import TimeGranularity, TimeWindowPreset
from agentic_lens.domain.models import (
    AgenticFactLedger,
    AgenticMetricsLedger,
    CodeDeliveredStats,
    ContextWindowHealth,
    FleetSummary,
    GroupedMetric,
    InvocationFact,
    ModelMetric,
    RoleMetric,
    SubagentTurnVelocity,
    TestToCodeRatio,
    ThreadExecutionRecord,
    ThreadRoleMetric,
    TicketComparisonRow,
    TimeBucketMetric,
    ToolExecutionHealth,
    WorkflowEfficiency,
)

CANONICAL_ROLE_ORDER: tuple[str, ...] = (
    "orchestrator",
    "planner",
    "coder",
    "tester",
    "reviewer",
)
SUBAGENT_ROLE_ORDER: tuple[str, ...] = (
    "planner",
    "coder",
    "tester",
    "reviewer",
)


def compute_totals(invocations: Sequence[InvocationFact]) -> GroupedMetric:
    """Compute aggregate invocation metrics from atomic invocation facts."""
    if not invocations:
        return GroupedMetric(
            invocations=0,
            input_tokens=0,
            output_tokens=0,
            cache_read_tokens=0,
            cache_hit_rate_pct=0.0,
            cost_usd=0.0,
            peak_context_tokens=0,
            context_health="HEALTHY",
        )

    by_role_sessions: dict[str, set[str]] = defaultdict(set)
    for inv in invocations:
        by_role_sessions[inv.role].add(inv.id)

    total_invocations = sum(len(sids) for sids in by_role_sessions.values())
    total_input = sum(inv.tokens.input for inv in invocations)
    total_output = sum(inv.tokens.output for inv in invocations)
    total_cache = sum(inv.tokens.cache_read for inv in invocations)
    total_cost = sum(inv.cost_usd for inv in invocations)
    total_peak = max((inv.tokens.peak_context for inv in invocations), default=0)

    total_cached_or_input = total_input + total_cache
    cache_hit_rate_pct = (
        round(total_cache / total_cached_or_input * 100, 1)
        if total_cached_or_input > 0
        else 0.0
    )
    health_status = get_base_health_status(total_peak)

    return GroupedMetric(
        invocations=total_invocations,
        input_tokens=total_input,
        output_tokens=total_output,
        cache_read_tokens=total_cache,
        cache_hit_rate_pct=cache_hit_rate_pct,
        cost_usd=round(total_cost, 4),
        peak_context_tokens=total_peak,
        context_health=health_status,
    )


def compute_workflow_efficiency(ledger: AgenticFactLedger) -> WorkflowEfficiency:
    """Derive workflow efficiency profile from atomic ledger facts."""
    invs = ledger.invocations
    tools = ledger.tools
    code = ledger.code

    active_duration_seconds = sum(inv.active_duration_seconds for inv in invs)
    dur_hours, dur_rem = divmod(active_duration_seconds, 3600)
    dur_mins, dur_rem_secs = divmod(dur_rem, 60)
    active_duration_formatted = (
        f"{dur_hours}h {dur_mins}m" if dur_hours > 0 else f"{dur_mins}m {dur_rem_secs}s"
    )

    role_costs: dict[str, float] = defaultdict(float)
    role_peaks: dict[str, int] = defaultdict(int)
    role_turns: dict[str, list[int]] = defaultdict(list)
    for inv in invs:
        role_costs[inv.role] += inv.cost_usd
        role_peaks[inv.role] = max(role_peaks[inv.role], inv.tokens.peak_context)
        if inv.parent_id is not None:
            role_turns[inv.role].append(inv.turns)

    sub_costs = {r: role_costs[r] for r in SUBAGENT_ROLE_ORDER if r in role_costs}
    tot_sub_cost = sum(sub_costs.values())
    role_effort_distribution_pct: dict[str, float] = {}
    for r in SUBAGENT_ROLE_ORDER:
        if r in sub_costs and tot_sub_cost > 0:
            role_effort_distribution_pct[r] = round(
                sub_costs[r] / tot_sub_cost * 100, 1
            )

    all_sub_turns = [
        t for r in SUBAGENT_ROLE_ORDER if r in role_turns for t in role_turns[r]
    ]
    avg_sub_turns = (
        round(sum(all_sub_turns) / len(all_sub_turns), 1) if all_sub_turns else 0.0
    )
    peak_sub_turns = max(all_sub_turns, default=0)
    vel_by_role: dict[str, float] = {}
    for r in SUBAGENT_ROLE_ORDER:
        if role_turns.get(r):
            vel_by_role[r] = round(sum(role_turns[r]) / len(role_turns[r]), 1)
    subagent_turn_velocity = SubagentTurnVelocity(
        avg_turns_per_invocation=avg_sub_turns,
        peak_turns=peak_sub_turns,
        by_role=vel_by_role,
    )

    test_ratio_str = (
        f"{code.test_lines_added / code.prod_lines_added:.1f}:1"
        if code.prod_lines_added > 0
        else f"{code.test_lines_added}:0"
    )
    test_to_production_ratio = TestToCodeRatio(
        ratio=test_ratio_str,
        test_lines_added=code.test_lines_added,
        prod_lines_added=code.prod_lines_added,
    )

    tot_reasoning = sum(inv.tokens.reasoning for inv in invs)
    tot_output = sum(inv.tokens.output for inv in invs)
    gen_total = tot_output + tot_reasoning
    reasoning_ratio_pct = (
        round(tot_reasoning / gen_total * 100, 1) if gen_total > 0 else 0.0
    )

    code_delivered = CodeDeliveredStats(
        lines_added=code.lines_added,
        lines_deleted=code.lines_deleted,
        files_changed=code.files_changed,
    )

    tot_tool_calls = sum(t.calls for t in tools)
    tot_tool_errors = sum(t.errors for t in tools)
    tool_succ = (
        round(((tot_tool_calls - tot_tool_errors) / tot_tool_calls * 100), 1)
        if tot_tool_calls > 0
        else 100.0
    )
    tool_execution = ToolExecutionHealth(
        total_calls=tot_tool_calls,
        errors_and_corrections=tot_tool_errors,
        success_rate_pct=tool_succ,
    )

    orch_peak = role_peaks.get("orchestrator", 0)
    orch_sat_pct = (
        round(orch_peak / HEALTHY_CEILING_TOKENS * 100, 1)
        if HEALTHY_CEILING_TOKENS > 0
        else 0.0
    )
    tot_compactions = sum(inv.compactions for inv in invs)
    sub_roles_present = [r for r in SUBAGENT_ROLE_ORDER if r in role_peaks]
    if sub_roles_present:
        sub_max_r = max(sub_roles_present, key=lambda r: role_peaks[r])
        sub_peak = role_peaks[sub_max_r]
        sub_role_str = sub_max_r
    else:
        sub_peak = 0
        sub_role_str = "none"

    overall_peak = max(orch_peak, sub_peak)
    cwh_status = get_base_health_status(overall_peak)
    context_window_health = ContextWindowHealth(
        healthy_ceiling_tokens=HEALTHY_CEILING_TOKENS,
        status=cwh_status,
        orchestrator_peak_tokens=orch_peak,
        orchestrator_saturation_pct=orch_sat_pct,
        subagents_peak_tokens=sub_peak,
        subagents_peak_role=sub_role_str,
        compactions=tot_compactions,
        peak_by_role=dict(role_peaks),
    )

    return WorkflowEfficiency(
        sdlc_iterations=code.sdlc_iterations,
        active_duration_seconds=active_duration_seconds,
        active_duration_formatted=active_duration_formatted,
        role_effort_distribution_pct=role_effort_distribution_pct,
        subagent_turn_velocity=subagent_turn_velocity,
        test_to_production_ratio=test_to_production_ratio,
        reasoning_tokens=tot_reasoning,
        reasoning_ratio_pct=reasoning_ratio_pct,
        code_delivered=code_delivered,
        tool_execution=tool_execution,
        context_window_health=context_window_health,
    )


def compute_role_metrics(
    invocations: Sequence[InvocationFact],
) -> list[RoleMetric]:
    """Compute aggregated metrics grouped by agent role."""
    by_role: dict[str, list[InvocationFact]] = defaultdict(list)
    for inv in invocations:
        by_role[inv.role].append(inv)

    ordered_roles = [r for r in CANONICAL_ROLE_ORDER if r in by_role] + sorted(
        r for r in by_role if r not in CANONICAL_ROLE_ORDER
    )

    metrics: list[RoleMetric] = []
    for role in ordered_roles:
        r_invs = by_role[role]
        sids = {inv.id for inv in r_invs}
        inp = sum(inv.tokens.input for inv in r_invs)
        out = sum(inv.tokens.output for inv in r_invs)
        cache = sum(inv.tokens.cache_read for inv in r_invs)
        cost = sum(inv.cost_usd for inv in r_invs)
        peak = max((inv.tokens.peak_context for inv in r_invs), default=0)

        tot_tok = inp + cache
        hit_pct = round(cache / tot_tok * 100, 1) if tot_tok > 0 else 0.0
        health = get_health_label(peak)

        metrics.append(
            RoleMetric(
                role=role,
                invocations=len(sids),
                input_tokens=inp,
                output_tokens=out,
                cache_read_tokens=cache,
                cache_hit_rate_pct=hit_pct,
                cost_usd=round(cost, 4),
                peak_context_tokens=peak,
                context_health=health,
            )
        )
    return metrics


def compute_model_metrics(
    invocations: Sequence[InvocationFact],
) -> list[ModelMetric]:
    """Compute aggregated metrics grouped by model, sorted by cost descending."""
    by_model: dict[str, list[InvocationFact]] = defaultdict(list)
    for inv in invocations:
        by_model[inv.model].append(inv)

    metrics: list[ModelMetric] = []
    for model, m_invs in by_model.items():
        sids = {inv.id for inv in m_invs}
        inp = sum(inv.tokens.input for inv in m_invs)
        out = sum(inv.tokens.output for inv in m_invs)
        cache = sum(inv.tokens.cache_read for inv in m_invs)
        cost = sum(inv.cost_usd for inv in m_invs)
        peak = max((inv.tokens.peak_context for inv in m_invs), default=0)

        tot_tok = inp + cache
        hit_pct = round(cache / tot_tok * 100, 1) if tot_tok > 0 else 0.0
        health = get_health_label(peak)

        metrics.append(
            ModelMetric(
                model=model,
                invocations=len(sids),
                input_tokens=inp,
                output_tokens=out,
                cache_read_tokens=cache,
                cache_hit_rate_pct=hit_pct,
                cost_usd=round(cost, 4),
                peak_context_tokens=peak,
                context_health=health,
            )
        )
    return sorted(metrics, key=lambda m: m.cost_usd, reverse=True)


def compute_thread_summaries(
    invocations: Sequence[InvocationFact],
) -> list[ThreadExecutionRecord]:
    """Group invocations by root session into hierarchical thread execution records."""
    if not invocations:
        return []

    id_to_parent: dict[str, str | None] = {inv.id: inv.parent_id for inv in invocations}
    id_to_user: dict[str, str] = {inv.id: inv.user for inv in invocations}

    def resolve_root(inv: InvocationFact) -> str:
        curr = inv.id
        visited: set[str] = set()
        while (
            curr in id_to_parent
            and id_to_parent[curr] is not None
            and curr not in visited
        ):
            visited.add(curr)
            parent = id_to_parent[curr]
            if parent is None:
                break
            curr = parent
        if curr not in visited and curr not in id_to_parent and inv.parent_id:
            return inv.parent_id
        return curr

    thread_groups: dict[str, list[InvocationFact]] = defaultdict(list)
    for inv in invocations:
        root_id = resolve_root(inv)
        thread_groups[root_id].append(inv)

    records: list[ThreadExecutionRecord] = []
    for root_id, thread_invs in thread_groups.items():
        thread_user = id_to_user.get(
            root_id, thread_invs[0].user if thread_invs else "unknown"
        )
        compactions = sum(inv.compactions for inv in thread_invs)

        by_role: dict[str, list[InvocationFact]] = defaultdict(list)
        for inv in thread_invs:
            by_role[inv.role].append(inv)

        ordered_roles = [r for r in CANONICAL_ROLE_ORDER if r in by_role] + sorted(
            r for r in by_role if r not in CANONICAL_ROLE_ORDER
        )

        role_metrics: list[ThreadRoleMetric] = []
        for role in ordered_roles:
            r_invs = by_role[role]
            by_model: dict[str, list[InvocationFact]] = defaultdict(list)
            for inv in r_invs:
                by_model[inv.model].append(inv)

            model_metrics: list[ModelMetric] = []
            for model, m_invs in by_model.items():
                sids = {inv.id for inv in m_invs}
                inp = sum(inv.tokens.input for inv in m_invs)
                out = sum(inv.tokens.output for inv in m_invs)
                cache = sum(inv.tokens.cache_read for inv in m_invs)
                cost = sum(inv.cost_usd for inv in m_invs)
                peak = max((inv.tokens.peak_context for inv in m_invs), default=0)

                tot_tok = inp + cache
                hit_pct = round(cache / tot_tok * 100, 1) if tot_tok > 0 else 0.0
                health = get_health_label(peak)

                model_metrics.append(
                    ModelMetric(
                        model=model,
                        invocations=len(sids),
                        input_tokens=inp,
                        output_tokens=out,
                        cache_read_tokens=cache,
                        cache_hit_rate_pct=hit_pct,
                        cost_usd=round(cost, 4),
                        peak_context_tokens=peak,
                        context_health=health,
                    )
                )
            model_metrics.sort(key=lambda m: m.cost_usd, reverse=True)
            role_metrics.append(ThreadRoleMetric(role=role, models=model_metrics))

        records.append(
            ThreadExecutionRecord(
                session_id=root_id,
                user=thread_user,
                roles=role_metrics,
                compactions=compactions,
            )
        )

    return sorted(records, key=lambda r: r.session_id)


def derive_metrics_ledger(fact_ledger: AgenticFactLedger) -> AgenticMetricsLedger:
    """Derive full metrics ledger projections from an atomic fact ledger."""
    totals = compute_totals(fact_ledger.invocations)
    eff = compute_workflow_efficiency(fact_ledger)
    roles = compute_role_metrics(fact_ledger.invocations)
    models = compute_model_metrics(fact_ledger.invocations)
    threads = compute_thread_summaries(fact_ledger.invocations)

    return AgenticMetricsLedger(
        ticket_id=fact_ledger.ticket_id,
        ticket_title=fact_ledger.ticket_title,
        ticket_type=fact_ledger.ticket_type,
        started_at=fact_ledger.started_at,
        completed_at=fact_ledger.completed_at,
        updated_at=fact_ledger.updated_at,
        facts=fact_ledger,
        workflow_efficiency=eff,
        threads=threads,
        totals=totals,
        roles=roles,
        models=models,
    )


def calculate_fleet_summary(
    ledgers: Sequence[AgenticMetricsLedger],
) -> FleetSummary:
    """Calculate fleet-wide aggregated metrics across all scanned ledgers."""
    total_tickets = len(ledgers)
    if total_tickets == 0:
        return FleetSummary(
            total_tickets=0,
            total_cost_usd=0.0,
            total_invocations=0,
            total_input_tokens=0,
            total_output_tokens=0,
            total_cache_read_tokens=0,
            overall_cache_hit_rate_pct=0.0,
            total_duration_seconds=0,
            formatted_total_duration="0m 0s",
            total_lines_added=0,
            total_lines_deleted=0,
            total_files_changed=0,
            total_reasoning_tokens=0,
            overall_reasoning_ratio_pct=0.0,
            role_costs={},
            model_costs={},
        )

    total_cost_usd = sum(ledger.totals.cost_usd for ledger in ledgers)
    total_invocations = sum(ledger.totals.invocations for ledger in ledgers)
    total_input_tokens = sum(ledger.totals.input_tokens for ledger in ledgers)
    total_output_tokens = sum(ledger.totals.output_tokens for ledger in ledgers)
    total_cache_read_tokens = sum(ledger.totals.cache_read_tokens for ledger in ledgers)

    total_cached_or_input = total_input_tokens + total_cache_read_tokens
    overall_cache_hit_rate_pct = (
        round(total_cache_read_tokens / total_cached_or_input * 100, 1)
        if total_cached_or_input > 0
        else 0.0
    )

    total_duration_seconds = sum(
        ledger.workflow_efficiency.active_duration_seconds for ledger in ledgers
    )
    dur_hours, dur_rem = divmod(total_duration_seconds, 3600)
    dur_mins, dur_rem_secs = divmod(dur_rem, 60)
    formatted_total_duration = (
        f"{dur_hours}h {dur_mins}m" if dur_hours > 0 else f"{dur_mins}m {dur_rem_secs}s"
    )

    total_lines_added = sum(
        ledger.workflow_efficiency.code_delivered.lines_added for ledger in ledgers
    )
    total_lines_deleted = sum(
        ledger.workflow_efficiency.code_delivered.lines_deleted for ledger in ledgers
    )
    total_files_changed = sum(
        ledger.workflow_efficiency.code_delivered.files_changed for ledger in ledgers
    )
    total_reasoning_tokens = sum(
        ledger.workflow_efficiency.reasoning_tokens for ledger in ledgers
    )

    gen_total = total_output_tokens + total_reasoning_tokens
    overall_reasoning_ratio_pct = (
        round(total_reasoning_tokens / gen_total * 100, 1) if gen_total > 0 else 0.0
    )

    role_costs: dict[str, float] = {}
    for ledger in ledgers:
        for r in ledger.roles:
            role_costs[r.role] = round(role_costs.get(r.role, 0.0) + r.cost_usd, 4)

    model_costs: dict[str, float] = {}
    for ledger in ledgers:
        for m in ledger.models:
            model_costs[m.model] = round(model_costs.get(m.model, 0.0) + m.cost_usd, 4)

    return FleetSummary(
        total_tickets=total_tickets,
        total_cost_usd=round(total_cost_usd, 4),
        total_invocations=total_invocations,
        total_input_tokens=total_input_tokens,
        total_output_tokens=total_output_tokens,
        total_cache_read_tokens=total_cache_read_tokens,
        overall_cache_hit_rate_pct=overall_cache_hit_rate_pct,
        total_duration_seconds=total_duration_seconds,
        formatted_total_duration=formatted_total_duration,
        total_lines_added=total_lines_added,
        total_lines_deleted=total_lines_deleted,
        total_files_changed=total_files_changed,
        total_reasoning_tokens=total_reasoning_tokens,
        overall_reasoning_ratio_pct=overall_reasoning_ratio_pct,
        role_costs=role_costs,
        model_costs=model_costs,
    )


def compute_lead_time_seconds(
    started_at: str | None, completed_at: str | None
) -> int | None:
    """Calculate elapsed wall-clock turnaround duration in seconds from start to completion."""
    if not started_at or not completed_at:
        return None
    try:
        start_dt = datetime.fromisoformat(started_at)
        completed_dt = datetime.fromisoformat(completed_at)
        return int((completed_dt - start_dt).total_seconds())
    except (ValueError, TypeError):
        return None


def format_lead_time(seconds: int | None) -> str:
    """Format elapsed lead time in seconds into human-readable duration."""
    if seconds is None:
        return "In Progress"
    if seconds < 0:
        return "0s"
    if seconds >= 86400:
        days, rem = divmod(seconds, 86400)
        hours = rem // 3600
        return f"{days}d {hours}h"
    if seconds >= 3600:
        hours, rem = divmod(seconds, 3600)
        mins = rem // 60
        return f"{hours}h {mins}m"
    mins, rem_secs = divmod(seconds, 60)
    return f"{mins}m {rem_secs}s"


def build_comparison_rows(
    ledgers: Sequence[AgenticMetricsLedger],
) -> list[TicketComparisonRow]:
    """Build normalized comparison rows for cross-ticket matrix visualization."""
    rows: list[TicketComparisonRow] = []
    for ledger in ledgers:
        tid = ledger.ticket_id or "unknown"
        eff = ledger.workflow_efficiency
        tot = ledger.totals
        cwh = eff.context_window_health
        lead_secs = compute_lead_time_seconds(ledger.started_at, ledger.completed_at)
        rows.append(
            TicketComparisonRow(
                ticket_id=tid,
                updated_at=ledger.updated_at,
                sdlc_iterations=eff.sdlc_iterations,
                active_duration_seconds=eff.active_duration_seconds,
                active_duration_formatted=eff.active_duration_formatted,
                cost_usd=round(tot.cost_usd, 4),
                cache_hit_rate_pct=tot.cache_hit_rate_pct,
                test_to_code_ratio=eff.test_to_production_ratio.ratio,
                lines_added=eff.code_delivered.lines_added,
                lines_deleted=eff.code_delivered.lines_deleted,
                files_changed=eff.code_delivered.files_changed,
                tool_success_rate_pct=eff.tool_execution.success_rate_pct,
                context_health_status=cwh.status,
                orchestrator_saturation_pct=cwh.orchestrator_saturation_pct,
                peak_context_tokens=cwh.orchestrator_peak_tokens,
                ticket_title=ledger.ticket_title or "N/A",
                started_at=ledger.started_at,
                completed_at=ledger.completed_at,
                lead_duration_seconds=lead_secs,
                lead_duration_formatted=format_lead_time(lead_secs),
            )
        )
    return sorted(
        rows,
        key=lambda r: (
            r.completed_at is not None,
            r.completed_at or "",
            r.started_at or "",
            r.ticket_id,
        ),
        reverse=True,
    )


def calculate_role_effort_rollup(
    ledgers: Sequence[AgenticMetricsLedger],
) -> dict[str, dict[str, float | int]]:
    """Group invocations, tokens, and cost across all tickets by agent role."""
    rollup: dict[str, dict[str, float | int]] = {}
    for ledger in ledgers:
        for r in ledger.roles:
            if r.role not in rollup:
                rollup[r.role] = {
                    "invocations": 0,
                    "input_tokens": 0,
                    "output_tokens": 0,
                    "cache_read_tokens": 0,
                    "cost_usd": 0.0,
                }
            curr = rollup[r.role]
            curr["invocations"] = int(curr["invocations"]) + r.invocations
            curr["input_tokens"] = int(curr["input_tokens"]) + r.input_tokens
            curr["output_tokens"] = int(curr["output_tokens"]) + r.output_tokens
            curr["cache_read_tokens"] = (
                int(curr["cache_read_tokens"]) + r.cache_read_tokens
            )
            curr["cost_usd"] = round(float(curr["cost_usd"]) + r.cost_usd, 4)
    return rollup


def resolve_date_preset(
    preset: str,
    reference_date: date | None = None,
) -> tuple[date | None, date | None]:
    """Resolve a time window preset string into start and end dates."""
    ref = reference_date if reference_date is not None else datetime.now(UTC).date()
    if preset in (TimeWindowPreset.LAST_7_DAYS.value, "Last 7 Days"):
        return (ref - timedelta(days=7), ref)
    if preset in (TimeWindowPreset.LAST_30_DAYS.value, "Last 30 Days"):
        return (ref - timedelta(days=30), ref)
    if preset in (TimeWindowPreset.LAST_90_DAYS.value, "Last 90 Days"):
        return (ref - timedelta(days=90), ref)
    return (None, None)


def _parse_iso_date(ts: str | None) -> date | None:
    """Defensively parse an ISO 8601 timestamp string into a date object."""
    if not ts or len(ts) < 10:
        return None
    try:
        return date.fromisoformat(ts[:10])
    except (ValueError, TypeError):
        return None


def get_ticket_effective_date(
    ledger: AgenticMetricsLedger | AgenticFactLedger,
) -> date | None:
    """Resolve the authoritative effective date for a ticket (completed_at -> started_at -> updated_at)."""
    return (
        _parse_iso_date(ledger.completed_at)
        or _parse_iso_date(ledger.started_at)
        or _parse_iso_date(ledger.updated_at)
    )


def filter_ledgers_by_date_range(
    ledgers: Sequence[AgenticMetricsLedger],
    start_date: date | None,
    end_date: date | None,
) -> list[AgenticMetricsLedger]:
    """Filter ledgers based on their effective date within inclusive bounds."""
    if start_date is None and end_date is None:
        return list(ledgers)

    filtered: list[AgenticMetricsLedger] = []
    for ledger in ledgers:
        ledger_date = get_ticket_effective_date(ledger)
        if ledger_date is None:
            continue
        if start_date is not None and ledger_date < start_date:
            continue
        if end_date is not None and ledger_date > end_date:
            continue
        filtered.append(ledger)
    return filtered


def aggregate_metrics_over_time(
    ledgers: Sequence[AgenticMetricsLedger],
    granularity: str = "daily",
    start_date: date | None = None,
    end_date: date | None = None,
) -> list[TimeBucketMetric]:
    """Aggregate metrics chronologically into daily, weekly, or cumulative buckets."""
    is_weekly = granularity.lower() in (TimeGranularity.WEEKLY.value, "weekly")

    class _BucketAccumulator:
        __slots__ = (
            "active_duration_seconds",
            "cache_read_tokens",
            "cost_usd",
            "files_changed",
            "input_tokens",
            "invocations",
            "lines_added",
            "lines_deleted",
            "output_tokens",
            "peak_context_tokens",
            "reasoning_tokens",
        )

        def __init__(self) -> None:
            self.invocations: int = 0
            self.cost_usd: float = 0.0
            self.input_tokens: int = 0
            self.output_tokens: int = 0
            self.cache_read_tokens: int = 0
            self.reasoning_tokens: int = 0
            self.active_duration_seconds: int = 0
            self.peak_context_tokens: int = 0
            self.lines_added: int = 0
            self.lines_deleted: int = 0
            self.files_changed: int = 0

    accumulators: dict[str, _BucketAccumulator] = defaultdict(_BucketAccumulator)

    def get_bucket_key(event_date: date) -> str:
        if is_weekly:
            monday = event_date - timedelta(days=event_date.weekday())
            return monday.isoformat()
        return event_date.isoformat()

    for ledger in ledgers:
        for inv in ledger.facts.invocations:
            inv_date = (
                _parse_iso_date(inv.completed_at)
                or _parse_iso_date(inv.started_at)
                or _parse_iso_date(ledger.completed_at)
                or _parse_iso_date(ledger.started_at)
                or _parse_iso_date(ledger.updated_at)
            )
            if inv_date is None:
                continue
            if start_date is not None and inv_date < start_date:
                continue
            if end_date is not None and inv_date > end_date:
                continue

            bkey = get_bucket_key(inv_date)
            acc = accumulators[bkey]
            acc.invocations += 1
            acc.cost_usd += inv.cost_usd
            acc.input_tokens += inv.tokens.input
            acc.output_tokens += inv.tokens.output
            acc.cache_read_tokens += inv.tokens.cache_read
            acc.reasoning_tokens += inv.tokens.reasoning
            acc.active_duration_seconds += inv.active_duration_seconds
            acc.peak_context_tokens = max(
                acc.peak_context_tokens, inv.tokens.peak_context
            )

        code_date = (
            _parse_iso_date(ledger.completed_at)
            or _parse_iso_date(ledger.started_at)
            or _parse_iso_date(ledger.updated_at)
        )
        if code_date is not None:
            in_range = True
            if start_date is not None and code_date < start_date:
                in_range = False
            if end_date is not None and code_date > end_date:
                in_range = False

            if in_range:
                bkey = get_bucket_key(code_date)
                acc = accumulators[bkey]
                code = ledger.workflow_efficiency.code_delivered
                acc.lines_added += code.lines_added
                acc.lines_deleted += code.lines_deleted
                acc.files_changed += code.files_changed

    if not accumulators:
        return []

    sorted_keys = sorted(accumulators.keys())

    running_cum_cost: float = 0.0
    running_cum_input: int = 0
    running_cum_output: int = 0
    running_cum_lines_added: int = 0
    running_cum_lines_deleted: int = 0
    running_cum_invocations: int = 0

    results: list[TimeBucketMetric] = []
    for bkey in sorted_keys:
        acc = accumulators[bkey]
        running_cum_cost += acc.cost_usd
        running_cum_input += acc.input_tokens
        running_cum_output += acc.output_tokens
        running_cum_lines_added += acc.lines_added
        running_cum_lines_deleted += acc.lines_deleted
        running_cum_invocations += acc.invocations

        cached_or_inp = acc.input_tokens + acc.cache_read_tokens
        hit_pct = (
            round(acc.cache_read_tokens / cached_or_inp * 100, 1)
            if cached_or_inp > 0
            else 0.0
        )

        results.append(
            TimeBucketMetric(
                bucket_key=bkey,
                bucket_date=bkey,
                invocations=acc.invocations,
                cost_usd=round(acc.cost_usd, 4),
                input_tokens=acc.input_tokens,
                output_tokens=acc.output_tokens,
                cache_read_tokens=acc.cache_read_tokens,
                reasoning_tokens=acc.reasoning_tokens,
                cache_hit_rate_pct=hit_pct,
                lines_added=acc.lines_added,
                lines_deleted=acc.lines_deleted,
                files_changed=acc.files_changed,
                net_lines_added=acc.lines_added - acc.lines_deleted,
                peak_context_tokens=acc.peak_context_tokens,
                active_duration_seconds=acc.active_duration_seconds,
                cumulative_cost_usd=round(running_cum_cost, 4),
                cumulative_input_tokens=running_cum_input,
                cumulative_output_tokens=running_cum_output,
                cumulative_lines_added=running_cum_lines_added,
                cumulative_lines_deleted=running_cum_lines_deleted,
                cumulative_invocations=running_cum_invocations,
            )
        )

    return results
