from __future__ import annotations

import sqlite3
import sys
import tempfile
from pathlib import Path
from unittest.mock import patch

# Ensure repo root is on sys.path for scripts module import
sys.path.insert(0, str(Path(__file__).resolve().parents[3]))

from scripts.extract_agentic_metrics import (
    AgenticFactLedger,
    CodeFacts,
    InvocationFact,
    InvocationTokens,
    extract_session_metrics,
    format_markdown_table,
    infer_ticket_title,
    load_existing_ledger,
    parse_numstat_output,
    resolve_active_root_session,
)


def test_parse_numstat_output() -> None:
    raw = (
        "100\t10\tsrc/domain/storage/models.py\n"
        "200\t20\ttests/unit/domain/storage/test_models.py\n"
        "-\t-\tbinary.dat\n"
    )
    parsed = parse_numstat_output(raw)
    assert len(parsed) == 2
    assert parsed[0] == (100, 10, "src/domain/storage/models.py")
    assert parsed[1] == (200, 20, "tests/unit/domain/storage/test_models.py")


def test_infer_ticket_title_contract_fallback() -> None:
    with tempfile.TemporaryDirectory() as tmpdir:
        ws = Path(tmpdir) / "workspaces" / "DPS-99999"
        ws.mkdir(parents=True)
        summary = ws / "SUMMARY.md"
        summary.write_text("# DPS-99999 - Summary\n\nSome body\n")
        contract = ws / "CONTRACT.md"
        contract.write_text("# CONTRACT: DPS-99999 Feature Name Here\n\nSection\n")

        with patch("scripts.extract_agentic_metrics.Path") as mock_path:
            # Let Path("workspaces/DPS-99999/...") resolve to our temp dir
            def side_effect(p: str | Path) -> Path:
                if str(p).startswith("workspaces/DPS-99999/"):
                    rel = str(p).replace("workspaces/DPS-99999/", "")
                    return ws / rel
                return Path(p)

            mock_path.side_effect = side_effect
            title = infer_ticket_title("DPS-99999")
            assert title == "Feature Name Here"


def test_format_markdown_table_with_general_subagent() -> None:
    ledger = AgenticFactLedger(
        ticket_id="DPS-TEST",
        ticket_title="Test Feature",
        ticket_type="feature",
        started_at="2026-10-15T00:00:00Z",
        completed_at="2026-10-15T00:30:00Z",
        updated_at="2026-10-15T01:00:00Z",
        code=CodeFacts(
            lines_added=500,
            lines_deleted=10,
            test_lines_added=300,
            test_lines_deleted=5,
            prod_lines_added=200,
            prod_lines_deleted=5,
            files_changed=8,
            sdlc_iterations=1,
        ),
        invocations=[
            InvocationFact(
                id="orch_1",
                parent_id=None,
                role="orchestrator",
                model="gemini-3.8-flash",
                user="alice",
                started_at="2026-10-15T00:00:00Z",
                completed_at="2026-10-15T00:30:00Z",
                active_duration_seconds=1800,
                turns=50,
                tokens=InvocationTokens(
                    input=100000,
                    output=5000,
                    cache_read=50000,
                    reasoning=2000,
                    peak_context=120000,
                ),
                cost_usd=1.50,
                compactions=0,
            ),
            InvocationFact(
                id="sub_1",
                parent_id="orch_1",
                role="general",
                model="gemini-3.8-flash",
                user="alice",
                started_at="2026-10-15T00:10:00Z",
                completed_at="2026-10-15T00:20:00Z",
                active_duration_seconds=600,
                turns=20,
                tokens=InvocationTokens(
                    input=80000,
                    output=4000,
                    cache_read=40000,
                    reasoning=1000,
                    peak_context=95000,
                ),
                cost_usd=0.75,
                compactions=0,
            ),
        ],
        tools=[],
    )

    md = format_markdown_table(ledger)
    assert "- **Role Effort Distribution (Subagents):** General 100.0%" in md
    assert (
        "- **Subagent Turn Velocity & Thrashing:** 20.0 avg turns/run (peak: 20 turns) | General 20.0"
        in md
    )
    assert "Subagents peak 95k (General)" in md
    assert "- **Test-to-Production Code Ratio:** 1.5:1 (+300 test lines / +200 prod lines)" in md
    assert "- **Code Delivered:** +500 / -10 lines across 8 code files (excluding docs)" in md


def test_resolve_active_root_session_does_not_cross_contaminate() -> None:
    conn = sqlite3.connect(":memory:")
    conn.execute(
        "CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, title TEXT, time_updated INTEGER)"
    )
    conn.execute(
        "CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER)"
    )
    # Insert another user's unrelated session
    conn.execute(
        "INSERT INTO session VALUES ('unrelated_ses', NULL, 'Unrelated ticket DPS-12345', 1000)"
    )
    conn.execute("INSERT INTO message VALUES ('msg_1', 'unrelated_ses', 1000)")

    # When querying for DPS-11173 which does not exist in DB, must return None
    assert resolve_active_root_session(conn, ticket_id="DPS-11173") is None

    # When querying with no ticket and no nonce, fallback to latest message is allowed
    assert resolve_active_root_session(conn) == "unrelated_ses"


def test_ledger_completed_at_to_dict_and_load() -> None:
    ledger = AgenticFactLedger(
        ticket_id="DPS-9999",
        ticket_title="Title",
        ticket_type="chore",
        started_at="2026-10-15T00:00:00Z",
        completed_at="2026-10-15T02:00:00Z",
        updated_at="2026-10-15T02:30:00Z",
        code=CodeFacts(0, 0, 0, 0, 0, 0, 0, 1),
        invocations=[],
        tools=[],
    )
    d = ledger.to_dict()
    assert d["started_at"] == "2026-10-15T00:00:00Z"
    assert d["completed_at"] == "2026-10-15T02:00:00Z"

    with tempfile.TemporaryDirectory() as tmpdir:
        ledger_path = Path(tmpdir) / "agentic_metrics.json"
        import json
        ledger_path.write_text(json.dumps(d))

        (
            cached_ticket,
            _title,
            _type,
            cached_started_at,
            cached_completed_at,
            _threads,
            _code,
            _invs,
            _tools,
        ) = load_existing_ledger(ledger_path)

        assert cached_ticket == "DPS-9999"
        assert cached_started_at == "2026-10-15T00:00:00Z"
        assert cached_completed_at == "2026-10-15T02:00:00Z"


def test_extract_session_metrics_computes_completed_at_when_missing() -> None:
    # Test that when completed_at is not in the json ledger, it is derived from invocations
    legacy_json = {
        "ticket_id": "DPS-1234",
        "ticket_title": "Legacy",
        "ticket_type": "bug",
        "started_at": "2026-10-01T10:00:00Z",
        "updated_at": "2026-10-01T12:00:00Z",
        "code": {
            "lines_added": 10,
            "lines_deleted": 0,
            "test_lines_added": 5,
            "test_lines_deleted": 0,
            "prod_lines_added": 5,
            "prod_lines_deleted": 0,
            "files_changed": 1,
            "sdlc_iterations": 1,
        },
        "invocations": [
            {
                "id": "ses_1",
                "parent_id": None,
                "role": "orchestrator",
                "model": "model_1",
                "user": "user_1",
                "started_at": "2026-10-01T10:05:00Z",
                "completed_at": "2026-10-01T11:45:00Z",
                "active_duration_seconds": 100,
                "turns": 5,
                "tokens": {
                    "input": 10,
                    "output": 10,
                    "cache_read": 0,
                    "reasoning": 0,
                    "peak_context": 10,
                },
                "cost_usd": 0.01,
                "compactions": 0,
            }
        ],
        "tools": [],
    }
    with tempfile.TemporaryDirectory() as tmpdir:
        ledger_path = Path(tmpdir) / "agentic_metrics.json"
        import json
        ledger_path.write_text(json.dumps(legacy_json))

        with patch("scripts.extract_agentic_metrics.Path.home") as mock_home:
            mock_home.return_value = Path(tmpdir)
            ledger = extract_session_metrics(
                ticket_id="DPS-1234",
                existing_ledger_path=ledger_path,
            )
            assert ledger.completed_at == "2026-10-01T11:45:00Z"
