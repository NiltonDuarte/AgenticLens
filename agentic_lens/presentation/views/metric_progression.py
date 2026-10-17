from __future__ import annotations

from collections.abc import Sequence
from datetime import date

import altair as alt
import pandas as pd
import streamlit as st

from agentic_lens.domain.aggregations import (
    aggregate_metrics_over_time,
    filter_ledgers_by_date_range,
    resolve_date_preset,
)
from agentic_lens.domain.constants import HEALTHY_CEILING_TOKENS
from agentic_lens.domain.enums import TimeGranularity, TimeWindowPreset
from agentic_lens.domain.models import AgenticMetricsLedger, TimeBucketMetric
from agentic_lens.presentation.views.formatters import (
    format_currency,
    format_number,
    format_percentage,
    format_tokens,
)


def render_metric_progression(ledgers: Sequence[AgenticMetricsLedger]) -> None:
    """Render temporal progression trends, time series charts, and cumulative metrics."""
    st.subheader("Metric Progression Over Time")
    st.caption(
        "Temporal progression and velocity trends tracking completed events and deliverables across daily, weekly, and cumulative horizons."
    )

    with st.expander(
        "About Metric Progression & Temporal Analysis Reference", expanded=False
    ):
        st.markdown(
            """
### Metric Progression Overview
Metric Progression provides longitudinal temporal analysis of multi-agent software development lifecycle (SDLC) velocity, token throughput, prompt cache efficiency, code delivery volume, and context saturation over time.

---

### Progression Dimensions & Metric Formulas

- **Financial & Invocation Velocity**:
  - *Period Spend (USD)*: Aggregate model invocation cost incurred within each time bucket.
  - *Invocation Volume*: Total count of LLM API requests executed in each bucket.
  - *Cumulative Trajectory*: Running sum of expenditure and API calls across successive time periods.
- **Token Composition & Throughput**:
  - Tracks categorical token volume per period: Prompt Input Tokens, Completion Output Tokens, Prompt Cache Read Tokens, and Reasoning Tokens.
- **Prompt Cache Hit Rate Trajectory (%)**:
  $$\\text{Cache Hit Rate} = \\frac{\\text{Cache Read Tokens}}{\\text{Input Tokens} + \\text{Cache Read Tokens}} \\times 100$$
  Evaluated against standard reference thresholds: 50% baseline and 80% target.
- **Code Delivery & Net Growth**:
  - *Periodic Lines Added / Deleted*: Volume of codebase additions (+) and deletions (-) per time bucket.
  - *Cumulative Net Code*: Running net code growth ($$\\sum \\text{Lines Added} - \\sum \\text{Lines Deleted}$$).
- **Peak Context Window Saturation**:
  - Maximum context window tokens retained during any single invocation within the period, evaluated against the 200,000 token healthy ceiling.

---

### Completion Date Semantics & Temporal Alignment
All invocation events and deliverable code statistics are bucketed by their authoritative completion date (`completed_at`). For in-progress sessions, timestamps fall back to `started_at` or `updated_at`.

---

### Aggregation Granularity Mechanics
- **Daily**: Buckets events by individual calendar date (`YYYY-MM-DD`).
- **Weekly**: Groups events into 7-day buckets starting on ISO Monday.
- **Cumulative**: Computes running totals from project inception through each period.
"""
        )

    if not ledgers:
        st.info("No ticket ledgers available for metric progression analysis.")
        return

    # Top Control Bar: Time Window and Granularity
    filter_col1, filter_col2, filter_col3 = st.columns([1, 1, 1])

    with filter_col1:
        preset = st.selectbox(
            "Time Window",
            options=[
                TimeWindowPreset.PROJECT_LIFETIME.value,
                TimeWindowPreset.LAST_7_DAYS.value,
                TimeWindowPreset.LAST_30_DAYS.value,
                TimeWindowPreset.LAST_90_DAYS.value,
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
            matched_ledgers = filter_ledgers_by_date_range(
                ledgers, start_date, end_date
            )
            if start_date is not None and end_date is not None:
                st.caption(
                    f"Active Window (Completion-Dated): {start_date.isoformat()} to {end_date.isoformat()} "
                    f"({len(matched_ledgers)} tickets matched)"
                )
            else:
                st.caption(
                    f"Active Window (Completion-Dated): Project Lifetime ({len(matched_ledgers)} tickets matched)"
                )

    with filter_col3:
        granularity_label = st.radio(
            "Granularity",
            options=["Daily", "Weekly", "Cumulative"],
            horizontal=True,
            index=0,
        )

    granularity_val = (
        TimeGranularity.WEEKLY.value
        if granularity_label == "Weekly"
        else (
            TimeGranularity.CUMULATIVE.value
            if granularity_label == "Cumulative"
            else TimeGranularity.DAILY.value
        )
    )

    metrics_series: list[TimeBucketMetric] = aggregate_metrics_over_time(
        ledgers,
        granularity=granularity_val,
        start_date=start_date,
        end_date=end_date,
    )

    if not metrics_series:
        st.info("No progression data available for the selected time window.")
        return

    # Active Window KPI Banner
    total_spend = sum(m.cost_usd for m in metrics_series)
    total_invs = sum(m.invocations for m in metrics_series)
    total_tokens = sum(
        m.input_tokens + m.output_tokens + m.reasoning_tokens for m in metrics_series
    )
    total_cache_read = sum(m.cache_read_tokens for m in metrics_series)
    total_input = sum(m.input_tokens for m in metrics_series)
    cached_or_inp = total_input + total_cache_read
    hit_rate = (
        round(total_cache_read / cached_or_inp * 100, 1) if cached_or_inp > 0 else 0.0
    )
    net_lines = sum(m.lines_added for m in metrics_series) - sum(
        m.lines_deleted for m in metrics_series
    )

    st.markdown("**Active Window Summary KPI**")
    k1, k2, k3, k4, k5 = st.columns(5)
    with k1:
        st.metric(
            label="Filtered Spend (USD)",
            value=format_currency(total_spend),
            help="Total spend across active progression buckets.",
        )
    with k2:
        st.metric(
            label="Invocations",
            value=format_number(total_invs),
            help="Total invocation calls in active progression window.",
        )
    with k3:
        st.metric(
            label="Total Tokens",
            value=format_tokens(total_tokens),
            help="Total prompt, completion, and reasoning tokens processed.",
        )
    with k4:
        st.metric(
            label="Overall Cache Hit Rate",
            value=format_percentage(hit_rate),
            help="Cache read ratio across active progression window.",
        )
    with k5:
        st.metric(
            label="Net Code Delivered",
            value=f"{net_lines:+,}",
            help=f"Lines added ({sum(m.lines_added for m in metrics_series):,}) minus lines deleted ({sum(m.lines_deleted for m in metrics_series):,}).",
        )

    st.divider()

    df_progression = pd.DataFrame(
        [
            {
                "bucket_key": m.bucket_key,
                "bucket_date": m.bucket_date,
                "invocations": m.invocations,
                "cost_usd": m.cost_usd,
                "input_tokens": m.input_tokens,
                "output_tokens": m.output_tokens,
                "cache_read_tokens": m.cache_read_tokens,
                "reasoning_tokens": m.reasoning_tokens,
                "cache_hit_rate_pct": m.cache_hit_rate_pct,
                "lines_added": m.lines_added,
                "lines_deleted": m.lines_deleted,
                "files_changed": m.files_changed,
                "net_lines_added": m.net_lines_added,
                "peak_context_tokens": m.peak_context_tokens,
                "active_duration_seconds": m.active_duration_seconds,
                "cumulative_cost_usd": m.cumulative_cost_usd,
                "cumulative_input_tokens": m.cumulative_input_tokens,
                "cumulative_output_tokens": m.cumulative_output_tokens,
                "cumulative_lines_added": m.cumulative_lines_added,
                "cumulative_lines_deleted": m.cumulative_lines_deleted,
                "cumulative_invocations": m.cumulative_invocations,
            }
            for m in metrics_series
        ]
    )

    tab1, tab2, tab3, tab4 = st.tabs(
        [
            "Financial & Invocations",
            "Token Throughput & Cache Efficiency",
            "Code Delivery & Growth",
            "Context Saturation",
        ]
    )

    with tab1:
        c1, c2 = st.columns(2)
        with c1:
            st.markdown("**Spend Progression (USD)**")
            if granularity_val == TimeGranularity.CUMULATIVE.value:
                spend_chart = (
                    alt.Chart(df_progression)
                    .mark_area(opacity=0.6, color="#1f77b4")
                    .encode(
                        x=alt.X("bucket_date:N", title="Time Period"),
                        y=alt.Y("cumulative_cost_usd:Q", title="Cumulative Cost (USD)"),
                        tooltip=[
                            alt.Tooltip("bucket_date:N", title="Period"),
                            alt.Tooltip(
                                "cumulative_cost_usd:Q",
                                title="Cumulative Cost",
                                format="$.4f",
                            ),
                            alt.Tooltip(
                                "cost_usd:Q", title="Period Cost", format="$.4f"
                            ),
                        ],
                    )
                    .properties(height=320)
                )
            else:
                spend_chart = (
                    alt.Chart(df_progression)
                    .mark_bar(color="#1f77b4")
                    .encode(
                        x=alt.X("bucket_date:N", title="Time Period"),
                        y=alt.Y("cost_usd:Q", title="Cost (USD)"),
                        tooltip=[
                            alt.Tooltip("bucket_date:N", title="Period"),
                            alt.Tooltip(
                                "cost_usd:Q", title="Cost (USD)", format="$.4f"
                            ),
                            alt.Tooltip("invocations:Q", title="Invocations"),
                        ],
                    )
                    .properties(height=320)
                )
            st.altair_chart(spend_chart, width="stretch")

        with c2:
            st.markdown("**Invocation Volume & Duration**")
            if granularity_val == TimeGranularity.CUMULATIVE.value:
                inv_chart = (
                    alt.Chart(df_progression)
                    .mark_line(point=True, color="#ff7f0e")
                    .encode(
                        x=alt.X("bucket_date:N", title="Time Period"),
                        y=alt.Y(
                            "cumulative_invocations:Q",
                            title="Cumulative Invocations",
                        ),
                        tooltip=[
                            alt.Tooltip("bucket_date:N", title="Period"),
                            alt.Tooltip(
                                "cumulative_invocations:Q",
                                title="Cumulative Invocations",
                            ),
                            alt.Tooltip("invocations:Q", title="Period Invocations"),
                        ],
                    )
                    .properties(height=320)
                )
            else:
                inv_chart = (
                    alt.Chart(df_progression)
                    .mark_bar(color="#ff7f0e")
                    .encode(
                        x=alt.X("bucket_date:N", title="Time Period"),
                        y=alt.Y("invocations:Q", title="Invocations"),
                        tooltip=[
                            alt.Tooltip("bucket_date:N", title="Period"),
                            alt.Tooltip("invocations:Q", title="Invocations"),
                            alt.Tooltip(
                                "active_duration_seconds:Q",
                                title="Active Duration (s)",
                            ),
                        ],
                    )
                    .properties(height=320)
                )
            st.altair_chart(inv_chart, width="stretch")

    with tab2:
        c1, c2 = st.columns(2)
        with c1:
            st.markdown("**Token Composition by Type**")
            df_tokens_long = pd.melt(
                df_progression,
                id_vars=["bucket_date"],
                value_vars=[
                    "input_tokens",
                    "output_tokens",
                    "cache_read_tokens",
                    "reasoning_tokens",
                ],
                var_name="Token Type",
                value_name="Tokens",
            )
            token_label_map = {
                "input_tokens": "Input Tokens",
                "output_tokens": "Output Tokens",
                "cache_read_tokens": "Cache Read Tokens",
                "reasoning_tokens": "Reasoning Tokens",
            }
            df_tokens_long["Token Type"] = df_tokens_long["Token Type"].map(
                token_label_map
            )
            token_chart = (
                alt.Chart(df_tokens_long)
                .mark_bar()
                .encode(
                    x=alt.X("bucket_date:N", title="Time Period"),
                    y=alt.Y("Tokens:Q", title="Tokens"),
                    color=alt.Color(
                        "Token Type:N",
                        title="Token Type",
                        scale=alt.Scale(
                            domain=[
                                "Input Tokens",
                                "Output Tokens",
                                "Cache Read Tokens",
                                "Reasoning Tokens",
                            ],
                            range=["#3366cc", "#dc3912", "#ff9900", "#109618"],
                        ),
                    ),
                    tooltip=[
                        "bucket_date:N",
                        "Token Type:N",
                        alt.Tooltip("Tokens:Q", format=","),
                    ],
                )
                .properties(height=320)
            )
            st.altair_chart(token_chart, width="stretch")

        with c2:
            st.markdown("**Prompt Cache Hit Rate Trend (%)**")
            cache_chart = (
                alt.Chart(df_progression)
                .mark_line(point=True, color="#2ca02c")
                .encode(
                    x=alt.X("bucket_date:N", title="Time Period"),
                    y=alt.Y(
                        "cache_hit_rate_pct:Q",
                        title="Cache Hit Rate (%)",
                        scale=alt.Scale(domain=[0, 100]),
                    ),
                    tooltip=[
                        alt.Tooltip("bucket_date:N", title="Period"),
                        alt.Tooltip(
                            "cache_hit_rate_pct:Q",
                            title="Hit Rate (%)",
                            format=".1f",
                        ),
                        alt.Tooltip(
                            "cache_read_tokens:Q",
                            title="Cache Read Tokens",
                            format=",",
                        ),
                        alt.Tooltip("input_tokens:Q", title="Input Tokens", format=","),
                    ],
                )
            )
            rule_50 = (
                alt.Chart(pd.DataFrame({"y": [50.0]}))
                .mark_rule(strokeDash=[4, 4], color="#f0ad4e")
                .encode(y="y:Q")
            )
            rule_80 = (
                alt.Chart(pd.DataFrame({"y": [80.0]}))
                .mark_rule(strokeDash=[4, 4], color="#5cb85c")
                .encode(y="y:Q")
            )
            combined_cache_chart = (cache_chart + rule_50 + rule_80).properties(
                height=320
            )
            st.altair_chart(combined_cache_chart, width="stretch")

    with tab3:
        c1, c2 = st.columns(2)
        with c1:
            st.markdown("**Code Delivery Volume (Lines)**")
            df_code_long = pd.melt(
                df_progression,
                id_vars=["bucket_date"],
                value_vars=["lines_added", "lines_deleted"],
                var_name="Change Type",
                value_name="Lines",
            )
            df_code_long["Change Type"] = df_code_long["Change Type"].map(
                {"lines_added": "Lines Added (+)", "lines_deleted": "Lines Deleted (-)"}
            )
            code_chart = (
                alt.Chart(df_code_long)
                .mark_bar()
                .encode(
                    x=alt.X("bucket_date:N", title="Time Period"),
                    y=alt.Y("Lines:Q", title="Lines of Code"),
                    color=alt.Color(
                        "Change Type:N",
                        title="Change Type",
                        scale=alt.Scale(
                            domain=["Lines Added (+)", "Lines Deleted (-)"],
                            range=["#28a745", "#dc3545"],
                        ),
                    ),
                    xOffset=alt.XOffset("Change Type:N"),
                    tooltip=[
                        "bucket_date:N",
                        "Change Type:N",
                        alt.Tooltip("Lines:Q", format=","),
                    ],
                )
                .properties(height=320)
            )
            st.altair_chart(code_chart, width="stretch")

        with c2:
            st.markdown("**Cumulative Code Growth (Net Lines)**")
            df_progression_net = df_progression.copy()
            df_progression_net["cumulative_net_lines"] = (
                df_progression_net["cumulative_lines_added"]
                - df_progression_net["cumulative_lines_deleted"]
            )
            growth_chart = (
                alt.Chart(df_progression_net)
                .mark_area(opacity=0.5, color="#6f42c1")
                .encode(
                    x=alt.X("bucket_date:N", title="Time Period"),
                    y=alt.Y("cumulative_net_lines:Q", title="Cumulative Net Lines"),
                    tooltip=[
                        alt.Tooltip("bucket_date:N", title="Period"),
                        alt.Tooltip(
                            "cumulative_net_lines:Q",
                            title="Cumulative Net Lines",
                            format=",",
                        ),
                        alt.Tooltip(
                            "cumulative_lines_added:Q",
                            title="Cumulative Added",
                            format=",",
                        ),
                        alt.Tooltip(
                            "cumulative_lines_deleted:Q",
                            title="Cumulative Deleted",
                            format=",",
                        ),
                    ],
                )
                .properties(height=320)
            )
            st.altair_chart(growth_chart, width="stretch")

    with tab4:
        st.markdown("**Peak Context Window Saturation (Tokens)**")
        peak_chart = (
            alt.Chart(df_progression)
            .mark_bar(color="#17a2b8")
            .encode(
                x=alt.X("bucket_date:N", title="Time Period"),
                y=alt.Y("peak_context_tokens:Q", title="Peak Context Tokens"),
                tooltip=[
                    alt.Tooltip("bucket_date:N", title="Period"),
                    alt.Tooltip(
                        "peak_context_tokens:Q",
                        title="Peak Context Tokens",
                        format=",",
                    ),
                ],
            )
        )
        ceiling_rule = (
            alt.Chart(pd.DataFrame({"y": [HEALTHY_CEILING_TOKENS]}))
            .mark_rule(strokeDash=[4, 4], color="#d9534f")
            .encode(y="y:Q")
        )
        combined_peak_chart = (peak_chart + ceiling_rule).properties(height=320)
        st.altair_chart(combined_peak_chart, width="stretch")

    st.divider()

    st.subheader("Data Table and Export")
    st.dataframe(df_progression, width="stretch", hide_index=True)
    csv_bytes = df_progression.to_csv(index=False).encode("utf-8")
    st.download_button(
        label="Download Progression CSV",
        data=csv_bytes,
        file_name="metric_progression.csv",
        mime="text/csv",
        width="stretch",
    )
