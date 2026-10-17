from __future__ import annotations

import argparse
import json
import subprocess
import sys
from dataclasses import asdict
from pathlib import Path

from agentic_lens.adapters.repository import (
    resolve_workspaces_dir,
    scan_workspaces,
)
from agentic_lens.domain.models import AgenticMetricsLedger
from agentic_lens.presentation.views.formatters import (
    format_markdown_metrics_table,
    update_summary_file,
)


def _serialize_ledger(ledger: AgenticMetricsLedger) -> dict[str, object]:
    """Serialize domain metrics ledger into standard JSON dictionary."""
    return {
        "ticket_id": ledger.ticket_id,
        "ticket_title": ledger.ticket_title,
        "ticket_type": ledger.ticket_type,
        "started_at": ledger.started_at,
        "completed_at": ledger.completed_at,
        "updated_at": ledger.updated_at,
        "workflow_efficiency": asdict(ledger.workflow_efficiency),
        "totals": asdict(ledger.totals),
        "roles": [asdict(r) for r in ledger.roles],
        "models": [asdict(m) for m in ledger.models],
    }


def main() -> None:
    parser = argparse.ArgumentParser(
        prog="agentic-lens",
        description="AgenticLens: Developer Telemetry and Lifecycle Metrics CLI",
    )
    parser.add_argument(
        "--ticket-id",
        type=str,
        default=None,
        help="Target ticket identifier to filter and report on",
    )
    parser.add_argument(
        "--workspaces",
        type=str,
        default=None,
        help="Path to root workspaces directory containing agentic_metrics.json ledgers",
    )
    parser.add_argument(
        "--append",
        type=str,
        default=None,
        help="Path to markdown summary file to idempotently update with metrics table",
    )
    parser.add_argument(
        "--json-output",
        type=str,
        default=None,
        help="Path to write computed metrics ledger as JSON",
    )
    parser.add_argument(
        "--list",
        action="store_true",
        help="List all discovered ticket workspaces and summary metrics",
    )
    parser.add_argument(
        "--ui",
        action="store_true",
        help="Launch the interactive Streamlit dashboard on port 8501",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=8501,
        help="Port to bind Streamlit dashboard when running --ui (default: 8501)",
    )

    args = parser.parse_args()

    if args.ui:
        app_path = Path(__file__).parent / "app.py"
        cmd = [
            sys.executable,
            "-m",
            "streamlit",
            "run",
            str(app_path),
            f"--server.port={args.port}",
            "--server.address=0.0.0.0",
        ]
        print(f"Launching AgenticLens UI on http://localhost:{args.port}...")
        try:
            subprocess.run(cmd, check=True)
        except KeyboardInterrupt:
            print("\nShutting down AgenticLens UI.")
        return

    workspaces_dir = resolve_workspaces_dir(args.workspaces)
    report = scan_workspaces(workspaces_dir)

    if not report.ledgers:
        print(f"No valid agentic_metrics.json ledgers found in {workspaces_dir}")
        sys.exit(1)

    if args.list:
        print(f"Found {len(report.ledgers)} ticket workspaces in {workspaces_dir}:\n")
        header = f"{'Ticket ID':<15} {'Duration':<12} {'Cost':<10} {'Health':<12} {'Iterations':<10} {'Title'}"
        print(header)
        print("-" * len(header))
        for ledger in report.ledgers:
            tid = ledger.ticket_id or "UNKNOWN"
            dur = ledger.workflow_efficiency.active_duration_formatted
            cost = f"${ledger.totals.cost_usd:.2f}"
            health = ledger.workflow_efficiency.context_window_health.status
            iters = str(ledger.workflow_efficiency.sdlc_iterations)
            title = ledger.ticket_title or "N/A"
            print(f"{tid:<15} {dur:<12} {cost:<10} {health:<12} {iters:<10} {title}")
        return

    # Filter target ledger
    target_ledger: AgenticMetricsLedger | None = None
    if args.ticket_id:
        target_ledger = next(
            (l for l in report.ledgers if l.ticket_id == args.ticket_id),
            None,
        )
        if target_ledger is None:
            print(
                f"Error: Ticket '{args.ticket_id}' not found in scanned workspaces ({workspaces_dir})"
            )
            sys.exit(1)
    else:
        # Default to most recently updated ledger
        target_ledger = max(report.ledgers, key=lambda l: l.updated_at)

    table_md = format_markdown_metrics_table(target_ledger)

    if args.append:
        dest_path = Path(args.append)
        update_summary_file(dest_path, table_md)
        print(f"AgenticLens metrics table successfully updated in {dest_path}")

    if args.json_output:
        out_path = Path(args.json_output)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        serialized = _serialize_ledger(target_ledger)
        out_path.write_text(json.dumps(serialized, indent=2) + "\n", encoding="utf-8")
        print(f"AgenticLens computed ledger written to {out_path}")

    if not args.append and not args.json_output:
        print(table_md)


if __name__ == "__main__":
    main()
