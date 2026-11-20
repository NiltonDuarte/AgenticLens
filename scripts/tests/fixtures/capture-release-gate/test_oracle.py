"""Focused regressions for native release gate assertions."""

import unittest
from collections import Counter
from shlex import join

from gate import birth_commit, first_completion_commit
from oracle import (
    Owner, artifact_matches_commit, capture_id, check_birth, check_complete, check_held, check_pending_identity,
    check_rollup, snapshot_matches_pending_record, unix_time,
)
from provider import FixtureModel, TICKET, TITLE, request_stage


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


class NativeCaptureControlOracleTest(unittest.TestCase):
    def setUp(self):
        self.model = FixtureModel("/fixture/bin/manager", "/fixture/project", "/fixture/destination")
        self.model.binding_engine = "/fixture/bin/engine"
        self.key = "a" * 64
        self.store = "/fixture/codex stores"
        self.model.capture_control = "verified-root", self.key, 7, self.store
        self.receipt = f"{'b' * 64}.7.{'c' * 64}"

    def test_bound_stage_does_not_reuse_unbound_tool_history(self):
        self.assertEqual(request_stage([
            {"role": "user", "content": "AL96:DISCARD"},
            {"role": "user", "content": "AL96:DISCARD_BOUND"},
        ]), "DISCARD_BOUND")

    def test_codex_control_rejects_a_stale_or_different_root_invitation(self):
        def invitation(root, receipt):
            command = [self.model.binding_engine, "capture", "submit-binding", "--repo", self.model.project,
                       "--store", self.store, "--session-id", root, "--request", receipt,
                       "--capture-id", self.key, "--capture-action", "discard"]
            return "To discard this capture when the user directs it, execute: " + join(command)

        stale = invitation("verified-root", self.receipt.replace(".7.", ".6."))
        outsider = invitation("another-root", self.receipt)
        with self.assertRaisesRegex(ValueError, "expected capture and revision"):
            self.model.native_control(stale + "\n" + outsider, "discard", True)
        current = invitation("verified-root", self.receipt)
        command, _ = self.model.native_control("\n".join((stale, current, outsider)), "discard", True)
        self.assertEqual(command[6:11], [self.store, "--session-id", "verified-root", "--request", self.receipt])

    def test_opencode_control_rejects_a_stale_or_malformed_receipt(self):
        invitation = f"To restore this capture when the user directs it, call agenticlens_restore with capture ID {self.key} and request "
        stale = invitation + self.receipt.replace(".7.", ".6.")
        with self.assertRaisesRegex(ValueError, "expected capture and revision"):
            self.model.native_control(stale + "\n" + invitation + self.receipt + "0", "restore", False)
        _, receipt = self.model.native_control(stale + "\n" + invitation + self.receipt, "restore", False)
        self.assertEqual(receipt, self.receipt)


class PendingSnapshotOracleTest(unittest.TestCase):
    def test_first_ledger_ignores_open_activity_after_its_native_cutoff(self):
        cutoff = "1970-01-01T00:00:11.713790Z"
        reconciliation = {"observed_at": "1970-01-01T00:00:12.150Z", "reconciliation": {"started_at": cutoff}}
        first = {"observed_at": "1970-01-01T00:00:12.280Z", "owners": [{"session_id": "root"}]}
        owner = Owner(None, completed=1, latest_completion=9.521, model_millis=Counter(fixture=2000),
                      open=1, open_intervals={"stream": ("fixture", 10000, 12080)},
                      usage=Counter(input=23, output=11), tools=Counter(bash=1))
        root = {
            "id": "root", "source": "opencode",
            "invocations": [{"active_duration_seconds": 2, "tokens": dict(owner.usage)}],
            "tools": [{"name": "bash", "calls": 1, "errors": 0}],
        }
        settled = {"sessions": [root]}
        ledger = {**settled, "ticket_id": TICKET, "ticket_title": TITLE, "ticket_type": "fix",
                  "updated_at": "1970-01-01T00:00:12.200Z"}
        snapshot = ledger, "- **Active Agent Duration:** 0m 2s (recorded duration)"
        native = {"root": owner}
        evidence = check_birth(snapshot, "opencode", native, "root", "unrelated", first, reconciliation, settled)
        self.assertEqual(evidence["ticket_measured_seconds"], 2)
        self.assertEqual(evidence["ticket_provisional_seconds"], 0)
        with self.assertRaisesRegex(RuntimeError, "no native evidence at the held cutoff"):
            check_held(snapshot, native, "root", first, reconciliation)

        premature = {**root, "invocations": [{
            **root["invocations"][0], "active_duration_seconds": 4, "provisional_seconds": 2,
        }]}
        inflated = {**ledger, "sessions": [premature]}, "- **Active Agent Duration:** 0m 4s (recorded duration including 2s provisional)"
        with self.assertRaisesRegex(RuntimeError, "held duration differs"):
            check_birth(inflated, "opencode", native, "root", "unrelated", first, reconciliation, settled)
        owner.latest_completion = 12.081
        with self.assertRaisesRegex(RuntimeError, "native completion is later than the held cutoff"):
            check_birth(snapshot, "opencode", native, "root", "unrelated", first, reconciliation, settled)

        # The completed source message confirms it was in flight at the
        # cutoff, even though the next streaming part had not begun.
        owner.completed = 2
        owner.model_millis["fixture"] += 2081
        owner.completed_intervals = {
            "earlier": ("fixture", 7521, 9521),
            "repair": ("fixture", 10000, 12081),
        }
        evidence = check_birth(inflated, "opencode", native, "root", "unrelated", first, reconciliation, settled)
        self.assertEqual(evidence["ticket_measured_seconds"], 2)
        self.assertEqual(evidence["ticket_provisional_seconds"], 2)
        self.assertEqual(evidence["native_intervals"]["root"]["native_completed_intervals"], 1)
        prematurely_measured = {**premature, "invocations": [{
            **root["invocations"][0], "active_duration_seconds": 4,
        }]}
        wrong = {**ledger, "sessions": [prematurely_measured]}, "- **Active Agent Duration:** 0m 4s (recorded duration)"
        with self.assertRaisesRegex(RuntimeError, "held duration differs"):
            check_birth(wrong, "opencode", native, "root", "unrelated", first, reconciliation, settled)

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
