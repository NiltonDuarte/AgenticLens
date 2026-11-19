"""Exercise the installed integrations and report the packaged release gate."""

import argparse
import json
from pathlib import Path
import shutil
import sqlite3
import subprocess
import tempfile
import time

from oracle import artifact_matches_commit, check_complete, check_held, native_id, provisional, published, source_owners, tree, unix_time
from project import Project, require, require_package
from provider import FixtureModel, TICKET

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


def verified_binding(project, root):
    for owner in status(project).get("registrations", []):
        if owner["identity"]["session_id"] == root and owner.get("binding") == "explicit_agent_response":
            return owner.get("binding_accepted_at")
    return None


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
    current = publication_diagnostics(status(project).get("diagnostics", []))
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


def completion_seen(pass_record, root, instant):
    if pass_record["outcome"] != "committed":
        return False
    return any(owner["session_id"] == root and owner.get("checkpoint") and
               unix_time(owner["checkpoint"]) >= instant - 0.001 for owner in pass_record.get("owners", []))


def wait_first_snapshot(project, binding, process):
    begin = unix_time(binding)
    first_seen = None

    def observed_first_commit():
        nonlocal first_seen
        rows = status(project).get("diagnostics", [])
        commits = [item for item in publication_diagnostics(rows) if item["outcome"] == "committed"]
        if not commits:
            return None
        if first_seen is None:
            first_seen = commits[0]
            require(first_seen["destination"] == str(project.ledger) and first_seen["summary"] == str(project.summary),
                    "First committed publication points to a different destination or summary")
        # The birth commit may have been followed by a normal refresh before
        # binding was observed here. Use its durable timestamp and validate
        # the current pair against its own commit.
        return first_seen if observe_publication_pair(project, commits[-1], rows) else None

    deadline = time.monotonic() + FIRST_SECONDS - (time.time() - begin)
    commitment = event_until(deadline, observed_first_commit,
                             "No first committed publication after verified binding", process)
    delay = unix_time(commitment["observed_at"]) - begin
    require(0 <= delay <= FIRST_SECONDS, f"First commit outside binding budget: {delay:.3f}s")
    return commitment["observed_at"], delay


def wait_held_snapshot(project, model, root, process):
    deadline = time.monotonic() + FIRST_SECONDS

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
        assert_activity(expected, project.source, root, model.nested_supported)
        return expected, time.time()

    expected, observed_at = event_until(deadline, inspect_native_hold,
                                       "Native source did not show a held root and all three completed descendants", process)

    def inspect_held_pair():
        rows = status(project).get("diagnostics", [])
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
    first_seen = None

    def observed_terminal_commit():
        nonlocal first_seen
        rows = status(project).get("diagnostics", [])
        passes = [record for record in publication_diagnostics(rows) if record["sequence"] > since_sequence]
        commits = [record for record in passes if completion_seen(record, root, target)]
        if not commits:
            return None
        if first_seen is None:
            first_seen = commits[0]
        require(passes[-1]["sequence"] == first_seen["sequence"],
                "Could not associate the current ledger with the first committed publication that saw native completion")
        pair = observe_publication_pair(project, first_seen, rows)
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


def assert_activity(owners, source, root, nested_supported):
    root_tools = owners[root].tools
    require(nested_supported is True, "Reviewed native version did not offer a nested delegation tool")
    require(len(owners) == 4, f"Expected three native descendants, observed {len(owners) - 1}")
    children = [owner for owner in owners.values() if owner.parent == root]
    require(len(children) == 2, "Native root did not delegate distinct implementation and verification children")
    if source == "codex":
        require(root_tools["spawn_agent"] == 2 and root_tools["wait_agent"] == 2,
                "Codex root did not make two spawn and wait calls")
    else:
        require(root_tools["task"] == 2 and root_tools["agenticlens_bind"] == 1,
                "OpenCode root did not use native synchronous tasks and the binding tool")
    require(sum(owner.tools["spawn_agent" if source == "codex" else "task"] for owner in children) == 1,
            "Implementation child did not exercise the available native nested delegation")
    require(all(owner.tools["exec_command" if source == "codex" else "bash"] > 0 for owner in children),
            "Native children did not perform the fixture edit and verification")


def exercise(project, model, evidence):
    processes = []
    try:
        unrelated_process, unrelated_log = launch(project, "UNRELATED")
        processes.append(unrelated_process)
        wait_success(unrelated_process, "unrelated session")
        unrelated = native_id(unrelated_log)
        require(unrelated, "Native unrelated session did not create an exact ID")
        evidence["excluded_native_session"] = unrelated
        require(published(project) is None and not project.ledger.exists(), "Unrelated session created a ledger")

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
        wait_success(root_process, "development session")
        owners = source_owners(project)
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
        evidence.update({
            "published_tree": {row["id"]: row.get("parent_id") for row in terminal[0]["sessions"]},
            "native_completion_epoch": target, "first_completion_commit": first_terminal_pass["observed_at"],
            "completion_to_commit_seconds": round(delay, 3),
        })
        evidence["after_completion"] = check_complete(terminal, project.source, expected, root, unrelated)
        assert_complete_coverage(first_terminal_pass, terminal_reconciliation, expected)

        continued, continued_log = launch(project, "RESUME", root)
        processes.append(continued)
        wait_success(continued, "normal resume")
        require(native_id(continued_log) == root, "Resume created a different native root")
        final_source = tree(source_owners(project), root)
        latest_completion = final_source[root].latest_completion
        require(latest_completion > target and final_source[root].usage["input"] > expected[root].usage["input"],
                "Native resume did not add completed activity and usage")
        final, final_commit, final_reconciliation, continued_delay = wait_terminal_snapshot(
            project, root, latest_completion, first_terminal_pass["sequence"], terminal[0]["updated_at"],
        )
        evidence["after_resume"] = check_complete(final, project.source, final_source, root, unrelated)
        assert_complete_coverage(final_commit, final_reconciliation, final_source)
        earlier_by_id = {row["id"]: row for row in terminal[0]["sessions"] if row["id"] != root}
        require({row["id"]: row for row in final[0]["sessions"] if row["id"] != root} == earlier_by_id,
                "Refreshing the root changed earlier child session records")
        require(not model.errors, f"Fixture provider errors: {model.errors}")
        evidence["resumed_completion_to_commit_seconds"] = round(continued_delay, 3)
    except Exception as error:
        additional = f" Provider diagnostic: {model.errors}." if model.errors else ""
        raise RuntimeError(f"{error}.{additional} Fixture logs: {project.root}") from error
    finally:
        model.release.set()
        for process in processes:
            stop(process)


def run_combination(parent, repository, assets, candidate, source, path, expected_version, evidence=None):
    project = Project(parent / source, source, candidate, path, repository)
    version = project.run(project.runtime, "--version")
    require(version == expected_version, f"{source} requires {expected_version!r}, found {version!r}")
    evidence = {} if evidence is None else evidence
    evidence.update({"source": source, "harness_version": version})
    model = FixtureModel(project.manager, project.checkout)
    server = model.begin()
    try:
        project.configure_client(server.server_port)
        project.install(assets)
        model.binding_engine = str(project.engine)
        if source == "codex":
            project.trust_codex_hooks()
        project.prepare()
        exercise(project, model, evidence)
        return {"result": "passed", **evidence}
    finally:
        model.release.set()
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
            results.append(run_combination(run_root, repository, assets, args.candidate, source, path, version, evidence))
        except (OSError, RuntimeError, subprocess.SubprocessError, sqlite3.Error) as error:
            detail = error.stderr if isinstance(error, subprocess.CalledProcessError) else str(error)
            results.append({"result": "failed", "detail": detail, **evidence})
        print(json.dumps(results[-1], sort_keys=True), flush=True)
    passed = all(result["result"] == "passed" for result in results)
    print(json.dumps({
        "candidate": args.candidate, "archive_sha256": assets[2][1], "manager_sha256": assets[0][1],
        "engine_sha256": assets[1][1], "result": "passed" if passed else "failed",
        "manual_only_prior_evidence": ["OpenCode 2.2.0 standalone"],
    }, sort_keys=True))
    if passed and not args.keep_fixture:
        shutil.rmtree(run_root)
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
