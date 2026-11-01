# CLI Reference

Use the canonical executable path `.agents/skills/agentic-lens-operator/bin/agentic-lens` (or `agentic-lens.exe` on Windows) for every command. This is the path installed by the setup skill. If it is missing, explain that setup is needed rather than searching other directories or silently downloading or updating it. Quote the path if the repository path contains spaces.

Examples below use `agentic-lens` as shorthand for the executable path you resolved. To confirm syntax for the installed version, run `<binary> help`, `<binary> help <command>`, or `<binary> <command> --help` where supported. CLI behavior can evolve; prefer installed help over this reference when it differs, and inspect matching release source only as a last resort.

## Command Map

| Command | Use |
|---|---|
| `agentic-lens` or root flags | Scan ledgers, show a ticket's metrics, or inspect anomalies. |
| `init` | Create shared and machine-local configuration files. |
| `generate` | Extract session telemetry and Git facts into a metrics ledger. |
| `capture` | Install owned lifecycle adapters, bind exact sessions, inspect local state and reconcile source snapshots. |
| `backfill` | Recompute historical facts, recover source-backed migration steps, or migrate schema. |
| `migrate` | Apply stepwise schema changes to existing workspace ledgers. |
| `server` | Run the embedded dashboard and live observability service. |
| `version` | Identify the installed build and schema version. |
| `update` | Replace installed platform binaries with a selected public release. |

## Root Scan and Ticket Inspection

```bash
agentic-lens [--workspaces DIR] [--list] [--ticket-id ID] [--anomalies] [--append FILE] [--json-output FILE]
```

- `--workspaces DIR`: root containing ticket directories and `agentic_metrics.json`; defaults to configured `workspaces.dir`.
- `--list`: list discovered ledgers with duration, cost, context health, iterations, anomalies, and title. Do not combine this with a ticket-specific inspection expectation; it exits after listing.
- `--ticket-id ID`: select a ledger. Without this, the latest-updated ledger is selected.
- `--anomalies`: print anomaly diagnostics for the selected ledger and exit.
- `--append FILE`: update a Markdown file's metrics table for the selected ledger. This writes the file.
- `--json-output FILE`: write selected computed metrics as JSON. This writes the file.
- `-v`, `--version`: show build information instead of scanning.

Examples:

```bash
agentic-lens --list --workspaces "alspaces"
agentic-lens --ticket-id AL-15 --anomalies --workspaces "alspaces"
agentic-lens --ticket-id AL-15 --append "alspaces/AL-15/SUMMARY.md"
```

Scanning requires a valid project configuration and an existing workspace directory containing at least one valid ledger. `--append` and `--json-output` are write operations; get authorization before using them unless already directly requested.

## `init`

```bash
agentic-lens init --project-name NAME --ticket-prefixes LIST --src-folders LIST [--workspaces DIR] [--db-path PATH] [--output FILE] [--force] [--dry-run]
```

- `--project-name NAME` (required): dashboard/project display name.
- `--ticket-prefixes LIST` (required): comma-separated identifiers, for example `PAY,OPS`. Prefixes are normalized to uppercase.
- `--src-folders LIST` (required): comma-separated Git source roots, for example `backend,web/src`.
- `--workspaces DIR` (default `workspaces`): workspace root.
- `--db-path PATH` (default `~/.local/share/opencode/opencode.db`): OpenCode database path in shared config for compatibility; prefer per-machine source paths in local config.
- `--output FILE` (default `.agenticlens-config.yaml`): destination for shared project config. It cannot target `.agenticlens-config.local.yaml`.
- `--dry-run`: print both generated YAML files without writing.
- `--force`: overwrite existing shared and local config files. This replaces both outputs and can erase local path settings; inspect both files first and obtain authorization.

`init` creates `.agenticlens-config.yaml` and `.agenticlens-config.local.yaml`. It does not edit `.gitignore`; ensure the machine-local file is ignored according to repository policy. Prefer a surgical edit for an existing installation rather than `init --force`.

## `generate`

```bash
agentic-lens generate --ticket-id ID --ticket-title TITLE --ticket-type TYPE [options]
```

- `--ticket-id ID` (required): ticket/session attribution.
- `--ticket-title TITLE` (required unless safely recovered from the existing target ledger): descriptive title.
- `--ticket-type TYPE` (required unless inferred from configured prefix; use a conventional type such as `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, or `build`).
- `--strict-type`: fail if a valid type is omitted or cannot be inferred.
- `--source opencode|codex`: choose telemetry harness. Otherwise use machine-local `telemetry.default_source`, then OpenCode. A failed selected source does not silently switch harnesses.
- `--nonce VALUE`: exact run nonce for OpenCode session discovery. Supply a literal unique value directly; do not use shell variable or command substitution.
- `--session-id ID`: explicit root session identifier. Use when known; child sessions may be discovered according to the source adapter.
- `--code-ref REF` (default `HEAD`): committed Git ref used for code metrics, compared with `code_metrics.base_branch`; uncommitted changes are excluded.
- `--started-at RFC3339`: override start time when required by a recovery workflow.
- `--completed-at RFC3339`: override completion time; otherwise current UTC time is used.
- `--json-output FILE`: write/update the structured ledger. Existing ledgers are validated and merged where supported.
- `--append FILE`: idempotently update the Markdown metrics table.
- `--db-path PATH`: explicit OpenCode SQLite source path.
- `--codex-home PATH`: explicit Codex home/rollout directory; defaults to `$CODEX_HOME` or `~/.codex`.
- `--capture-dir DIR` (default `.agenticlens-capture`): local lifecycle registry to reconcile before final generation. Use the installation's exact state directory. Registered sessions bound to the ticket are refreshed from their actual source, including independent roots and both harnesses; no ownership is inferred from the directory. When a registry exists, generation writes `capture_report.json` beside the ledger, with snapshot cutoff, outcomes and ledger digest.
- `--anomalies`: print and record anomaly results for the target ticket.
- `--allow-empty`: permits empty generation with no root session/invocations; intended as an exceptional/testing escape hatch, not normal recovery.

Example:

```bash
agentic-lens generate --source opencode --nonce A7F39B2C --ticket-id PAY-42 --ticket-title "Reduce checkout latency" --ticket-type perf --json-output workspaces/PAY-42/agentic_metrics.json --append workspaces/PAY-42/SUMMARY.md
```

Generation writes the requested ledger and/or summary. Confirm those destinations before execution. Use a literal nonce. For Codex, select a known session ID or literal nonce as required by that installed version's help; do not guess a session based only on ticket text.

## `capture` (native release handler)

Resolve the native executable beside the wrapper, for example `agentic-lens-linux-amd64`, `agentic-lens-darwin-arm64` or `agentic-lens-windows-amd64.exe`. Installed callbacks call it directly. A manual `capture` invocation through the wrapper also bypasses missing-binary downloads and all update checks; absent native handlers fail visibly. Never put the updating wrapper in hook configuration.

```bash
agentic-lens capture install --source codex --store "/absolute/Codex home" --repo "/absolute/repo" --state-dir "/absolute/repo/.agenticlens-capture" --dry-run
agentic-lens capture status --state-dir "/absolute/repo/.agenticlens-capture"
agentic-lens capture reconcile --ticket-id PAY-42 --state-dir "/absolute/repo/.agenticlens-capture" --timeout 30s
```

Capture commands return JSON by default; `capture protocol` prints the handler protocol number. `capture <command> --help` lists supported flags. `install` creates additive owned Codex hooks or an OpenCode plugin, an ownership manifest and an ignored local registry destination. Source is required. `--store` is the exact Codex home or OpenCode database; supply it explicitly after checking harness configuration. Direct install may resolve the selected harness's configured local path when omitted; review the resolved value before applying. `--binary PATH` selects a compatible native handler; the default is the running executable. `--repo` defaults to the current directory and `--state-dir` to `.agenticlens-capture`. Dry run previews owned file changes without writing; repeat without it only after setup changes are authorized.

Agent interaction and binding:

- Codex lifecycle feedback supplies a binding request with exact source, session, state directory and a per-session challenge. The agent responds with the supplied native `capture bind` command, adding the ticket explicitly supplied by the user. It must not guess the session or ticket. Start/prompt callbacks can deliver context; shutdown callbacks alone cannot request missing information.
- OpenCode's per-session system transform supplies the request. The `agenticlens_bind` tool accepts `ticket` and `request`; its execution context supplies the authoritative session identity. The agent cannot replace the runtime identity through tool arguments.
- When ticket context is missing, ask the user and retain an unassigned registration. A proven parent may supply inherited binding. Independent roots need their own explicit response. Model, effort, usage and cost still come from authoritative harness records; an agent assertion cannot manufacture them.

For an explicitly authorized operator repair, inspect the local request with `capture status`, verify the exact source/session identity, then bind it:

```bash
agentic-lens capture bind --source codex --store "/absolute/Codex home" --state-dir "/absolute/repo/.agenticlens-capture" --session-id "<EXACT_SOURCE_SESSION_ID>" --ticket-id PAY-42 --request "<EXACT_SESSION_REQUEST_TOKEN>"
```

Do not publish request tokens or local status dumps. Wrong/missing request tokens and conflicting ticket assignments fail. `capture request` is the native session-specific feedback primitive used by the OpenCode plugin, not a heuristic discovery command. `capture receive` accepts bounded hook JSON on stdin; event handlers hand off durable local work and do not perform a full extraction or publish ticket artifacts.

Reconciliation is explicit and bounded. `--max-sessions` defaults to 256 and `--max-source-bytes` to 64 MiB per pass; `--timeout` defaults to 2s and can be increased up to 5m for an operator pass. Pending work/source flush is distinct from failed, unsupported, unassigned and unresolved parent evidence. Preserve previous accepted facts on source failure; repeated cumulative snapshots must not be added twice. Use `generate --capture-dir` for final publication; `capture reconcile` only refreshes local snapshots and prints outcomes.

After a crash, restore the native handler/source access if needed and rerun reconciliation against the same local state. Do not delete accepted registrations to conceal a gap. After `update`, inspect native protocol/status, preview `capture install` with the original source/store/destinations, resolve conflicting local edits and apply the reviewed upgrade. Installation alone does not trust hooks or prove actual callbacks; review harness trust/plugin loading and verify real session IDs, the agent response and source facts. See [support and local state limits](configuration.md#lifecycle-capture-state-and-support).

Disable or remove only the owned activation entries/files:

```bash
agentic-lens capture disable --source codex --store "/absolute/Codex home" --repo "/absolute/repo" --state-dir "/absolute/repo/.agenticlens-capture" --dry-run
agentic-lens capture uninstall --source codex --store "/absolute/Codex home" --repo "/absolute/repo" --state-dir "/absolute/repo/.agenticlens-capture" --dry-run
```

Repeat the selected command without `--dry-run` when authorized. Both commands deactivate owned callbacks while retaining local registrations/manifests for recovery; they do not purge source data or final ledgers. Unrelated configuration and conflicting local edits are preserved. A handler/configuration failure leaves coding available but capture degraded. Never auto-approve or bypass harness trust to repair capture.

## `backfill`

```bash
agentic-lens backfill --source opencode|codex|both [options]
```

- `--source` (required): `opencode`, `codex`, or `both` (`all` is accepted by the current CLI). Choose sources that actually own sessions in the target ledger.
- `--workspaces DIR`: workspace root; defaults to configured directory.
- `--ticket-id ID`: restrict work to one workspace. Recommended for targeted recovery.
- `--codex-home PATH`: local Codex rollout directory; otherwise local config, `$CODEX_HOME`, or `~/.codex`.
- `--db-path PATH`: local OpenCode SQLite DB; otherwise local config and then configured/default OpenCode path.
- `--recompute-db`: re-query OpenCode session data and tool facts from the local database. Run on the machine that has the original sessions. Pair with `--source opencode` or `both`.
- `--recompute-git`: recompute code facts and iterations from Git history. Requires both `--code-ref REF` and `--ticket-id ID`.
- `--code-ref REF`: Git ref compared against configured base; required with `--recompute-git`.
- `--migrate-schema`: upgrade ledger schema during backfill. This is an alternative combined workflow to running schema migration separately.
- `--dry-run`: preview without persisting changes.
- `--update-summaries`: refresh Markdown tables after persisted changes.
- `--force`: rerun migration/recompute even when the ledger appears complete. Use only when specifically needed and authorized.
- `--allow-fallbacks`: permits heuristic session matching (title or workspace commit range) when exact IDs are not found. This can select the wrong session; explain the tradeoff and obtain authorization before using it.

For pending `source_backfill`, use the source-bearing developer's machine, the correct harness, and the appropriate local source path. If no original source is available or the source cannot be found, leave the ledger partial and coordinate with the named session owner; do not manually clear migration state.

Examples:

```bash
agentic-lens backfill --source opencode --ticket-id PAY-42 --recompute-db --dry-run
agentic-lens backfill --source opencode --ticket-id PAY-42 --recompute-db --update-summaries
agentic-lens backfill --source codex --codex-home ~/.codex --ticket-id PAY-42 --dry-run
agentic-lens backfill --source both --ticket-id PAY-42 --migrate-schema --dry-run
```

The non-dry-run forms modify metrics JSON and may update summaries. The `--dry-run` preview does not complete migration.

## `migrate`

```bash
agentic-lens migrate [--workspaces DIR] [--ticket-id ID] [--target-version N] [--dry-run] [--backup] [--update-summaries]
```

- `--workspaces DIR`: workspace root.
- `--ticket-id ID`: limit migration to one workspace; omit to scan all matching workspaces.
- `--target-version N` (default: installed CLI's current schema version): target schema integer. Do not select a lower or arbitrary version without inspecting release guidance; migrations may not support downgrade.
- `--dry-run`: report planned changes without writing ledgers.
- `--backup`: save an original `.bak` copy before modifying a ledger.
- `--update-summaries`: refresh workspace `SUMMARY.md` tables after migration.

`migrate` changes ledger schema, not the source harness database or installed binary. A successful shape migration may intentionally leave `migration.status: partial` and `source_backfill` pending. Follow with source-backed `backfill` on machines containing original sessions. Treat the ledger as incomplete until pending migration steps clear. Use `--dry-run` to inspect first; actual migration writes JSON and optionally summaries/backups.

## `server`

```bash
agentic-lens server [--port N] [--workspaces DIR] [--db-path PATH]
```

- `--port N`: listen port, 1–65535. Default is `8085`, overridden by valid `PORT` environment variable.
- `--workspaces DIR`: workspace root; must exist, else server startup fails.
- `--db-path PATH`: OpenCode SQLite source override for live session monitoring.

Starts the embedded UI and HTTP/SSE service; leave the process running while using the dashboard. Open `http://localhost:<port>/`. The server reads session data from the configured local SQLite DB using read-only access. If no local DB is found, static/historical dashboard features may work while live views have no source data. Stop with the normal interrupt for the shell environment.

## `version`

```bash
agentic-lens version [--json|--short]
```

- no flag: human-readable version/build metadata.
- `--json`: structured version, commit, build date, and schema metadata; use this to identify matching source.
- `--short`: version string only; useful for release tag matching.

No config or ledger mutation occurs.

## `update` (wrapper)

```bash
agentic-lens update [--version TAG]
```

- no argument: resolve and install the latest public release.
- `--version TAG`: install that release tag, such as `v1.2.3`.

The wrapper stages release candidates before replacement. If the installed host handler exposes a capture protocol, the candidate must return that same protocol or the update fails with the old binary and active hooks preserved. Existing foreign-platform binaries remain unchanged in that case until upgraded/verified on their native host. Successful replacements retain backups until the transaction completes; failed replacement restores owned binary changes. It does not migrate ledgers or modify project YAML. For existing lifecycle installations, inspect native `capture protocol`/`capture status` and review `capture install --dry-run` with the original source/store/destinations before applying any needed owned-hook upgrade; binary update alone does not verify capture compatibility. Updating is an external network and filesystem mutation. Require direct user authorization unless already specifically requested, then verify with `agentic-lens version --json`. For source inspection, use the tag matching the installed version, not the latest tag by default.

The downloader's `--force --version TAG` and `--local-bin` upgrade paths apply the same host-handler preflight. When opt-in capture setup is requested, it additionally requires the supported capture protocol before first activation and restores the previous native binary/wrapper if subsequent native setup fails. The native installer separately rolls back owned configuration writes; it must not overwrite conflicting user changes. These checks preserve protocol compatibility, not a claim that an untested harness/version or foreign-platform executable ran successfully.
