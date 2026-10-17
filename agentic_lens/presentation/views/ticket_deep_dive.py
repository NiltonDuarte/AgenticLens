from __future__ import annotations

from collections.abc import Sequence
from pathlib import Path

import altair as alt
import pandas as pd
import streamlit as st

from agentic_lens.adapters.repository import read_ticket_artifacts
from agentic_lens.domain.aggregations import compute_lead_time_seconds
from agentic_lens.domain.models import AgenticMetricsLedger
from agentic_lens.presentation.views.formatters import (
    format_currency,
    format_iso_datetime,
    format_lead_time,
    format_number,
    format_percentage,
    format_ticket_label,
    format_tokens,
)


def render_ticket_deep_dive(
    ledgers: Sequence[AgenticMetricsLedger],
    workspaces_dir: Path | str | None = None,
) -> None:
    """Render single-ticket detailed metrics, context health, and thread logs."""
    if not ledgers:
        st.info("No ticket ledgers available.")
        return

    ticket_map = {
        ledger.ticket_id or f"Unknown-{idx}": ledger
        for idx, ledger in enumerate(ledgers)
    }
    sorted_ticket_keys = sorted(ticket_map.keys())

    selected_ticket_id = st.selectbox(
        "Select Ticket Workspace",
        options=sorted_ticket_keys,
        index=0,
        format_func=lambda tid: format_ticket_label(tid, ticket_map[tid].ticket_title),
    )

    ledger = ticket_map[selected_ticket_id]
    eff = ledger.workflow_efficiency
    totals = ledger.totals
    cwh = eff.context_window_health
    lead_secs = compute_lead_time_seconds(ledger.started_at, ledger.completed_at)

    if ledger.ticket_title:
        st.subheader(
            f"Ticket Efficiency Profile: {selected_ticket_id} - {ledger.ticket_title}"
        )
    else:
        st.subheader(f"Ticket Efficiency Profile: {selected_ticket_id}")

    started_str = format_iso_datetime(ledger.started_at)
    completed_str = (
        format_iso_datetime(ledger.completed_at)
        if ledger.completed_at
        else "In Progress"
    )
    updated_str = format_iso_datetime(ledger.updated_at)
    st.caption(
        f"Completed: {completed_str} | Started: {started_str} | "
        f"Type: {ledger.ticket_type or 'N/A'} | Extraction Updated: {updated_str}"
    )

    with st.expander(
        "About Ticket Efficiency Deep Dive & Telemetry Reference", expanded=False
    ):
        st.markdown(
            """
### Ticket Efficiency Deep Dive Overview
The Ticket Efficiency Deep Dive provides a granular lifecycle autopsy for an individual ticket deliverable, evaluating subagent role interactions, model selections, context window saturation dynamics, turn velocity, tool reliability, and workspace artifacts.

---

### Core Metric Definitions & Component Breakdown

- **Lifecycle Milestones & Turnaround**:
  - *Completed*: Final deliverable completion timestamp (`completed_at`). Active tickets indicate `In Progress`.
  - *Started*: Ticket lifecycle kickoff timestamp (`started_at`).
  - *Lead Time*: Wall-clock calendar duration elapsed ($$\\text{completed\\_at} - \\text{started\\_at}$$).
  - *Active Duration*: Cumulative compute time spent actively processing LLM inferences across all subagents.
- **SDLC Iterations**: Number of feedback cycles executed between planner, coder, tester, and reviewer subagents.
- **Test:Prod Ratio**: Ratio of test code lines added to production code lines added ($$\\text{test\\_lines\\_added} : \\text{prod\\_lines\\_added}$$).
- **Code Delivery**: Detailed accounting of lines added (+), lines deleted (-), files changed, and test vs production breakdown.
- **Reasoning & Generation**:
  - *Reasoning Tokens*: Internal chain-of-thought tokens generated during planning and execution.
  - *Reasoning Ratio (%)*: Proportion of generated tokens dedicated to reasoning ($$\\text{reasoning} / (\\text{output} + \\text{reasoning}) \\times 100$$).
- **Context Window Health & Compactions**:
  - *Orchestrator Peak*: Peak token count retained in the orchestrator's context window.
  - *Saturation (%)*: Peak usage relative to the 200,000 token ceiling ($$\\text{peak\\_tokens} / 200,000 \\times 100$$).
  - *Subagents Peak*: Highest context peak observed among delegated subagents (`planner`, `coder`, `tester`, `reviewer`).
  - *Compactions*: Count of context memory truncation/compaction events.
- **Subagent Turn Velocity**:
  - *Avg Turns per Invocation*: Average conversation turns exchanged per subagent delegation.
  - *Peak Turns*: Maximum turns executed in a single subagent session.
- **Tool Execution Health**:
  - Total tool calls, execution errors, and success percentage ($$(\\text{calls} - \\text{errors}) / \\text{calls} \\times 100$$).

---

### Completion Date Semantics
Deliverable status is anchored to the authoritative completion timestamp (`completed_at`), establishing the exact historical milestone when all acceptance criteria and artifacts were satisfied.

---

### Operational Baselines & Thresholds
- **Context Ceiling**: Healthy context ceiling is 200,000 tokens per session.
- **Turn Velocity**: Baseline turn velocity target is <= 5 turns per subagent invocation.
- **Tool Reliability**: Target tool execution success rate is > 95%.
"""
        )

    # Top Row Cards: Core Efficiency
    e1, e2, e3, e4, e5, e6 = st.columns(6)
    with e1:
        st.metric(label="SDLC Iterations", value=eff.sdlc_iterations)
    with e2:
        st.metric(
            label="Lead Time",
            value=format_lead_time(lead_secs),
            help="Elapsed wall-clock turnaround duration from ticket start to completion.",
        )
    with e3:
        st.metric(
            label="Active Duration",
            value=eff.active_duration_formatted,
            help="Cumulative active model execution time across all invocations.",
        )
    with e4:
        st.metric(label="Total Cost (USD)", value=format_currency(totals.cost_usd))
    with e5:
        st.metric(
            label="Cache Hit Rate", value=format_percentage(totals.cache_hit_rate_pct)
        )
    with e6:
        st.metric(label="Test:Prod Ratio", value=eff.test_to_production_ratio.ratio)

    st.divider()

    # Second Row: Code Delivery, Reasoning, Tool Health
    d1, d2, d3 = st.columns(3)
    with d1:
        st.markdown("**Code Delivery**")
        st.write(f"- Lines Added: +{format_number(eff.code_delivered.lines_added)}")
        st.write(f"- Lines Deleted: -{format_number(eff.code_delivered.lines_deleted)}")
        st.write(f"- Files Changed: {format_number(eff.code_delivered.files_changed)}")
        st.write(
            f"- Test Lines: +{format_number(eff.test_to_production_ratio.test_lines_added)}"
        )
        st.write(
            f"- Production Lines: +{format_number(eff.test_to_production_ratio.prod_lines_added)}"
        )
    with d2:
        st.markdown("**Reasoning & Generation**")
        st.write(f"- Reasoning Tokens: {format_tokens(eff.reasoning_tokens)}")
        st.write(
            f"- Reasoning Ratio: {format_percentage(eff.reasoning_ratio_pct)} of output"
        )
        st.write(f"- Total Output Tokens: {format_tokens(totals.output_tokens)}")
        st.write(f"- Total Invocations: {format_number(totals.invocations)}")
    with d3:
        st.markdown("**Tool Execution Health**")
        st.write(f"- Total Tool Calls: {format_number(eff.tool_execution.total_calls)}")
        st.write(
            f"- Errors & Corrections: {format_number(eff.tool_execution.errors_and_corrections)}"
        )
        st.write(
            f"- Success Rate: {format_percentage(eff.tool_execution.success_rate_pct)}"
        )

    st.divider()

    # Context Window Health Section
    st.subheader("Context Window Health & Saturation")
    cw_col1, cw_col2 = st.columns([1, 2])
    with cw_col1:
        st.metric(
            label="Context Health Status",
            value=cwh.status,
            delta=f"{cwh.orchestrator_saturation_pct:.1f}% Saturation",
            delta_color="off",
            delta_arrow="off",
            help="Context window saturation relative to the healthy ceiling (200k tokens).",
        )
        st.write(
            f"- Healthy Ceiling: {format_tokens(cwh.healthy_ceiling_tokens)} tokens"
        )
        st.write(
            f"- Orchestrator Peak: {format_tokens(cwh.orchestrator_peak_tokens)} tokens"
        )
        st.write(
            f"- Subagents Peak: {format_tokens(cwh.subagents_peak_tokens)} tokens ({cwh.subagents_peak_role})"
        )
        st.write(f"- Compactions: {cwh.compactions}")

        # Saturation Progress Bar (normalized to ceiling)
        saturation_ratio = min(cwh.orchestrator_saturation_pct / 100.0, 1.0)
        st.progress(
            saturation_ratio,
            text=f"Orchestrator Saturation: {cwh.orchestrator_saturation_pct:.1f}%",
        )

    with cw_col2:
        st.markdown("**Peak Context Window Tokens by Role**")
        if cwh.peak_by_role:
            peak_role_df = pd.DataFrame(
                [
                    {"Role": role, "Peak Tokens": toks}
                    for role, toks in cwh.peak_by_role.items()
                ]
            )
            peak_chart = (
                alt.Chart(peak_role_df)
                .mark_bar()
                .encode(
                    x=alt.X("Peak Tokens:Q", title="Peak Context Tokens"),
                    y=alt.Y("Role:N", sort="-x", title="Role"),
                    color=alt.condition(
                        alt.datum["Peak Tokens"] > cwh.healthy_ceiling_tokens,
                        alt.value("#d9534f"),
                        alt.value("#337ab7"),
                    ),
                    tooltip=["Role", "Peak Tokens"],
                )
                .properties(height=220)
            )
            st.altair_chart(peak_chart, width="stretch")
        else:
            st.info("No peak token distribution data available.")

    st.divider()

    # Subagent Turn Velocity Section
    st.subheader("Subagent Turn Velocity & Thrashing")
    vel = eff.subagent_turn_velocity
    v1, v2 = st.columns([1, 2])
    with v1:
        st.metric(
            label="Avg Turns per Invocation",
            value=f"{vel.avg_turns_per_invocation:.1f}",
        )
        st.metric(label="Peak Turns", value=vel.peak_turns)
    with v2:
        if vel.by_role:
            vel_df = pd.DataFrame(
                [
                    {"Role": role, "Avg Turns": turns}
                    for role, turns in vel.by_role.items()
                ]
            )
            vel_chart = (
                alt.Chart(vel_df)
                .mark_bar()
                .encode(
                    x=alt.X("Avg Turns:Q", title="Average Turns"),
                    y=alt.Y("Role:N", sort="-x", title="Role"),
                    tooltip=["Role", "Avg Turns"],
                )
                .properties(height=200)
            )
            st.altair_chart(vel_chart, width="stretch")
        else:
            st.info("No turn velocity breakdown available.")

    st.divider()

    # Breakdown by Role Table
    st.subheader("Metrics by Subagent Role")
    if ledger.roles:
        role_table = [
            {
                "Role": r.role,
                "Invocations": r.invocations,
                "Input Tokens": format_tokens(r.input_tokens),
                "Output Tokens": format_tokens(r.output_tokens),
                "Cache Read Tokens": format_tokens(r.cache_read_tokens),
                "Cache Hit (%)": format_percentage(r.cache_hit_rate_pct),
                "Peak Context": format_tokens(r.peak_context_tokens),
                "Health": r.context_health,
                "Cost (USD)": format_currency(r.cost_usd),
            }
            for r in ledger.roles
        ]
        st.dataframe(pd.DataFrame(role_table), width="stretch", hide_index=True)
    else:
        st.info("No role metrics available.")

    st.divider()

    # Breakdown by Model Table
    st.subheader("Metrics by Model")
    if ledger.models:
        model_table = [
            {
                "Model": m.model,
                "Invocations": m.invocations,
                "Input Tokens": format_tokens(m.input_tokens),
                "Output Tokens": format_tokens(m.output_tokens),
                "Cache Read Tokens": format_tokens(m.cache_read_tokens),
                "Cache Hit (%)": format_percentage(m.cache_hit_rate_pct),
                "Peak Context": format_tokens(m.peak_context_tokens),
                "Health": m.context_health,
                "Cost (USD)": format_currency(m.cost_usd),
            }
            for m in ledger.models
        ]
        st.dataframe(pd.DataFrame(model_table), width="stretch", hide_index=True)
    else:
        st.info("No model metrics available.")

    st.divider()

    # Session Threads Log
    st.subheader("Session Thread Records")
    if ledger.threads:
        for t_idx, thread in enumerate(ledger.threads, start=1):
            with st.expander(
                f"Thread {t_idx}: Session {thread.session_id} (User: {thread.user}, Compactions: {thread.compactions})"
            ):
                for role_metric in thread.roles:
                    st.markdown(f"**Role: {role_metric.role}**")
                    if role_metric.models:
                        thread_model_data = [
                            {
                                "Model": m.model,
                                "Invocations": m.invocations,
                                "Input Tokens": format_tokens(m.input_tokens),
                                "Output Tokens": format_tokens(m.output_tokens),
                                "Cache Read Tokens": format_tokens(m.cache_read_tokens),
                                "Cache Hit (%)": format_percentage(
                                    m.cache_hit_rate_pct
                                ),
                                "Cost (USD)": format_currency(m.cost_usd),
                                "Peak Context": format_tokens(m.peak_context_tokens),
                                "Health": m.context_health,
                            }
                            for m in role_metric.models
                        ]
                        st.dataframe(
                            pd.DataFrame(thread_model_data),
                            width="stretch",
                            hide_index=True,
                        )
    else:
        st.info("No session thread records found.")

    st.divider()

    # Workspace Artifacts Section
    st.subheader("Workspace Artifacts")
    artifacts = read_ticket_artifacts(selected_ticket_id, base_path=workspaces_dir)

    contract_tab, summary_tab = st.tabs(
        ["Contract (CONTRACT.md)", "Delivery Summary (SUMMARY.md)"]
    )
    with contract_tab:
        if artifacts.contract_content and artifacts.contract_content.strip():
            st.markdown(artifacts.contract_content)
        else:
            st.info(f"No CONTRACT.md found for workspace {selected_ticket_id}.")

    with summary_tab:
        if artifacts.summary_content and artifacts.summary_content.strip():
            st.markdown(artifacts.summary_content)
        else:
            st.info(f"No SUMMARY.md found for workspace {selected_ticket_id}.")
