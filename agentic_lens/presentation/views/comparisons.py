from __future__ import annotations

from collections.abc import Sequence

import altair as alt
import pandas as pd
import streamlit as st

from agentic_lens.domain.aggregations import build_comparison_rows
from agentic_lens.domain.models import AgenticMetricsLedger
from agentic_lens.presentation.views.formatters import format_iso_datetime


def render_comparisons(ledgers: Sequence[AgenticMetricsLedger]) -> None:
    """Render cross-ticket tabular comparison matrix and correlation charts."""
    st.subheader("Cross-Ticket Comparison Matrix")
    st.caption(
        "Cross-ticket benchmarking matrix and comparative analytics ordered by deliverable completion date."
    )

    with st.expander(
        "About Cross-Ticket Comparisons & Benchmark Reference", expanded=False
    ):
        st.markdown(
            """
### Cross-Ticket Comparisons Overview
Cross-Ticket Comparisons provides multi-dimensional benchmarking across ticket deliverables, contrasting turnaround lead time, active compute duration, cost efficiency, test coverage rigor, tool execution reliability, and context saturation.

---

### Comparison Matrix Columns & Benchmark Formulas

- **Completed & Started**:
  - *Completed*: Authoritative timestamp marking deliverable completion (`completed_at`). Active tickets display as `In Progress`.
  - *Started*: Timestamp marking ticket initiation (`started_at`).
- **Lead Time vs Active Duration**:
  - *Lead Time*: Elapsed wall-clock calendar duration ($$\\text{completed\\_at} - \\text{started\\_at}$$).
  - *Active Duration*: Cumulative compute latency spent actively executing LLM inferences across all subagents.
- **Iterations**: Total number of planner-coder-tester feedback loops executed to finalize the ticket.
- **Cost (USD)**: Aggregate LLM inference expenditure across all subagent sessions and model tiers.
- **Prompt Cache Hit Rate (%)**: Proportion of prompt tokens retrieved from cache ($$\\text{cache\\_read} / (\\text{input} + \\text{cache\\_read}) \\times 100$$).
- **Test:Prod Ratio**: Ratio of test code lines added to production code lines added ($$\\text{test\\_lines\\_added} : \\text{prod\\_lines\\_added}$$).
- **Code Delivered**: Volume metrics including lines added (+), lines deleted (-), and total files changed.
- **Tool Success Rate (%)**: Tool execution reliability metric:
  $$\\text{Tool Success Rate} = \\frac{\\text{Total Calls} - \\text{Errors}}{\\text{Total Calls}} \\times 100$$
- **Context Health & Orchestrator Saturation (%)**:
  - *Context Health*: Categorical health rating (`HEALTHY`, `SATURATED`, `CRITICAL`).
  - *Orchestrator Saturation*: Orchestrator peak context window usage relative to the 200,000 token ceiling ($$\\text{peak\\_tokens} / 200,000 \\times 100$$).

---

### Completion Date Semantics & Row Sorting
Comparison rows are sorted primarily by deliverable completion timestamp (`completed_at` descending), placing the most recently finalized deliverables first, followed by active in-progress tickets.

---

### Operational Benchmarks & Baselines
- **Tool Reliability Baseline**: Target tool execution success rate is > 95%.
- **Context Window Ceiling**: Healthy saturation ceiling is 200,000 tokens (<= 100%).
- **Test Coverage Rigor Baseline**: Target test-to-production ratio is >= 1.0:1.
"""
        )

    rows = build_comparison_rows(ledgers)
    if not rows:
        st.info("No ticket records available for comparison.")
        return

    table_data = [
        {
            "Ticket ID": r.ticket_id,
            "Title": r.ticket_title,
            "Completed": (
                format_iso_datetime(r.completed_at) if r.completed_at else "In Progress"
            ),
            "Started": format_iso_datetime(r.started_at),
            "Lead Time": r.lead_duration_formatted,
            "Active Duration": r.active_duration_formatted,
            "Iterations": r.sdlc_iterations,
            "Cost (USD)": r.cost_usd,
            "Cache Hit Rate (%)": r.cache_hit_rate_pct,
            "Test:Prod Ratio": r.test_to_code_ratio,
            "Lines Added": r.lines_added,
            "Lines Deleted": r.lines_deleted,
            "Files Changed": r.files_changed,
            "Tool Success Rate (%)": r.tool_success_rate_pct,
            "Context Health": r.context_health_status,
            "Orchestrator Saturation (%)": r.orchestrator_saturation_pct,
        }
        for r in rows
    ]

    df_comparison = pd.DataFrame(table_data)
    st.dataframe(
        df_comparison,
        width="stretch",
        hide_index=True,
    )

    st.divider()

    st.subheader("Comparative Analytics")

    comp_col1, comp_col2 = st.columns(2)

    with comp_col1:
        st.markdown("**Lead Time vs Active Execution Duration (Minutes)**")
        dur_data = [
            {
                "Ticket ID": r.ticket_id,
                "Title": r.ticket_title,
                "Lead Time (m)": round((r.lead_duration_seconds or 0) / 60, 1),
                "Active Duration (m)": round(r.active_duration_seconds / 60, 1),
                "Lead Time": r.lead_duration_formatted,
                "Active Duration": r.active_duration_formatted,
                "Status": "Completed" if r.completed_at else "In Progress",
            }
            for r in rows
        ]
        df_dur = pd.DataFrame(dur_data)
        df_dur_melted = pd.melt(
            df_dur,
            id_vars=[
                "Ticket ID",
                "Title",
                "Lead Time",
                "Active Duration",
                "Status",
            ],
            value_vars=["Lead Time (m)", "Active Duration (m)"],
            var_name="Duration Metric",
            value_name="Minutes",
        )
        dur_chart = (
            alt.Chart(df_dur_melted)
            .mark_bar()
            .encode(
                x=alt.X("Ticket ID:N", title="Ticket"),
                y=alt.Y("Minutes:Q", title="Duration (Minutes)"),
                color=alt.Color(
                    "Duration Metric:N",
                    title="Duration Metric",
                    scale=alt.Scale(
                        domain=["Lead Time (m)", "Active Duration (m)"],
                        range=["#17a2b8", "#ff7f0e"],
                    ),
                ),
                xOffset=alt.XOffset("Duration Metric:N"),
                tooltip=[
                    "Ticket ID",
                    "Title",
                    "Duration Metric",
                    alt.Tooltip("Minutes:Q", format=".1f"),
                    "Lead Time",
                    "Active Duration",
                    "Status",
                ],
            )
            .properties(height=300)
        )
        st.altair_chart(dur_chart, width="stretch")

    with comp_col2:
        st.markdown("**Tool Execution Success Rate by Ticket**")
        tool_chart = (
            alt.Chart(df_comparison)
            .mark_bar()
            .encode(
                x=alt.X("Ticket ID:N", title="Ticket"),
                y=alt.Y(
                    "Tool Success Rate (%):Q",
                    scale=alt.Scale(domain=[90, 100]),
                    title="Success Rate (%)",
                ),
                color=alt.Color(
                    "Tool Success Rate (%):Q",
                    scale=alt.Scale(scheme="blues"),
                    legend=None,
                ),
                tooltip=["Ticket ID", "Title", "Tool Success Rate (%)"],
            )
            .properties(height=300)
        )
        st.altair_chart(tool_chart, width="stretch")

    st.divider()

    comp_col3, comp_col4 = st.columns(2)

    with comp_col3:
        st.markdown("**Orchestrator Context Window Saturation by Ticket**")
        sat_chart = (
            alt.Chart(df_comparison)
            .mark_bar()
            .encode(
                x=alt.X("Ticket ID:N", title="Ticket"),
                y=alt.Y("Orchestrator Saturation (%):Q", title="Saturation (%)"),
                color=alt.condition(
                    alt.datum["Orchestrator Saturation (%)"] > 100,
                    alt.value("#d9534f"),
                    alt.value("#5cb85c"),
                ),
                tooltip=[
                    "Ticket ID",
                    "Title",
                    "Orchestrator Saturation (%)",
                    "Context Health",
                ],
            )
            .properties(height=300)
        )
        st.altair_chart(sat_chart, width="stretch")

    with comp_col4:
        st.markdown("**Code Delivery Volume vs Financial Cost**")
        bubble_chart = (
            alt.Chart(df_comparison)
            .mark_circle(size=120)
            .encode(
                x=alt.X("Lines Added:Q", title="Lines Added"),
                y=alt.Y("Cost (USD):Q", title="Cost (USD)"),
                color=alt.Color("Context Health:N", title="Context Health"),
                tooltip=[
                    "Ticket ID",
                    "Title",
                    "Lines Added",
                    "Lines Deleted",
                    "Cost (USD)",
                    "Active Duration",
                ],
            )
            .properties(height=300)
        )
        st.altair_chart(bubble_chart, width="stretch")
