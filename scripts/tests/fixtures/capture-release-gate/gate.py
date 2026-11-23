"""Exercise the installed integrations and report the packaged release gate."""

import argparse
from concurrent.futures import ThreadPoolExecutor
from contextlib import contextmanager, nullcontext
import fcntl
from http.client import HTTPConnection, HTTPException
import json
import os
from pathlib import Path
import shutil
import signal
import socket
import sqlite3
import subprocess
import sys
import tempfile
import time

from oracle import (
    artifact_matches_commit, capture_id, check_birth, check_complete, check_complete_pending, check_completed_facts,
    check_held, check_held_pending,
    check_pending_identity, json_lines, native_id, provisional, published, snapshot_matches_pending_record,
    source_owners, tree,
    unix_time,
)
from project import Project, require, require_package
from provider import FINAL_INPUT_TOKENS, FINAL_OUTPUT_TOKENS, FixtureModel, TICKET, TITLE, manual_generate

FIRST_SECONDS = 60
SETTLE_SECONDS = 30
POLL_SECONDS = 0.2


class FixtureProcess:
    def __init__(self, command, **options):
        # Keep a separately owned group leader until cleanup, even when the
        # native client exits while one of its descendants is still active.
        self.anchor = subprocess.Popen(
            (sys.executable, "-B", "-c", "import sys; sys.stdin.buffer.read()"),
            stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, process_group=0,
        )
        try:
            self.process = subprocess.Popen(command, **options, process_group=self.anchor.pid)
        except BaseException:
            self.anchor.stdin.close()
            self.anchor.wait(timeout=5)
            raise

    def poll(self):
        return self.process.poll()

    def wait(self, timeout):
        return self.process.wait(timeout=timeout)

    def stop(self):
        def send(sig):
            try:
                os.killpg(self.anchor.pid, sig)
            except ProcessLookupError:
                pass

        # The group's PID is reserved by our unreaped anchor during both
        # signals. Do not gate cleanup on the native client's exit status.
        send(signal.SIGTERM)
        try:
            self.process.wait(timeout=1)
        except subprocess.TimeoutExpired:
            pass
        finally:
            send(signal.SIGKILL)
            self.anchor.stdin.close()
            self.anchor.wait(timeout=5)
        self.process.wait(timeout=5)


def event_until(deadline, predicate, message, process=None):
    while time.monotonic() < deadline:
        value = predicate()
        if value:
            return value
        if process is not None and process.poll() is not None:
            raise RuntimeError(f"{message}. Native process exited with code {process.returncode}")
        time.sleep(POLL_SECONDS)
    raise RuntimeError(message)


def launch(project, stage, session=None):
    log = project.root / f"{stage.lower()}.jsonl"
    errors = project.root / f"{stage.lower()}.stderr"
    with log.open("w") as stdout, errors.open("w") as stderr:
        launch_process = FixtureProcess if project.completion_feedback else subprocess.Popen
        process = launch_process(project.native_command(stage, session), cwd=project.checkout, env=project.env,
                                 stdout=stdout, stderr=stderr, stdin=subprocess.DEVNULL)
    return process, log


def stop(process):
    if isinstance(process, FixtureProcess):
        process.stop()
        return
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait(timeout=5)


def wait_success(process, stage, timeout=45):
    require(process.wait(timeout=timeout) == 0, f"Native {stage} failed, see fixture logs")


def status(project):
    return json.loads(project.capture("diagnostics", "--ticket-id", TICKET))


def verified_registration(project, root):
    for owner in status(project).get("registrations", []):
        if owner["identity"]["session_id"] == root and owner.get("binding") == "explicit_agent_response":
            return owner
    return None


def verified_binding(project, root):
    registration = verified_registration(project, root)
    return registration and registration.get("binding_accepted_at")


def publication_diagnostics(rows):
    return sorted((row for row in rows if row.get("ticket_id") == TICKET and row["phase"] == "publication"),
                  key=lambda row: row["sequence"])


def observe_publication_pair(project, commitment, rows):
    reconciliation = max(
        (row for row in rows if row.get("ticket_id") == TICKET and row["phase"] == "reconciliation"
         and row["sequence"] < commitment["sequence"]),
        key=lambda row: row["sequence"], default=None,
    )
    snapshot = published(project)
    if reconciliation is None or snapshot is None or not artifact_matches_commit(snapshot[0], commitment, reconciliation):
        return None
    current = publication_diagnostics(project.diagnostic_history())
    if not current or current[-1]["sequence"] != commitment["sequence"]:
        return None
    require(commitment["destination"] == str(project.ledger) and commitment["summary"] == str(project.summary),
            "Publication diagnostic points to a different destination or summary")
    return snapshot, commitment, reconciliation


def assert_complete_coverage(commitment, reconciliation, owners):
    coverage = reconciliation.get("reconciliation", {})
    require(coverage.get("attempt_outcome") == "completed" and not coverage.get("pending_children")
            and not coverage.get("unresolved_parent_links") and not coverage.get("child_discovery_pending"),
            f"Publication has incomplete reconciliation: {coverage.get('attempt_outcome')}, "
            f"pending children={coverage.get('pending_children')}, "
            f"unresolved parent links={coverage.get('unresolved_parent_links')}")
    published = {owner["session_id"]: owner["refresh_state"] for owner in commitment.get("owners", [])}
    require(set(published) == set(owners) and all(value == "refreshed" for value in published.values()),
            "Committed publication contains unresolved or stale owner coverage")


def checkpoint_seen(pass_record, root, instant):
    return any(owner["session_id"] == root and owner.get("checkpoint") and
               unix_time(owner["checkpoint"]) >= instant - 0.001 for owner in pass_record.get("owners", []))


def first_completion_commit(passes, root, instant):
    return next((record for record in passes
                 if record["outcome"] == "committed" and checkpoint_seen(record, root, instant)), None)


def first_eligible_completion_publication(rows, root, instant, source_read_end, previous_sequence):
    reconciliation = None
    for row in rows:
        if row.get("ticket_id") != TICKET:
            continue
        if row["phase"] == "reconciliation":
            reconciliation = row
        elif (row["phase"] == "publication" and row["outcome"] == "committed" and
              row["sequence"] > previous_sequence and reconciliation is not None):
            boundary = reconciliation.get("reconciliation", {})
            started = boundary.get("started_at")
            if started and unix_time(started) >= source_read_end:
                return row, reconciliation, "source_visible_before_reconciliation"
            # The worker can see terminal evidence during the observation
            # window before our independent poll completes. Check such a
            # commit, but do not use its self-reported checkpoint to skip a
            # later commit that was independently known to be eligible.
            cutoff = boundary.get("cutoff")
            if cutoff and unix_time(cutoff) >= instant and checkpoint_seen(row, root, instant):
                return row, reconciliation, "overlapping_native_observation"
    return None


def wait_first_snapshot(project, binding, process):
    begin = unix_time(binding)

    def observed_first_commit():
        rows = project.diagnostic_history()
        commits = [item for item in publication_diagnostics(rows) if item["outcome"] == "committed"]
        if not commits:
            return None
        first = commits[0]
        require(first["destination"] == str(project.ledger) and first["summary"] == str(project.summary),
                "First committed publication points to a different destination or summary")
        # The birth commit may have been followed by a normal refresh before
        # binding was observed here. Use its durable timestamp and validate
        # the current pair against its own commit.
        return first if observe_publication_pair(project, commits[-1], rows) else None

    deadline = time.monotonic() + FIRST_SECONDS - (time.time() - begin)
    commitment = event_until(deadline, observed_first_commit,
                             "No first committed publication after the destination became eligible", process)
    delay = unix_time(commitment["observed_at"]) - begin
    require(0 <= delay <= FIRST_SECONDS, f"First commit outside publication budget: {delay:.3f}s")
    return commitment["observed_at"], delay


def birth_commit(commits):
    require(len(commits) <= 1, "The first canonical ledger was superseded before its contents could be verified")
    return commits[0] if commits else None


def wait_birth_snapshot(project, model, root, path, ready_at, process, check_index):
    def observed_birth():
        check_index("first canonical publication")
        pending_exists = path.is_file()
        rows = project.diagnostic_history()
        first = birth_commit([row for row in publication_diagnostics(rows) if row["outcome"] == "committed"])
        require(pending_exists or first, "Pending snapshot disappeared before durable canonical publication")
        if first is None or not model.repair_holding.is_set():
            return None
        pair = observe_publication_pair(project, first, rows)
        if pair is None:
            return None
        native = source_owners(project)
        if root not in native or not native[root].open_intervals:
            return None
        return pair, tree(native, root)

    deadline = time.monotonic() + FIRST_SECONDS - (time.time() - ready_at)
    pair, native = event_until(deadline, observed_birth,
                               "The first canonical ledger could not be observed during the held native repair", process)
    _, commitment, _ = pair
    delay = unix_time(commitment["observed_at"]) - ready_at
    require(0 <= delay <= FIRST_SECONDS, f"First commit outside publication budget: {delay:.3f}s")
    return pair, native, delay


def wait_native_hold(project, model, root, process, deadline):
    def inspect_native_hold():
        if not model.holding.is_set():
            return None
        owners = source_owners(project)
        if root not in owners:
            return None
        expected = tree(owners, root)
        if len(expected) != 4 or not expected[root].open_intervals or any(
            owner.open or owner.completed == 0 for session_id, owner in expected.items() if session_id != root
        ):
            return None
        assert_activity(expected, project.source, root, model.nested_supported, bound=project.prepared)
        return expected, time.time()

    return event_until(deadline, inspect_native_hold,
                       "Native source did not show a held root and all three completed descendants", process)


def wait_held_snapshot(project, model, root, process):
    deadline = time.monotonic() + FIRST_SECONDS
    expected, observed_at = wait_native_hold(project, model, root, process, deadline)
    def inspect_held_pair():
        rows = project.diagnostic_history()
        publication = publication_diagnostics(rows)
        if not publication or publication[-1]["outcome"] != "committed":
            return None
        pair = observe_publication_pair(project, publication[-1], rows)
        if pair is None:
            return None
        snapshot, commitment, reconciliation = pair
        if unix_time(reconciliation["reconciliation"]["started_at"]) < observed_at:
            return None
        if provisional(snapshot[0]).get(root, 0) <= 0:
            return None
        assert_complete_coverage(commitment, reconciliation, expected)
        return check_held(snapshot, expected, root, commitment, reconciliation), commitment

    return event_until(deadline, inspect_held_pair,
                       "A held native turn never published source-verified provisional duration", process)


def wait_terminal_snapshot(project, root, target, since_sequence, earlier=None):
    deadline = time.monotonic() + SETTLE_SECONDS - (time.time() - target)

    def observed_terminal_commit():
        rows = project.diagnostic_history()
        passes = [record for record in publication_diagnostics(rows) if record["sequence"] > since_sequence]
        first = first_completion_commit(passes, root, target)
        if first is None:
            return None
        require(passes[-1]["sequence"] == first["sequence"],
                "Could not associate the current ledger with the first committed publication that saw native completion")
        pair = observe_publication_pair(project, first, rows)
        if pair is None:
            return None
        snapshot, commitment, reconciliation = pair
        ledger, _ = snapshot
        if earlier and ledger["updated_at"] <= earlier:
            return None
        delay = unix_time(commitment["observed_at"]) - target
        require(0 <= delay <= SETTLE_SECONDS, f"Completion-to-commit delay outside budget: {delay:.3f}s")
        return snapshot, commitment, reconciliation, delay

    return event_until(deadline, observed_terminal_commit, "Automatic publication did not settle with native completion in time")


def assert_activity(owners, source, root, nested_supported, bound=True):
    root_tools = owners[root].tools
    require(nested_supported is True, "Reviewed native version did not offer a nested delegation tool")
    require(len(owners) == 4, f"Expected three native descendants, observed {len(owners) - 1}")
    children = [owner for owner in owners.values() if owner.parent == root]
    require(len(children) == 2, "Native root did not delegate distinct implementation and verification children")
    if source == "codex":
        require(root_tools["spawn_agent"] == 2 and root_tools["wait_agent"] == 2,
                "Codex root did not make two spawn and wait calls")
    else:
        require(root_tools["task"] == 2 and root_tools["agenticlens_bind"] == int(bound),
                "OpenCode root's native task or binding tool counts were unexpected")
    require(sum(owner.tools["spawn_agent" if source == "codex" else "task"] for owner in children) == 1,
            "Implementation child did not exercise the available native nested delegation")
    require(all(owner.tools["exec_command" if source == "codex" else "bash"] > 0 for owner in children),
            "Native children did not perform the fixture edit and verification")


def completed_source(project, root, after):
    owners = source_owners(project)
    native = owners.get(root)
    return owners if native and native.open == 0 and native.latest_completion > after else None


def pending_diagnostics(project, root, path, state):
    return [
        row for row in project.diagnostic_history(None)
        if row["phase"] == "pending_snapshot" and row["outcome"] == state
        and row.get("destination") == str(path)
        and [(owner["source"], owner["session_id"]) for owner in row["owners"]] == [(project.source, root)]
    ]


def no_canonical_output(project):
    require(not project.summary.exists() and not any(project.checkout.rglob("agentic_metrics.json")),
            "A canonical artifact appeared while publication had no usable destination")
    require(not any(row["outcome"] == "committed" for row in publication_diagnostics(project.diagnostic_history())),
            "Worker reported a canonical commit before the destination was available")


def observe_pending(project, root, path, state, eligible, first=False):
    no_canonical_output(project)
    try:
        body = path.read_bytes()
    except FileNotFoundError:
        return None
    snapshot = json.loads(body)
    eligible_records = [row for row in pending_diagnostics(project, root, path, state) if eligible(row)]
    if not eligible_records:
        return None
    if first:
        record = eligible_records[0]
        if not snapshot_matches_pending_record(snapshot, record):
            require(not any(snapshot_matches_pending_record(snapshot, row) for row in eligible_records[1:]),
                    "A later pending snapshot replaced the first snapshot that saw native completion")
            return None
    else:
        record = next((row for row in reversed(eligible_records) if snapshot_matches_pending_record(snapshot, row)), None)
        if record is None:
            return None
    if body != path.read_bytes():
        return None
    check_pending_identity(snapshot, project.source, root, state)
    return snapshot, record


def wait_held_pending(project, model, root, path, unrelated, process):
    deadline = time.monotonic() + FIRST_SECONDS
    expected, observed_at = wait_native_hold(project, model, root, process, deadline)

    def inspect():
        observed = observe_pending(project, root, path, "awaiting_ticket",
                                   lambda row: unix_time(row["reconciliation"]["started_at"]) >= observed_at)
        if observed and provisional(observed[0]).get(root, 0) > 0:
            snapshot, diagnostic = observed
            return check_held_pending(snapshot, project.source, expected, root, unrelated, diagnostic), diagnostic
        return None

    return event_until(deadline, inspect, "No pending snapshot with positive native provisional duration", process)


def wait_completed_pending(project, root, path, state, target, since_sequence, first=False):
    return event_until(time.monotonic() + SETTLE_SECONDS - (time.time() - target),
                       lambda: observe_pending(project, root, path, state,
                                               lambda row: row["sequence"] > since_sequence
                                               and checkpoint_seen(row, root, target), first),
                       f"Native completion did not settle while {state}")


def exact_children(snapshot, root):
    return {row["id"]: row for row in snapshot["sessions"] if row["id"] != root}


def native_turn(project, root, stage, processes):
    turn, log = launch(project, stage, root)
    processes.append(turn)
    wait_success(turn, stage)
    require(native_id(log) == root, f"{stage} ran in a different native session")


def capture_status(project, root, key, ticket):
    control = json.loads(project.capture("status", "--capture-id", key))
    require((control.get("capture_id"), control.get("root_session_id"), control.get("source"),
             control.get("ticket_id", "")) == (key, root, project.source, ticket),
            "Local capture status changed its identity or binding")
    return control


def publication_boundary(project):
    return max((row["sequence"] for row in publication_diagnostics(project.diagnostic_history())), default=0)


@contextmanager
def canonical_discard_baseline(project):
    directory = os.open(project.workspace, os.O_RDONLY | os.O_DIRECTORY)
    try:
        def lock_directory():
            try:
                fcntl.flock(directory, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                return False
            return True

        # Establish the byte baseline after any already admitted active write.
        event_until(time.monotonic() + SETTLE_SECONDS, lock_directory,
                    "Could not acquire the canonical directory before native discard")
        yield project.ledger.read_bytes(), project.summary.read_bytes()
    finally:
        os.close(directory)


def assert_canonical_unchanged(project, baseline, stage):
    ledger, summary = baseline
    require(project.ledger.read_bytes() == ledger, f"{stage} changed the accepted canonical ledger")
    require(project.summary.read_bytes() == summary, f"{stage} changed the accepted canonical summary")


def discard_cycle(project, model, root, key, path, processes, check_index, before, bound=False):
    suffix, ticket = ("_BOUND", TICKET) if bound else ("", "")
    tool = "exec_command" if project.source == "codex" else "bash"
    active = capture_status(project, root, key, ticket)
    require(active["state"] == "active", "Capture was not active before native discard")
    model.capture_control = root, key, active["revision"], str(project.codex_home)

    def discarded_status(stage):
        require(discard_turn.poll() in (None, 0), "Native discard failed, see fixture logs")
        control = capture_status(project, root, key, ticket)
        if bound:
            assert_canonical_unchanged(project, retained, stage)
        if control["state"] == "discarded":
            require(control["revision"] > active["revision"], "Native discard did not advance the capture revision")
            return control
        return None

    with canonical_discard_baseline(project) if bound else nullcontext() as retained:
        discard_turn, discard_log = launch(project, "DISCARD" + suffix, root)
        processes.append(discard_turn)
        event_until(time.monotonic() + SETTLE_SECONDS, lambda: discarded_status("Persisting discard"),
                    "Native discard did not persist in local capture status")

    # The directory is released so native cleanup can acquire it. Keep checking
    # the files until cleanup has settled and the native turn has finished.
    def discarded_cleanup_finished():
        control = discarded_status("Acknowledging discard cleanup")
        if control and not control["cleanup_pending"]:
            require(not path.exists(), "Discard acknowledged cleanup with its pending snapshot still present")
            if discard_turn.poll() == 0:
                return control
        return None

    discarded = event_until(time.monotonic() + SETTLE_SECONDS, discarded_cleanup_finished,
                            "Native discard did not settle and finish its turn")
    require(native_id(discard_log) == root, "Native discard ran in a different session")
    check_index("native discard" + suffix)
    if not bound:
        no_canonical_output(project)

    before_during = tree(source_owners(project), root)
    require(before_during[root].latest_completion > before[root].latest_completion,
            "The native discard turn added no completed activity")
    request_before = status(project)["publication"]["Generation"]
    pending_state = "awaiting_destination" if bound else "awaiting_ticket"
    pending_before = max((row["sequence"] for row in pending_diagnostics(project, root, path, pending_state)), default=0)
    publication_before = publication_boundary(project)

    def assert_suppressed_files():
        require(not path.exists(), "The worker recreated a discarded pending snapshot")
        if bound:
            assert_canonical_unchanged(project, retained, "Activity during discard")
        else:
            require(not project.ledger.exists() and not project.summary.exists(),
                    "Activity during discard produced a canonical artifact")

    during_started = time.time()
    during_turn, during_log = launch(project, "DURING_DISCARD" + suffix, root)
    processes.append(during_turn)

    def during_turn_finished():
        assert_suppressed_files()
        return during_turn.poll() is not None

    event_until(time.monotonic() + 45, during_turn_finished, "Native work during discard did not complete")
    wait_success(during_turn, "work during discard" + suffix)
    require(native_id(during_log) == root, "Native work during discard ran in a different session")
    during = tree(source_owners(project), root)
    require(during[root].latest_completion > before_during[root].latest_completion and during[root].open == 0
            and during[root].tools[tool] == before_during[root].tools[tool] + 1,
            "Native work during discard did not complete an extra verified tool call")

    def worker_observed_discarded_activity():
        control = capture_status(project, root, key, ticket)
        require(control == discarded, "Resumed native activity changed the discarded disposition")
        assert_suppressed_files()
        require(not any(row["sequence"] > pending_before for row in pending_diagnostics(project, root, path, pending_state)),
                "The worker reported a refresh of the discarded pending snapshot")
        if bound:
            require(not any(row["outcome"] == "committed" and row["sequence"] > publication_before
                            for row in publication_diagnostics(project.diagnostic_history())),
                    "The worker reported a canonical commit while the capture was discarded")
        else:
            no_canonical_output(project)
        check_index("resumed activity during discard" + suffix)
        publication = status(project)["publication"]
        if (publication["Generation"] > request_before and publication["Acknowledged"] == publication["Generation"]
                and publication["RequestedAt"] >= int(during_started * 1000)):
            return publication["Acknowledged"]
        return None

    observed_generation = event_until(time.monotonic() + SETTLE_SECONDS, worker_observed_discarded_activity,
                                      "The worker did not process a native wake from activity during discard")
    model.capture_control = root, key, discarded["revision"], str(project.codex_home)
    native_turn(project, root, "RESTORE" + suffix, processes)
    restored_source = tree(source_owners(project), root)
    require(restored_source[root].latest_completion > during[root].latest_completion and restored_source[root].open == 0,
            "The native restore turn added no completed activity")

    def restored_status():
        control = capture_status(project, root, key, ticket)
        return control if (control["revision"] > discarded["revision"] and control["state"] == "active"
                           and not control["coverage_pending"] and not control["cleanup_pending"]) else None

    restored = event_until(time.monotonic() + SETTLE_SECONDS, restored_status,
                           "Native restore did not finish reconciliation for the exact capture")
    check_index("native restore" + suffix)
    return restored_source, {
        "discard_revision": discarded["revision"], "restored_revision": restored["revision"],
        "processed_discarded_generation": observed_generation,
        "pending_remained_absent_during_discard": True,
        "native_completion_before_discard": before[root].latest_completion,
        "native_completion_during_discard": during[root].latest_completion,
        "native_completion_after_restore": restored_source[root].latest_completion,
        "native_tool_calls_before_discard": sum(before[root].tools.values()),
        "native_tool_calls_during_discard": sum(during[root].tools.values()),
        "native_tool_calls_after_restore": sum(restored_source[root].tools.values()),
        "pending_publication_boundary": pending_before,
        "canonical_publication_boundary": publication_before,
    }


def resume_and_check(project, root, unrelated, earlier, earlier_commit, earlier_source, processes, evidence):
    continued, continued_log = launch(project, "RESUME", root)
    processes.append(continued)
    wait_success(continued, "normal resume")
    require(native_id(continued_log) == root, "Resume created a different native root")
    final_source = tree(source_owners(project), root)
    latest_completion = final_source[root].latest_completion
    require(latest_completion > earlier_source[root].latest_completion
            and final_source[root].usage["input"] > earlier_source[root].usage["input"],
            "Native resume did not add completed activity and usage")
    final, final_commit, final_reconciliation, continued_delay = wait_terminal_snapshot(
        project, root, latest_completion, earlier_commit["sequence"], earlier[0]["updated_at"],
    )
    evidence["after_resume"] = check_complete(final, project.source, final_source, root, unrelated)
    assert_complete_coverage(final_commit, final_reconciliation, final_source)
    require(exact_children(final[0], root) == exact_children(earlier[0], root),
            "Refreshing the root changed earlier child session records")
    evidence["resumed_completion_to_commit_seconds"] = round(continued_delay, 3)


def exercise_pending(project, model, evidence, unrelated, processes, index_before):
    def check_index(stage):
        require(project.index_entries() == index_before, f"Git index entries or blob IDs changed during {stage}")

    check_index("unrelated native activity")
    no_canonical_output(project)
    require(not project.workspace.exists(), "The late destination already exists")
    process, log = launch(project, "UNBOUND")
    processes.append(process)
    root = event_until(time.monotonic() + FIRST_SECONDS, lambda: native_id(log),
                       "Unbound native root ID did not become observable", process)
    key = capture_id(project.source, root)
    path = project.pending / (key + ".json")
    before, held_record = wait_held_pending(project, model, root, path, unrelated, process)
    check_index("open unbound capture")
    require(verified_registration(project, root) is None, "Unbound root acquired a ticket before any response")
    require(project.run("git", "status", "--porcelain", "--untracked-files=all", "--", path) ==
            f"?? {path.relative_to(project.checkout)}", "Pending capture is ignored or staged in Git")
    evidence.update({
        "pending_capture_id": key, "pending_file": str(path.relative_to(project.checkout)),
        "before_completion": before, "pending_visible_unstaged": True,
    })
    model.release.set()
    native = event_until(
        time.monotonic() + 45, lambda: completed_source(project, root, unix_time(held_record["observed_at"])),
        "Unbound native session did not record completion",
    )
    wait_success(process, "unbound development")
    expected = tree(native, root)
    require(unrelated in native, "Unrelated native smoke disappeared from the source")
    assert_activity(expected, project.source, root, model.nested_supported, bound=False)
    require((project.checkout / "fixture.py").read_text() == "def answer():\n    return 42\n",
            "Implementation child did not make the required code edit")
    target = expected[root].latest_completion
    unbound, unbound_record = wait_completed_pending(
        project, root, path, "awaiting_ticket", target, held_record["sequence"], first=True,
    )
    check_index("settled unbound capture")
    delay = unix_time(unbound_record["observed_at"]) - target
    require(0 <= delay <= SETTLE_SECONDS, f"Unbound pending settlement outside budget: {delay:.3f}s")
    evidence.update({
        "native_tree": {session_id: owner.parent for session_id, owner in expected.items()},
        "nested_delegation_offered": model.nested_supported, "native_completion_epoch": target,
        "first_completion_pending": unbound_record["observed_at"], "completion_to_pending_seconds": round(delay, 3),
        "awaiting_ticket": check_complete_pending(unbound, project.source, expected, root, unrelated),
    })

    restored_source, unbound_discard = discard_cycle(project, model, root, key, path, processes, check_index, expected)
    target = restored_source[root].latest_completion
    restored, restored_record = wait_completed_pending(
        project, root, path, "awaiting_ticket", target, unbound_discard["pending_publication_boundary"],
    )
    require(exact_children(restored, root) == exact_children(unbound, root),
            "Unbound restoration changed previously accepted child facts")
    unbound_discard["restored_snapshot"] = check_complete_pending(restored, project.source, restored_source, root, unrelated)
    evidence["unbound_discard_restore"] = unbound_discard

    native_turn(project, root, "BIND", processes)
    registration = event_until(time.monotonic() + FIRST_SECONDS, lambda: verified_registration(project, root),
                               "Resumed native root did not verify its binding")
    require((registration["ticket_id"], registration["ticket_title"], registration["ticket_type"]) == (TICKET, TITLE, "fix")
            and not registration.get("publication_destination"), "Native binding metadata or destination differs from its response")
    bound_source = tree(source_owners(project), root)
    bound_target = bound_source[root].latest_completion
    require(bound_target > target, "Native binding turn has no completion")
    bound, bound_record = wait_completed_pending(
        project, root, path, "awaiting_destination", bound_target, restored_record["sequence"],
    )
    check_index("binding without a destination")
    require(exact_children(bound, root) == exact_children(unbound, root),
            "Binding modified the children captured before binding")
    evidence["binding_accepted_at"] = registration["binding_accepted_at"]
    evidence["pending_binding_observed_at"] = bound_record["observed_at"]
    evidence["awaiting_destination"] = check_complete_pending(bound, project.source, bound_source, root, unrelated)

    native_turn(project, root, "DESTINATION", processes)

    def destination_registered():
        row = verified_registration(project, root)
        return row and row.get("publication_destination") == str(project.workspace)

    event_until(time.monotonic() + FIRST_SECONDS, destination_registered,
                "The native destination challenge did not persist the exact directory hint")
    hint_source = tree(source_owners(project), root)
    rejected, rejected_record = wait_completed_pending(
        project, root, path, "awaiting_destination", hint_source[root].latest_completion, bound_record["sequence"],
    )

    def blocked_destination():
        require(path.is_file() and not project.workspace.exists(), "The unavailable destination changed pending capture")
        no_canonical_output(project)
        return next((row for row in publication_diagnostics(project.diagnostic_history())
                     if row["outcome"] == "failed" and row["sequence"] > rejected_record["sequence"]
                     and str(project.workspace) in row.get("detail", "")), None)

    blocked = event_until(time.monotonic() + FIRST_SECONDS, blocked_destination,
                          "The unavailable native destination did not produce a publication failure")
    check_index("unavailable destination")
    check_complete_pending(rejected, project.source, hint_source, root, unrelated)
    evidence["unavailable_destination_diagnostic"] = blocked["detail"]
    evidence["unavailable_destination_observed_at"] = blocked["observed_at"]

    # Repair makes the already-indicated path eligible during a normal native
    # turn. No second hint, operator preparation, or direct worker invocation.
    retry_started = time.time()
    repair, repair_log = launch(project, "RETRY", root)
    processes.append(repair)
    pair, birth_source, first_delay = wait_birth_snapshot(project, model, root, path, retry_started, repair, check_index)
    birth, first_commit, first_reconciliation = pair
    require(native_id(repair_log) == root and project.workspace.is_dir(),
            "The resumed root did not create its indicated directory")
    assert_complete_coverage(first_commit, first_reconciliation, birth_source)
    evidence["first_publication"] = check_birth(
        birth, project.source, birth_source, root, unrelated, first_commit, first_reconciliation, restored,
    )

    def promoted():
        check_index("pending cleanup")
        row = next(iter(pending_diagnostics(project, root, path, "promoted")), None)
        if row and not path.exists():
            require(first_commit["sequence"] < row["sequence"]
                    and unix_time(first_commit["observed_at"]) <= unix_time(row["observed_at"]),
                    "Pending cleanup does not follow its durable canonical commit")
            return row
        return None

    promotion = event_until(time.monotonic() + SETTLE_SECONDS, promoted,
                            "The pending snapshot was not removed after durable publication", repair)
    model.repair_release.set()
    wait_success(repair, "native destination repair")
    published_source = tree(source_owners(project), root)
    published_target = published_source[root].latest_completion
    require(published_target > unix_time(first_commit["observed_at"]), "Native repair did not complete after first publication")
    terminal, terminal_commit, reconciliation, delay = wait_terminal_snapshot(
        project, root, published_target, first_commit["sequence"], birth[0]["updated_at"],
    )
    check_index("completed repair publication")
    assert_complete_coverage(terminal_commit, reconciliation, published_source)
    require(exact_children(terminal[0], root) == exact_children(birth[0], root),
            "Repair completion changed children from the first publication")
    evidence.update({
        "first_published_at": first_commit["observed_at"], "destination_repair_to_first_publication_seconds": round(first_delay, 3),
        "pending_removed_at": promotion["observed_at"], "first_completion_commit": terminal_commit["observed_at"],
        "native_publication_completion_epoch": published_target, "completion_to_commit_seconds": round(delay, 3),
        "published_tree": {row["id"]: row.get("parent_id") for row in terminal[0]["sessions"]},
        "after_completion": check_complete(terminal, project.source, published_source, root, unrelated),
    })
    restored_source, bound_discard = discard_cycle(
        project, model, root, key, path, processes, check_index, published_source, bound=True,
    )
    restored, restored_commit, restored_reconciliation, _ = wait_terminal_snapshot(
        project, root, restored_source[root].latest_completion, bound_discard["canonical_publication_boundary"],
        terminal[0]["updated_at"],
    )
    assert_complete_coverage(restored_commit, restored_reconciliation, restored_source)
    require(exact_children(restored[0], root) == exact_children(terminal[0], root),
            "Bound restoration changed previously accepted child facts")
    bound_discard["restored_publication"] = check_complete(restored, project.source, restored_source, root, unrelated)
    bound_discard["canonical_bytes_preserved_during_discard"] = True
    evidence["bound_discard_restore"] = bound_discard
    resume_and_check(project, root, unrelated, restored, restored_commit, restored_source, processes, evidence)
    check_index("subsequent canonical refresh")
    evidence["git_index_unchanged"] = True
    require(not path.exists(), "An already published root reappeared in pending capture")


def exercise(project, model, evidence):
    processes = []
    try:
        index_before = None if project.prepared else project.index_entries()
        unrelated_process, unrelated_log = launch(project, "UNRELATED")
        processes.append(unrelated_process)
        wait_success(unrelated_process, "unrelated session")
        unrelated = native_id(unrelated_log)
        require(unrelated, "Native unrelated session did not create an exact ID")
        evidence["excluded_native_session"] = unrelated
        require(published(project) is None and not project.ledger.exists(), "Unrelated session created a ledger")
        project.observe_diagnostic_history()
        if not project.prepared:
            exercise_pending(project, model, evidence, unrelated, processes, index_before)
            require(not model.errors, f"Fixture provider errors: {model.errors}")
            return

        root_process, root_log = launch(project, "ROOT")
        processes.append(root_process)
        root = event_until(time.monotonic() + FIRST_SECONDS, lambda: native_id(root_log),
                           "Native root ID did not become observable", root_process)
        binding = event_until(time.monotonic() + FIRST_SECONDS, lambda: verified_binding(project, root),
                              "Native root did not reach verified explicit binding", root_process)
        first_published_at, first_delay = wait_first_snapshot(project, binding, root_process)
        evidence.update({
            "binding_accepted_at": binding, "first_published_at": first_published_at,
            "binding_to_first_publication_seconds": round(first_delay, 3),
        })

        before, held_commit = wait_held_snapshot(project, model, root, root_process)
        evidence["before_completion"] = before
        model.release.set()
        owners = event_until(
            time.monotonic() + 45,
            lambda: completed_source(project, root, unix_time(held_commit["observed_at"])),
            "Native development session did not record completion", root_process,
        )
        require(unrelated in owners and root in owners, "Native source is missing an observed exact root")
        expected = tree(owners, root)
        evidence["native_tree"] = {session_id: owner.parent for session_id, owner in expected.items()}
        evidence["nested_delegation_offered"] = model.nested_supported
        assert_activity(expected, project.source, root, model.nested_supported)
        require((project.checkout / "fixture.py").read_text() == "def answer():\n    return 42\n",
                "Implementation child did not make the required code edit")
        target = expected[root].latest_completion
        terminal, first_terminal_pass, terminal_reconciliation, delay = wait_terminal_snapshot(
            project, root, target, held_commit["sequence"],
        )
        wait_success(root_process, "development session")
        evidence.update({
            "published_tree": {row["id"]: row.get("parent_id") for row in terminal[0]["sessions"]},
            "native_completion_epoch": target, "first_completion_commit": first_terminal_pass["observed_at"],
            "completion_to_commit_seconds": round(delay, 3),
        })
        evidence["after_completion"] = check_complete(terminal, project.source, expected, root, unrelated)
        assert_complete_coverage(first_terminal_pass, terminal_reconciliation, expected)

        resume_and_check(project, root, unrelated, terminal, first_terminal_pass, expected, processes, evidence)
        require(not model.errors, f"Fixture provider errors: {model.errors}")
    except Exception as error:
        additional = f" Provider diagnostic: {model.errors}." if model.errors else ""
        raise RuntimeError(f"{error}.{additional} Fixture logs: {project.root}") from error
    finally:
        model.release.set()
        model.repair_release.set()
        for process in processes:
            stop(process)


def launch_completion_reader(project):
    with socket.socket() as listener:
        listener.bind(("127.0.0.1", 0))
        port = listener.getsockname()[1]
    with (project.root / "reader.log").open("w") as log:
        process = FixtureProcess(
            (project.manager, "--al-offline", "--al-project", project.checkout, "server", "--port", str(port)),
            cwd=project.checkout, env=project.env, stdout=log, stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL,
        )

    def query():
        conn = HTTPConnection("127.0.0.1", port, timeout=1)
        try:
            conn.request("GET", f"/api/workspaces/{TICKET}")
            with conn.getresponse() as response:
                if response.status != 200:
                    return None
                return json.load(response)["ledger"]
        except (OSError, HTTPException):
            return None
        finally:
            conn.close()

    return process, query


def final_completion_with_usage(project, root, interval):
    owner = source_owners(project).get(root)
    if owner is None or interval not in owner.completed_intervals or owner.open != 0 or owner.turns <= 1:
        return None
    expected = (FINAL_INPUT_TOKENS, FINAL_OUTPUT_TOKENS)
    _, start_ms, end_ms = owner.completed_intervals[interval]
    if project.source == "codex":
        final = owner.last_token_count
        return owner if final and final[:2] == expected and final[2] >= start_ms / 1000 else None
    # An OpenCode tool reply can finish its assistant message before the
    # application writes the later final response for the same user turn.
    last, (_, _, last_end_ms) = max(owner.completed_intervals.items(), key=lambda item: item[1][2])
    return owner if last_end_ms > end_ms and owner.message_usage.get(last) == expected else None


def observe_settled_completion(project, process, query, root, interval, held, release, evidence):
    api_reads = []
    native = None
    publication = None

    def sample_api():
        started = time.time()
        response = query()
        if response is not None:
            api_reads.append((started, time.time(), response))

    sample_api()
    source_pending_start = time.time()
    require(final_completion_with_usage(project, root, interval) is None,
            "The fixture reached native completion before release")
    last_native_pending = source_pending_start, time.time()
    release_at = time.time()
    evidence["fixture_released_checkpoint_epoch"] = release_at
    release.touch()

    def observe():
        nonlocal native, publication, last_native_pending
        sample_api()
        if "native_client_return_observed_epoch" not in evidence and process.poll() is not None:
            evidence["native_client_return_observed_epoch"] = time.time()
        if native is None:
            source_read_start = time.time()
            native = final_completion_with_usage(project, root, interval)
            source_read_at = time.time()
            if native is None:
                last_native_pending = source_read_start, source_read_at
                return None
            evidence.update({
                "interval_end_epoch": native.completed_intervals[interval][2] / 1000,
                "native_final_completion_epoch": native.latest_completion,
                "source_last_pending_read_bounds_epoch": last_native_pending,
                "source_read_bounds_epoch": (source_read_start, source_read_at),
            })
        if publication is None:
            rows = project.diagnostic_history()
            eligible = first_eligible_completion_publication(
                rows, root, native.latest_completion, evidence["source_read_bounds_epoch"][1], held["sequence"],
            )
            if eligible is None:
                return None
            first, reconciliation, basis = eligible
            require(checkpoint_seen(first, root, native.latest_completion),
                    "First independently eligible publication retained a stale native completion checkpoint")
            commits = [row for row in publication_diagnostics(rows) if row["outcome"] == "committed"]
            require(commits[-1]["sequence"] == first["sequence"],
                    "Could not tie the current ledger to the first native-completion publication")
            publication = observe_publication_pair(project, first, rows)
            if publication is None:
                return None
            evidence.update({
                "first_publication_eligibility": basis,
                "first_publication_reconciliation_started_epoch": unix_time(reconciliation["reconciliation"]["started_at"]),
            })
        snapshot, first, reconciliation = publication
        ledger, _ = snapshot
        older = None
        for query_start, query_at, response in api_reads:
            response_updated = unix_time(response["updated_at"])
            updated = unix_time(ledger["updated_at"])
            if response_updated < updated:
                older = (query_start, query_at)
                continue
            if response_updated != updated:
                continue
            rows = response["facts"]["sessions"]
            require(len(rows) == 1 and rows[0]["id"] == root and rows[0]["source"] == project.source and
                    rows[0]["invocations"] == ledger["sessions"][0]["invocations"],
                    "Ticket detail API did not expose the first settled owner's invocation facts")
            evidence["first_settled_query_bounds_epoch"] = (query_start, query_at)
            if older:
                evidence["api_last_unsettled_read_bounds_epoch"] = older
            return native, snapshot, first, reconciliation
        return None

    return event_until(time.monotonic() + SETTLE_SECONDS, observe,
                       "Native final usage or matching API publication did not become observable in time")


def exercise_completion_feedback(project, model, evidence):
    processes = []
    release = project.checkout / ".fixture-release"
    try:
        unrelated_process, unrelated_log = launch(project, "UNRELATED")
        processes.append(unrelated_process)
        wait_success(unrelated_process, "unrelated session", timeout=120)
        unrelated = native_id(unrelated_log)
        require(unrelated, "Native unrelated session did not expose its ID")
        project.observe_diagnostic_history()
        (project.checkout / "selected" / "manual").mkdir()
        initial_head = project.run("git", "rev-parse", "HEAD")
        staged = project.run("git", "ls-files", "--stage", "--", "staged-fixture.txt")
        require(staged, "Fixture did not start with independently staged work")
        binding, log = launch(project, "BIND")
        processes.append(binding)
        root = event_until(time.monotonic() + FIRST_SECONDS, lambda: native_id(log),
                           "Native fixture binding did not expose a root ID", binding)
        wait_success(binding, "binding")
        accepted = event_until(time.monotonic() + FIRST_SECONDS, lambda: verified_binding(project, root),
                               "Native fixture did not accept binding")
        # Binding completed normally before this stage. The automatic
        # worker's first publication may still be outstanding.
        wait_first_snapshot(project, accepted, None)
        reader, query = launch_completion_reader(project)
        processes.append(reader)
        event_until(time.monotonic() + 10, query, "Packaged ticket detail API did not become queryable", reader)
        model.completion_root = root
        process, completion_log = launch(project, "COMPLETION", root)
        processes.append(process)
        ready = project.checkout / ".fixture-checkpoint-ready"
        # The long native work is deliberately over three minutes. Startup,
        # tool scheduling and the generation pass have their own overhead.
        event_until(time.monotonic() + 320, ready.exists, "Native checkpoint did not complete", process)
        commit_time = float((project.checkout / ".fixture-commit-time").read_text())
        head = project.run("git", "rev-parse", "HEAD")
        require(project.run("git", "rev-parse", "HEAD^") == initial_head and
                project.run("git", "diff-tree", "--no-commit-id", "--name-only", "-r", head) == "fixture.py" and
                project.run("git", "show", "HEAD:fixture.py") == "def answer():\n    return 42",
                "Native tool did not commit the expected fixture work")
        require(project.run("git", "ls-files", "--stage", "--", "staged-fixture.txt") == staged,
                "Native tool altered the unrelated staged work")
        evidence.update({"git_commit": head, "git_commit_observed_at": commit_time, "session_id": root})
        require(native_id(completion_log) == root, "Native tool ran in a different source owner")
        before = source_owners(project)[root]
        require(len(before.open_intervals) == 1, "Native source did not identify one open checkpoint interval")
        interval, (_, start_ms, _) = next(iter(before.open_intervals.items()))
        kind = "turn" if project.source == "codex" else "assistant message"
        warning = (project.checkout / ".fixture-generate-stderr").read_text()
        warnings = [line for line in warning.splitlines() if line.startswith("AgenticLens checkpoint:")]
        exact = [line for line in warnings if f"source={project.source}; session={root};" in line and
                 f"open {kind} {interval}," in line and "threshold 180s" in line]
        require(len(exact) == 1 and "invoking caller is unverified" in exact[0],
                "Preview did not emit the qualified warning for the exact native owner and interval")
        model.expected_completion_feedback = exact[0]
        manual = project.checkout / "selected" / "manual" / "agentic_metrics.json"
        preview = json.loads(manual.read_text())
        require(provisional(preview).get(root, 0) >= 180, "Manual preview did not preserve the provisional interval")
        evidence.update({"interval_id": interval, "interval_start_epoch": start_ms / 1000,
                         "checkpoint_stderr": exact[0], "provisional_seconds_before": provisional(preview)[root]})

        def automatic_preview():
            rows = project.diagnostic_history()
            commits = [row for row in publication_diagnostics(rows) if row["outcome"] == "committed"]
            if not commits:
                return None
            pair = observe_publication_pair(project, commits[-1], rows)
            if pair and unix_time(pair[0][0]["updated_at"]) >= start_ms / 1000 and provisional(pair[0][0]).get(root, 0) > 0:
                return pair[1]
            return None

        held = event_until(time.monotonic() + 16, automatic_preview,
                           "Automatic capture did not publish a provisional snapshot for the open interval", process)
        native, settled, first, reconciliation = observe_settled_completion(
            project, process, query, root, interval, held, release, evidence,
        )
        ledger, _ = settled
        wait_success(process, "natural completion after the checkpoint")
        evidence.setdefault("native_client_return_observed_epoch", time.time())
        require(model.feedback_in_model_input.is_set(), "Checkpoint warning did not reach the provider tool-result context")
        events = list(json_lines(completion_log))
        if project.source == "codex":
            replies = [event["item"]["text"] for event in events if event["type"] == "item.completed" and
                       event["item"]["type"] == "agent_message"]
        else:
            replies = [event["part"]["text"] for event in events if event["type"] == "text"]
        require("Fixture complete." in replies, "Native client did not report the provider's final response")
        evidence.update({"checkpoint_in_provider_tool_result": True, "normal_response_after_tool_result": True})
        require(not model.errors, f"Fixture provider errors: {model.errors}")
        owners = {root: native}
        check_complete(settled, project.source, owners, root, unrelated)
        assert_complete_coverage(first, reconciliation, owners)
        turns = sum(inv["turns"] for inv in ledger["sessions"][0]["invocations"])
        require(turns == native.turns, f"Settled automatic turns {turns} differ from native records {native.turns}")
        project.run(*manual_generate(project.manager, project.checkout, project.source, root))
        manual_ledger = json.loads(manual.read_text())
        # Manual generation must be independently repeatable from outside
        # the completed turn using the explicitly selected source and owner.
        check_completed_facts(manual_ledger, project.source, owners, root, unrelated)
        require(sum(inv["turns"] for inv in manual_ledger["sessions"][0]["invocations"]) == turns,
                "Manual post-completion generation does not agree with the settled turn count")
        published_at = unix_time(first["observed_at"])
        source_read_at = evidence["source_read_bounds_epoch"][1]
        evidence.update({
            "first_settled_publication_epoch": published_at,
            "commit_to_interval_completion_seconds": round(evidence["interval_end_epoch"] - commit_time, 3),
            "commit_to_native_final_completion_seconds": round(native.latest_completion - commit_time, 3),
            "native_final_completion_to_source_read_seconds": round(source_read_at - native.latest_completion, 3),
            "source_read_to_settled_query_seconds": round(evidence["first_settled_query_bounds_epoch"][1] - source_read_at, 3),
            "poll_interval_seconds": POLL_SECONDS, "settled_turns": turns, "settled_usage": dict(native.usage),
            "native_final_reply_usage": {"input": FINAL_INPUT_TOKENS, "output": FINAL_OUTPUT_TOKENS},
            "provisional_seconds_after": provisional(ledger)[root], "external_manual_refresh_verified": True,
        })
    except Exception as error:
        extra = f" Provider diagnostic: {model.errors}." if model.errors else ""
        raise RuntimeError(f"{error}.{extra} Fixture logs: {project.root}") from error
    finally:
        release.touch()
        for process in processes:
            stop(process)


def run_combination(parent, repository, assets, candidate, source, path, expected_version, evidence, prepared,
                    completion_feedback):
    project = Project(parent / source, source, candidate, path, repository, prepared, completion_feedback)
    version = project.run(project.runtime, "--version")
    require(version == expected_version, f"{source} requires {expected_version!r}, found {version!r}")
    evidence.update({"source": source, "harness_version": version})
    model = FixtureModel(project.manager, project.checkout, project.workspace)
    server = model.begin()
    try:
        project.configure_client(server.server_port)
        project.install(assets)
        model.binding_engine = str(project.engine)
        if source == "codex":
            project.trust_codex_hooks(timeout=45 if completion_feedback else 20)
        if prepared:
            project.prepare()
        if completion_feedback:
            exercise_completion_feedback(project, model, evidence)
        else:
            exercise(project, model, evidence)
        return {"result": "passed", **evidence}
    finally:
        model.release.set()
        model.repair_release.set()
        server.shutdown()
        server.server_close()
        if not completion_feedback and source == "codex":
            project.run(project.runtime, "app-server", "daemon", "stop", timeout=15)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--release-dir", required=True, type=Path,
                        help="Packaged host candidate directory containing SHA256SUMS and the archive")
    parser.add_argument("--candidate", required=True, help="Exact v3 release candidate tag")
    parser.add_argument("--codex", default="codex", help="Native Codex executable")
    parser.add_argument("--opencode", default="opencode", help="Native OpenCode executable")
    parser.add_argument("--keep-fixture", action="store_true", help="Keep the private fixture after a successful run")
    parser.add_argument("--fixture-parent", type=Path,
                        help="Existing directory for creating the isolated private fixture")
    scenarios = parser.add_mutually_exclusive_group()
    scenarios.add_argument("--prepared-destination", action="store_true",
                           help="Reproduce the earlier operator-prepared workflow for a historical candidate")
    scenarios.add_argument("--completion-feedback", action="store_true",
                           help="Check a long native code commit, manual checkpoint, and automatic completion refresh")
    args = parser.parse_args()
    repository = Path(__file__).resolve().parents[4]
    assets = require_package(args.release_dir.resolve(), args.candidate)
    run_root = Path(tempfile.mkdtemp(prefix="agenticlens-capture-release-gate-", dir=args.fixture_parent)).resolve()
    results = []
    print(f"Private fixture directory {run_root}", flush=True)
    clients = (("codex", args.codex, "codex-cli 0.170.0"), ("opencode", args.opencode, "1.19.5"))

    def run_client(client):
        source, command, version = client
        evidence = {"source": source, "expected_version": version}
        try:
            path = shutil.which(command)
            require(path, f"Missing required {source} executable {command}")
            return run_combination(run_root, repository, assets, args.candidate, source, path, version, evidence,
                                   args.prepared_destination or args.completion_feedback, args.completion_feedback)
        except (OSError, RuntimeError, subprocess.SubprocessError, sqlite3.Error) as error:
            detail = error.stderr if isinstance(error, subprocess.CalledProcessError) else str(error)
            return {"result": "failed", "detail": detail, **evidence}

    if args.completion_feedback:
        with ThreadPoolExecutor(max_workers=len(clients)) as executor:
            results = list(executor.map(run_client, clients))
    else:
        results = [run_client(client) for client in clients]
    for result in results:
        print(json.dumps(result, sort_keys=True), flush=True)
    passed = all(result["result"] == "passed" for result in results)
    print(json.dumps({
        "candidate": args.candidate, "archive_sha256": assets[2][1], "manager_sha256": assets[0][1],
        "engine_sha256": assets[1][1], "result": "passed" if passed else "failed",
        "scenario": "completion_feedback" if args.completion_feedback else
                    "prepared_destination" if args.prepared_destination else "pending_capture",
        "manual_only_prior_evidence": ["OpenCode 2.2.0 standalone"],
    }, sort_keys=True))
    # Opt-in clients can leave a bounded detached capture pass finishing its
    # native publication. Keep that private state available until it retires.
    if passed and not args.keep_fixture and not args.completion_feedback:
        shutil.rmtree(run_root)
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
