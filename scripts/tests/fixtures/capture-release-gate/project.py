"""Build and install an isolated adopting project using the release assets."""

from contextlib import closing
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shlex
import sqlite3
import subprocess
import time
import tomllib

from provider import MODEL, TICKET


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def write_json(path, value):
    path.write_text(json.dumps(value, indent=2) + "\n")


def require_package(release, version):
    require(re.fullmatch(r"v3\.\d+\.\d+-rc\.[0-9]+\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*", version), "Expected an exact v3 candidate tag")
    os_name = platform.system().lower()
    arch = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(platform.machine())
    require(os_name in ("linux", "darwin") and arch, "Unsupported release host")
    platform_name = f"{os_name}-{arch}"
    manifest = (release / "SHA256SUMS").read_text().splitlines()

    def asset(filename):
        lines = [line.split() for line in manifest if line.split() and line.split()[-1].lstrip("*").removeprefix("bin/") == filename]
        require(len(lines) == 1 and len(lines[0]) == 2, f"Expected exactly one manifest entry for {filename}")
        digest, manifest_path = lines[0]
        path = release / manifest_path.lstrip("*")
        if not path.is_file() and manifest_path.lstrip("*").startswith("bin/"):
            path = release / filename
        require(re.fullmatch("[0-9a-fA-F]{64}", digest), f"Invalid digest for {filename}")
        with path.open("rb") as body:
            require(hashlib.file_digest(body, "sha256").hexdigest() == digest.lower(),
                    f"Packaged artifact checksum mismatch for {filename}")
        return path, digest.lower()

    archive = f"agentic-lens-{version}-{platform_name}.tar.gz"
    archive_entry = asset(archive)
    return asset(f"agentic-lens-{platform_name}"), asset(f"agentic-lens-engine-{platform_name}"), archive_entry


class Project:
    def __init__(self, parent, source, version, binary, repository):
        self.source = source
        self.version = version
        self.runtime = binary
        self.repository = repository
        self.root = parent
        self.checkout = parent / "project"
        self.home = parent / "home"
        self.workspace = self.checkout / "selected" / "ordinary-directory"
        self.ledger = self.workspace / "agentic_metrics.json"
        self.summary = self.workspace / "delivery.md"
        self.data = parent / "data"
        self.codex_home = self.home / ".codex"
        self.opencode_db = self.data / "opencode.db"
        self.manager = self.home / ".local" / "bin" / "agentic-lens"
        self.engine = None
        self.diagnostic_registry = None
        self.env = {
            "HOME": str(self.home), "PATH": str(self.manager.parent) + os.pathsep + os.environ["PATH"],
            "LANG": "C.UTF-8", "TERM": "xterm-256color", "CODEX_HOME": str(self.codex_home),
            "XDG_CONFIG_HOME": str(parent / "config"), "XDG_DATA_HOME": str(self.data),
            "XDG_STATE_HOME": str(parent / "state"), "XDG_CACHE_HOME": str(parent / "cache"),
            "OPENCODE_DB": str(self.opencode_db), "OPENCODE_DISABLE_AUTOUPDATE": "true",
            "OPENCODE_DISABLE_LSP_DOWNLOAD": "true", "OPENCODE_DISABLE_MODELS_FETCH": "true",
            "OPENCODE_DISABLE_TERMINAL_TITLE": "true",
            "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": os.devnull,
        }
        for path in (self.workspace, self.codex_home, self.data, parent / "config"):
            path.mkdir(parents=True)
        (self.checkout / ".gitignore").write_text(".agenticlens-config.local.yaml\n")
        (self.checkout / "fixture.py").write_text("def answer():\n    return 41\n")
        self.run("git", "init", "-q", "-b", "main")
        self.run("git", "config", "user.name", "Capture gate")
        self.run("git", "config", "user.email", "capture-gate@example.invalid")
        self.run("git", "add", ".gitignore", "fixture.py")
        self.run("git", "commit", "-qm", "Initial test project")
        self.run("git", "update-ref", "refs/remotes/origin/main", "HEAD")
        write_json(self.checkout / ".agenticlens-config.yaml", {
            "agenticlens": {"version": version},
            "project_name": f"AL-96 {source} release gate",
            "ticket": {
                "case_insensitive": False,
                "prefixes": [{"prefix": "TEST-", "description": "Isolated release-gate ticket"}],
            },
            "workspaces": {
                "roots": [{"path": "selected", "discovery": {"max_depth": 1, "max_directories": 300}}],
                "artifacts": [{"path": "delivery.md", "title": "Delivery", "description": "Generated fixture summary"}],
            },
            "generation": {
                "metrics": {"skills": True, "delegations": True, "anomalies": False},
                "summary": {"enabled": True, "filename": "delivery.md"},
                "feedback": {"provisional_threshold_seconds": 0},
            },
            "code_metrics": {
                "base_branch": "origin/main",
                "production": {"include": ["fixture.py"], "exclude": []},
                "tests": {"include": ["test_*.py"], "exclude": []},
            },
            "database": {"busy_timeout_ms": 1000},
            "live": {"active_threshold_seconds": 60, "poll_interval_ms": 1000},
            "anomaly_thresholds": {
                "stall_threshold_seconds": 30, "error_rate_min_calls": 5,
                "error_rate_threshold": 0.35, "error_rate_critical_threshold": 0.7,
                "consecutive_errors_streak_warning": 4, "consecutive_errors_streak_critical": 8,
                "context_ceiling_warning_tokens": 100000, "context_ceiling_critical_tokens": 200000,
                "turn_usage_warning_tokens": 30000, "turn_usage_critical_tokens": 60000,
                "context_growth_warning_tokens": 40000, "context_growth_critical_tokens": 80000,
                "tool_thrashing_min_failures": 3, "tool_thrashing_critical_failures": 5,
            },
            "role_mappings": {},
            "role_order": [],
        })
        source_config = {"home": str(self.codex_home)} if source == "codex" else {"database_path": str(self.opencode_db)}
        write_json(self.checkout / ".agenticlens-config.local.yaml", {
            "agenticlens": {"state_dir": str(self.home / ".agenticlens" / "state")},
            "telemetry": {
                "default_source": source,
                "sources": {source: source_config},
                "capture": {
                    "enabled": True, "state_dir": str(self.home / ".agenticlens" / "capture"),
                    "refresh_ms": 15000, "debounce_ms": 500, "idle_ms": 16000,
                    "max_lifetime_ms": 180000, "pass_timeout_ms": 10000,
                    "max_sessions": 32, "max_source_bytes": 16777216, "commit_warning": False,
                },
            },
        })

    def run(self, *args, timeout=60):
        return subprocess.run(args, cwd=self.checkout, env=self.env, text=True, stdin=subprocess.DEVNULL,
                              capture_output=True, check=True, timeout=timeout).stdout.strip()

    def configure_client(self, port):
        if self.source == "codex":
            (self.codex_home / "config.toml").write_text(f"""model = "{MODEL}"
model_provider = "gate"
approval_policy = "never"
sandbox_mode = "workspace-write"
[agents]
max_depth = 2
max_threads = 4
[sandbox_workspace_write]
exclude_slash_tmp = true
exclude_tmpdir_env_var = true
[model_providers.gate]
name = "Loopback fixture"
base_url = "http://127.0.0.1:{port}/v1"
wire_api = "responses"
request_max_retries = 0
stream_max_retries = 0
supports_websockets = false
""")
        else:
            write_json(self.checkout / "opencode.json", {
                "model": "al96local/mock-model",
                "provider": {"al96local": {
                    "npm": "@ai-sdk/openai-compatible",
                    "name": "AL-96 loopback fixture",
                    "options": {"baseURL": f"http://127.0.0.1:{port}/v1", "apiKey": "loopback-fixture"},
                    "models": {"mock-model": {
                        "name": "AL-96 fixture", "tool_call": True, "limit": {"context": 8192, "output": 1024},
                    }},
                }},
                "permission": {"bash": "allow", "task": "allow"},
                "subagent_depth": 2,
            })

    def install(self, assets):
        manager, engine, _archive = assets
        args = ["bash", self.repository / ".agents/skills/agentic-lens-setup/scripts/download.sh",
                "--project", self.checkout, "--version", self.version, "--manager-version", self.version,
                "--local-manager-bin", manager[0], "--manager-sha256", manager[1],
                "--local-bin", engine[0], "--sha256", engine[1], "--capture-source", self.source]
        if self.source == "codex":
            args.extend(("--capture-store", self.codex_home))
        self.run(*args)
        engine_name = engine[0].name
        host_platform = engine_name.removeprefix("agentic-lens-engine-")
        self.engine = self.home / ".agenticlens" / "bin" / self.version / host_platform / engine_name
        for installed, (_, digest) in ((self.manager, manager), (self.engine, engine)):
            with installed.open("rb") as body:
                require(hashlib.file_digest(body, "sha256").hexdigest() == digest,
                        f"Installed {installed.name} differs from the package manifest")
        require(self.run(self.manager, "--manager-version", "--short") == self.version,
                "Installed manager version differs from candidate")
        version = self.run(self.manager, "--al-offline", "--al-project", self.checkout, "version", "--short")
        require(version == self.version, "Selected installed engine version differs from candidate")

    def capture(self, operation, *args):
        return self.run(self.manager, "--al-offline", "--al-project", self.checkout,
                        "capture", operation, "--repo", self.checkout, *args, timeout=10)

    def prepare(self):
        require(not self.ledger.exists() and not self.summary.exists(), "First destination is already occupied")
        self.capture("prepare-publication", "--ticket-id", TICKET, "--json-output", self.ledger)
        require(not self.ledger.exists() and not self.summary.exists(), "Preparing destination seeded output")

    def observe_diagnostic_history(self):
        projects = self.home / ".agenticlens" / "capture" / "projects"
        candidates = list(projects.glob("*/registry.db"))
        require(len(candidates) == 1, "Expected one native capture registry in the isolated fixture")
        registry = candidates[0]
        require(registry.is_file() and registry.resolve() == registry and registry.is_relative_to(self.root),
                "Refusing to attach diagnostic observation outside the physical fixture registry")
        with closing(sqlite3.connect(registry.as_uri() + "?mode=rw", uri=True, timeout=5)) as db:
            db.executescript("""
                BEGIN IMMEDIATE;
                CREATE TABLE release_gate_diagnostics(sequence INTEGER PRIMARY KEY, ticket TEXT NOT NULL, record BLOB NOT NULL);
                CREATE TRIGGER release_gate_copy_diagnostic AFTER INSERT ON capture_diagnostics
                WHEN NEW.phase IN ('publication', 'reconciliation')
                BEGIN
                    INSERT INTO release_gate_diagnostics(sequence,ticket,record) VALUES(NEW.sequence,NEW.ticket,NEW.record);
                END;
                COMMIT;
            """)
        self.diagnostic_registry = registry

    def diagnostic_history(self):
        require(self.diagnostic_registry is not None, "Fixture diagnostic history observation was not installed")
        uri = self.diagnostic_registry.as_uri() + "?mode=ro"
        with closing(sqlite3.connect(uri, uri=True, timeout=1)) as db:
            rows = db.execute("SELECT sequence,record FROM release_gate_diagnostics WHERE ticket=? ORDER BY sequence", (TICKET,))
            return [dict(json.loads(record), sequence=sequence) for sequence, record in rows]

    def trust_codex_hooks(self):
        # The normal first-launch screen saves native trust for the inspected
        # hooks in the isolated home. The fixture never writes trust hashes.
        hooks_path = self.codex_home / "hooks.json"
        hooks = json.loads(hooks_path.read_text())["hooks"]
        expected = {"SessionStart", "SessionEnd", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "PostToolUse"}
        require(set(hooks) == expected, "Installed Codex hook events differ from the reviewed fixture")
        for event, groups in hooks.items():
            command = "codex-binding-wake" if event == "PostToolUse" else "codex-hook"
            require(len(groups) == 1 and len(groups[0]["hooks"]) == 1 and
                    shlex.split(groups[0]["hooks"][0]["command"]) == [str(self.manager), command],
                    f"Unexpected command or group in Codex hook {event}")
        tmux = ("tmux", "-S", str(self.root / "trust.sock"), "-f", os.devnull)
        terminal_env = dict(self.env, TERM="tmux-256color")
        self.run(*tmux, "new-session", "-d", "-x", "130", "-y", "36", "-c", self.checkout,
                 "env", "-i", *(f"{key}={value}" for key, value in terminal_env.items()),
                 self.runtime, "-s", "workspace-write")

        screen = ""
        folder_response_at = float("-inf")
        hooks_approved = False
        config_path = self.codex_home / "config.toml"
        deadline = time.monotonic() + 20
        try:
            while time.monotonic() < deadline:
                screen = self.run(*tmux, "capture-pane", "-p")
                if "Folder access" in screen and time.monotonic() - folder_response_at >= 1:
                    require("› 1. Trust and continue" in screen, "Codex workspace trust selection changed")
                    self.run(*tmux, "send-keys", "Enter")
                    folder_response_at = time.monotonic()
                if not hooks_approved and "hooks are new or changed." in screen:
                    require(f"{len(expected)} hooks are new or changed." in screen and "Trust all and continue" in screen,
                            "Codex reported an unexpected hook trust prompt")
                    self.run(*tmux, "send-keys", "Down", "Enter")
                    hooks_approved = True
                trust = tomllib.loads(config_path.read_text()).get("hooks", {}).get("state", {})
                if hooks_approved and len(trust) == len(expected) and all(
                    key.startswith(str(hooks_path) + ":") and record.get("trusted_hash", "").startswith("sha256:")
                    for key, record in trust.items()
                ):
                    return
                time.sleep(0.1)
            trace = self.root / "codex-trust.terminal"
            trace.write_text(screen)
            raise RuntimeError(f"Codex did not persist trust from its native hook approval screen. See {trace}")
        finally:
            subprocess.run((*tmux, "kill-server"), cwd=self.checkout, env=self.env, check=False,
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=5)

    def native_command(self, stage, session=None):
        if self.source == "codex":
            args = [self.runtime, "exec", "--trust", "--json"]
            if session:
                args.extend(("resume", session))
            else:
                args.extend(("--sandbox", "workspace-write"))
        else:
            args = [self.runtime, "run", "--format", "json", "--model", "al96local/mock-model", "--auto"]
            args.extend(("--session", session) if session else ("--title", f"AL96:{stage}"))
        args.append({
            "ROOT": f"AL96:ROOT Bind ticket {TICKET} with title 'Native capture release gate' and type fix. Delegate the fixture edit and verification, then wait.",
            "UNRELATED": "AL96:UNRELATED Read fixture.py. This session has no ticket.",
            "RESUME": "AL96:RESUME Verify the edited fixture.py on this normal continuation.",
        }[stage])
        return args
