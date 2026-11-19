"""Exercise the installed integrations and report the packaged release gate."""

import argparse
import json
from pathlib import Path
import shutil
import sqlite3
import subprocess
import tempfile
import time

from oracle import (
    artifact_matches_commit, capture_id, check_birth, check_complete, check_complete_pending, check_held, check_held_pending,
    check_pending_identity, native_id, provisional, published, snapshot_matches_pending_record, source_owners, tree,
    unix_time,
)
from project import Project, require, require_package
from provider import FixtureModel, TICKET, TITLE

FIRST_SECONDS = 60
SETTLE_SECONDS = 30
POLL_SECONDS = 0.2


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
        process = subprocess.Popen(project.native_command(stage, session), cwd=project.checkout, env=project.env,
                                   stdout=stdout, stderr=stderr, stdin=subprocess.DEVNULL)
    return process, log


def stop(process):
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


def wait_success(process, stage):
    require(process.wait(timeout=45) == 0, f"Native {stage} failed, see fixture logs")


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

    def native_turn(stage):
        turn, turn_log = launch(project, stage, root)
        processes.append(turn)
        wait_success(turn, stage)
        require(native_id(turn_log) == root, f"{stage} ran in a different native session")

    native_turn("BIND")
    registration = event_until(time.monotonic() + FIRST_SECONDS, lambda: verified_registration(project, root),
                               "Resumed native root did not verify its binding")
    require((registration["ticket_id"], registration["ticket_title"], registration["ticket_type"]) == (TICKET, TITLE, "fix")
            and not registration.get("publication_destination"), "Native binding metadata or destination differs from its response")
    bound_source = tree(source_owners(project), root)
    bound_target = bound_source[root].latest_completion
    require(bound_target > target, "Native binding turn has no completion")
    bound, bound_record = wait_completed_pending(
        project, root, path, "awaiting_destination", bound_target, unbound_record["sequence"],
    )
    check_index("binding without a destination")
    require(exact_children(bound, root) == exact_children(unbound, root),
            "Binding modified the children captured before binding")
    evidence["binding_accepted_at"] = registration["binding_accepted_at"]
    evidence["pending_binding_observed_at"] = bound_record["observed_at"]
    evidence["awaiting_destination"] = check_complete_pending(bound, project.source, bound_source, root, unrelated)

    native_turn("DESTINATION")

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
        birth, project.source, birth_source, root, unrelated, first_commit, first_reconciliation, unbound,
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
    resume_and_check(project, root, unrelated, terminal, terminal_commit, published_source, processes, evidence)
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


def run_combination(parent, repository, assets, candidate, source, path, expected_version, evidence, prepared):
    project = Project(parent / source, source, candidate, path, repository, prepared)
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
            project.trust_codex_hooks()
        if prepared:
            project.prepare()
        exercise(project, model, evidence)
        return {"result": "passed", **evidence}
    finally:
        model.release.set()
        model.repair_release.set()
        server.shutdown()
        server.server_close()
        if source == "codex":
            project.run(project.runtime, "app-server", "daemon", "stop", timeout=15)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--release-dir", required=True, type=Path,
                        help="Packaged host candidate directory containing SHA256SUMS and the archive")
    parser.add_argument("--candidate", required=True, help="Exact v3 release candidate tag")
    parser.add_argument("--codex", default="codex", help="Native Codex executable")
    parser.add_argument("--opencode", default="opencode", help="Native OpenCode executable")
    parser.add_argument("--keep-fixture", action="store_true", help="Keep the private fixture after a successful run")
    parser.add_argument("--prepared-destination", action="store_true",
                        help="Reproduce the earlier operator-prepared workflow for a historical candidate")
    args = parser.parse_args()
    repository = Path(__file__).resolve().parents[4]
    assets = require_package(args.release_dir.resolve(), args.candidate)
    run_root = Path(tempfile.mkdtemp(prefix="agenticlens-capture-release-gate-")).resolve()
    results = []
    print(f"Private fixture directory {run_root}", flush=True)
    for source, command, version in (
        ("codex", args.codex, "codex-cli 0.170.0"), ("opencode", args.opencode, "1.19.5"),
    ):
        evidence = {"source": source, "expected_version": version}
        try:
            path = shutil.which(command)
            require(path, f"Missing required {source} executable {command}")
            results.append(run_combination(run_root, repository, assets, args.candidate, source, path, version, evidence,
                                           args.prepared_destination))
        except (OSError, RuntimeError, subprocess.SubprocessError, sqlite3.Error) as error:
            detail = error.stderr if isinstance(error, subprocess.CalledProcessError) else str(error)
            results.append({"result": "failed", "detail": detail, **evidence})
        print(json.dumps(results[-1], sort_keys=True), flush=True)
    passed = all(result["result"] == "passed" for result in results)
    print(json.dumps({
        "candidate": args.candidate, "archive_sha256": assets[2][1], "manager_sha256": assets[0][1],
        "engine_sha256": assets[1][1], "result": "passed" if passed else "failed",
        "scenario": "prepared_destination" if args.prepared_destination else "pending_capture",
        "manual_only_prior_evidence": ["OpenCode 2.2.0 standalone"],
    }, sort_keys=True))
    if passed and not args.keep_fixture:
        shutil.rmtree(run_root)
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
