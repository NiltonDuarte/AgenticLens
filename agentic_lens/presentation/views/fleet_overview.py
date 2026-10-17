from __future__ import annotations

from collections.abc import Sequence
from datetime import date

import altair as alt
import pandas as pd
import streamlit as st

from agentic_lens.domain.aggregations import (
    calculate_fleet_summary,
    calculate_role_effort_rollup,
    filter_ledgers_by_date_range,
    resolve_date_preset,
)
from agentic_lens.domain.enums import TimeWindowPreset
from agentic_lens.domain.models import AgenticMetricsLedger, FleetSummary
from agentic_lens.presentation.views.formatters import (
    format_currency,
    format_number,
    format_percentage,
    format_tokens,
)


def render_fleet_overview(
    ledgers: Sequence[AgenticMetricsLedger],
    summary: FleetSummary | None = None,
) -> None:
    """Render fleet-wide KPI metrics and aggregate trend visualizations."""
    st.subheader("Fleet Executive Summary")
    st.caption(
        "Fleet-wide executive metrics, cost attribution, and subagent role effort filtered by deliverable completion date."
    )

    with st.expander(
        "About Fleet Executive Summary & Telemetry Reference", expanded=False
    ):
        st.markdown(
            """
### Fleet Executive Summary Overview
The Fleet Executive Summary provides an aggregated high-level view of multi-agent software development lifecycle (SDLC) telemetry across ticket workspaces within the selected completion date window.

---

### Core Metric Definitions & Formulas

- **Fleet Cost (USD)**: Cumulative financial expenditure across all LLM model invocations, computed from token pricing tiers (prompt input, prompt cache read, completion output, and reasoning tokens).
- **Total Invocations**: Sum of all LLM API calls executed by the orchestrator and subagents across all tickets in the active window.
- **Prompt Cache Hit Rate (%)**: Proportion of prompt input tokens served from prompt cache:
  $$\\text{Cache Hit Rate} = \\frac{\\text{Cache Read Tokens}}{\\text{Input Tokens} + \\text{Cache Read Tokens}} \\times 100$$
- **Active Duration vs Lead Time**:
  - *Active Duration*: Total cumulative execution time spent actively processing LLM requests across all subagents.
  - *Lead Time*: Wall-clock calendar duration elapsed from initial ticket kickoff (`started_at`) to final deliverable completion (`completed_at`).
- **Reasoning Ratio (%)**: Proportion of generated tokens dedicated to internal reasoning:
  $$\\text{Reasoning Ratio} = \\frac{\\text{Reasoning Tokens}}{\\text{Output Tokens} + \\text{Reasoning Tokens}} \\times 100$$

---

### Completion Date Filtering Semantics
The active time window filters ticket workspaces using the authoritative deliverable completion date (`completed_at`). For tickets currently in progress without a completion date, filters fall back to `started_at` or `updated_at`.

---

### Operational Baselines & Thresholds
- **Cache Efficiency**: Target cache hit rate benchmark is > 80% (acceptable baseline > 50%).
- **Context Ceiling**: Healthy context ceiling is 200,000 tokens per invocation session.
- **Subagent Role Attribution**: Work is distributed across canonical roles (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`).
"""
        )

    filter_col1, filter_col2 = st.columns([1, 2])
    with filter_col1:
        preset = st.selectbox(
            "Time Window",
            options=[
                TimeWindowPreset.LAST_30_DAYS.value,
                TimeWindowPreset.LAST_7_DAYS.value,
                TimeWindowPreset.LAST_90_DAYS.value,
                TimeWindowPreset.PROJECT_LIFETIME.value,
                TimeWindowPreset.CUSTOM_RANGE.value,
            ],
            index=0,
        )

    start_date: date | None = None
    end_date: date | None = None
    with filter_col2:
        if preset == TimeWindowPreset.CUSTOM_RANGE.value:
            date_range = st.date_input("Select Date Range", value=())
            if isinstance(date_range, tuple | list):
                if len(date_range) == 2:
                    start_date, end_date = date_range[0], date_range[1]
                elif len(date_range) == 1:
                    start_date, end_date = date_range[0], date_range[0]
            elif isinstance(date_range, date):
                start_date, end_date = date_range, date_range
        else:
            start_date, end_date = resolve_date_preset(preset)
            filtered_temp = filter_ledgers_by_date_range(ledgers, start_date, end_date)
            if start_date is not None and end_date is not None:
                st.caption(
                    f"Active Window (Completion-Dated): {start_date.isoformat()} to {end_date.isoformat()} "
                    f"({len(filtered_temp)} tickets matched)"
                )
            else:
                st.caption(
                    f"Active Window (Completion-Dated): Project Lifetime ({len(filtered_temp)} tickets matched)"
                )

    filtered_ledgers = filter_ledgers_by_date_range(ledgers, start_date, end_date)
    if not filtered_ledgers:
        st.info("No ticket ledgers found for the selected time window.")
        return

    active_summary = (
        calculate_fleet_summary(filtered_ledgers)
        if (summary is None or start_date is not None or end_date is not None)
        else summary
    )

    # First row metrics: Core Executive KPIs
    kpi1, kpi2, kpi3, kpi4, kpi5 = st.columns(5)
    with kpi1:
        st.metric(
            label="Total Tickets",
            value=active_summary.total_tickets,
            help="Total number of analyzed ticket workspaces in active window.",
        )
    with kpi2:
        st.metric(
            label="Fleet Cost (USD)",
            value=format_currency(active_summary.total_cost_usd),
            help="Aggregate LLM invocation cost across active tickets in USD.",
        )
    with kpi3:
        st.metric(
            label="Total Invocations",
            value=format_number(active_summary.total_invocations),
            help="Total number of model API calls made across active tickets.",
        )
    with kpi4:
        st.metric(
            label="Cache Hit Rate",
            value=format_percentage(active_summary.overall_cache_hit_rate_pct),
            help="Percentage of prompt input tokens served from prompt cache.",
        )
    with kpi5:
        st.metric(
            label="Active Duration",
            value=active_summary.formatted_total_duration,
            help="Cumulative active workflow execution duration across active tickets.",
        )

    st.divider()

    # Second row metrics: Token Totals and Reasoning Effort
    sub1, sub2, sub3, sub4 = st.columns(4)
    with sub1:
        st.metric(
            label="Input Tokens",
            value=format_tokens(active_summary.total_input_tokens),
            help="Total prompt input tokens processed across all models.",
        )
    with sub2:
        st.metric(
            label="Output Tokens",
            value=format_tokens(active_summary.total_output_tokens),
            help="Total completion output tokens generated across all models.",
        )
    with sub3:
        st.metric(
            label="Cache Read Tokens",
            value=format_tokens(active_summary.total_cache_read_tokens),
            help="Total input tokens retrieved from prompt cache.",
        )
    with sub4:
        st.metric(
            label="Reasoning Tokens",
            value=format_tokens(active_summary.total_reasoning_tokens),
            help=(
                "Total reasoning tokens generated. Overall reasoning ratio: "
                f"{format_percentage(active_summary.overall_reasoning_ratio_pct)} of total output."
            ),
        )

    st.divider()

    chart_col1, chart_col2 = st.columns(2)

    with chart_col1:
        st.subheader("Cost by Ticket (USD)")
        ticket_cost_data = [
            {
                "Ticket ID": ledger.ticket_id or "Unknown",
                "Title": ledger.ticket_title or "N/A",
                "Cost USD": ledger.totals.cost_usd,
            }
            for ledger in filtered_ledgers
        ]
        if ticket_cost_data:
            df_ticket_costs = pd.DataFrame(ticket_cost_data)
            cost_chart = (
                alt.Chart(df_ticket_costs)
                .mark_bar()
                .encode(
                    x=alt.X("Ticket ID:N", sort="-y", title="Ticket"),
                    y=alt.Y("Cost USD:Q", title="Cost (USD)"),
                    tooltip=[
                        "Ticket ID",
                        "Title",
                        alt.Tooltip("Cost USD:Q", format="$.4f"),
                    ],
                )
                .properties(height=320)
            )
            st.altair_chart(cost_chart, width="stretch")
        else:
            st.info("No ticket cost data available.")

    with chart_col2:
        st.subheader("Cost by Model (USD)")
        if active_summary.model_costs:
            df_model_costs = pd.DataFrame(
                [
                    {"Model": model, "Cost USD": cost}
                    for model, cost in active_summary.model_costs.items()
                ]
            )
            model_chart = (
                alt.Chart(df_model_costs)
                .mark_bar()
                .encode(
                    x=alt.X("Cost USD:Q", title="Cost (USD)"),
                    y=alt.Y("Model:N", sort="-x", title="Model"),
                    tooltip=["Model", alt.Tooltip("Cost USD:Q", format="$.4f")],
                )
                .properties(height=320)
            )
            st.altair_chart(model_chart, width="stretch")
        else:
            st.info("No model cost data available.")

    st.divider()

    st.subheader("Subagent Role Effort Rollup")
    role_rollup = calculate_role_effort_rollup(filtered_ledgers)
    if role_rollup:
        role_rows = []
        for role_name, stats in role_rollup.items():
            role_rows.append(
                {
                    "Role": role_name,
                    "Invocations": stats["invocations"],
                    "Input Tokens": format_tokens(int(stats["input_tokens"])),
                    "Output Tokens": format_tokens(int(stats["output_tokens"])),
                    "Cache Read Tokens": format_tokens(int(stats["cache_read_tokens"])),
                    "Cost (USD)": format_currency(float(stats["cost_usd"])),
                }
            )
        role_order = ["orchestrator", "planner", "coder", "tester", "reviewer"]
        df_roles = pd.DataFrame(role_rows)
        df_roles["sort_key"] = df_roles["Role"].apply(
            lambda r: role_order.index(r) if r in role_order else 99
        )
        df_roles = df_roles.sort_values("sort_key").drop(columns=["sort_key"])
        st.dataframe(df_roles, width="stretch", hide_index=True)
    else:
        st.info("No role rollup data available.")
