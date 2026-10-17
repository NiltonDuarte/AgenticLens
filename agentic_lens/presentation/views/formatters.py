from __future__ import annotations

from datetime import UTC, datetime
from pathlib import Path

from agentic_lens.domain.constants import (
    HEALTHY_CEILING_TOKENS,
    format_tokens_k,
    get_health_label,
)
from agentic_lens.domain.models import AgenticMetricsLedger


def format_iso_datetime(ts: str | None) -> str:
    """Format ISO 8601 timestamp string into standard YYYY-MM-DD HH:MM UTC format."""
    if not ts or not ts.strip():
        return "N/A"
    clean_ts = ts.strip()
    try:
        dt = datetime.fromisoformat(clean_ts)
        dt_utc = dt if dt.tzinfo is not None else dt.replace(tzinfo=UTC)
        return dt_utc.astimezone(UTC).strftime("%Y-%m-%d %H:%M UTC")
    except (ValueError, TypeError):
        return clean_ts


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


def format_currency(amount: float) -> str:
    """Format USD currency amount."""
    return f"${amount:,.4f}" if amount < 10.0 else f"${amount:,.2f}"


def format_percentage(pct: float) -> str:
    """Format percentage value."""
    return f"{pct:.1f}%"


def format_tokens(tokens: int) -> str:
    """Format token count with abbreviation."""
    return format_tokens_k(tokens)


def format_duration(seconds: int) -> str:
    """Format duration in seconds into human-readable hours and minutes."""
    hours, rem = divmod(seconds, 3600)
    mins, secs = divmod(rem, 60)
    if hours > 0:
        return f"{hours}h {mins}m"
    return f"{mins}m {secs}s"


def format_number(val: int) -> str:
    """Format integer with thousands separator."""
    return f"{val:,}"


def format_compact_number(val: int) -> str:
    """Format integer into compact SI notation (e.g. 11.9k, 1.5M)."""
    abs_val = abs(val)
    if abs_val >= 1_000_000:
        num = abs_val / 1_000_000
        formatted = f"{int(num)}M" if num == int(num) else f"{num:.1f}M"
    elif abs_val >= 1_000:
        num = abs_val / 1_000
        formatted = f"{int(num)}k" if num == int(num) else f"{num:.1f}k"
    else:
        formatted = str(abs_val)

    return f"-{formatted}" if val < 0 else formatted


def format_code_delivery(lines_added: int, lines_deleted: int) -> str:
    """Format code delivered lines added and deleted with compact notation."""
    return f"+{format_compact_number(lines_added)} / -{format_compact_number(lines_deleted)}"


def format_ticket_label(ticket_id: str | None, ticket_title: str | None) -> str:
    """Format combined ticket identifier and human-readable title."""
    id_clean = ticket_id.strip() if ticket_id else ""
    title_clean = ticket_title.strip() if ticket_title else ""
    if id_clean and title_clean:
        return f"{id_clean} - {title_clean}"
    if id_clean:
        return id_clean
    if title_clean:
        return title_clean
    return "Unknown"


def format_markdown_metrics_table(ledger: AgenticMetricsLedger) -> str:
    """Reconstruct standard review markdown table directly from domain ledger."""
    we = ledger.workflow_efficiency
    cwh = we.context_window_health
    totals = ledger.totals

    contributors = sorted(
        {i.user for i in ledger.facts.invocations if i.user and i.user != "unknown"}
    )

    role_dist_parts = [
        f"{role.title()} {pct:.1f}%"
        for role, pct in we.role_effort_distribution_pct.items()
        if role != "orchestrator" and pct > 0
    ]
    role_dist_str = ", ".join(role_dist_parts) if role_dist_parts else "N/A"

    vel_parts = [
        f"{role.title()} {turns:.1f}"
        for role, turns in we.subagent_turn_velocity.by_role.items()
        if turns > 0
    ]
    vel_str = f" | {', '.join(vel_parts)}" if vel_parts else ""

    orch_tok_str = format_tokens_k(cwh.orchestrator_peak_tokens)
    sub_tok_str = format_tokens_k(cwh.subagents_peak_tokens)
    orch_pct_str = f"{round(cwh.orchestrator_saturation_pct)}%"

    lines = [
        "### Workflow Efficiency",
        f"- **Software Development Life Cycle Iterations:** {we.sdlc_iterations}",
        f"- **Active Agent Duration:** {we.active_duration_formatted} (excludes user wait time)",
        (
            f"- **Context Window Health:** {cwh.status} — Orchestrator peak {orch_tok_str} tokens "
            f"({orch_pct_str} of {format_tokens_k(HEALTHY_CEILING_TOKENS)} healthy ceiling, "
            f"{cwh.compactions} compactions) | Subagents peak {sub_tok_str} ({cwh.subagents_peak_role.title()})"
        ),
        f"- **Role Effort Distribution (Subagents):** {role_dist_str}",
        (
            f"- **Subagent Turn Velocity & Thrashing:** {we.subagent_turn_velocity.avg_turns_per_invocation:.1f} "
            f"avg turns/run (peak: {we.subagent_turn_velocity.peak_turns} turns){vel_str}"
        ),
        (
            f"- **Test-to-Production Code Ratio:** {we.test_to_production_ratio.ratio} "
            f"(+{we.test_to_production_ratio.test_lines_added:,} test lines / "
            f"+{we.test_to_production_ratio.prod_lines_added:,} prod lines)"
        ),
        (
            f"- **Code Delivered:** +{we.code_delivered.lines_added:,} / "
            f"-{we.code_delivered.lines_deleted:,} lines across "
            f"{we.code_delivered.files_changed} code files (excluding docs)"
        ),
        (
            f"- **Reasoning Effort:** {we.reasoning_tokens:,} reasoning tokens "
            f"({we.reasoning_ratio_pct:.1f}% of generated output)"
        ),
        (
            f"- **Tool Health & Self-Correction:** {we.tool_execution.total_calls:,} tool calls "
            f"({we.tool_execution.errors_and_corrections} self-corrections, "
            f"{we.tool_execution.success_rate_pct:.1f}% success rate)"
        ),
        "",
        "### Agentic Execution Metrics",
        "",
    ]

    if contributors:
        lines.extend([f"**Contributors:** {', '.join(contributors)}", ""])

    lines.extend(
        [
            "#### By Subagent Role",
            "| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |",
            "|---|---|---|---|---|---|---|---|---|",
        ]
    )

    for r in ledger.roles:
        lines.append(
            f"| **{r.role}** | {r.invocations} | {r.input_tokens:,} | "
            f"{r.output_tokens:,} | {r.cache_read_tokens:,} | "
            f"{r.cache_hit_rate_pct:.1f}% | {format_tokens_k(r.peak_context_tokens)} | "
            f"{get_health_label(r.peak_context_tokens)} | ${r.cost_usd:.2f} |"
        )

    lines.append(
        f"| **Total** | **{totals.invocations}** | **{totals.input_tokens:,}** | "
        f"**{totals.output_tokens:,}** | **{totals.cache_read_tokens:,}** | "
        f"**{totals.cache_hit_rate_pct:.1f}%** | **{format_tokens_k(totals.peak_context_tokens)}** | "
        f"**{totals.context_health}** | **${totals.cost_usd:.2f}** |"
    )

    lines.extend(
        [
            "",
            "#### By Model",
            "| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |",
            "|---|---|---|---|---|---|---|",
        ]
    )

    for m in sorted(ledger.models, key=lambda x: x.cost_usd, reverse=True):
        lines.append(
            f"| **{m.model}** | {m.invocations} | {m.input_tokens:,} | "
            f"{m.output_tokens:,} | {m.cache_read_tokens:,} | "
            f"{m.cache_hit_rate_pct:.1f}% | ${m.cost_usd:.2f} |"
        )

    lines.append(
        f"| **Total** | **{totals.invocations}** | **{totals.input_tokens:,}** | "
        f"**{totals.output_tokens:,}** | **{totals.cache_read_tokens:,}** | "
        f"**{totals.cache_hit_rate_pct:.1f}%** | **${totals.cost_usd:.2f}** |"
    )
    lines.append("")

    return "\n".join(lines)


def update_summary_file(dest: Path, table_md: str) -> None:
    """Idempotently update or append the metrics table inside the summary file."""
    dest.parent.mkdir(parents=True, exist_ok=True)
    header = "### Workflow Efficiency"
    legacy_header = "### Agentic Execution Metrics"

    if dest.exists():
        content = dest.read_text(encoding="utf-8")
        target_idx = -1
        if header in content:
            target_idx = content.find(header)
        elif legacy_header in content:
            target_idx = content.find(legacy_header)

        if target_idx != -1:
            next_heading = content.find("\n## ", target_idx + len(header))
            if next_heading != -1:
                new_content = (
                    content[:target_idx].rstrip()
                    + "\n\n"
                    + table_md.strip()
                    + "\n\n"
                    + content[next_heading:].lstrip()
                )
            else:
                new_content = (
                    content[:target_idx].rstrip() + "\n\n" + table_md.strip() + "\n"
                )
            dest.write_text(new_content, encoding="utf-8")
            return

    with open(dest, "a", encoding="utf-8") as f:
        f.write("\n" + table_md.strip() + "\n")
