from __future__ import annotations

import json
import logging
from pathlib import Path

from pydantic import ValidationError

from agentic_lens.adapters.schemas import AgenticFactLedgerSchema
from agentic_lens.domain.aggregations import derive_metrics_ledger
from agentic_lens.domain.models import (
    AgenticFactLedger,
    AgenticMetricsLedger,
    ScanReport,
    TicketArtifacts,
)

logger = logging.getLogger(__name__)


def resolve_workspaces_dir(override_path: Path | str | None = None) -> Path:
    """Resolve workspaces directory from argument, environment, config, or parent lookup."""
    if override_path is not None:
        return Path(override_path)

    # 1. Container volume mount path
    container_dir = Path("/app/workspaces")
    if container_dir.is_dir():
        return container_dir

    # 2. Configured workspace dir in .agenticlens/config.json
    for directory in [Path.cwd(), *Path.cwd().parents]:
        config_file = directory / ".agenticlens" / "config.json"
        if config_file.is_file():
            try:
                data = json.loads(config_file.read_text(encoding="utf-8"))
                ws_dir = data.get("workspaces_dir")
                if ws_dir:
                    candidate = directory / ws_dir
                    if candidate.is_dir():
                        return candidate
            except (json.JSONDecodeError, OSError) as err:
                logger.debug("Failed reading .agenticlens/config.json at %s: %s", config_file, err)

    # 3. Current working directory
    cwd_dir = Path.cwd() / "workspaces"
    if cwd_dir.is_dir():
        return cwd_dir

    # 4. Parent traversal (e.g. running from apps/agentic-lens)
    for parent in Path.cwd().parents:
        candidate = parent / "workspaces"
        if candidate.is_dir():
            return candidate

    return Path("workspaces")


def scan_workspaces(base_path: Path | str | None = None) -> ScanReport:
    """Scan workspaces directory for all agentic_metrics.json ledgers."""
    target_dir = resolve_workspaces_dir(base_path)
    logger.info("Scanning for agentic metrics in %s", target_dir)

    ledgers: list[AgenticMetricsLedger] = []
    errors: list[str] = []

    if not target_dir.is_dir():
        msg = f"Workspaces directory not found: {target_dir}"
        logger.warning(msg)
        return ScanReport(ledgers=(), errors=(msg,))

    for metric_file in sorted(target_dir.glob("**/agentic_metrics.json")):
        try:
            content = metric_file.read_text(encoding="utf-8")
            fact_schema = AgenticFactLedgerSchema.model_validate_json(content)
            fact_ledger = fact_schema.to_domain()

            # Deduce ticket_id from parent directory if omitted
            if not fact_ledger.ticket_id:
                parent_name = metric_file.parent.name
                if parent_name != "workspaces":
                    fact_ledger = AgenticFactLedger(
                        ticket_id=parent_name,
                        ticket_title=fact_ledger.ticket_title,
                        ticket_type=fact_ledger.ticket_type,
                        started_at=fact_ledger.started_at,
                        completed_at=fact_ledger.completed_at,
                        updated_at=fact_ledger.updated_at,
                        code=fact_ledger.code,
                        invocations=fact_ledger.invocations,
                        tools=fact_ledger.tools,
                    )

            ledger = derive_metrics_ledger(fact_ledger)
            ledgers.append(ledger)
            logger.info(
                "Successfully loaded agentic metrics for %s from %s",
                ledger.ticket_id,
                metric_file,
            )
        except (ValidationError, OSError, ValueError) as err:
            err_msg = f"Failed to load {metric_file}: {err}"
            logger.warning(err_msg)
            errors.append(err_msg)

    return ScanReport(ledgers=tuple(ledgers), errors=tuple(errors))


def read_ticket_artifacts(
    ticket_id: str, base_path: Path | str | None = None
) -> TicketArtifacts:
    """Read CONTRACT.md and SUMMARY.md artifacts for a ticket workspace safely."""
    target_dir = resolve_workspaces_dir(base_path).resolve()
    ticket_dir = (target_dir / ticket_id).resolve()

    # Prevent path traversal outside the workspaces root
    if not ticket_dir.is_relative_to(target_dir) or ticket_dir == target_dir:
        logger.warning(
            "Invalid ticket identifier or path traversal attempt: %s", ticket_id
        )
        return TicketArtifacts(ticket_id=ticket_id)

    if not ticket_dir.is_dir():
        logger.warning("Ticket directory not found: %s", ticket_dir)
        return TicketArtifacts(ticket_id=ticket_id)

    contract_file = ticket_dir / "CONTRACT.md"
    contract_content: str | None = None
    if contract_file.is_file():
        try:
            contract_content = contract_file.read_text(encoding="utf-8")
            logger.debug("Successfully read CONTRACT.md for %s", ticket_id)
        except (OSError, UnicodeDecodeError) as err:
            logger.warning("Failed to read CONTRACT.md for %s: %s", ticket_id, err)

    summary_file = ticket_dir / "SUMMARY.md"
    summary_content: str | None = None
    if summary_file.is_file():
        try:
            summary_content = summary_file.read_text(encoding="utf-8")
            logger.debug("Successfully read SUMMARY.md for %s", ticket_id)
        except (OSError, UnicodeDecodeError) as err:
            logger.warning("Failed to read SUMMARY.md for %s: %s", ticket_id, err)

    return TicketArtifacts(
        ticket_id=ticket_id,
        contract_content=contract_content,
        summary_content=summary_content,
    )
