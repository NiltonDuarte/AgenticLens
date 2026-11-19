"""Focused regressions for native release gate assertions."""

import unittest
from collections import Counter

from gate import birth_commit, first_completion_commit
from oracle import (
    Owner, artifact_matches_commit, capture_id, check_birth, check_complete, check_held, check_pending_identity,
    check_rollup, snapshot_matches_pending_record, unix_time,
)
from provider import TICKET, TITLE


def snapshot(rendered):
    return (
        {"sessions": [
            {"id": "root", "invocations": [{"active_duration_seconds": 4}]},
            {"id": "child", "invocations": [{"active_duration_seconds": 3}]},
        ]},
        f"- **Active Agent Duration:** {rendered} (recorded duration)",
    )


class TicketDurationOracleTest(unittest.TestCase):
    def test_rejects_altered_rendered_total_with_correct_provisional_suffix(self):
        matched = check_rollup(snapshot("0m 7s"), native_measured_seconds=7)
        self.assertEqual(matched["ticket_active_seconds"], 7)
        self.assertEqual(matched["ticket_measured_seconds"], 7)
        self.assertEqual(matched["ticket_provisional_seconds"], 0)
        with self.assertRaisesRegex(RuntimeError, "rendered active duration"):
            check_rollup(snapshot("0m 0s"), native_measured_seconds=7)

    def test_rejects_closed_native_child_with_provisional_duration(self):
        ledger = {"updated_at": "1970-01-01T00:00:15Z", "sessions": [
            {"id": "root", "invocations": [{"active_duration_seconds": 4, "provisional_seconds": 4}]},
            {"id": "child", "invocations": [{"active_duration_seconds": 3, "provisional_seconds": 1}]},
        ]}
        native = {
            "root": Owner(None, open=1, open_intervals={"turn": ("fixture", 10000, 10000)}),
            "child": Owner("root", completed=1, latest_completion=13, model_millis=Counter(fixture=3000)),
        }
        reconciliation = {"observed_at": "1970-01-01T00:00:14.5Z", "reconciliation": {"started_at": "1970-01-01T00:00:14Z"}}
        committed = {"observed_at": "1970-01-01T00:00:15.5Z", "owners": [
            {"session_id": "root"}, {"session_id": "child", "checkpoint": "1970-01-01T00:00:13Z"},
        ]}
        summary = "- **Active Agent Duration:** 0m 7s (recorded duration including 5s provisional)"
        with self.assertRaisesRegex(RuntimeError, "child: held duration differs"):
            check_held((ledger, summary), native, "root", committed, reconciliation)

    def test_rejects_ledger_from_a_later_commit_with_older_diagnostic(self):
        checkpoint = "2026-11-19T19:30:59Z"
        owners = [{"session_id": "root", "checkpoint": checkpoint}]
        first = {"sequence": 27, "outcome": "committed", "observed_at": "2026-11-19T19:31:01Z", "owners": owners}
        later = {"sequence": 29, "outcome": "committed", "observed_at": "2026-11-19T19:31:03Z", "owners": owners}
        first_reconciliation = {"observed_at": "2026-11-19T19:31:00Z"}
        later_reconciliation = {"observed_at": "2026-11-19T19:31:02Z"}
        selected = first_completion_commit([first, later], "root", unix_time(checkpoint))
        self.assertEqual(selected["sequence"], 27)
        self.assertTrue(artifact_matches_commit({"updated_at": "2026-11-19T19:31:00.5Z"}, selected, first_reconciliation))
        later_ledger = {"updated_at": "2026-11-19T19:31:02.5Z"}
        self.assertTrue(artifact_matches_commit(later_ledger, later, later_reconciliation))
        self.assertFalse(artifact_matches_commit(later_ledger, selected, first_reconciliation))

    def test_rejects_duplicate_tool_rows_with_matching_name_and_count(self):
        owner = Owner(None, completed=1, model_millis=Counter(fixture=7000), tools=Counter(bash=1))
        ledger = {"ticket_id": TICKET, "ticket_title": TITLE, "ticket_type": "fix", "sessions": [
            {"id": "root", "source": "opencode", "invocations": [{"active_duration_seconds": 7}],
             "tools": [{"name": "bash", "calls": 1, "errors": 0}] * 2},
        ]}
        summary = "- **Active Agent Duration:** 0m 7s (recorded duration)"
        with self.assertRaisesRegex(RuntimeError, "published duplicate tool rows"):
            check_complete((ledger, summary), "opencode", {"root": owner}, "root", "unrelated")


class PendingSnapshotOracleTest(unittest.TestCase):
    def test_rejects_empty_first_canonical_ledger_even_with_matching_commit_timestamps(self):
        started = "2026-11-19T19:31:00Z"
        ledger = {"ticket_id": TICKET, "ticket_title": TITLE, "ticket_type": "fix",
                  "updated_at": "2026-11-19T19:31:00.5Z", "sessions": []}
        reconciliation = {"observed_at": started, "reconciliation": {"started_at": started}}
        first = {"sequence": 27, "outcome": "committed", "observed_at": "2026-11-19T19:31:01Z", "owners": []}
        native = {"root": Owner(None), "implementation": Owner("root"), "verify": Owner("root"),
                  "nested": Owner("implementation")}
        self.assertTrue(artifact_matches_commit(ledger, first, reconciliation))
        summary = "- **Active Agent Duration:** 0m 0s (recorded duration)"
        with self.assertRaisesRegex(RuntimeError, "Expected native IDs"):
            check_birth((ledger, summary), "codex", native, "root", "unrelated", first, reconciliation, {})

    def test_later_corrected_snapshot_cannot_stand_in_for_first_commit(self):
        first = {"sequence": 27, "outcome": "committed"}
        later = {"sequence": 29, "outcome": "committed"}
        self.assertEqual(birth_commit([first]), first)
        with self.assertRaisesRegex(RuntimeError, "superseded before its contents could be verified"):
            birth_commit([first, later])

    def test_binding_preserves_capture_identity_and_requires_exact_metadata(self):
        unbound = {
            "version": 1, "capture_id": capture_id("codex", "root"),
            "source": "codex", "root_session_id": "root", "state": "awaiting_ticket",
        }
        check_pending_identity(unbound, "codex", "root", "awaiting_ticket")
        bound = dict(unbound, state="awaiting_destination", ticket_id=TICKET, ticket_title=TITLE, ticket_type="fix")
        check_pending_identity(bound, "codex", "root", "awaiting_destination")
        with self.assertRaisesRegex(RuntimeError, "capture identity"):
            check_pending_identity(dict(bound, capture_id=capture_id("opencode", "root")),
                                   "codex", "root", "awaiting_destination")
        with self.assertRaisesRegex(RuntimeError, "metadata"):
            check_pending_identity(dict(unbound, ticket_type="fix"), "codex", "root", "awaiting_ticket")

    def test_rejects_pending_snapshot_from_a_later_source_cutoff(self):
        first = {"outcome": "awaiting_ticket", "reconciliation": {"started_at": "2026-11-19T19:31:00Z"}}
        later = {"outcome": "awaiting_ticket", "reconciliation": {"started_at": "2026-11-19T19:31:02Z"}}
        snapshot = {"state": "awaiting_ticket", "source_cutoff": later["reconciliation"]["started_at"]}
        self.assertTrue(snapshot_matches_pending_record(snapshot, later))
        self.assertFalse(snapshot_matches_pending_record(snapshot, first))


if __name__ == "__main__":
    unittest.main()
