# CLI Reference

Use `agentic-lens --al-project <PROJECT>` through the installed native manager, including from project subdirectories. Ordinary dispatch is offline; `--al-offline` explicitly marks integration use. Both controls precede native arguments and are stripped. Native `capture --repo` must belong to the same selected project; use the explicit prefix when cwd belongs elsewhere. Arguments, stdin, JSON stdout and native exit codes are preserved. POSIX dispatch uses exec.

The default stable path is `~/.local/bin/agentic-lens`; native Windows uses `~/.agenticlens/agentic-lens.exe`. Installed integrations use its absolute path and their exact project context. They never depend on a repository binary or an interactive PATH. See [release authority](configuration.md#shared-release-authority-and-user-entrypoint).

Missing/invalid pins, unsupported platforms and missing exact cache leaves fail visibly without downloads, cache writes, latest selection or environment fallback. Repair explicitly with setup `download.sh --project <PROJECT>` or native `download.ps1 -Project <PROJECT>`, after authorization already covers those effects. The installer verifies checksums/native identity/protocol with rollback and preserves edited files/other versions. Local native repair requires an explicit expected SHA256. No Go/Node/Python/yq runtime is needed. Missing scripts require restoring the complete setup/operator bundle at one selected tag; do not reinitialize existing shared/local configuration.

Missing setup scripts must be restored from the existing shared pin; fresh setup requires an approved exact release. Follow the [setup skill](../../agentic-lens-setup/SKILL.md), preserving the selected bundle revision. Never repair from moving main or combine revisions.

## Command Map

| Command | Use |
|---|---|
| `agentic-lens` or root flags | Scan ledgers, show a ticket's metrics, or inspect anomalies. |
| `--manager-version --json` | Inspect installed manager identity without selecting an engine. |
| `install <EXACT_VERSION>` | Explicitly provision an engine; no project pin changes. |
| `self-update --version <EXACT_VERSION>` | Explicitly update the manager without changing engine pins. |
| `versions` | Inspect installed manager and cached engine versions. |
| `init` | Create shared and machine-local configuration files. |
| `generate` | Extract session telemetry and Git facts into a metrics ledger. |
| `capture` | Install owned lifecycle adapters, bind exact sessions, inspect local state and reconcile source snapshots. |
| `backfill` | Recompute historical facts, recover source-backed migration steps, or migrate schema. |
| `migrate` | Apply stepwise schema changes to existing workspace ledgers. |
| `server` | Run the embedded dashboard and live observability service. |
| `version` | Identify the installed build and schema version. |
| setup `download.sh --project PROJECT --force` | Explicitly repair the exact shared-pin cache leaf with rollback. |

## Root Scan and Ticket Inspection

```bash
agentic-lens [--workspaces DIR] [--list] [--ticket-id ID] [--anomalies] [--append FILE] [--json-output FILE]
```

- `--workspaces DIR`: root containing ticket directories and `agentic_metrics.json`; uses explicit --workspaces first, otherwise required shared workspaces.dir; no conventional directory is selected.
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
agentic-lens --al-project PROJECT init --settings SHARED.yaml --local-settings LOCAL.yaml --workspaces DIR --dry-run
```

Review all candidate values before applying. Shared/local YAML candidates use `.yaml` or `.yml`; explicit candidate filenames need not be discoverable project names. Runtime project discovery uses only canonical names.

- `--settings PATH`: explicit approved shared candidate, including prefix mappings/descriptions, base ref, independent category lists and consumed runtime/generation policy.
- `--local-settings PATH`: explicit approved local choices; fills missing values without replacing valid existing choices.
- `--workspaces DIR`: required explicit nonblank root for new shared configuration. Relative origins follow the selected output file for configured values and invocation cwd for CLI paths.
- `--project-name NAME`: explicit candidate override for project label/capture namespace. Changing a populated namespace is not state relocation.
- `--agenticlens-version EXACT_TAG`: approved exact pin override; otherwise retain the valid existing/candidate pin. Dev/unknown builds cannot invent a release.
- `--output FILE`: default `.agenticlens-config.yaml`; an explicit directory may use either canonical basename. Blank/arbitrary basenames or conflicting canonical files fail before writes. The local file is beside the shared output.
- `--capture-enabled true|false`, `--state-dir`, `--source`, `--codex-home`, `--db-path`: explicit local setup choices. Enabling workers requires their complete approved timing/budget policy; no integration is installed or activated by writing config.
- `--local-only`: complete missing local choices while preserving valid shared/local values; guided setup remains base-only.
- `--complete-scaffold`: complete only an exact unchanged installer-owned scaffold; not combined with `--force`.
- `--dry-run`: preview validated candidate outputs without writing.
- `--force`: explicit shared replacement, requiring authorization for that replacement; never a repair for absent worktree-local files.

Validate both outputs before writing. Keep machine-local configuration ignored according to repository policy; never commit source paths or credentials accidentally. Prefer a narrow authorized edit for an existing setting. `init --update-capture-config --capture-state-dir PATH --dry-run` previews an explicit selected-local state-path update, preserving unrelated content and using populated-state/compare-and-swap safeguards. Omitting the replacement does not invent a state directory. Retained state requires its explicit supported relocation workflow.

## `generate`

```bash
agentic-lens generate --ticket-id ID --ticket-title TITLE --ticket-type TYPE [options]
```

- `--ticket-id ID` (required): ticket/session attribution.
- `--ticket-title TITLE` (required unless safely recovered from the existing target ledger): descriptive title.
- `--ticket-type TYPE`: explicit value before recorded type, then agreeing configured prefix types; use `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, or `build`. Conflicts fail; an unavailable type stays unknown with a diagnostic unless `--strict-type` requires failure.
- `--strict-type`: fail when no valid type is available from the explicit value, recorded type or agreeing configured mappings.
- `--source opencode|codex`: choose telemetry harness. Otherwise use selected-local `telemetry.default_source`; missing or invalid selection fails without output. A failed selected source does not silently switch harnesses.
- `--nonce VALUE`: literal text identifying exactly one root; OpenCode groups matching records by root and does not treat SQL wildcards specially. Supply a literal unique value directly; do not use shell variable or command substitution.
- `--session-id ID`: exact existing root identifier; source-proven children may be discovered. For OpenCode, both supplied selectors must resolve to the same root; no newest-session tie-break or title/workspace search occurs.
- `--code-ref REF` (default `HEAD`): committed target measured from its merge base with configured `code_metrics.base_branch`; staged/uncommitted changes are excluded. Normal generation deliberately targets worktree HEAD.
- `--started-at RFC3339`: override start time when required by a recovery workflow.
- `--completed-at RFC3339`: override completion time; otherwise current UTC time is used.
- `--json-output FILE`: write/update the structured ledger. Existing ledgers are validated and merged where supported.
- `--append FILE`: explicit summary request overriding shared enablement/destination; idempotently updates its metrics table. Otherwise shared generation.summary selects a safe filename beside the ledger, or in the configured ticket workspace when no ledger path is given.
- `--db-path PATH`: explicit OpenCode SQLite source path.
- `--codex-home PATH`: explicit Codex home; otherwise selected-local telemetry.sources.codex.home, with no automatic environment/home discovery.
- `--capture-dir DIR`: capture-state override for automatic publication only. Manual generation accepts this flag for compatibility but never accesses its path or the registry. Manual generation needs project/metrics/source configuration and supported source selection, without capture installation, enablement, state or worker settings. Explicit ticket selection does not change automatic bindings. Shared publication locking and journal recovery still protect the ledger and summary.
- `--capture-auto`: automatic lifecycle publication from registered, ticket-bound owners only. Requires enabled capture configuration and valid worktree/registry ownership; cannot be combined with `--session-id` or `--nonce`.
- `--max-sessions COUNT`, `--max-source-bytes BYTES`: explicit positive automatic reconciliation budgets, taking precedence over selected local capture settings. The worker forwards admitted effective budgets to its generation child. Manual generation does not consume capture budgets; supplied values must still be positive.
- `--anomalies`: print and record anomaly results for the target ticket.

Example:

```bash
agentic-lens generate --source opencode --nonce A7F39B2C --ticket-id PAY-42 --ticket-title "Reduce checkout latency" --ticket-type perf --json-output workspaces/PAY-42/agentic_metrics.json --append workspaces/PAY-42/SUMMARY.md
```

Generation writes the requested ledger and/or summary. Confirm those destinations before execution. Use a literal nonce. For Codex, select a known session ID or literal nonce as required by that installed version's help; do not guess a session based only on ticket text.

## `capture` (native release handler)

Every lifecycle entrypoint uses the stable native manager in offline mode with matching `--al-project PROJECT` and native `--repo PROJECT`. The selected native worker and publication children retain their current executable for the pass; future manager calls observe the current shared pin. Never install a repository updater or a version-specific native path as a callback.

Capture commands return JSON by default; `capture protocol` prints native hook capability 3, distinct from persisted registry/manifest schema 2. Older capabilities cannot consume the shared named-state contract. `capture <command> --help` lists supported flags. `install` creates additive owned Codex hooks or an OpenCode plugin, an ownership manifest and a configured-name registry with exact-worktree coordination. Source is required. `--store` is the exact Codex home or OpenCode database; supply it explicitly after checking harness configuration. Direct install may resolve the selected harness's configured local path when omitted; review the resolved value before applying. `--binary PATH` must identify the exact shared-pin native cache leaf. `--wrapper PATH` selects the stable native manager entrypoint (default user location). `--repo` defaults to the current directory and `--state-dir` uses the required configured state location. Dry run previews owned file changes without writing; repeat without it only after setup changes are authorized.

For `capture install`, `--timeout` bounds setup and each required installation phase separately (default 2s). Optional harness version/features inspection shares a separate 3s allowance; unavailable inspection is recorded without consuming the following native-validation allowance. Caller cancellation still applies throughout. A native timeout or execution failure is reported separately from an observed version/protocol mismatch; required pin, protocol and wrapper checks remain mandatory. Installation can therefore take longer than one `--timeout` interval. Other capture commands retain their existing command/pass timeout semantics.

Agent interaction and binding:

- Codex lifecycle feedback supplies a binding request with exact source, session, state directory and a per-session challenge. The agent responds with the supplied native-manager `capture bind` command, adding the ticket explicitly supplied by the user. It must not guess the session or ticket. Start/prompt callbacks can deliver context; shutdown callbacks alone cannot request missing information.
- OpenCode's per-session system transform supplies the request. The `agenticlens_bind` tool accepts `ticket` and `request`; its execution context supplies the authoritative session identity. The agent cannot replace the runtime identity through tool arguments.
- When ticket context is missing, ask the user and retain an unassigned registration. A proven parent may supply inherited binding. Independent roots need their own explicit response. Model, effort, usage and cost still come from authoritative harness records; an agent assertion cannot manufacture them.

For an explicitly authorized operator repair, inspect the local request with `capture status`, verify the exact source/session identity, then bind it:

```bash
agentic-lens capture bind --source codex --store "/absolute/Codex home" --session-id "<EXACT_SOURCE_SESSION_ID>" --ticket-id PAY-42 --request "<EXACT_SESSION_REQUEST_TOKEN>"
```

Do not publish request tokens or local status dumps. Wrong/missing request tokens and conflicting ticket assignments fail. `capture request` is the native session-specific feedback primitive used by the OpenCode plugin, not a heuristic discovery command. `capture receive` accepts bounded hook JSON on stdin; event handlers hand off durable local work and do not perform a full extraction or publish ticket artifacts.

Reconciliation is explicit and bounded. Where supported, supply positive `--max-sessions` and `--max-source-bytes` overrides; otherwise the operation requires `telemetry.capture.max_sessions` and `telemetry.capture.max_source_bytes` in the selected local configuration. No session or byte budget is supplied automatically. `--timeout` defaults to 2s and can be increased up to 5m for an operator pass. Pending work/source flush is distinct from failed, unsupported, unassigned and unresolved parent evidence. Preserve previous accepted facts on source failure; repeated cumulative snapshots must not be added twice. Use manual `generate` with an explicit source/session selection for recovery publication; successful `capture reconcile` refreshes local snapshots, prints outcomes and schedules native publication when enabled.

After a crash, restore the native handler/source access if needed and rerun reconciliation against the same local state. Do not delete accepted registrations to conceal a gap. After an explicit pin change and installation, inspect protocol/status, preview `capture install` with the original source/store/destinations, resolve conflicting local edits and apply the reviewed upgrade. Installation alone does not trust hooks or prove actual callbacks; review harness trust/plugin loading and verify real session IDs, the agent response and source facts. See [support and local state limits](configuration.md#lifecycle-capture-state-and-support).

For explicit state relocation, preview `capture relocate-state --repo PROJECT --from-state-dir EXACT_OLD_DIRECTORY --to-state-dir NEW_BASE --dry-run`. The source is an exact existing registry directory; the target is the selected state base. Relocation preserves this worktree's bindings, accepted snapshots, pending generations and diagnostics, refuses live owned worker locks/unknown layouts, and retains source backups plus cutover recovery evidence. Review the reported configuration/integration actions before applying; copying a live SQLite main file or editing state_dir alone is insufficient. Schema2 scopes each worktree inside the configured-name database; the ledger remains v5. The relocation command supports only the inspected legacy registry protocol1 family into schema2 named state. Moving or renaming populated schema2 state is currently unsupported and fails without retiring the source or modifying configuration; preserve it and report the limitation rather than inventing a copy/rename repair. See [state ownership and recovery](configuration.md#lifecycle-capture-state-and-support).

Disable or remove only the owned activation entries/files:

```bash
agentic-lens capture disable --source codex --store "/absolute/Codex home" --repo "/absolute/repo" --dry-run
agentic-lens capture uninstall --source codex --store "/absolute/Codex home" --repo "/absolute/repo" --dry-run
```

Repeat the selected command without `--dry-run` when authorized. Both commands deactivate owned callbacks while retaining local registrations/manifests for recovery; they do not purge source data or final ledgers. Unrelated configuration and conflicting local edits are preserved. A handler/configuration failure leaves coding available but capture degraded. Never auto-approve or bypass harness trust to repair capture.

## `backfill`

```bash
agentic-lens backfill --source opencode|codex|both [options]
```

- `--source` (required): `opencode`, `codex`, or `both` (`all` is accepted by the current CLI). Choose sources that actually own sessions in the target ledger.
- `--workspaces DIR`: workspace root; defaults to configured directory.
- `--ticket-id ID`: restrict work to one workspace. Recommended for targeted recovery.
- `--codex-home PATH`: local Codex rollout directory; otherwise the selected-local Codex home; missing/invalid selection fails.
- `--db-path PATH`: local OpenCode SQLite DB; otherwise the selected-local OpenCode path; missing/invalid selection fails.
- `--recompute-db`: re-query OpenCode session data and tool facts from the local database. Run on the machine that has the original sessions. Pair with `--source opencode` or `both`.
- `--recompute-git`: recompute code facts and iterations from Git history. Requires both `--code-ref REF` and `--ticket-id ID`.
- `--code-ref REF`: Git ref compared against configured base; required with `--recompute-git`.
- `--migrate-schema`: upgrade ledger schema during backfill. This is an alternative combined workflow to running schema migration separately.
- `--dry-run`: preview without persisting changes.
- `--update-summaries`: request updates using the shared configured summary filename; validate the effective destination before publication.
- `--force`: rerun requested work even when apparently complete; never enables heuristic session discovery.
- `--allow-fallbacks`: explicit opt-in heuristic backfill recovery using literal title/configured-workspace evidence. Report heuristics and incomplete exact recovery; never choose ambiguous roots, overwrite known owners or replace another harness.

For pending `source_backfill`, use the source-bearing developer's machine, the correct harness, and the appropriate local source path. If no original source is available or the source cannot be found, leave the ledger partial and coordinate with the named session owner; do not manually clear migration state.

Examples:

```bash
agentic-lens backfill --source opencode --ticket-id PAY-42 --recompute-db --dry-run
agentic-lens backfill --source opencode --ticket-id PAY-42 --recompute-db --update-summaries
agentic-lens backfill --source codex --codex-home ~/.codex --ticket-id PAY-42 --dry-run
agentic-lens backfill --source both --ticket-id PAY-42 --migrate-schema --dry-run
```

The non-dry-run forms modify metrics JSON and may update summaries. The `--dry-run` preview does not complete migration. Completed migration metadata is removed after its audit is copied to the command report and printed with the ledger path. Pending/partial state remains resumable. Backfill creates no backups; retain ledger originals in Git or another backup before execution. For a schema upgrade with rollback copies, run `migrate --backup` first.

## `migrate`

```bash
agentic-lens migrate [--workspaces DIR] [--ticket-id ID] [--target-version N] [--dry-run] [--backup] [--update-summaries]
```

- `--workspaces DIR`: workspace root.
- `--ticket-id ID`: limit migration to one workspace; omit to scan all matching workspaces.
- `--target-version N` (default: installed CLI's current schema version): target schema integer. Do not select a lower or arbitrary version without inspecting release guidance; migrations may not support downgrade.
- `--dry-run`: report planned changes without writing ledgers.
- `--backup`: save an original `.bak` copy before modifying a ledger.
- `--update-summaries`: request shared configured summary updates after migration.

`migrate` changes ledger schema, not the source harness database or installed binary. Completed migrations with no pending steps omit the `migration` object; running this command on an already current ledger also removes its completed metadata and leaves other recorded values intact. The command prints completed steps and loss accounting with each ledger path, including during cleanup. Save that console output when the audit is needed later. Ordinary reads never perform this cleanup. `--backup` prints the actual backup path after creation; restore that `.bak` file over the ledger to revert its changes. Without `--backup`, recovery requires Git or another retained original. Console accounting cannot reconstruct discarded payloads. Dry runs create no backups and change no ledgers. A successful shape migration may intentionally leave `migration.status: partial` and `source_backfill` pending. Follow with source-backed `backfill` on machines containing original sessions. Treat the ledger as incomplete until pending migration steps clear. Use `--dry-run` to inspect first; actual migration writes JSON and optionally summaries/backups.

## `server`

```bash
agentic-lens server [--port N] [--workspaces DIR] [--db-path PATH]
```

- `--port N`: listen port, 1–65535. Explicit flag overrides deliberate default `8085`; no environment lookup. Invalid/in-use selected ports fail without probing.
- `--workspaces DIR`: workspace root; must exist, else server startup fails.
- `--db-path PATH`: OpenCode SQLite source override for live session monitoring.

Starts the embedded UI and HTTP/SSE service; leave the process running while using the dashboard. Use the actual advertised `http://127.0.0.1:<port>/` after successful bind. The server reads session data from the configured local SQLite DB using read-only access. Unavailable optional sources preserve historical dashboard data and expose unavailable hydration; required live policy errors clearly. Stop with the normal interrupt for the shell environment.

## `version`

```bash
agentic-lens version [--json|--short]
```

- no flag: human-readable version/build metadata.
- `--json`: structured version, commit, build date, and schema metadata; use this to identify matching source.
- `--short`: version string only; useful for release tag matching.

No config or ledger mutation occurs.

## Explicit upgrades and diagnostics

Runtime `update` does not install or choose a release. Change shared `agenticlens.version` explicitly, install that exact release through setup and preview an owned `capture install` refresh. Preserve source/state paths and human trust. Changed project template/version/source provenance or an incompatible/unowned wrapper invalidates live setup verification until a new actual callback/binding/source exchange is observed. Other version/platform cache leaves and in-flight native workers remain intact.

`capture diagnostics --repo PROJECT` reports durable owner refresh/publication evidence; `capture verify` reports actual installation-specific callback/agent/source verification. Status/check are read only and do not create state or start extraction/workers. `capture check` is an advisory bounded observer; unknown caller identity, conservative pending markers and normal accepted-snapshot activity tails alone do not warn. Missing bindings/metrics and genuine refresh/publication failures remain actionable.

Explicit `capture install-check --repo PROJECT --wrapper ENTRYPOINT --dry-run` previews the optional Git pre-commit warning. `uninstall-check` removes unchanged owned content. Existing arbitrary hooks cause safe refusal/manual composition guidance; effective hooksPath is honored without changes. Runtime Git scope comes from the committing worktree. Local `telemetry.capture.commit_warning: false` disables its output. AgenticLens's own check never blocks the commit.

Publication retains canonical output locks and recoverable journals. Existing authorized ledger/report/summary transactions can be recovered before publishing the supported ledger/summary set; no new report/digest claim is produced. Preserve interrupted journals for explicit recovery instead of deleting failure evidence. Deleting local capture state can lose bindings and pending observations; the native cache location does not change state retention.

### Explicit installation consistency and completion scope

Validated byte-identical native cache leaves retain their filesystem identity, including force-identical installation and in-place permission repair, so another project's admitted worker is not retired. A changed-content repair remains explicit. Pre-apply validation/activation failures roll back owned changes. Once external capture apply may have committed, an incomplete result, output-limit or timeout failure retains validated native/wrapper files and recoverable prior backups; it reports incomplete setup and directs status/owned reinstall recovery rather than claiming complete rollback or removing handlers referenced by hooks. Resolve those diagnostics before declaring setup healthy.

Printed scaffold completion runs inside the selected project and restores the caller's location. The native manager's --al-project selects release authority; it does not change ordinary native working-directory semantics. Keep the printed scoped command when supplying project facts, so shared and local configuration writes cannot target another same-version scaffold in the caller's directory.

Read-only inspection is paginated without evicting stored history. `capture status`/`diagnostics` expose diagnostics_continuation, current_diagnostics_continuation and diagnostic_bytes_read when work remains; continue history with --before-diagnostic <diagnostics_before>. --max-sessions bounds registration/history work and --max-source-bytes also bounds decoded diagnostic payload. Check/verify inspect relevant current ticket/owner heads; use diagnostic continuations to investigate older failures rather than clearing history or treating an incomplete page as healthy.

`diagnostic_metadata_complete: false` means current lookup metadata is missing, partial or behind recorded diagnostics. Historical records remain available, but current health cannot be certified. Run `capture reconcile --repo PROJECT --ticket-id TICKET` for bounded mutable maintenance; repeat if its progress remains partial, increasing `--max-sessions`, `--max-source-bytes` or `--timeout` when needed. Inspection and initial registration never backfill old payloads. New facts retain their sequence order while maintenance resumes. If a history page decodes no rows because its first record exceeds the byte budget, increase the budget and retry the same cursor.

Older reconciliation records can also lack a whole-attempt outcome. Their recorded owner reads remain facts, while current collection health stays unknown until a complete reconciliation. A fatal whole-attempt failure is independent of successful individual reads and remains visible through partial retries. Use a completed reconciliation to establish recovery; the ledger's accepted numeric facts remain preserved throughout.

## `format`

Opt-in whitespace formatting using the policy shipped by the invoked installed/pinned release. It neither downloads a release nor changes the pin, installs a Git hook, or mandates a CI check. No source stores, capture registry, hydration or network are needed. Facts, unknown members, scalar lexemes, schema and timestamps inside the ledger remain unchanged; summaries and retired capture reports remain untouched. Interrupted publication journals must be recovered with the original authorized generator before formatting.

Exactly one selector is required:

- `agentic-lens format --all`: all discovered ticket ledgers. An empty root succeeds with an explicit zero-ledger result.
- `agentic-lens format --ticket-id AL-42`: one existing ticket ledger.
- `agentic-lens format --ticket-ids AL-42,AL-43`: an explicit set; trims comma-separated items, rejects empty items and deduplicates IDs.

Add `--workspaces DIR` to override the configured root. `--dry-run` reports differences without writing. `--check` also writes nothing and exits 1 for formatting differences, 2 for selection/input/I/O failures, or 0 when every selected ledger already matches. These two modes are mutually exclusive. Default writes only changed files, preserves permissions and reports per-target results and totals. All selected inputs are validated before any replacement; later I/O failures return a nonzero partial result. Symbolic aliases within the workspace are coordinated; escaping aliases and unsafe hard-linked replacements fail explicitly. Only expected agentic_metrics.json files are selected; use disposable fixtures when evaluating the command instead of mass-formatting committed history.
