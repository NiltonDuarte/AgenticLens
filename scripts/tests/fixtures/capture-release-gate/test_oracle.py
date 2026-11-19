"""Focused regressions for the native release gate's ticket duration oracle."""

import unittest
from collections import Counter

from oracle import Owner, artifact_matches_commit, check_complete, check_held, check_rollup
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
        committed = {"observed_at": "2026-11-19T19:31:01Z", "owners": [
            {"session_id": "root", "checkpoint": "2026-11-19T19:30:59Z"},
        ]}
        reconciliation = {"observed_at": "2026-11-19T19:31:00Z"}
        self.assertTrue(artifact_matches_commit({"updated_at": "2026-11-19T19:31:00.5Z"}, committed, reconciliation))
        self.assertFalse(artifact_matches_commit({"updated_at": "2026-11-19T19:31:01.5Z"}, committed, reconciliation))

    def test_rejects_duplicate_tool_rows_with_matching_name_and_count(self):
        owner = Owner(None, completed=1, model_millis=Counter(fixture=7000), tools=Counter(bash=1))
        ledger = {"ticket_id": TICKET, "ticket_title": TITLE, "ticket_type": "fix", "sessions": [
            {"id": "root", "source": "opencode", "invocations": [{"active_duration_seconds": 7}],
             "tools": [{"name": "bash", "calls": 1, "errors": 0}] * 2},
        ]}
        summary = "- **Active Agent Duration:** 0m 7s (recorded duration)"
        with self.assertRaisesRegex(RuntimeError, "published duplicate tool rows"):
            check_complete((ledger, summary), "opencode", {"root": owner}, "root", "unrelated")


if __name__ == "__main__":
    unittest.main()
