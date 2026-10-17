from __future__ import annotations

from collections.abc import Sequence

import altair as alt
import pandas as pd
import streamlit as st

from agentic_lens.domain.aggregations import compute_lead_time_seconds
from agentic_lens.domain.models import AgenticMetricsLedger
from agentic_lens.presentation.views.formatters import (
    format_currency,
    format_number,
    format_percentage,
    format_ticket_label,
    format_tokens,
)


def build_invocations_dataframe(
    ledgers: Sequence[AgenticMetricsLedger],
) -> pd.DataFrame:
    """Flatten invocation facts across all ledgers into an event-level DataFrame."""
    rows: list[dict[str, object]] = []
    for ledger in ledgers:
        tid = ledger.ticket_id or "unknown"
        ticket_lbl = format_ticket_label(tid, ledger.ticket_title)
        for inv in ledger.facts.invocations:
            date_str = (
                inv.completed_at[:10]
                if inv.completed_at
                else (inv.started_at[:10] if inv.started_at else "N/A")
            )
            total_toks = inv.tokens.input + inv.tokens.output + inv.tokens.reasoning
            tot_cached_input = inv.tokens.input + inv.tokens.cache_read
            hit_pct = (
                round(inv.tokens.cache_read / tot_cached_input * 100, 1)
                if tot_cached_input > 0
                else 0.0
            )
            rows.append(
                {
                    "ticket_id": tid,
                    "ticket_title": ledger.ticket_title or "N/A",
                    "ticket_label": ticket_lbl,
                    "id": inv.id,
                    "parent_id": inv.parent_id or "root",
                    "is_root": inv.parent_id is None,
                    "role": inv.role,
                    "model": inv.model,
                    "user": inv.user,
                    "started_at": inv.started_at or "N/A",
                    "completed_at": inv.completed_at or "N/A",
                    "date": date_str,
                    "active_duration_seconds": inv.active_duration_seconds,
                    "turns": inv.turns,
                    "input_tokens": inv.tokens.input,
                    "output_tokens": inv.tokens.output,
                    "cache_read_tokens": inv.tokens.cache_read,
                    "reasoning_tokens": inv.tokens.reasoning,
                    "total_tokens": total_toks,
                    "cache_hit_rate_pct": hit_pct,
                    "peak_context_tokens": inv.tokens.peak_context,
                    "cost_usd": round(inv.cost_usd, 6),
                    "compactions": inv.compactions,
                }
            )
    if not rows:
        return pd.DataFrame(
            columns=[
                "ticket_id",
                "ticket_title",
                "ticket_label",
                "id",
                "parent_id",
                "is_root",
                "role",
                "model",
                "user",
                "started_at",
                "completed_at",
                "date",
                "active_duration_seconds",
                "turns",
                "input_tokens",
                "output_tokens",
                "cache_read_tokens",
                "reasoning_tokens",
                "total_tokens",
                "cache_hit_rate_pct",
                "peak_context_tokens",
                "cost_usd",
                "compactions",
            ]
        )
    return pd.DataFrame(rows)


def build_tools_dataframe(
    ledgers: Sequence[AgenticMetricsLedger],
) -> pd.DataFrame:
    """Flatten tool facts across all ledgers into a tool usage DataFrame."""
    rows: list[dict[str, object]] = []
    for ledger in ledgers:
        tid = ledger.ticket_id or "unknown"
        ticket_lbl = format_ticket_label(tid, ledger.ticket_title)
        for t in ledger.facts.tools:
            succ = max(t.calls - t.errors, 0)
            rate = round(succ / t.calls * 100, 1) if t.calls > 0 else 100.0
            rows.append(
                {
                    "ticket_id": tid,
                    "ticket_title": ledger.ticket_title or "N/A",
                    "ticket_label": ticket_lbl,
                    "tool_name": t.name,
                    "calls": t.calls,
                    "errors": t.errors,
                    "successes": succ,
                    "success_rate_pct": rate,
                }
            )
    if not rows:
        return pd.DataFrame(
            columns=[
                "ticket_id",
                "ticket_title",
                "ticket_label",
                "tool_name",
                "calls",
                "errors",
                "successes",
                "success_rate_pct",
            ]
        )
    return pd.DataFrame(rows)


def build_code_dataframe(
    ledgers: Sequence[AgenticMetricsLedger],
) -> pd.DataFrame:
    """Flatten code delivery and SDLC metrics across all ledgers into a DataFrame."""
    rows: list[dict[str, object]] = []
    for ledger in ledgers:
        tid = ledger.ticket_id or "unknown"
        ticket_lbl = format_ticket_label(tid, ledger.ticket_title)
        code = ledger.facts.code
        net_lines = code.lines_added - code.lines_deleted
        ratio = (
            round(code.test_lines_added / code.prod_lines_added, 2)
            if code.prod_lines_added > 0
            else 0.0
        )
        lead_secs = compute_lead_time_seconds(ledger.started_at, ledger.completed_at)
        date_str = (
            ledger.completed_at[:10]
            if ledger.completed_at
            else (ledger.started_at[:10] if ledger.started_at else "N/A")
        )
        rows.append(
            {
                "ticket_id": tid,
                "ticket_title": ledger.ticket_title or "N/A",
                "ticket_label": ticket_lbl,
                "ticket_type": ledger.ticket_type or "N/A",
                "started_at": ledger.started_at or "N/A",
                "completed_at": ledger.completed_at or "N/A",
                "updated_at": ledger.updated_at or "N/A",
                "date": date_str,
                "lead_time_seconds": lead_secs,
                "is_completed": ledger.completed_at is not None,
                "sdlc_iterations": code.sdlc_iterations,
                "files_changed": code.files_changed,
                "lines_added": code.lines_added,
                "lines_deleted": code.lines_deleted,
                "net_lines_added": net_lines,
                "prod_lines_added": code.prod_lines_added,
                "prod_lines_deleted": code.prod_lines_deleted,
                "test_lines_added": code.test_lines_added,
                "test_lines_deleted": code.test_lines_deleted,
                "test_to_code_ratio": ratio,
            }
        )
    if not rows:
        return pd.DataFrame(
            columns=[
                "ticket_id",
                "ticket_title",
                "ticket_label",
                "ticket_type",
                "started_at",
                "completed_at",
                "updated_at",
                "date",
                "lead_time_seconds",
                "is_completed",
                "sdlc_iterations",
                "files_changed",
                "lines_added",
                "lines_deleted",
                "net_lines_added",
                "prod_lines_added",
                "prod_lines_deleted",
                "test_lines_added",
                "test_lines_deleted",
                "test_to_code_ratio",
            ]
        )
    return pd.DataFrame(rows)


def render_analytics_studio(
    ledgers: Sequence[AgenticMetricsLedger],
) -> None:
    """Render the exploratory Analytics Studio view."""
    st.subheader("Analytics Studio")
    st.caption(
        "Ad-hoc exploratory data investigation, slicing, dynamic aggregations, and custom visualizations."
    )

    with st.expander("About Analytics Studio & Dataset Reference", expanded=False):
        st.markdown(
            """
### Analytics Studio Overview
Analytics Studio provides an interactive exploratory workbench for filtering, slicing, dynamically aggregating, and visualizing multi-agent software development lifecycle (SDLC) telemetry datasets across tickets.

---

### Dataset Architecture & Schemas

#### 1. Invocations (Event Level)
An atomic record for every LLM model invocation across agent sessions:
- **ticket_id / ticket_title / ticket_label**: Unique Jira ticket identifier, descriptive deliverable title, and formatted composite label.
- **id / parent_id / is_root**: Session hierarchy identifiers indicating root orchestrator vs subagent delegation.
- **role**: Subagent execution role (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`).
- **model**: Model identifier used for the invocation (e.g., `gemini-2.5-pro`, `gemini-2.5-flash`).
- **user**: User or subagent persona initiating the invocation.
- **started_at / completed_at**: ISO 8601 timestamps marking invocation commencement and finalization.
- **date**: Calendar date bucket derived primarily from `completed_at` (falling back to `started_at`).
- **active_duration_seconds**: Execution duration in seconds spent actively processing the LLM request.
- **turns**: Number of conversation turns executed in the session.
- **input_tokens / output_tokens / cache_read_tokens / reasoning_tokens**: Categorized token consumption metrics.
- **total_tokens**: Sum of input, output, and reasoning tokens.
- **cache_hit_rate_pct**: Prompt cache efficiency calculated as `cache_read_tokens / (input_tokens + cache_read_tokens) * 100`.
- **peak_context_tokens**: Maximum context window token count reached during the invocation.
- **cost_usd**: Financial cost in USD derived from published model pricing tiers.
- **compactions**: Number of context window compaction events executed.

#### 2. Tool Usage
Aggregated tool execution telemetry across ticket workspaces:
- **ticket_id / ticket_title / ticket_label**: Associated ticket identifier, title, and formatted composite label.
- **tool_name**: Name of the invoked tool (e.g., `read`, `edit`, `bash`, `grep`, `glob`).
- **calls**: Total invocation attempts for the tool.
- **errors**: Total failed executions or returned error responses.
- **successes**: Successful tool executions (`calls - errors`).
- **success_rate_pct**: Execution reliability metric computed as `(calls - errors) / calls * 100`.

#### 3. Code & SDLC
Repository code changes and delivery volume per ticket deliverable:
- **ticket_id / ticket_title / ticket_label / ticket_type**: Ticket identifier, title, formatted composite label, and classification.
- **started_at / completed_at / updated_at**: Ticket lifecycle milestone timestamps.
- **date**: Deliverable completion date extracted from `completed_at` (falling back to `started_at`).
- **lead_time_seconds**: Wall-clock calendar duration in seconds from `started_at` to `completed_at`.
- **is_completed**: Boolean flag indicating whether the deliverable is completed.
- **sdlc_iterations**: Number of planner-coder-tester feedback loops executed.
- **files_changed**: Total count of unique files modified.
- **lines_added / lines_deleted / net_lines_added**: Code change volumes (`net_lines_added = lines_added - lines_deleted`).
- **prod_lines_added / test_lines_added**: Breakdown of production vs test code additions.
- **test_to_code_ratio**: Rigor metric defined as `test_lines_added / prod_lines_added`.

---

### Completion Date Semantics
All date-based slicing and temporal attributes anchor on the completion timestamp (`completed_at`). For in-progress tickets where completion timestamp is pending, dates fall back to `started_at` or `updated_at`.

---

### Operational Mechanics & Aggregation Reference
- **Slice Filtering**: Combine multi-select categorical filters, numeric range sliders, and custom Pandas query syntax (e.g., `cost_usd > 0.05 and turns >= 5`).
- **Dynamic Grouping**: Group filtered records by any categorical dimensions and apply standard aggregation functions (`Sum`, `Mean`, `Median`, `Min`, `Max`, `Count`).
- **Visualization Modes**: Render multi-dimensional charts (Vertical/Horizontal Bars, Lines, Area, Scatter/Bubble Plots, Donut Charts) or tabular grids with CSV export.
"""
        )

    if not ledgers:
        st.info("No ticket ledgers available.")
        return

    ticket_title_map = {
        ledger.ticket_id: ledger.ticket_title for ledger in ledgers if ledger.ticket_id
    }

    # 1. Dataset Selection
    source_choice = st.radio(
        "Select Data Source",
        options=["Invocations (Event Level)", "Tool Usage", "Code & SDLC"],
        horizontal=True,
    )

    match source_choice:
        case "Invocations (Event Level)":
            raw_df = build_invocations_dataframe(ledgers)
        case "Tool Usage":
            raw_df = build_tools_dataframe(ledgers)
        case "Code & SDLC":
            raw_df = build_code_dataframe(ledgers)
        case _:
            raw_df = pd.DataFrame()

    if raw_df.empty:
        st.info(f"No records available for dataset: {source_choice}.")
        return

    # 2. Interactive Filtering
    with st.expander("Filter and Slice Dataset", expanded=True):
        filtered_df = raw_df.copy()

        if source_choice == "Invocations (Event Level)":
            f1, f2, f3, f4 = st.columns(4)
            ticket_opts = sorted(str(x) for x in raw_df["ticket_id"].unique())
            sel_tickets = f1.multiselect(
                "Ticket ID",
                options=ticket_opts,
                format_func=lambda tid: format_ticket_label(
                    tid, ticket_title_map.get(tid)
                ),
            )

            role_opts = sorted(str(x) for x in raw_df["role"].unique())
            sel_roles = f2.multiselect("Role", options=role_opts)

            model_opts = sorted(str(x) for x in raw_df["model"].unique())
            sel_models = f3.multiselect("Model", options=model_opts)

            user_opts = sorted(str(x) for x in raw_df["user"].unique())
            sel_users = f4.multiselect("User", options=user_opts)

            if sel_tickets:
                filtered_df = filtered_df[filtered_df["ticket_id"].isin(sel_tickets)]
            if sel_roles:
                filtered_df = filtered_df[filtered_df["role"].isin(sel_roles)]
            if sel_models:
                filtered_df = filtered_df[filtered_df["model"].isin(sel_models)]
            if sel_users:
                filtered_df = filtered_df[filtered_df["user"].isin(sel_users)]

            r1, r2, r3, r4 = st.columns(4)
            min_c = float(raw_df["cost_usd"].min())
            max_c = float(raw_df["cost_usd"].max())
            if min_c < max_c:
                c_range = r1.slider("Cost USD Range", min_c, max_c, (min_c, max_c))
                filtered_df = filtered_df[
                    (filtered_df["cost_usd"] >= c_range[0])
                    & (filtered_df["cost_usd"] <= c_range[1])
                ]

            min_t = int(raw_df["turns"].min())
            max_t = int(raw_df["turns"].max())
            if min_t < max_t:
                t_range = r2.slider("Turns Range", min_t, max_t, (min_t, max_t))
                filtered_df = filtered_df[
                    (filtered_df["turns"] >= t_range[0])
                    & (filtered_df["turns"] <= t_range[1])
                ]

            min_d = int(raw_df["active_duration_seconds"].min())
            max_d = int(raw_df["active_duration_seconds"].max())
            if min_d < max_d:
                d_range = r3.slider("Active Duration (s)", min_d, max_d, (min_d, max_d))
                filtered_df = filtered_df[
                    (filtered_df["active_duration_seconds"] >= d_range[0])
                    & (filtered_df["active_duration_seconds"] <= d_range[1])
                ]

            min_p = int(raw_df["peak_context_tokens"].min())
            max_p = int(raw_df["peak_context_tokens"].max())
            if min_p < max_p:
                p_range = r4.slider("Peak Context Tokens", min_p, max_p, (min_p, max_p))
                filtered_df = filtered_df[
                    (filtered_df["peak_context_tokens"] >= p_range[0])
                    & (filtered_df["peak_context_tokens"] <= p_range[1])
                ]

        elif source_choice == "Tool Usage":
            f1, f2 = st.columns(2)
            ticket_opts = sorted(str(x) for x in raw_df["ticket_id"].unique())
            sel_tickets = f1.multiselect(
                "Ticket ID",
                options=ticket_opts,
                format_func=lambda tid: format_ticket_label(
                    tid, ticket_title_map.get(tid)
                ),
            )

            tool_opts = sorted(str(x) for x in raw_df["tool_name"].unique())
            sel_tools = f2.multiselect("Tool Name", options=tool_opts)

            if sel_tickets:
                filtered_df = filtered_df[filtered_df["ticket_id"].isin(sel_tickets)]
            if sel_tools:
                filtered_df = filtered_df[filtered_df["tool_name"].isin(sel_tools)]

            r1, r2 = st.columns(2)
            min_calls = int(raw_df["calls"].min())
            max_calls = int(raw_df["calls"].max())
            if min_calls < max_calls:
                c_range = r1.slider(
                    "Calls Range", min_calls, max_calls, (min_calls, max_calls)
                )
                filtered_df = filtered_df[
                    (filtered_df["calls"] >= c_range[0])
                    & (filtered_df["calls"] <= c_range[1])
                ]

            min_errors = int(raw_df["errors"].min())
            max_errors = int(raw_df["errors"].max())
            if min_errors < max_errors:
                e_range = r2.slider(
                    "Errors Range", min_errors, max_errors, (min_errors, max_errors)
                )
                filtered_df = filtered_df[
                    (filtered_df["errors"] >= e_range[0])
                    & (filtered_df["errors"] <= e_range[1])
                ]

        else:
            f1, f2, f3 = st.columns(3)
            ticket_opts = sorted(str(x) for x in raw_df["ticket_id"].unique())
            sel_tickets = f1.multiselect(
                "Ticket ID",
                options=ticket_opts,
                format_func=lambda tid: format_ticket_label(
                    tid, ticket_title_map.get(tid)
                ),
            )

            type_opts = sorted(str(x) for x in raw_df["ticket_type"].unique())
            sel_types = f2.multiselect("Ticket Type", options=type_opts)

            completion_filter = f3.selectbox(
                "Completion Status",
                options=["All Tickets", "Completed Only", "In Progress Only"],
                index=0,
            )

            if sel_tickets:
                filtered_df = filtered_df[filtered_df["ticket_id"].isin(sel_tickets)]
            if sel_types:
                filtered_df = filtered_df[filtered_df["ticket_type"].isin(sel_types)]
            if completion_filter == "Completed Only":
                filtered_df = filtered_df[filtered_df["is_completed"].astype(bool)]
            elif completion_filter == "In Progress Only":
                filtered_df = filtered_df[~filtered_df["is_completed"].astype(bool)]

            r1, r2 = st.columns(2)
            min_lines = int(raw_df["lines_added"].min())
            max_lines = int(raw_df["lines_added"].max())
            if min_lines < max_lines:
                l_range = r1.slider(
                    "Lines Added Range", min_lines, max_lines, (min_lines, max_lines)
                )
                filtered_df = filtered_df[
                    (filtered_df["lines_added"] >= l_range[0])
                    & (filtered_df["lines_added"] <= l_range[1])
                ]

            min_iters = int(raw_df["sdlc_iterations"].min())
            max_iters = int(raw_df["sdlc_iterations"].max())
            if min_iters < max_iters:
                i_range = r2.slider(
                    "SDLC Iterations Range",
                    min_iters,
                    max_iters,
                    (min_iters, max_iters),
                )
                filtered_df = filtered_df[
                    (filtered_df["sdlc_iterations"] >= i_range[0])
                    & (filtered_df["sdlc_iterations"] <= i_range[1])
                ]

        query_str = st.text_input(
            "Custom Query Expression (Pandas query syntax)",
            value="",
            help="Example: cost_usd > 0.05 and turns >= 5",
        )
        if query_str.strip():
            try:
                filtered_df = filtered_df.query(query_str.strip())
            except Exception as err:  # noqa: BLE001
                st.warning(f"Invalid query expression: {err}. Showing unqueried slice.")

    # 3. Slice KPI Metric Cards
    if filtered_df.empty:
        st.info("No records match the current filter criteria.")
        return

    st.markdown("**Slice Summary KPI**")
    if source_choice == "Invocations (Event Level)":
        k1, k2, k3, k4, k5 = st.columns(5)
        k1.metric("Filtered Records", format_number(len(filtered_df)))
        k2.metric(
            "Total Cost",
            format_currency(float(filtered_df["cost_usd"].sum())),
        )
        k3.metric(
            "Total Tokens",
            format_tokens(int(filtered_df["total_tokens"].sum())),
        )
        k4.metric("Avg Turns", f"{filtered_df['turns'].mean():.1f}")
        dur_secs = int(filtered_df["active_duration_seconds"].sum())
        d_h, d_rem = divmod(dur_secs, 3600)
        d_m, d_s = divmod(d_rem, 60)
        d_fmt = f"{d_h}h {d_m}m" if d_h > 0 else f"{d_m}m {d_s}s"
        k5.metric("Active Duration", d_fmt)
    elif source_choice == "Tool Usage":
        k1, k2, k3, k4 = st.columns(4)
        tot_calls = int(filtered_df["calls"].sum())
        tot_errors = int(filtered_df["errors"].sum())
        succ_rate = (
            round((tot_calls - tot_errors) / tot_calls * 100, 1)
            if tot_calls > 0
            else 100.0
        )
        k1.metric("Filtered Records", format_number(len(filtered_df)))
        k2.metric("Total Calls", format_number(tot_calls))
        k3.metric("Total Errors", format_number(tot_errors))
        k4.metric("Success Rate", format_percentage(succ_rate))
    else:
        k1, k2, k3, k4, k5 = st.columns(5)
        k1.metric("Filtered Tickets", format_number(len(filtered_df)))
        k2.metric(
            "Files Changed",
            format_number(int(filtered_df["files_changed"].sum())),
        )
        k3.metric(
            "Lines Added",
            f"+{format_number(int(filtered_df['lines_added'].sum()))}",
        )
        k4.metric(
            "Lines Deleted",
            f"-{format_number(int(filtered_df['lines_deleted'].sum()))}",
        )
        k5.metric(
            "Avg SDLC Iterations",
            f"{filtered_df['sdlc_iterations'].mean():.1f}",
        )

    st.divider()

    # 4. Dynamic Aggregation & Grouping
    st.subheader("Data Aggregation and Grouping")
    view_mode = st.radio(
        "View Mode",
        options=["Raw Granular Records", "Aggregated Summary"],
        horizontal=True,
    )

    display_df = filtered_df
    if view_mode == "Aggregated Summary":
        cat_cols = [
            c
            for c in filtered_df.columns
            if not pd.api.types.is_numeric_dtype(filtered_df[c])
            or pd.api.types.is_bool_dtype(filtered_df[c])
        ]
        num_cols = [
            c
            for c in filtered_df.columns
            if pd.api.types.is_numeric_dtype(filtered_df[c])
            and not pd.api.types.is_bool_dtype(filtered_df[c])
        ]

        g_col1, g_col2, g_col3 = st.columns([2, 2, 1])
        default_groups = [cat_cols[0]] if cat_cols else []
        sel_groups = g_col1.multiselect(
            "Group By Dimensions", options=cat_cols, default=default_groups
        )

        default_metrics = [num_cols[0]] if num_cols else []
        sel_metrics = g_col2.multiselect(
            "Metrics to Aggregate", options=num_cols, default=default_metrics
        )

        agg_func_name = g_col3.selectbox(
            "Aggregation Function",
            options=["Sum", "Mean", "Median", "Min", "Max", "Count"],
            index=0,
        )

        agg_map = {
            "Sum": "sum",
            "Mean": "mean",
            "Median": "median",
            "Min": "min",
            "Max": "max",
            "Count": "count",
        }

        if sel_groups and sel_metrics:
            display_df = filtered_df.groupby(sel_groups, as_index=False)[
                sel_metrics
            ].agg(agg_map[agg_func_name])
            if agg_func_name in ("Mean", "Median"):
                display_df[sel_metrics] = display_df[sel_metrics].round(2)
        else:
            st.info(
                "Select at least one grouping dimension and one metric to aggregate."
            )
            display_df = filtered_df

    st.divider()

    # 5. Flexible Presentation Styles
    st.subheader("Visualization")
    avail_nums = [
        c
        for c in display_df.columns
        if pd.api.types.is_numeric_dtype(display_df[c])
        and not pd.api.types.is_bool_dtype(display_df[c])
    ]
    avail_all = list(display_df.columns)

    if not avail_nums:
        st.info("No numeric fields available to plot.")
    else:
        v_col1, v_col2 = st.columns([1, 3])
        style = v_col1.selectbox(
            "Presentation Style",
            options=[
                "Vertical Bar Chart",
                "Horizontal Bar Chart",
                "Line Chart",
                "Area Chart",
                "Scatter / Bubble Plot",
                "Donut Chart",
                "Data Table Only",
            ],
            index=0,
        )

        if style != "Data Table Only":
            with v_col2:
                c1, c2, c3, c4 = st.columns(4)
                x_col = c1.selectbox(
                    "X-Axis Field",
                    options=avail_all,
                    index=0,
                )
                y_col = c2.selectbox(
                    "Y-Axis Metric",
                    options=avail_nums,
                    index=0,
                )
                avail_splits = ["(None)"] + [
                    c for c in avail_all if c != y_col and c != x_col
                ]
                color_col = c3.selectbox(
                    "Split / Color Dimension",
                    options=avail_splits,
                    index=0,
                )

                size_col = "(None)"
                stack_bars = False
                if style == "Scatter / Bubble Plot":
                    size_col = c4.selectbox(
                        "Bubble Size Metric",
                        options=["(None)"] + avail_nums,
                        index=0,
                    )
                elif "Bar Chart" in style and color_col != "(None)":
                    stack_bars = c4.checkbox("Stack Bars", value=False)

            # Chart construction
            tooltips = [x_col, y_col]
            if color_col != "(None)" and color_col not in tooltips:
                tooltips.append(color_col)
            if size_col != "(None)" and size_col not in tooltips:
                tooltips.append(size_col)
            for t_col in ("ticket_id", "ticket_title", "ticket_label"):
                if t_col in display_df.columns and t_col not in tooltips:
                    tooltips.append(t_col)

            try:
                chart: alt.Chart | None = None
                if style == "Vertical Bar Chart":
                    x_enc = alt.X(f"{x_col}:N", title=x_col)
                    y_enc = alt.Y(f"{y_col}:Q", title=y_col)
                    if color_col != "(None)":
                        c_enc = alt.Color(f"{color_col}:N", title=color_col)
                        if not stack_bars:
                            chart = (
                                alt.Chart(display_df)
                                .mark_bar()
                                .encode(
                                    x=x_enc,
                                    y=y_enc,
                                    color=c_enc,
                                    xOffset=alt.XOffset(f"{color_col}:N"),
                                    tooltip=tooltips,
                                )
                            )
                        else:
                            chart = (
                                alt.Chart(display_df)
                                .mark_bar()
                                .encode(
                                    x=x_enc,
                                    y=y_enc,
                                    color=c_enc,
                                    tooltip=tooltips,
                                )
                            )
                    else:
                        chart = (
                            alt.Chart(display_df)
                            .mark_bar()
                            .encode(x=x_enc, y=y_enc, tooltip=tooltips)
                        )

                elif style == "Horizontal Bar Chart":
                    x_enc = alt.X(f"{y_col}:Q", title=y_col)
                    y_enc = alt.Y(f"{x_col}:N", title=x_col)
                    if color_col != "(None)":
                        c_enc = alt.Color(f"{color_col}:N", title=color_col)
                        if not stack_bars:
                            chart = (
                                alt.Chart(display_df)
                                .mark_bar()
                                .encode(
                                    x=x_enc,
                                    y=y_enc,
                                    color=c_enc,
                                    yOffset=alt.YOffset(f"{color_col}:N"),
                                    tooltip=tooltips,
                                )
                            )
                        else:
                            chart = (
                                alt.Chart(display_df)
                                .mark_bar()
                                .encode(
                                    x=x_enc,
                                    y=y_enc,
                                    color=c_enc,
                                    tooltip=tooltips,
                                )
                            )
                    else:
                        chart = (
                            alt.Chart(display_df)
                            .mark_bar()
                            .encode(x=x_enc, y=y_enc, tooltip=tooltips)
                        )

                elif style == "Line Chart":
                    x_type = (
                        "Q"
                        if pd.api.types.is_numeric_dtype(display_df[x_col])
                        and not pd.api.types.is_bool_dtype(display_df[x_col])
                        else "N"
                    )
                    x_enc = alt.X(f"{x_col}:{x_type}", title=x_col)
                    y_enc = alt.Y(f"{y_col}:Q", title=y_col)
                    if color_col != "(None)":
                        chart = (
                            alt.Chart(display_df)
                            .mark_line(point=True)
                            .encode(
                                x=x_enc,
                                y=y_enc,
                                color=alt.Color(f"{color_col}:N", title=color_col),
                                tooltip=tooltips,
                            )
                        )
                    else:
                        chart = (
                            alt.Chart(display_df)
                            .mark_line(point=True)
                            .encode(x=x_enc, y=y_enc, tooltip=tooltips)
                        )

                elif style == "Area Chart":
                    x_type = (
                        "Q"
                        if pd.api.types.is_numeric_dtype(display_df[x_col])
                        and not pd.api.types.is_bool_dtype(display_df[x_col])
                        else "N"
                    )
                    x_enc = alt.X(f"{x_col}:{x_type}", title=x_col)
                    y_enc = alt.Y(f"{y_col}:Q", title=y_col)
                    if color_col != "(None)":
                        chart = (
                            alt.Chart(display_df)
                            .mark_area(opacity=0.6)
                            .encode(
                                x=x_enc,
                                y=y_enc,
                                color=alt.Color(f"{color_col}:N", title=color_col),
                                tooltip=tooltips,
                            )
                        )
                    else:
                        chart = (
                            alt.Chart(display_df)
                            .mark_area(opacity=0.6)
                            .encode(x=x_enc, y=y_enc, tooltip=tooltips)
                        )

                elif style == "Scatter / Bubble Plot":
                    x_type = (
                        "Q" if pd.api.types.is_numeric_dtype(display_df[x_col]) else "N"
                    )
                    x_enc = alt.X(f"{x_col}:{x_type}", title=x_col)
                    y_enc = alt.Y(f"{y_col}:Q", title=y_col)
                    kwargs: dict[str, object] = {
                        "x": x_enc,
                        "y": y_enc,
                        "tooltip": tooltips,
                    }
                    if color_col != "(None)":
                        kwargs["color"] = alt.Color(f"{color_col}:N", title=color_col)
                    if size_col != "(None)":
                        kwargs["size"] = alt.Size(f"{size_col}:Q", title=size_col)
                    chart = alt.Chart(display_df).mark_circle().encode(**kwargs)

                elif style == "Donut Chart":
                    chart = (
                        alt.Chart(display_df)
                        .mark_arc(innerRadius=60)
                        .encode(
                            theta=alt.Theta(f"{y_col}:Q", title=y_col),
                            color=alt.Color(f"{x_col}:N", title=x_col),
                            tooltip=tooltips,
                        )
                    )

                if chart is not None:
                    chart = chart.properties(height=380)
                    st.altair_chart(chart, width="stretch")
            except Exception as err:  # noqa: BLE001
                st.error(f"Error constructing chart visualization: {err}")

    st.divider()

    # 6. Table & CSV Export
    st.subheader("Data Table and Export")
    st.dataframe(display_df, width="stretch", hide_index=True)
    csv_bytes = display_df.to_csv(index=False).encode("utf-8")
    st.download_button(
        label="Download CSV",
        data=csv_bytes,
        file_name=f"analytics_studio_{source_choice.lower().replace(' ', '_')}.csv",
        mime="text/csv",
        width="stretch",
    )
