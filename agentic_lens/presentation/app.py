from __future__ import annotations

from pathlib import Path

import streamlit as st

from agentic_lens.adapters.repository import resolve_workspaces_dir, scan_workspaces
from agentic_lens.domain.aggregations import calculate_fleet_summary
from agentic_lens.presentation.views.analytics_studio import render_analytics_studio
from agentic_lens.presentation.views.comparisons import render_comparisons
from agentic_lens.presentation.views.fleet_overview import render_fleet_overview
from agentic_lens.presentation.views.metric_progression import render_metric_progression
from agentic_lens.presentation.views.ticket_deep_dive import render_ticket_deep_dive

st.set_page_config(
    page_title="AgenticLens",
    layout="wide",
    initial_sidebar_state="expanded",
)


def main() -> None:
    st.title("AgenticLens")
    st.caption("Agentic Development Telemetry and Lifecycle Metrics Visualizer")

    default_dir = resolve_workspaces_dir()

    with st.sidebar:
        st.header("Configuration")
        workspace_input = st.text_input(
            "Workspaces Directory",
            value=str(default_dir),
            help="Path to root workspaces directory containing agentic_metrics.json ledgers.",
        )
        if st.button("Rescan Workspaces", width="stretch"):
            st.rerun()

        st.divider()
        st.header("Navigation")
        view_choice = st.radio(
            "Select Dashboard View",
            options=[
                "Fleet Overview",
                "Metric Progression",
                "Cross-Ticket Comparisons",
                "Ticket Deep Dive",
                "Analytics Studio",
            ],
            index=0,
        )

    target_path = Path(workspace_input.strip())
    report = scan_workspaces(target_path)

    if report.errors:
        with st.expander(f"Scan Warnings ({len(report.errors)} issues detected)"):
            for err in report.errors:
                st.warning(err)

    if not report.ledgers:
        st.info(
            f"No valid agentic_metrics.json ledgers found in: {target_path}. "
            "Please check the workspaces directory path or run metric extraction."
        )
        return

    match view_choice:
        case "Fleet Overview":
            summary = calculate_fleet_summary(report.ledgers)
            render_fleet_overview(report.ledgers, summary)
        case "Metric Progression":
            render_metric_progression(report.ledgers)
        case "Cross-Ticket Comparisons":
            render_comparisons(report.ledgers)
        case "Ticket Deep Dive":
            render_ticket_deep_dive(report.ledgers, workspaces_dir=target_path)
        case "Analytics Studio":
            render_analytics_studio(report.ledgers)
        case _:
            st.error(f"Unknown dashboard view: {view_choice}")


if __name__ == "__main__":
    main()
