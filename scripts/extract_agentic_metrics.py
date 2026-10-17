"""Extract session-specific agentic metrics from OpenCode DB into an atomic fact ledger."""

from __future__ import annotations

import argparse
import json
import re
import sqlite3
import subprocess
from collections import defaultdict
from dataclasses import asdict, dataclass, field
from datetime import UTC, datetime
from pathlib import Path
from typing import cast

HEALTHY_CEILING_TOKENS: int = 200_000


@dataclass(slots=True)
class _SessionModelAccumulator:
    inp: int = 0
    out: int = 0
    cache: int = 0
    reas: int = 0
    peak: int = 0
    cost: float = 0.0
    active_ms: int = 0
    first_ts: int | None = None
    last_ts: int | None = None
    asst_msgs: int = 0


@dataclass(slots=True)
class _RoleAggregation:
    session_ids: set[str] = field(default_factory=set)
    input: int = 0
    output: int = 0
    cache: int = 0
    cost: float = 0.0
    peak: int = 0
    turns: list[int] = field(default_factory=list)


@dataclass(slots=True)
class _ModelAggregation:
    session_ids: set[str] = field(default_factory=set)
    input: int = 0
    output: int = 0
    cache: int = 0
    cost: float = 0.0


def format_tokens_k(toks: int) -> str:
    """Format token count as human-readable string (e.g., 900.8k, 200k, 1.2M, 450)."""
    if toks >= 1_000_000:
        val = toks / 1_000_000
        return f"{val:.0f}M" if val.is_integer() else f"{val:.1f}M"
    if toks >= 1_000:
        val = toks / 1_000
        return f"{val:.0f}k" if val.is_integer() else f"{val:.1f}k"
    return str(toks)


def get_health_label(toks: int, ceiling: int = HEALTHY_CEILING_TOKENS) -> str:
    """Format granular health label with percentage for bloated/critical against ceiling."""
    if toks > 2 * ceiling:
        pct = round(toks / ceiling * 100)
        return f"CRITICAL ({pct}%)"
    if toks > ceiling:
        pct = round(toks / ceiling * 100)
        return f"BLOATED ({pct}%)"
    if toks >= 100_000:
        return "ELEVATED"
    return "HEALTHY"


def get_base_health_status(toks: int, ceiling: int = HEALTHY_CEILING_TOKENS) -> str:
    """Return base health category without percentage."""
    if toks > 2 * ceiling:
        return "CRITICAL"
    if toks > ceiling:
        return "BLOATED"
    if toks >= 100_000:
        return "ELEVATED"
    return "HEALTHY"


@dataclass(slots=True)
class InvocationTokens:
    input: int
    output: int
    cache_read: int
    reasoning: int
    peak_context: int


@dataclass(slots=True)
class InvocationFact:
    id: str
    parent_id: str | None
    role: str
    model: str
    user: str
    started_at: str | None
    completed_at: str | None
    active_duration_seconds: int
    turns: int
    tokens: InvocationTokens
    cost_usd: float
    compactions: int = 0


@dataclass(slots=True)
class ToolFact:
    name: str
    calls: int
    errors: int


@dataclass(slots=True)
class CodeFacts:
    lines_added: int
    lines_deleted: int
    test_lines_added: int
    test_lines_deleted: int
    prod_lines_added: int
    prod_lines_deleted: int
    files_changed: int
    sdlc_iterations: int


@dataclass(slots=True)
class AgenticFactLedger:
    ticket_id: str | None
    ticket_title: str | None
    ticket_type: str | None
    started_at: str | None
    completed_at: str | None
    updated_at: str
    code: CodeFacts
    invocations: list[InvocationFact]
    tools: list[ToolFact]

    def to_dict(self) -> dict[str, object]:
        return {
            "ticket_id": self.ticket_id,
            "ticket_title": self.ticket_title,
            "ticket_type": self.ticket_type,
            "started_at": self.started_at,
            "completed_at": self.completed_at,
            "updated_at": self.updated_at,
            "code": asdict(self.code),
            "invocations": [asdict(i) for i in self.invocations],
            "tools": [asdict(t) for t in self.tools],
        }


def get_git_user() -> str:
    """Fetch user.email directly from git config to avoid proxy redactions."""
    try:
        res = subprocess.run(
            ["git", "config", "user.email"],
            capture_output=True,
            text=True,
            check=False,
        )
        email = res.stdout.strip()
        if email:
            return email
    except OSError:
        pass

    try:
        res = subprocess.run(
            ["git", "config", "user.name"],
            capture_output=True,
            text=True,
            check=False,
        )
        name = res.stdout.strip()
        if name:
            return name
    except OSError:
        pass

    return "unknown"


def resolve_active_root_session(
    conn: sqlite3.Connection,
    nonce: str | None = None,
    ticket_id: str | None = None,
    explicit_session_id: str | None = None,
) -> str | None:
    """Resolve active root session ID without fragile fuzzy string matches."""
    if explicit_session_id:
        return explicit_session_id

    clean_nonce = nonce.strip() if nonce else None
    clean_ticket = ticket_id.strip() if ticket_id else None

    # Priority 1: Exact Nonce Matching (Zero collision across parallel threads/folders)
    if clean_nonce:
        query_nonce = """
            SELECT DISTINCT s.id
            FROM part p
            JOIN session s ON (p.session_id = s.id OR s.id = (SELECT parent_id FROM session WHERE id = p.session_id))
            WHERE s.parent_id IS NULL AND p.data LIKE ?
            ORDER BY p.time_created DESC LIMIT 1
        """
        row_nonce: tuple[str] | None = conn.execute(
            query_nonce, (f"%{clean_nonce}%",)
        ).fetchone()
        if row_nonce:
            return str(row_nonce[0])

    # Priority 2: Ticket ID matching in session title
    if clean_ticket:
        query_title = """
            SELECT id
            FROM session
            WHERE parent_id IS NULL AND title LIKE ?
            ORDER BY time_updated DESC LIMIT 1
        """
        row_title: tuple[str] | None = conn.execute(
            query_title, (f"%{clean_ticket}%",)
        ).fetchone()
        if row_title:
            return str(row_title[0])

    # Priority 3: Fallback to most recent message ONLY IF no ticket or nonce was specified
    if not clean_nonce and not clean_ticket:
        query_latest = """
            SELECT s.id, s.parent_id
            FROM message m
            JOIN session s ON m.session_id = s.id
            ORDER BY m.time_created DESC LIMIT 1
        """
        latest_row: tuple[str, str | None] | None = conn.execute(
            query_latest
        ).fetchone()
        if latest_row:
            sess_id, parent_id = latest_row
            return str(parent_id if parent_id else sess_id)

    return None


def parse_numstat_output(stdout: str) -> list[tuple[int, int, str]]:
    """Parse git numstat output into list of (added, deleted, filepath)."""
    results: list[tuple[int, int, str]] = []
    for line in stdout.splitlines():
        parts = line.strip().split(maxsplit=2)
        if len(parts) >= 3 and parts[0].isdigit() and parts[1].isdigit():
            results.append((int(parts[0]), int(parts[1]), parts[2]))
    return results


def get_ticket_numstat(ticket_id: str | None) -> list[tuple[int, int, str]]:
    """Retrieve git numstat entries for a ticket, excluding docs and manifests."""
    if not ticket_id:
        return []

    clean_ticket = ticket_id.strip()

    # 1. Check if active branch is working on this ticket
    res_curr = subprocess.run(
        ["git", "branch", "--show-current"],
        capture_output=True,
        text=True,
        check=False,
    )
    curr_branch = res_curr.stdout.strip()
    if curr_branch and clean_ticket.lower() in curr_branch.lower():
        res_diff = subprocess.run(
            [
                "git",
                "diff",
                "--numstat",
                "origin/main...HEAD",
                "--",
                ":(exclude)*.md",
                ":(exclude)*.json",
                ":(exclude)*.lock",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        parsed = parse_numstat_output(res_diff.stdout)
        if parsed:
            return parsed

    # 2. Check merged commits on origin/main where ticket appears in subject line
    res_log = subprocess.run(
        [
            "git",
            "log",
            "origin/main",
            f"--grep={clean_ticket}",
            "--format=%H\t%s",
            "-n",
            "25",
        ],
        capture_output=True,
        text=True,
        check=False,
    )
    for line in res_log.stdout.splitlines():
        if not line:
            continue
        parts = line.split("\t", 1)
        if len(parts) == 2:
            sha, subj = parts
            if re.search(rf"\b{re.escape(clean_ticket)}\b", subj, re.IGNORECASE):
                res_diff = subprocess.run(
                    [
                        "git",
                        "diff",
                        "--numstat",
                        f"{sha}^1",
                        sha,
                        "--",
                        ":(exclude)*.md",
                        ":(exclude)*.json",
                        ":(exclude)*.lock",
                    ],
                    capture_output=True,
                    text=True,
                    check=False,
                )
                parsed = parse_numstat_output(res_diff.stdout)
                if parsed:
                    return parsed

    # 3. Check any local or remote feature branch matching ticket ID
    res_br = subprocess.run(
        ["git", "branch", "-a", "--list", f"*{clean_ticket}*"],
        capture_output=True,
        text=True,
        check=False,
    )
    for b in res_br.stdout.splitlines():
        b_clean = b.strip().lstrip("*+ ").strip()
        if b_clean and "HEAD" not in b_clean:
            res = subprocess.run(
                [
                    "git",
                    "diff",
                    "--numstat",
                    f"origin/main...{b_clean}",
                    "--",
                    ":(exclude)*.md",
                    ":(exclude)*.json",
                    ":(exclude)*.lock",
                ],
                capture_output=True,
                text=True,
                check=False,
            )
            parsed = parse_numstat_output(res.stdout)
            if any(
                r[2].startswith(("src/", "tests/", "infra/", "cli/"))
                for r in parsed
            ):
                return parsed

    # 4. Check merged commits on origin/main mentioning ticket in body (excluding tool/metrics commits)
    for line in res_log.stdout.splitlines():
        if not line:
            continue
        parts = line.split("\t", 1)
        if len(parts) == 2:
            sha, subj = parts
            if re.search(
                r"\b(refactor|feat|chore)\((agentic/)?metrics\)",
                subj,
                re.IGNORECASE,
            ):
                continue
            res_diff = subprocess.run(
                [
                    "git",
                    "diff",
                    "--numstat",
                    f"{sha}^1",
                    sha,
                    "--",
                    ":(exclude)*.md",
                    ":(exclude)*.json",
                    ":(exclude)*.lock",
                ],
                capture_output=True,
                text=True,
                check=False,
            )
            parsed = parse_numstat_output(res_diff.stdout)
            if any(
                r[2].startswith(("src/", "tests/", "infra/", "cli/"))
                for r in parsed
            ):
                return parsed

    return []


def compute_sdlc_iterations(ticket_id: str | None) -> int:
    """Calculate iterations count from ticket workspace."""
    if not ticket_id:
        return 1

    w_dir = Path(f"workspaces/{ticket_id}")
    if (w_dir / "iterations").exists():
        subdirs = [d for d in (w_dir / "iterations").iterdir() if d.is_dir()]
        if subdirs:
            return len(subdirs)

    return 1


def infer_ticket_type(ticket_id: str | None) -> str | None:
    """Infer ticket type from git branch name or commit messages."""
    if not ticket_id:
        return None

    clean_ticket = ticket_id.strip()

    # 1. Search git branches
    try:
        res = subprocess.run(
            ["git", "branch", "-a", "--list", f"*{clean_ticket}*"],
            capture_output=True,
            text=True,
            check=False,
        )
        for line in res.stdout.strip().splitlines():
            clean = line.strip().lstrip("*+").strip()
            if "/" in clean:
                parts = [p for p in clean.split("/") if p != "remotes"]
                prefix = (
                    parts[0].lower()
                    if parts[0] != "origin"
                    else (parts[1].lower() if len(parts) > 1 else "")
                )
                if prefix in ("feat", "feature"):
                    return "feature"
                if prefix in ("fix", "bug", "bugfix"):
                    return "bug"
                if prefix in ("chore", "ci", "task"):
                    return "chore"
                if prefix in ("refactor",):
                    return "refactor"
                if prefix in ("docs",):
                    return "documentation"
    except OSError:
        pass

    # 2. Search git log commits mentioning ticket
    try:
        res = subprocess.run(
            ["git", "log", "--all", f"--grep={clean_ticket}", "--oneline", "-n", "5"],
            capture_output=True,
            text=True,
            check=False,
        )
        for line in res.stdout.strip().splitlines():
            line_lower = line.lower()
            if (
                "feat(" in line_lower
                or "feat:" in line_lower
                or "feature" in line_lower
            ):
                return "feature"
            if "fix(" in line_lower or "fix:" in line_lower or "bug" in line_lower:
                return "bug"
            if (
                "chore(" in line_lower
                or "chore:" in line_lower
                or "ci(" in line_lower
                or "ci:" in line_lower
            ):
                return "chore"
            if "refactor(" in line_lower or "refactor:" in line_lower:
                return "refactor"
    except OSError:
        pass

    return "feature"


def infer_ticket_title(ticket_id: str | None) -> str | None:
    """Infer a concise, human-readable ticket title from delivery artifacts or git history."""
    if not ticket_id:
        return None

    clean_ticket = ticket_id.strip()

    # 1. Parse top heading from SUMMARY.md in ticket workspace if present
    sum_path = Path(f"workspaces/{clean_ticket}/SUMMARY.md")
    if sum_path.exists():
        try:
            first_line = (
                sum_path.read_text(encoding="utf-8").splitlines()[0].strip()
            )
            clean = re.sub(r"^#\s*", "", first_line)
            clean = re.sub(
                r"Delivery Summary:?\s*", "", clean, flags=re.IGNORECASE
            )
            clean = re.sub(r"Summary:?\s*", "", clean, flags=re.IGNORECASE)
            clean = re.sub(
                rf"{clean_ticket}:?\s*", "", clean, flags=re.IGNORECASE
            )
            clean = re.sub(r"\s*-\s*Summary$", "", clean, flags=re.IGNORECASE)
            clean = re.sub(
                r"\s*Delivery Summary$", "", clean, flags=re.IGNORECASE
            )
            clean = clean.strip(": -")
            if clean:
                return clean
        except OSError:
            pass

    # 2. Parse top heading from CONTRACT.md in ticket workspace if present
    contract_path = Path(f"workspaces/{clean_ticket}/CONTRACT.md")
    if contract_path.exists():
        try:
            first_line = (
                contract_path.read_text(encoding="utf-8")
                .splitlines()[0]
                .strip()
            )
            clean = re.sub(r"^#\s*", "", first_line)
            clean = re.sub(r"CONTRACT:?\s*", "", clean, flags=re.IGNORECASE)
            clean = re.sub(
                rf"{clean_ticket}:?\s*", "", clean, flags=re.IGNORECASE
            )
            clean = re.sub(r"\s*-\s*Contract$", "", clean, flags=re.IGNORECASE)
            clean = clean.strip(": -")
            if clean:
                return clean
        except OSError:
            pass

    # 3. Search git commit logs mentioning ticket
    try:
        res = subprocess.run(
            [
                "git",
                "log",
                "--all",
                f"--grep={clean_ticket}",
                "--oneline",
                "-n",
                "15",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
        for line in res.stdout.splitlines():
            if re.search(
                r"\b(refactor|feat|chore)\((agentic/)?metrics\)",
                line,
                re.IGNORECASE,
            ):
                continue
            m = re.search(
                rf"\b{re.escape(clean_ticket)}[:\s\-]+([^(#\n]+)", line
            )
            if m:
                clean = m.group(1).strip()
                clean = re.sub(r"^\[SLCP\]\s*", "", clean, flags=re.IGNORECASE)
                clean = clean.strip(": -")
                if clean and not clean.startswith("DPS-"):
                    return clean
    except OSError:
        pass

    return None


def compute_code_facts(
    ticket_id: str | None,
    fallback: CodeFacts | None = None,
) -> CodeFacts:
    """Calculate code changes, test/prod breakdown, and SDLC iterations from git."""
    numstat = get_ticket_numstat(ticket_id)
    if not numstat and fallback is not None:
        return fallback

    added = sum(r[0] for r in numstat)
    deleted = sum(r[1] for r in numstat)
    files = len(numstat)
    test_added = sum(r[0] for r in numstat if r[2].startswith("tests/"))
    test_deleted = sum(r[1] for r in numstat if r[2].startswith("tests/"))
    prod_added = sum(r[0] for r in numstat if r[2].startswith("src/"))
    prod_deleted = sum(r[1] for r in numstat if r[2].startswith("src/"))
    sdlc_iters = compute_sdlc_iterations(ticket_id)

    return CodeFacts(
        lines_added=added,
        lines_deleted=deleted,
        test_lines_added=test_added,
        test_lines_deleted=test_deleted,
        prod_lines_added=prod_added,
        prod_lines_deleted=prod_deleted,
        files_changed=files,
        sdlc_iterations=sdlc_iters,
    )


def load_existing_ledger(
    path: Path | None,
) -> tuple[
    str | None,
    str | None,
    str | None,
    str | None,
    str | None,
    list[str],
    CodeFacts | None,
    list[InvocationFact],
    list[ToolFact],
]:
    """Load existing facts from disk ledger, supporting both new schema and legacy schema."""
    if not path or not path.exists():
        return None, None, None, None, None, [], None, [], []

    try:
        content: object = json.loads(path.read_text(encoding="utf-8"))
        if not isinstance(content, dict):
            return None, None, None, None, None, [], None, [], []

        cd = cast(dict[str, object], content)
        cached_ticket = str(cd["ticket_id"]) if cd.get("ticket_id") else None
        cached_ticket_title = (
            str(cd["ticket_title"]) if cd.get("ticket_title") else None
        )
        cached_ticket_type = str(cd["ticket_type"]) if cd.get("ticket_type") else None
        cached_started_at = str(cd["started_at"]) if cd.get("started_at") else None
        cached_completed_at = (
            str(cd["completed_at"]) if cd.get("completed_at") else None
        )

        # Case 1: Already new atomic fact schema
        if "invocations" in cd and isinstance(cd["invocations"], list):
            invocations_list: list[InvocationFact] = []
            for item in cast(list[object], cd["invocations"]):
                if isinstance(item, dict):
                    id_dict = cast(dict[str, object], item)
                    toks_dict = cast(dict[str, object], id_dict.get("tokens", {}))
                    invocations_list.append(
                        InvocationFact(
                            id=str(id_dict.get("id", "")),
                            parent_id=str(id_dict["parent_id"])
                            if id_dict.get("parent_id")
                            else None,
                            role=str(id_dict.get("role", "orchestrator")),
                            model=str(id_dict.get("model", "default")),
                            user=str(id_dict.get("user", "unknown")),
                            started_at=str(id_dict["started_at"])
                            if id_dict.get("started_at")
                            else None,
                            completed_at=str(id_dict["completed_at"])
                            if id_dict.get("completed_at")
                            else None,
                            active_duration_seconds=int(
                                str(id_dict.get("active_duration_seconds", 0))
                            ),
                            turns=int(str(id_dict.get("turns", 0))),
                            tokens=InvocationTokens(
                                input=int(str(toks_dict.get("input", 0))),
                                output=int(str(toks_dict.get("output", 0))),
                                cache_read=int(str(toks_dict.get("cache_read", 0))),
                                reasoning=int(str(toks_dict.get("reasoning", 0))),
                                peak_context=int(str(toks_dict.get("peak_context", 0))),
                            ),
                            cost_usd=float(str(id_dict.get("cost_usd", 0.0))),
                            compactions=int(str(id_dict.get("compactions", 0))),
                        )
                    )

            tools_list: list[ToolFact] = []
            if "tools" in cd and isinstance(cd["tools"], list):
                for t_item in cast(list[object], cd["tools"]):
                    if isinstance(t_item, dict):
                        td = cast(dict[str, object], t_item)
                        tools_list.append(
                            ToolFact(
                                name=str(td.get("name", "")),
                                calls=int(str(td.get("calls", 0))),
                                errors=int(str(td.get("errors", 0))),
                            )
                        )

            code_facts: CodeFacts | None = None
            if "code" in cd and isinstance(cd["code"], dict):
                code_d = cast(dict[str, object], cd["code"])
                code_facts = CodeFacts(
                    lines_added=int(str(code_d.get("lines_added", 0))),
                    lines_deleted=int(str(code_d.get("lines_deleted", 0))),
                    test_lines_added=int(str(code_d.get("test_lines_added", 0))),
                    test_lines_deleted=int(str(code_d.get("test_lines_deleted", 0))),
                    prod_lines_added=int(str(code_d.get("prod_lines_added", 0))),
                    prod_lines_deleted=int(str(code_d.get("prod_lines_deleted", 0))),
                    files_changed=int(str(code_d.get("files_changed", 0))),
                    sdlc_iterations=int(str(code_d.get("sdlc_iterations", 1))),
                )

            root_thread_ids = [i.id for i in invocations_list if i.parent_id is None]
            return (
                cached_ticket,
                cached_ticket_title,
                cached_ticket_type,
                cached_started_at,
                cached_completed_at,
                root_thread_ids,
                code_facts,
                invocations_list,
                tools_list,
            )

        # Case 2: Legacy schema with "threads" and "workflow_efficiency"
        root_thread_ids: list[str] = []
        raw_threads = cd.get("threads")
        if isinstance(raw_threads, list):
            for t_item in cast(list[object], raw_threads):
                if isinstance(t_item, dict):
                    td = cast(dict[str, object], t_item)
                    tid = str(td.get("session_id", ""))
                    if tid:
                        root_thread_ids.append(tid)

        code_facts = None
        raw_eff = cd.get("workflow_efficiency")
        if isinstance(raw_eff, dict):
            eff_d = cast(dict[str, object], raw_eff)
            cd_info = cast(dict[str, object], eff_d.get("code_delivered", {}))
            tr_info = cast(dict[str, object], eff_d.get("test_to_production_ratio", {}))
            code_facts = CodeFacts(
                lines_added=int(str(cd_info.get("lines_added", 0))),
                lines_deleted=int(str(cd_info.get("lines_deleted", 0))),
                test_lines_added=int(str(tr_info.get("test_lines_added", 0))),
                test_lines_deleted=0,
                prod_lines_added=int(str(tr_info.get("prod_lines_added", 0))),
                prod_lines_deleted=0,
                files_changed=int(str(cd_info.get("files_changed", 0))),
                sdlc_iterations=int(str(eff_d.get("sdlc_iterations", 1))),
            )

        return (
            cached_ticket,
            cached_ticket_title,
            cached_ticket_type,
            cached_started_at,
            cached_completed_at,
            root_thread_ids,
            code_facts,
            [],
            [],
        )
    except (json.JSONDecodeError, OSError):
        return None, None, None, None, None, [], None, [], []


def extract_facts_from_db(
    conn: sqlite3.Connection,
    thread_ids: list[str],
    user: str,
    user_by_session_id: dict[str, str],
) -> tuple[list[InvocationFact], list[ToolFact], str | None, str | None]:
    """Query atomic invocation and tool facts from OpenCode DB for all specified root threads."""
    if not thread_ids:
        return [], [], None, None

    placeholders = ",".join("?" for _ in thread_ids)
    params = thread_ids + thread_ids

    # 1. Earliest start time and latest completed time
    q_times = f"""
        SELECT MIN(time_created), MAX(time_updated)
        FROM session
        WHERE id IN ({placeholders}) OR parent_id IN ({placeholders})
    """
    row_times: tuple[int | None, int | None] | None = conn.execute(
        q_times, params
    ).fetchone()
    started_at_iso = (
        datetime.fromtimestamp(row_times[0] / 1000, UTC).isoformat()
        if row_times and row_times[0] is not None
        else None
    )
    completed_at_iso = (
        datetime.fromtimestamp(row_times[1] / 1000, UTC).isoformat()
        if row_times and row_times[1] is not None
        else None
    )

    # 2. Compactions per session
    q_comp = f"""
        SELECT p.session_id, COUNT(*)
        FROM part p
        JOIN session s ON p.session_id = s.id
        WHERE (s.id IN ({placeholders}) OR s.parent_id IN ({placeholders}))
          AND json_extract(p.data, '$.type') = 'compaction'
        GROUP BY p.session_id
    """
    comp_map: dict[str, int] = dict(conn.execute(q_comp, params).fetchall())

    # 3. Sessions
    q_sess = f"""
        SELECT s.id, s.parent_id, s.agent, s.model, s.time_created
        FROM session s
        WHERE s.id IN ({placeholders}) OR s.parent_id IN ({placeholders})
        ORDER BY s.time_created ASC
    """
    sessions = conn.execute(q_sess, params).fetchall()

    invocations: list[InvocationFact] = []

    for sid, pid, agent, s_model, t_created in sessions:
        role = (
            "orchestrator"
            if pid is None or agent in ("build", "orchestrator")
            else str(agent or "orchestrator")
        )
        session_user = user_by_session_id.get(sid, user)

        q_msgs = """
            SELECT json_extract(m.data, '$.modelID'), m.data, json_extract(m.data, '$.role')
            FROM message m
            WHERE m.session_id = ?
            ORDER BY m.time_created ASC
        """
        msgs = conn.execute(q_msgs, (sid,)).fetchall()

        by_model: dict[str, _SessionModelAccumulator] = {}
        total_session_msgs = len(msgs)

        for mid, m_data, m_role in msgs:
            if m_role == "assistant":
                m_name = str(mid) if mid else "default"
                if m_name not in by_model:
                    by_model[m_name] = _SessionModelAccumulator()
                bm = by_model[m_name]
                bm.asst_msgs += 1
                if m_data:
                    try:
                        raw_d: object = json.loads(m_data)
                        if isinstance(raw_d, dict):
                            d = cast(dict[str, object], raw_d)
                            t_info = d.get("time")
                            if isinstance(t_info, dict):
                                td = cast(dict[str, object], t_info)
                                c = td.get("created")
                                comp = td.get("completed")
                                if isinstance(c, (int, float)):
                                    if bm.first_ts is None or c < bm.first_ts:
                                        bm.first_ts = int(c)
                                    if isinstance(comp, (int, float)):
                                        if bm.last_ts is None or comp > bm.last_ts:
                                            bm.last_ts = int(comp)
                                        if comp >= c:
                                            bm.active_ms += int(comp - c)
                                    elif bm.last_ts is None or c > bm.last_ts:
                                        bm.last_ts = int(c)

                            toks = d.get("tokens")
                            if isinstance(toks, dict):
                                toks_dict = cast(dict[str, object], toks)
                                inp = int(str(toks_dict.get("input", 0)))
                                out = int(str(toks_dict.get("output", 0)))
                                cache_info = toks_dict.get("cache")
                                cache = 0
                                if isinstance(cache_info, dict):
                                    cache = int(
                                        str(
                                            cast(dict[str, object], cache_info).get(
                                                "read", 0
                                            )
                                        )
                                    )
                                reas = int(str(toks_dict.get("reasoning", 0)))
                                bm.inp += inp
                                bm.out += out
                                bm.cache += cache
                                bm.reas += reas
                                bm.peak = max(bm.peak, inp + cache)

                            cost_raw = d.get("cost")
                            cost = float(str(cost_raw)) if cost_raw is not None else 0.0
                            bm.cost += cost
                    except json.JSONDecodeError, TypeError:
                        pass

        if not by_model:
            fallback_model = "default"
            if s_model:
                try:
                    raw_sm: object = json.loads(s_model)
                    if isinstance(raw_sm, dict):
                        sm_dict = cast(dict[str, object], raw_sm)
                        if "id" in sm_dict and sm_dict["id"] is not None:
                            fallback_model = str(sm_dict["id"])
                        else:
                            fallback_model = str(s_model)
                    else:
                        fallback_model = str(s_model)
                except json.JSONDecodeError, TypeError:
                    fallback_model = str(s_model)
            by_model[fallback_model] = _SessionModelAccumulator(
                first_ts=t_created,
                last_ts=t_created,
            )

        session_compactions = comp_map.get(sid, 0)
        for idx, (m_name, st) in enumerate(by_model.items()):
            f_ts = st.first_ts or t_created
            l_ts = st.last_ts or t_created
            st_iso = (
                datetime.fromtimestamp(f_ts / 1000, UTC).isoformat() if f_ts else None
            )
            end_iso = (
                datetime.fromtimestamp(l_ts / 1000, UTC).isoformat() if l_ts else None
            )
            invocations.append(
                InvocationFact(
                    id=sid,
                    parent_id=pid,
                    role=role,
                    model=m_name,
                    user=session_user,
                    started_at=st_iso,
                    completed_at=end_iso,
                    active_duration_seconds=round(st.active_ms / 1000),
                    turns=total_session_msgs if len(by_model) == 1 else st.asst_msgs,
                    tokens=InvocationTokens(
                        input=st.inp,
                        output=st.out,
                        cache_read=st.cache,
                        reasoning=st.reas,
                        peak_context=st.peak,
                    ),
                    cost_usd=round(st.cost, 6),
                    compactions=session_compactions if idx == 0 else 0,
                )
            )

    # 4. Tool calls
    q_tools = f"""
        SELECT p.data
        FROM part p
        JOIN session s ON p.session_id = s.id
        WHERE (s.id IN ({placeholders}) OR s.parent_id IN ({placeholders}))
          AND json_extract(p.data, '$.type') = 'tool'
    """
    t_map: dict[str, dict[str, int]] = {}
    for (p_raw,) in conn.execute(q_tools, params).fetchall():
        if p_raw:
            try:
                raw_p: object = json.loads(p_raw)
                if isinstance(raw_p, dict):
                    pd = cast(dict[str, object], raw_p)
                    t_name = str(pd.get("tool") or "unknown")
                    if t_name not in t_map:
                        t_map[t_name] = {"calls": 0, "errors": 0}
                    t_map[t_name]["calls"] += 1
                    st_obj = pd.get("state")
                    if isinstance(st_obj, dict):
                        st_dict = cast(dict[str, object], st_obj)
                        s_stat = str(st_dict.get("status", ""))
                        s_out = str(st_dict.get("output", ""))
                        if (
                            s_stat == "error"
                            or "Traceback" in s_out
                            or "Error:" in s_out[:50]
                        ):
                            t_map[t_name]["errors"] += 1
            except json.JSONDecodeError, TypeError:
                pass

    tools = [
        ToolFact(name=t_name, calls=stats["calls"], errors=stats["errors"])
        for t_name, stats in sorted(
            t_map.items(), key=lambda x: x[1]["calls"], reverse=True
        )
    ]

    inv_max_completed = max(
        (i.completed_at for i in invocations if i.completed_at), default=None
    )
    completed_at_iso = inv_max_completed or completed_at_iso

    return invocations, tools, started_at_iso, completed_at_iso


def extract_session_metrics(
    nonce: str | None = None,
    explicit_session_id: str | None = None,
    ticket_id: str | None = None,
    ticket_title: str | None = None,
    ticket_type: str | None = None,
    started_at: str | None = None,
    completed_at: str | None = None,
    existing_ledger_path: Path | None = None,
) -> AgenticFactLedger:
    """Query and aggregate agent metrics into an atomic fact ledger."""
    opencode_db = Path.home() / ".local" / "share" / "opencode" / "opencode.db"

    (
        cached_ticket,
        cached_ticket_title,
        cached_ticket_type,
        cached_started_at,
        cached_completed_at,
        root_thread_ids,
        cached_code,
        cached_invocations,
        cached_tools,
    ) = load_existing_ledger(existing_ledger_path)

    active_ticket = ticket_id or cached_ticket
    active_ticket_title = (
        ticket_title or cached_ticket_title or infer_ticket_title(active_ticket)
    )
    active_ticket_type = (
        ticket_type or cached_ticket_type or infer_ticket_type(active_ticket)
    )
    current_user = get_git_user()
    active_started_at: str | None = started_at or cached_started_at
    active_completed_at: str | None = completed_at or cached_completed_at

    known_thread_ids = list(root_thread_ids)
    user_by_session_id = {i.id: i.user for i in cached_invocations}

    invocations: list[InvocationFact] = list(cached_invocations)
    tools: list[ToolFact] = list(cached_tools)

    if opencode_db.exists():
        try:
            conn = sqlite3.connect(str(opencode_db))
            root_id = resolve_active_root_session(
                conn=conn,
                nonce=nonce,
                ticket_id=active_ticket,
                explicit_session_id=explicit_session_id,
            )

            if root_id and root_id not in known_thread_ids:
                known_thread_ids.append(root_id)

            if known_thread_ids:
                (
                    db_invocations,
                    db_tools,
                    db_started_at,
                    db_completed_at,
                ) = extract_facts_from_db(
                    conn=conn,
                    thread_ids=known_thread_ids,
                    user=current_user,
                    user_by_session_id=user_by_session_id,
                )
                if db_invocations:
                    invocations = db_invocations
                    tools = db_tools
                    active_started_at = (
                        started_at or db_started_at or active_started_at
                    )
                    active_completed_at = (
                        completed_at or db_completed_at or active_completed_at
                    )
        except sqlite3.Error:
            pass

    if invocations and not started_at:
        inv_min_started = min(
            (i.started_at for i in invocations if i.started_at), default=None
        )
        if inv_min_started:
            active_started_at = min(
                filter(None, [active_started_at, inv_min_started])
            )

    if invocations and not completed_at:
        inv_max_completed = max(
            (i.completed_at for i in invocations if i.completed_at), default=None
        )
        if inv_max_completed:
            active_completed_at = max(
                filter(None, [active_completed_at, inv_max_completed])
            )

    code_facts = compute_code_facts(active_ticket, fallback=cached_code)
    updated_at_iso = datetime.now(UTC).isoformat()

    return AgenticFactLedger(
        ticket_id=active_ticket,
        ticket_title=active_ticket_title,
        ticket_type=active_ticket_type,
        started_at=active_started_at,
        completed_at=active_completed_at,
        updated_at=updated_at_iso,
        code=code_facts,
        invocations=invocations,
        tools=tools,
    )


def format_markdown_table(ledger: AgenticFactLedger) -> str:
    """Dynamically reconstruct the standard summary markdown review table from atomic facts."""
    invs = ledger.invocations
    tools = ledger.tools
    code = ledger.code

    if not invs:
        return ""

    contributors = sorted({i.user for i in invs if i.user and i.user != "unknown"})

    role_order = [
        "orchestrator",
        "planner",
        "coder",
        "tester",
        "reviewer",
        "general",
        "explore",
    ]

    by_role: dict[str, _RoleAggregation] = defaultdict(_RoleAggregation)
    for i in invs:
        r = i.role
        by_role[r].session_ids.add(i.id)
        by_role[r].input += i.tokens.input
        by_role[r].output += i.tokens.output
        by_role[r].cache += i.tokens.cache_read
        by_role[r].cost += i.cost_usd
        by_role[r].peak = max(by_role[r].peak, i.tokens.peak_context)
        if i.parent_id is not None:
            by_role[r].turns.append(i.turns)

    standard_subagents = ["planner", "coder", "tester", "reviewer"]
    subagent_order = [
        r for r in standard_subagents if r in by_role and r != "orchestrator"
    ] + [
        r
        for r in by_role
        if r != "orchestrator" and r not in standard_subagents
    ]

    tot_invocations = sum(len(by_role[r].session_ids) for r in by_role)
    tot_input = sum(i.tokens.input for i in invs)
    tot_output = sum(i.tokens.output for i in invs)
    tot_cache = sum(i.tokens.cache_read for i in invs)
    tot_cost = sum(i.cost_usd for i in invs)
    tot_peak = max((i.tokens.peak_context for i in invs), default=0)
    tot_tokens_all = tot_input + tot_cache
    tot_hit_pct = (tot_cache / tot_tokens_all * 100) if tot_tokens_all > 0 else 0.0
    tot_health = get_base_health_status(tot_peak)

    dur_secs = sum(i.active_duration_seconds for i in invs)
    dur_hours, dur_rem = divmod(dur_secs, 3600)
    dur_mins, dur_rem_secs = divmod(dur_rem, 60)
    formatted_duration = (
        f"{dur_hours}h {dur_mins}m" if dur_hours > 0 else f"{dur_mins}m {dur_rem_secs}s"
    )

    tot_reasoning = sum(i.tokens.reasoning for i in invs)
    gen_total = tot_output + tot_reasoning
    reasoning_ratio = (
        round((tot_reasoning / gen_total * 100), 1) if gen_total > 0 else 0.0
    )

    tot_tool_calls = sum(t.calls for t in tools)
    tot_tool_errors = sum(t.errors for t in tools)
    tool_succ = (
        round(((tot_tool_calls - tot_tool_errors) / tot_tool_calls * 100), 1)
        if tot_tool_calls > 0
        else 100.0
    )

    test_ratio_str = (
        f"{code.test_lines_added / code.prod_lines_added:.1f}:1"
        if code.prod_lines_added > 0
        else f"{code.test_lines_added}:0"
    )
    sub_costs = {r: by_role[r].cost for r in subagent_order if r in by_role}
    tot_sub_cost = sum(sub_costs.values())
    role_dist_parts = [
        f"{r.title()} {round(sub_costs[r] / tot_sub_cost * 100, 1)}%"
        for r in subagent_order
        if r in sub_costs and tot_sub_cost > 0
    ]
    role_dist_str = ", ".join(role_dist_parts) if role_dist_parts else "N/A"

    all_sub_turns = [
        t for r in subagent_order if r in by_role for t in by_role[r].turns
    ]
    avg_sub_turns = (
        round(sum(all_sub_turns) / len(all_sub_turns), 1) if all_sub_turns else 0.0
    )
    peak_sub_turns = max(all_sub_turns) if all_sub_turns else 0
    vel_parts = [
        f"{r.title()} {round(sum(by_role[r].turns) / len(by_role[r].turns), 1)}"
        for r in subagent_order
        if r in by_role and by_role[r].turns
    ]
    vel_str = f" | {', '.join(vel_parts)}" if vel_parts else ""

    orch_peak = by_role["orchestrator"].peak if "orchestrator" in by_role else 0
    orch_tok_str = format_tokens_k(orch_peak)
    orch_pct_str = f"{round(orch_peak / HEALTHY_CEILING_TOKENS * 100)}%"
    tot_compactions = sum(i.compactions for i in invs)

    sub_roles_present = [r for r in subagent_order if r in by_role]
    if sub_roles_present:
        sub_max_r = max(sub_roles_present, key=lambda r: by_role[r].peak)
        sub_tok_str = format_tokens_k(by_role[sub_max_r].peak)
        sub_role_str = sub_max_r.title()
        sub_peak = by_role[sub_max_r].peak
    else:
        sub_tok_str = "0"
        sub_role_str = "N/A"
        sub_peak = 0

    overall_peak = max(orch_peak, sub_peak)
    cwh_status = get_base_health_status(overall_peak)

    lines = [
        "### Workflow Efficiency",
        f"- **Software Development Life Cycle Iterations:** {code.sdlc_iterations}",
        f"- **Active Agent Duration:** {formatted_duration} (excludes user wait time)",
        f"- **Context Window Health:** {cwh_status} — Orchestrator peak {orch_tok_str} tokens ({orch_pct_str} of {format_tokens_k(HEALTHY_CEILING_TOKENS)} healthy ceiling, {tot_compactions} compactions) | Subagents peak {sub_tok_str} ({sub_role_str})",
        f"- **Role Effort Distribution (Subagents):** {role_dist_str}",
        f"- **Subagent Turn Velocity & Thrashing:** {avg_sub_turns} avg turns/run (peak: {peak_sub_turns} turns){vel_str}",
        f"- **Test-to-Production Code Ratio:** {test_ratio_str} (+{code.test_lines_added:,} test lines / +{code.prod_lines_added:,} prod lines)",
        f"- **Code Delivered:** +{code.lines_added:,} / -{code.lines_deleted:,} lines across {code.files_changed} code files (excluding docs)",
        f"- **Reasoning Effort:** {tot_reasoning:,} reasoning tokens ({reasoning_ratio:.1f}% of generated output)",
        f"- **Tool Health & Self-Correction:** {tot_tool_calls:,} tool calls ({tot_tool_errors} self-corrections, {tool_succ:.1f}% success rate)",
        "",
        "### Agentic Execution Metrics",
        "",
    ]
    if contributors:
        lines.extend([f"**Contributors:** {', '.join(contributors)}", ""])

    lines.extend(
        [
            "#### By Subagent Role",
            "| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |",
            "|---|---|---|---|---|---|---|---|---|",
        ]
    )

    ordered_role_keys = [r for r in role_order if r in by_role] + [
        r for r in by_role if r not in role_order
    ]
    for r in ordered_role_keys:
        rd = by_role[r]
        tot_tok = rd.input + rd.cache
        hit_pct = (rd.cache / tot_tok * 100) if tot_tok > 0 else 0.0
        lines.append(
            f"| **{r}** | {len(rd.session_ids)} | {rd.input:,} | "
            f"{rd.output:,} | {rd.cache:,} | "
            f"{hit_pct:.1f}% | {format_tokens_k(rd.peak)} | "
            f"{get_health_label(rd.peak)} | ${rd.cost:.2f} |"
        )

    lines.append(
        f"| **Total** | **{tot_invocations}** | **{tot_input:,}** | "
        f"**{tot_output:,}** | **{tot_cache:,}** | "
        f"**{tot_hit_pct:.1f}%** | **{format_tokens_k(tot_peak)}** | "
        f"**{tot_health}** | **${tot_cost:.2f}** |"
    )

    by_model: dict[str, _ModelAggregation] = defaultdict(_ModelAggregation)
    for i in invs:
        m = i.model
        by_model[m].session_ids.add(i.id)
        by_model[m].input += i.tokens.input
        by_model[m].output += i.tokens.output
        by_model[m].cache += i.tokens.cache_read
        by_model[m].cost += i.cost_usd

    lines.extend(
        [
            "",
            "#### By Model",
            "| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |",
            "|---|---|---|---|---|---|---|",
        ]
    )

    for m, md in sorted(by_model.items(), key=lambda x: x[1].cost, reverse=True):
        m_tot_tok = md.input + md.cache
        m_hit_pct = (md.cache / m_tot_tok * 100) if m_tot_tok > 0 else 0.0
        lines.append(
            f"| **{m}** | {len(md.session_ids)} | {md.input:,} | "
            f"{md.output:,} | {md.cache:,} | "
            f"{m_hit_pct:.1f}% | ${md.cost:.2f} |"
        )

    lines.append(
        f"| **Total** | **{tot_invocations}** | **{tot_input:,}** | "
        f"**{tot_output:,}** | **{tot_cache:,}** | "
        f"**{tot_hit_pct:.1f}%** | **${tot_cost:.2f}** |"
    )
    lines.append("")
    return "\n".join(lines)


def update_summary_file(dest: Path, table_md: str) -> None:
    """Idempotently update or append the metrics table inside the summary file."""
    dest.parent.mkdir(parents=True, exist_ok=True)
    header = "### Workflow Efficiency"
    legacy_header = "### Agentic Execution Metrics"

    if dest.exists():
        content = dest.read_text(encoding="utf-8")
        target_idx = -1
        if header in content:
            target_idx = content.find(header)
        elif legacy_header in content:
            target_idx = content.find(legacy_header)

        if target_idx != -1:
            next_heading = content.find("\n## ", target_idx + len(header))
            if next_heading != -1:
                new_content = (
                    content[:target_idx].rstrip()
                    + "\n\n"
                    + table_md.strip()
                    + "\n\n"
                    + content[next_heading:].lstrip()
                )
            else:
                new_content = (
                    content[:target_idx].rstrip() + "\n\n" + table_md.strip() + "\n"
                )
            dest.write_text(new_content, encoding="utf-8")
            print(f"Metrics table successfully updated in {dest}")
            return

    with open(dest, "a", encoding="utf-8") as f:
        f.write("\n" + table_md.strip() + "\n")
    print(f"Metrics table successfully appended to {dest}")


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Extract session agentic metrics from OpenCode DB into an atomic fact ledger."
    )
    parser.add_argument(
        "--nonce",
        default=None,
        help="Unique random execution nonce to lock session resolution to",
    )
    parser.add_argument(
        "--ticket-id",
        default=None,
        help="Jira ticket ID to lock session resolution to",
    )
    parser.add_argument(
        "--ticket-title",
        default=None,
        help="Human-readable ticket title or descriptive name",
    )
    parser.add_argument(
        "--ticket-type",
        default=None,
        help="Ticket type classification (e.g., feature, bug, chore, refactor)",
    )
    parser.add_argument(
        "--session-id",
        default=None,
        help="Explicit root session ID",
    )
    parser.add_argument(
        "--started-at",
        default=None,
        help="Explicit start timestamp ISO format",
    )
    parser.add_argument(
        "--completed-at",
        default=None,
        help="Explicit completion timestamp ISO format",
    )
    parser.add_argument(
        "--append",
        default=None,
        help="Path to markdown file to append/update metrics table in",
    )
    parser.add_argument(
        "--json-output",
        default=None,
        help="Path to write/update structured JSON ledger",
    )
    args = parser.parse_args()

    json_path = Path(args.json_output) if args.json_output else None

    ledger = extract_session_metrics(
        nonce=args.nonce,
        explicit_session_id=args.session_id,
        ticket_id=args.ticket_id,
        ticket_title=args.ticket_title,
        ticket_type=args.ticket_type,
        started_at=args.started_at,
        completed_at=args.completed_at,
        existing_ledger_path=json_path,
    )
    table_md = format_markdown_table(ledger)

    if json_path:
        json_path.parent.mkdir(parents=True, exist_ok=True)
        with open(json_path, "w", encoding="utf-8") as f:
            json.dump(ledger.to_dict(), f, indent=2)
            f.write("\n")
        print(f"Structured metrics ledger written to {json_path}")

    if args.append:
        update_summary_file(Path(args.append), table_md)
    elif not args.json_output:
        print(table_md)


if __name__ == "__main__":
    main()
