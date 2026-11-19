"""Independent native source observations and published ledger assertions."""

from collections import Counter
from contextlib import closing
from dataclasses import dataclass, field
from datetime import datetime
from hashlib import sha256
import json
import math
from pathlib import Path
import re
import sqlite3

from project import require
from provider import TICKET, TITLE


def unix_time(stamp):
    return datetime.fromisoformat(stamp).timestamp()


def duration_seconds(milliseconds):
    return (milliseconds + 500) // 1000


@dataclass
class Owner:
    parent: str | None
    tools: Counter = field(default_factory=Counter)
    usage: Counter = field(default_factory=Counter)
    model_millis: Counter = field(default_factory=Counter)
    open_intervals: dict[str, tuple[str, int, int]] = field(default_factory=dict)
    latest_completion: float = 0
    completed: int = 0
    open: int = 0
    cache_write: int | None = None

    @property
    def seconds(self):
        return sum(duration_seconds(milliseconds) for milliseconds in self.model_millis.values())


def json_lines(path):
    lines = path.read_text().splitlines()
    for index, line in enumerate(lines):
        try:
            yield json.loads(line)
        except json.JSONDecodeError:
            # A reader can observe the current native append mid-write.
            if index + 1 != len(lines):
                raise


def native_id(log):
    for row in json_lines(log):
        if row.get("type") == "thread.started":
            return row["thread_id"]
        if row.get("sessionID"):
            return row["sessionID"]
    return None


def codex_owners(home):
    owners = {}
    for path in (home / "sessions").rglob("rollout-*.jsonl"):
        events = list(json_lines(path))
        if not events or events[0]["type"] != "session_meta":
            continue
        meta = events[0]["payload"]
        session_id = meta["id"]
        src = meta.get("source")
        parent = src.get("subagent", {}).get("thread_spawn", {}).get("parent_thread_id") if isinstance(src, dict) else None
        owner = Owner(parent)
        starts = {}
        calls = {}
        version_model = ""
        for event in events:
            payload = event["payload"]
            kind = payload.get("type")
            if event["type"] == "turn_context":
                version_model = payload["model"]
            if event["type"] == "event_msg" and kind == "task_started":
                starts[payload["turn_id"]] = unix_time(event["timestamp"])
            if event["type"] == "event_msg" and kind == "task_complete":
                end = unix_time(event["timestamp"])
                started = starts.pop(payload["turn_id"])
                owner.model_millis[version_model] += round(1000 * (end - started))
                owner.latest_completion = max(owner.latest_completion, end)
                owner.completed += 1
            if event["type"] == "event_msg" and kind == "token_count":
                usage = (payload.get("info") or {}).get("total_token_usage")
                if usage:
                    owner.usage = Counter({
                        "input": usage["input_tokens"], "output": usage["output_tokens"],
                        "cache_read": usage["cached_input_tokens"], "reasoning": usage["reasoning_output_tokens"],
                    })
                    owner.cache_write = usage.get("cache_write_input_tokens")
            if event["type"] == "response_item" and kind == "function_call" and payload["call_id"].startswith("call_al96_"):
                calls[payload["call_id"]] = payload["name"]
        owner.tools.update(calls.values())
        owner.open = len(starts)
        owner.open_intervals = {
            turn_id: (version_model, round(start * 1000), round(start * 1000)) for turn_id, start in starts.items()
        }
        require(session_id not in owners, f"Duplicate native Codex rollout for {session_id}")
        owners[session_id] = owner
    return owners


def opencode_owners(path):
    if not path.exists():
        return {}
    try:
        with closing(sqlite3.connect(path.as_uri() + "?mode=ro", uri=True, timeout=1)) as db:
            db.execute("BEGIN")
            owners = {session_id: Owner(parent) for session_id, parent in db.execute("SELECT id, parent_id FROM session")}
            open_messages = {}
            for message_id, session_id, text in db.execute("SELECT id, session_id, data FROM message"):
                message = json.loads(text)
                if message.get("role") != "assistant":
                    continue
                owner = owners[session_id]
                timing = message["time"]
                if timing.get("completed") is not None:
                    owner.model_millis[message["modelID"]] += timing["completed"] - timing["created"]
                    owner.latest_completion = max(owner.latest_completion, timing["completed"] / 1000)
                    owner.completed += 1
                else:
                    owner.open += 1
                    open_messages[message_id] = (message["modelID"], timing["created"])
                tokens = message["tokens"]
                cache_write = tokens["cache"]["write"]
                require(type(cache_write) is int and cache_write >= 0, "Invalid native OpenCode cache-write counter")
                owner.cache_write = (owner.cache_write or 0) + cache_write
                owner.usage.update({
                    "input": tokens["input"], "output": tokens["output"],
                    "cache_read": tokens["cache"]["read"], "reasoning": tokens["reasoning"],
                })
            for message_id, session_id, text in db.execute("SELECT message_id, session_id, data FROM part"):
                part = json.loads(text)
                if part.get("type") == "tool":
                    owners[session_id].tools[part["tool"]] += 1
                if message_id not in open_messages:
                    continue
                model, start = open_messages[message_id]
                if part["type"] in ("text", "reasoning"):
                    timing = part.get("time") or {}
                    activity_start = timing.get("start") if "end" not in timing else None
                elif part["type"] == "tool" and part.get("state", {}).get("status") == "running":
                    activity_start = part["state"].get("time", {}).get("start")
                else:
                    activity_start = None
                if activity_start is not None:
                    owners[session_id].open_intervals[message_id] = (model, start, activity_start)
    except sqlite3.OperationalError as error:
        if "no such table: session" not in str(error):
            raise
        return {}
    return owners


def source_owners(project):
    if project.source == "codex":
        return codex_owners(project.codex_home)
    return opencode_owners(project.opencode_db)


def tree(owners, root):
    selected = {root}
    while True:
        expanded = selected | {session_id for session_id, owner in owners.items() if owner.parent in selected}
        if expanded == selected:
            return {session_id: owners[session_id] for session_id in selected}
        selected = expanded


def published(project):
    pending = Path(str(project.ledger) + ".publication-pending")
    if pending.exists() or not project.ledger.is_file() or not project.summary.is_file():
        return None
    ledger = json.loads(project.ledger.read_text())
    summary = project.summary.read_text()
    # Require a consistent observation when a new commit races these reads.
    if pending.exists() or ledger != json.loads(project.ledger.read_text()):
        return None
    return ledger, summary


def provisional(ledger):
    return {row["id"]: sum(i.get("provisional_seconds", 0) for i in row["invocations"]) for row in ledger["sessions"]}


def artifact_matches_commit(ledger, commitment, reconciliation):
    updated = unix_time(ledger["updated_at"])
    source_completions = [unix_time(owner["checkpoint"]) for owner in commitment.get("owners", []) if owner.get("checkpoint")]
    return (unix_time(reconciliation["observed_at"]) <= updated <= unix_time(commitment["observed_at"])
            and all(checkpoint <= updated for checkpoint in source_completions))


def capture_id(source, root):
    return sha256(f"{source}\0{root}".encode()).hexdigest()


def check_pending_identity(snapshot, source, root, state):
    require((snapshot["version"], snapshot["capture_id"], snapshot["source"],
             snapshot["root_session_id"], snapshot["state"]) == (1, capture_id(source, root), source, root, state),
            "Pending snapshot has wrong version, capture identity or state")
    metadata = tuple(snapshot.get(field) for field in ("ticket_id", "ticket_title", "ticket_type"))
    expected = (None, None, None) if state == "awaiting_ticket" else (TICKET, TITLE, "fix")
    require(metadata == expected, f"Pending snapshot metadata differs from the verified {state} state")


def snapshot_matches_pending_record(snapshot, diagnostic):
    return (snapshot.get("source_cutoff") == diagnostic["reconciliation"]["started_at"]
            and snapshot.get("state") == diagnostic["outcome"]
            and snapshot.get("ticket_id") == (diagnostic.get("ticket_id") or None))


def check_held(snapshot, expected, root, commitment, reconciliation, require_provisional=True):
    ledger, _ = snapshot
    rollup = check_rollup(snapshot)
    cutoff = reconciliation["reconciliation"]["started_at"]
    checkpoints = {item["session_id"]: item.get("checkpoint") for item in commitment.get("owners", [])}
    return {
        **rollup, "native_intervals": held_intervals(rollup, expected, root, cutoff, checkpoints, require_provisional),
        "ledger_updated_at": ledger["updated_at"], "source_duration_cutoff": cutoff,
        "source_reconciliation_completed_at": reconciliation["observed_at"],
        "publication_observed_at": commitment["observed_at"],
    }


def check_held_pending(snapshot, source, expected, root, unrelated, diagnostic):
    check_tree(snapshot, source, expected, root, unrelated)
    rollup = duration_rollup(snapshot)
    cutoff = snapshot["source_cutoff"]
    return {
        **rollup, "native_intervals": held_intervals(rollup, expected, root, cutoff),
        "source_duration_cutoff": cutoff, "pending_snapshot_observed_at": diagnostic["observed_at"],
    }


def held_intervals(rollup, expected, root, cutoff, checkpoints=None, require_provisional=True):
    recorded = rollup["per_session"]
    require(set(recorded) == set(expected), "Held ledger and native tree have different owners")
    cutoff_ms = math.floor(unix_time(cutoff) * 1000)
    native = {}
    for session_id, owner in expected.items():
        require((owner.open == 1 and len(owner.open_intervals) == 1) if session_id == root else
                (owner.open == 0 and owner.completed > 0),
                f"{session_id}: native held interval or child completion is unproven")
        require(round(owner.latest_completion * 1000) <= cutoff_ms,
                f"{session_id}: native completion is later than the held cutoff")
        durations = owner.model_millis.copy()
        for model, start, activity in owner.open_intervals.values():
            require(start <= activity <= cutoff_ms, f"{session_id}: no native evidence at the held cutoff")
            durations[model] += cutoff_ms - start
        active = sum(duration_seconds(milliseconds) for milliseconds in durations.values())
        measured = owner.seconds
        fact = recorded[session_id]
        require(fact == {"active_seconds": active, "measured_seconds": measured, "provisional_seconds": active - measured},
                f"{session_id}: held duration differs from the native intervals at cutoff {cutoff}")
        checkpoint = checkpoints[session_id] if checkpoints is not None else None
        if checkpoints is not None and session_id != root:
            require(checkpoint and abs(unix_time(checkpoint) - owner.latest_completion) < 0.001,
                    f"{session_id}: published held snapshot lacks its native child completion checkpoint")
        native[session_id] = {
            **fact, "native_completed_intervals": owner.completed, "native_latest_completion_epoch": owner.latest_completion or None,
            "open_intervals": [
                {"id": identifier, "model": model, "started_epoch_ms": start, "activity_epoch_ms": activity}
                for identifier, (model, start, activity) in owner.open_intervals.items()
            ],
        }
        if checkpoints is not None:
            native[session_id]["published_completion_checkpoint"] = checkpoint
    if require_provisional:
        require(recorded[root]["provisional_seconds"] > 0, "Held root did not publish positive provisional duration")
    return native


def duration_rollup(ledger, native_measured_seconds=None):
    per_owner = {}
    for session in ledger["sessions"]:
        active = sum(invocation["active_duration_seconds"] for invocation in session["invocations"])
        provisional_duration = sum(invocation.get("provisional_seconds", 0) for invocation in session["invocations"])
        require(0 <= provisional_duration <= active, f"{session['id']}: invalid active and provisional duration")
        per_owner[session["id"]] = {
            "active_seconds": active, "measured_seconds": active - provisional_duration,
            "provisional_seconds": provisional_duration,
        }
    require(len(per_owner) == len(ledger["sessions"]), "Ticket duration rollup contains duplicate session IDs")
    active_seconds = sum(owner["active_seconds"] for owner in per_owner.values())
    provisional_seconds = sum(owner["provisional_seconds"] for owner in per_owner.values())
    measured_seconds = active_seconds - provisional_seconds
    if native_measured_seconds is not None:
        require(provisional_seconds == 0, "Ticket provisional total has not settled")
        require(measured_seconds == native_measured_seconds,
                f"Ticket measured duration {measured_seconds}s disagrees with native total {native_measured_seconds}s")

    evidence = {
        "per_session": per_owner, "ticket_active_seconds": active_seconds, "ticket_measured_seconds": measured_seconds,
        "ticket_provisional_seconds": provisional_seconds,
    }
    if native_measured_seconds is not None:
        evidence["native_measured_seconds"] = native_measured_seconds
    return evidence


def check_rollup(snapshot, native_measured_seconds=None):
    ledger, summary = snapshot
    rollup = duration_rollup(ledger, native_measured_seconds)
    active_seconds = rollup["ticket_active_seconds"]
    provisional_seconds = rollup["ticket_provisional_seconds"]
    rows = re.findall(
        r"^- \*\*Active Agent Duration:\*\* ([^\n]+?) \(recorded duration(?: including ([1-9]\d*)s provisional)?\)$",
        summary, re.MULTILINE,
    )
    require(len(rows) == 1, "Generated summary must contain one ticket duration row")
    rendered_duration, rendered_provisional = rows[0]
    hours, remainder = divmod(active_seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    expected_display = f"{hours}h {minutes}m" if hours else f"{minutes}m {seconds}s"
    require(rendered_duration == expected_display,
            f"Ticket rendered active duration {rendered_duration!r} disagrees with ledger duration {expected_display!r}")
    require(int(rendered_provisional or "0") == provisional_seconds,
            f"Ticket rendered provisional duration disagrees with ledger total {provisional_seconds}s")

    return {**rollup, "rendered_active_duration": rendered_duration}


def check_tree(ledger, source, expected, root, unrelated):
    observed = {row["id"]: row for row in ledger["sessions"]}
    require(len(observed) == len(ledger["sessions"]), "Published ledger contains duplicate session IDs")
    require(unrelated not in observed, f"Unrelated session {unrelated} was captured on {TICKET}")
    require(set(observed) == set(expected), f"Expected native IDs {sorted(expected)}, published {sorted(observed)}")
    require(observed[root].get("parent_id") is None, "Published root has a parent")
    for session_id, owner in expected.items():
        require(observed[session_id]["source"] == source and observed[session_id].get("parent_id") == owner.parent,
                f"Wrong source or parent for {session_id}")
    return observed


def check_completed_facts(ledger, source, expected, root, unrelated):
    observed = check_tree(ledger, source, expected, root, unrelated)
    evidence = check_usage_and_tools(observed, expected)
    for session_id, owner in expected.items():
        row = observed[session_id]
        require(owner.completed > 0 and owner.open == 0, f"Native completion unconfirmed for {session_id}")
        invocations = row["invocations"]
        active_seconds = sum(i["active_duration_seconds"] for i in invocations)
        require(active_seconds == owner.seconds, f"{session_id}: expected {owner.seconds} active seconds, observed {active_seconds}")
        require(not sum(i.get("provisional_seconds", 0) for i in invocations), f"{session_id}: completed owner retains provisional time")
    return evidence


def check_usage_and_tools(observed, expected):
    for session_id, owner in expected.items():
        row = observed[session_id]
        invocations = row["invocations"]
        tools = {tool["name"]: tool["calls"] for tool in row["tools"]}
        require(len(tools) == len(row["tools"]), f"{session_id}: published duplicate tool rows")
        require(tools == dict(owner.tools), f"{session_id}: native tools {dict(owner.tools)}, published {tools}")
        errors = sum(tool["errors"] for tool in row["tools"])
        require(errors == 0, f"{session_id}: fixture had {errors} recorded tool errors")
        totals = {key: sum(inv["tokens"][key] for inv in invocations) for key in owner.usage}
        require(all("cache_write" not in invocation["tokens"] for invocation in invocations),
                f"{session_id}: ledger fabricated an unsupported cache-write key")
        require(owner.usage["input"] > 0 and owner.usage["output"] > 0 and totals == dict(owner.usage),
                f"{session_id}: native tokens {dict(owner.usage)}, published {totals}")
    return {
        "native_cache_write_by_session": {session_id: owner.cache_write for session_id, owner in expected.items()},
        "ledger_cache_write": "unpersisted",
    }


def check_ticket_metadata(ledger):
    require((ledger["ticket_id"], ledger["ticket_title"], ledger["ticket_type"]) == (TICKET, TITLE, "fix"),
            "Published binding metadata differs from the native agent's response")


def check_birth(snapshot, source, expected, root, unrelated, commitment, reconciliation, pending):
    ledger, _ = snapshot
    check_ticket_metadata(ledger)
    observed = check_tree(ledger, source, expected, root, unrelated)
    require({session_id: row for session_id, row in observed.items() if session_id != root} ==
            {row["id"]: row for row in pending["sessions"] if row["id"] != root},
            "First canonical publication did not preserve children captured before binding")
    evidence = check_held(snapshot, expected, root, commitment, reconciliation, require_provisional=False)
    pending_root = duration_rollup(pending)["per_session"][root]
    require(pending_root["provisional_seconds"] == 0 and
            evidence["per_session"][root]["measured_seconds"] >= pending_root["measured_seconds"] > 0,
            "First canonical publication lost the root's measured duration from before binding")
    return {**evidence, **check_usage_and_tools(observed, expected)}


def check_complete(snapshot, source, expected, root, unrelated):
    ledger, _ = snapshot
    check_ticket_metadata(ledger)
    return {
        **check_completed_facts(ledger, source, expected, root, unrelated),
        **check_rollup(snapshot, native_measured_seconds=sum(owner.seconds for owner in expected.values())),
    }


def check_complete_pending(snapshot, source, expected, root, unrelated):
    return {
        **check_completed_facts(snapshot, source, expected, root, unrelated),
        **duration_rollup(snapshot, native_measured_seconds=sum(owner.seconds for owner in expected.values())),
    }
