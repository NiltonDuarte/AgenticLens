# Configuration Reference and Editing

AgenticLens uses shared project configuration and optional machine-local source configuration:

- `.agenticlens-config.yaml`: project identity, ticket parsing, workspace layout, metrics rules, dashboard behavior, and role presentation. Usually tracked.
- `.agenticlens-config.local.yaml`: machine-specific telemetry source locations and default source. Usually ignored by Git; verify the adopting repository's policy before changing or committing it.
- `.agenticlens-config.local.example.yaml`: portable sanitized example when present.

Never put credentials or secrets in either file. Local source paths identify the current machine's own OpenCode database or Codex home; they do not migrate or copy that data.

## Read Before Editing

1. Locate both files and read the relevant sections, plus any example file.
2. Determine which setting belongs in shared config versus local config. Do not move machine-specific paths into a tracked shared file.
3. Preserve comments, unrelated keys, types, and existing values. Change only the requested key.
4. Explain the exact file/key/value and expected effect. Edit only if the user directly requested or already allowed that change; otherwise ask permission.
5. After an authorized edit, inspect the diff and check YAML syntax with an available YAML parser. Use a config-consuming read-only command such as `agentic-lens --list` only when a valid ledger/workspace exists. There is no dedicated config validation command.

Do not use `agentic-lens init --force` to change one setting: it overwrites shared and local generated configuration. `init --dry-run` previews generated defaults but is not a round-trip editor and may not preserve existing custom values.

## Source Selection and Path Precedence

For telemetry extraction, explicit command flags take precedence over machine-local source paths. The default harness is selected by `generate --source`, then `telemetry.default_source`, then OpenCode. If the selected source is unavailable, the CLI fails rather than silently switching to another harness.

OpenCode uses a SQLite database path; Codex uses a home directory containing rollout files. Relevant keys:

```yaml
telemetry:
  default_source: opencode
  sources:
    opencode:
      database_path: ~/.local/share/opencode/opencode.db
    codex:
      home: ~/.codex
```

OpenCode path resolution generally prefers CLI override, machine-local `telemetry.sources.opencode.database_path`, legacy shared `database.path`, then adapter default. For Codex, use `--codex-home`, local `telemetry.sources.codex.home`, `$CODEX_HOME`, then `~/.codex` as supported by the installed version. Paths are local to the machine executing the CLI.

## Shared Configuration Keys

| Key | Purpose and guidance |
|---|---|
| `project_name` | Project label shown in CLI/dashboard. |
| `ticket.prefixes` | Prefix strings or objects with `prefix`, `description`, optional default `ticket_type`. |
| `ticket.case_insensitive` | Whether ticket prefix matching ignores case; default true. |
| `ticket.custom_patterns` | Additional regexes for nonstandard ticket IDs. |
| `workspaces.dir` | Root directory containing ticket workspaces and ledgers. |
| `workspaces.artifacts` | Relative artifact paths and labels shown in ticket detail. |
| `database.path` | Legacy/shared OpenCode DB path; prefer local source path for per-machine setup. |
| `database.query_timeout_ms` | Query timeout in milliseconds. |
| `database.busy_timeout_ms` | SQLite lock wait timeout in milliseconds. |
| `database.read_only` | OpenCode DB access mode; preserve read-only behavior unless the user asks for an explicitly supported change. |
| `live.active_threshold_seconds` | Window used to classify sessions as active. |
| `live.poll_interval_ms` | Live server polling interval. |
| `live.max_trace_buffer_messages` | Accepted legacy setting; does not truncate live trace responses. |
| `live.anomaly_thresholds.*` | Stall seconds, tool error ratio, context exhaustion ratio, and other supported anomaly limits. |
| `code_metrics.base_branch` | Git comparison base; when omitted, the CLI probes common upstream/main/master refs. |
| `code_metrics.source_roots` | Directories included as source roots. |
| `code_metrics.include` / `exclude` | Glob patterns to include or omit from code metrics. |
| `code_metrics.test_patterns` / `prod_patterns` | File classification patterns. |
| `role_mappings` | Map raw agent personas to canonical reporting categories. |
| `role_order` | Preferred display order; unlisted roles follow alphabetical order. |

Use the installed version's config example and help/source as authoritative if a key or default differs. Unknown or invalid values may fail config loading; avoid inventing keys.

## Safe Editing Pattern

For a directly authorized, narrow edit, preserve the rest of the file and show the diff:

```yaml
telemetry:
  default_source: codex
  sources:
    opencode:
      database_path: ~/.local/share/opencode/opencode.db
    codex:
      home: ~/.codex
```

Do not replace the entire configuration with this excerpt. Keep machine-local files ignored by Git. Before changing a path, check that it exists and belongs to the intended local harness. If the source is missing, report it; do not silently substitute another developer's or harness's data.

## Lifecycle Capture State and Support

Lifecycle installation is separate from binary availability and shared telemetry configuration. `capture install` uses the native release handler and records owned hook/plugin entries and handler protocol under the selected local state directory. Pass the selected harness's exact `--store` and explicit `--state-dir` when setup uses non-default paths. Automatic publication settings use `telemetry.capture`; see the policy below. Source paths and native executable paths are machine-local and must be reviewed before sharing generated hook/plugin configuration.

The default state directory is `.agenticlens-capture`, ignored in the adopting repository by additive setup. It holds a local SQLite registry, binding requests, pending/accepted snapshots and ownership manifests. Keep it out of Git and avoid copying status output with request tokens into ticket artifacts. Raw transcripts remain in the harness source store; callbacks retain normalized registration data rather than prompt/tool payloads. Collection is local and hook commands do not fetch releases or invoke a model.

Only an explicit session response or proven parent establishes ticket binding. Codex sends the request through lifecycle model context; OpenCode uses a per-session system transform and runtime-identified binding tool. A missing response stays unassigned. A missing parent remains unresolved independently of whether the session itself was captured. No directory/branch/role/time matching or interface thread/task ID substitution is permitted. Harness records remain authoritative for identity, native parents and usage/model facts.

Capability baselines inspected during AL-68 development are Codex CLI 0.166.0 and OpenCode 1.19.5. Inspection of versions/types/configuration is not a capture support claim. Before marking a local path verified, record the actual platform/interface/version, trust or plugin enablement, hook payload semantics, successful agent request/response, native parent/child capture and independent reviewer roots compared against source records. Do not extrapolate to cloud interfaces, untested harness versions/platforms or another OpenCode schema family. Consult the installed handler's diagnostic output and the matching release's runtime evidence; unsupported/unverified paths stay degraded. Harness trust belongs to the user and must not be enabled or bypassed automatically.

Observed runtime coverage on Linux amd64 is limited to the following paths; it is not a promise of complete capture on every interface:

| Harness/path | Source-verified behavior | Remaining degraded coverage |
| --- | --- | --- |
| Codex 0.166.0, normal trusted exec/resume | Two independent roots completed the agent binding exchange; a completed native child was captured through its source-proven parent and SubagentStop; resume retained identity and unchanged reconciliation preserved snapshots. | SubagentStart was not observed on this native exec path. Running/interrupted children before Stop remain unverified. When source records exist, explicitly select the verified parent source ID during final generation to reconcile its native source tree; keep incomplete coverage visible. |
| Codex 0.166.0, user-started session with source classification vscode | AgenticLens context reached the agent, which requested the missing ticket; Stop registered it as unassigned without guessing. | This does not establish native-child coverage for that interface. |
| OpenCode 1.19.5, CLI/local plugin | Runtime-identified agent binding, two independent roots/native child, exact source parent/usage and unchanged repeats were verified. | Additional runtime/schema families, versions and native non-Linux execution remain unverified. |

Codex exec roots verified startup context, UserPromptSubmit binding acknowledgments, SubagentStop handoff and SessionEnd registration. The stored registration keeps its latest callback rather than a complete event history. Do not infer execution of unobserved callbacks or manufacture their payloads. Model/effort/cost remain unknown when absent from authoritative records.

`capture status` and `capture reconcile` expose registration and extraction outcomes separately from installation. Check unassigned registrations, pending writes/source flush, source access/parse failures, unsupported capabilities and unresolved parent evidence individually. None means measured zero activity. If callbacks never ran or a source record is unavailable, do not claim complete coverage. Stop/idle is a checkpoint, not ticket completion.

Automatic native publication normally updates the ledger and summary. For manual recovery, use the installation's exact `--capture-dir`. When detailed reports are enabled, inspect `capture_report.json` alongside `agentic_metrics.json`; otherwise inspect compact summary coverage and treat retained reports as stale. The separate versioned report records the snapshot cutoff and source/binding coverage limits and is tied to the ledger by digest, preserving canonical ledger schema v5. Verify the digest before relying on a previously stored report; if final report publication fails after ledger writing, keep the failure visible and regenerate against the same registry to recover. Existing other-source facts and unavailable sessions must survive refresh failures. Installed native callbacks schedule automatic publication; manual generate remains a recovery path.

`capture disable` and `capture uninstall` remove owned activation entries/files and retain local state for recovery. Review dry-run changes, preserve edited owned files until conflicts are resolved and keep unrelated hooks intact. Reinstall/upgrade uses the ownership manifest; never erase or recreate the whole harness configuration as a repair. Restore a missing compatible native binary explicitly, then verify its protocol, installation and a real callback/agent binding exchange. Binary wrapper updates stage candidates, preserve the installed host handler's capture protocol and reject incompatible replacements before changing the native file. Downloader setup failures restore prior native binary/wrapper changes as well as the installer's owned configuration rollback. Existing foreign-platform binaries remain unchanged when their compatibility cannot be probed here. These updates do not automatically migrate hooks or trust them.


### Automatic publication policy

Installed lifecycle callbacks hand off to a bounded native worker; contributing agents only answer the explicit session-ticket binding request, including the user-supplied human-readable title (`--ticket-title`) and optional conventional type (`--ticket-type`). Missing titles remain pending because schema v5 requires them. No contributing agent needs to run `generate`. The command remains available for manual recovery. Automatic publication uses the same canonical-output lock as manual generation, before reading baseline or source facts.

Settings under `telemetry.capture` resolve CLI state override, then `.agenticlens-config.local.yaml`, then portable `.agenticlens-config.yaml`, then defaults. Relative state paths resolve against `--repo`, not callback cwd. Source paths retain their existing resolver. Native callbacks capture absolute project/state/source/binary paths; review and reinstall owned hooks if those paths change. No Git hooks, CI settings or harness trust changes are installed.

```yaml
telemetry:
  capture:
    enabled: true
    state_dir: .agenticlens-capture
    refresh_ms: 15000
    debounce_ms: 500
    idle_ms: 120000
    max_lifetime_ms: 1800000
    pass_timeout_ms: 30000
    max_sessions: 256
    max_source_bytes: 67108864
    detailed_report: true
    summary: true
```

External/home state is opt-in. State is namespaced by the canonical exact project directory under `state_dir/projects/<location-key>`; symlink aliases coordinate. Distinct directories with the same ticket and source store remain separate. Unscoped legacy state is retained for explicit source-verified recovery. Copying state never transfers project ownership. Removing local state loses explicit associations, so preserve it until recovery is complete. No automatic retention cleanup removes pending associations or accepted facts.

The worker uses durable request generations and acknowledges only the generation observed before its pass. Kernel process locks release on crashes. Later callbacks restart workers; workers periodically refresh active sessions, debounce bursts, back off after failed publication, and exit after idle terminal sessions. A finite maximum lifetime transfers healthy active work to a fresh native worker. Failed publication stops at that deadline with pending work retained for the next lifecycle or reconcile event, avoiding an endless failure process chain. Check `capture status --repo PROJECT` and local `worker.log` for failures. This is native child-process scheduling, with no separately installed service, model recursion or callback network/self-update work.

The ledger remains schema v5 at `<workspaces.dir>/<ticket>/agentic_metrics.json`. Top-level `completed_at` is the observability snapshot finalization boundary, equal to artifact `updated_at`; it does not assert source-session termination or ticket merge. Source invocation completion timestamps remain unchanged. Each owner has its own source-read checkpoint; there is no simultaneous source-read guarantee. Source failures preserve accepted facts and visibly degrade coverage.

`detailed_report: false` suppresses new report production and removes the summary link. Existing reports are retained and explicitly stale. When enabled, `capture_report.json` includes the ledger digest; verify it before relying on a stored report. Compact summary counts remain available. Native/runtime support requires actual hook/plugin smoke evidence; configuration and synthetic probes alone do not establish it.

Disable semantics: YAML `enabled: false` is reloaded at worker entry, before each pass/ticket and before automatic publication commit/handoff. Owned disable/uninstall marks only that source inactive; another active source integration remains usable. A sleeping worker observes disablement on its next `refresh_ms` check. An already committing pass may finish within `pass_timeout_ms`; no new pass or worker chain follows deactivation. Dry-run never deactivates, and manual generate remains available. Accepted associations and facts are retained.

Installed state discovery uses owned project-local `.codex/agenticlens-installation.json` and `.opencode/agenticlens-installation.json` manifest mirrors plus legacy owned hook/plugin paths. These mirrors are updated transactionally with owned installation changes; preserve them with the installed configuration. A changed YAML state path is diagnosed even when callbacks retain explicit frozen paths. Default status/publication fails with the old state path rather than silently opening a new registry. Use the diagnosed `--state-dir` for status/recovery/removal or restore the configured base to the installed one and reinstall owned hooks. Explicit CLI state overrides retain precedence with a drift diagnostic. Do not copy/import registries or reset associations as a repair.
