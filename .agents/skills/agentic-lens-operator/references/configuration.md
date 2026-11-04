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

## Shared Release Authority and User Entrypoint

```yaml
agenticlens:
  version: v1.2.3 # replace with the explicitly selected published release
```

The shared exact pin is required by the common user-level wrapper. Local `agenticlens.install_root` may move the default `~/.agenticlens` installation root; native leaves are `bin/<version>/<executing-os-arch>/<artifact>`. This cache is separate from telemetry source/state/ledger paths. Local `agenticlens.version` keys are prohibited. Conflicting `AGENTIC_LENS_VERSION` or explicit installation versions fail; neither changes the shared pin. RELEASE_VERSION identifies a skill bundle only. The retired AGENTIC_LENS_BIN_DIR path override is replaced by local install_root and the installer's wrapper-only --dest option.

Bootstrap supports a plain root agenticlens block, exactly two-space plain children version/install_root, literal plain/single/double-quoted scalars, whitespace-separated comments, BOM and LF/CRLF. Exact release syntax is vMAJOR.MINOR.PATCH with optional safe prerelease/build identifiers and a 128-character limit. Installation-root scalars are limited to 4096 UTF-8 bytes. No escapes/embedded quotes, duplicate root/child keys, tab indentation, aliases/anchors/merges/tags, block scalars, quoted keys, directives/documents or flow pin mappings are supported. The unsupported-construct rule applies throughout the document. Native YAML validates the remaining configuration after selection. Inputs must be regular files of at most 1 MiB; never evaluate them. Absolute POSIX/Windows drive/UNC roots or ~/ are supported by their executing entrypoint; single-quote Windows paths containing backslashes. Project/cache traversal resolves symlinks before later .. components.

Selection walks from the caller or explicit --al-project directory to the nearest yaml/yml/json config without crossing a nearer .git. YAML/YML share the profile; a nearer JSON file fails with explicit migration guidance rather than selecting a parent YAML. Native legacy direct JSON loading remains available. Ordinary and --al-offline dispatch use only already installed exact native leaves and do not install/update. Installed agent/Git hooks use the absolute stable common wrapper. Windows-local PowerShell 5.1 dispatch/installation is tested; UNC script authorization and generic console cancellation remain separate limitations. Persistent PATH or execution-policy changes are never automatic.

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

Lifecycle installation is separate from binary availability and shared telemetry configuration. `capture install` enters through the common wrapper and uses the selected native handler and records owned hook/plugin entries and handler protocol under the selected local state directory. Pass the selected harness's exact `--store` and explicit `--state-dir` when setup uses non-default paths. Automatic publication settings use `telemetry.capture`; see the policy below. Source paths and stable wrapper/native executable paths are machine-local and must be reviewed before sharing generated hook/plugin configuration.

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

Automatic publication updates the canonical v5 ledger and SUMMARY.md coverage. No capture_report.json is produced. Durable database diagnostics distinguish refreshed/preserved owners, source attempts, failure and actual output replacement; `capture status`/`diagnostics` inspect them after restart. Deprecated detailed_report is accepted with a visible no-effect warning. Existing sidecars remain obsolete until explicitly removed by the owner. Valid available facts survive individual refresh failures.

`capture disable` and `capture uninstall` remove owned activation entries/files and retain local state for recovery. Review dry-run changes, preserve edited owned files until conflicts are resolved and keep unrelated hooks intact. Reinstall/upgrade uses the ownership manifest; never erase or recreate the whole harness configuration as a repair. Restore a missing compatible native binary explicitly, then verify its protocol, installation and a real callback/agent binding exchange. Binary wrapper updates stage candidates, preserve the installed host handler's capture protocol and reject incompatible replacements before changing the native file. Downloader setup failures restore prior native binary/wrapper changes as well as the installer's owned configuration rollback. Existing foreign-platform binaries remain unchanged when their compatibility cannot be probed here. These updates do not automatically migrate hooks or trust them.


### Automatic publication policy

Installed lifecycle callbacks hand off to a bounded native worker; contributing agents only answer the explicit session-ticket binding request, including the user-supplied human-readable title (`--ticket-title`) and optional conventional type (`--ticket-type`). Missing titles remain pending because schema v5 requires them. No contributing agent needs to run `generate`. The command remains available for manual recovery. Automatic publication uses the same canonical-output lock as manual generation, before reading baseline or source facts.

Settings under `telemetry.capture` merge portable `.agenticlens-config.yaml` with `.agenticlens-config.local.yaml` overrides. Numeric policy fields retain their defaults; `state_dir` must be explicitly configured and valid before a CLI state override can apply. Relative state paths resolve against `--repo`, not callback cwd. Source paths retain their existing resolver. Native callbacks capture absolute project/state/source/binary paths; review and reinstall owned hooks if those paths change. No Git hooks, CI settings or harness trust changes are installed.

```yaml
telemetry:
  capture:
    enabled: true
    state_dir: ~/.agenticlens/state
    refresh_ms: 15000
    debounce_ms: 500
    idle_ms: 120000
    max_lifetime_ms: 1800000
    pass_timeout_ms: 30000
    max_sessions: 256
    max_source_bytes: 67108864
    commit_warning: true
    summary: true
```

Configuration creation writes `telemetry.capture.state_dir: ~/.agenticlens/state`. Operational loading requires an explicit nonblank path after shared/local overlays; missing or invalid configuration fails even if a CLI path override is supplied. Loading never inserts a fallback, creates directories or relocates state. The directory need not exist before the first authorized write. Relative explicit paths belong to the configuration root; `~` expands using the executing user's home.

The shared database is `state_dir/projects/<configured-project-name>/registry.db`. The namespace comes from required `project_name`; unsafe literal names use the supported reversible representation, never a checkout-derived project hash. Exact name metadata rejects filesystem aliases/collisions. Worktrees sharing that name concurrently use separate exact-worktree records, bindings, source ownership, publication requests/cursors and worker coordination. For the same ticket, publication may retain unchanged committed history from a verified worktree of the same actual Git repository. New capture and workers remain isolated and never acknowledge another worktree's pending work. A shared project name alone grants no history access.

Aliases for the configured state base are supported. An alias inside its project namespace must stay within that physical base and preserve the `projects/<configured-project-name>` suffix; escaping or differently shaped targets are rejected. Configure the intended base explicitly when using another layout.

Home state survives worktree deletion, but missing sources or publication destinations still require explicit recovery. No deleted checkout is recreated or rebound by name. Existing explicit project-local/external paths remain unchanged on narrow configuration update; relocation is explicit, preserves pending/failed work and retained original data, and refuses unsupported layouts or active owned workers. Copying a live database or changing a YAML path alone is not a migration. No automatic cleanup removes registrations or pending observations.

The relocation command supports only the inspected legacy registry protocol1 family into schema2 named state. Moving or renaming populated schema2 state is currently unsupported and fails without retiring the source or modifying configuration; preserve it and report the limitation rather than inventing a copy/rename repair.

The worker uses durable request generations and acknowledges only the generation observed before its pass. Kernel process locks release on crashes. Later callbacks restart workers; workers periodically refresh active sessions, debounce bursts, back off after failed publication, and exit after idle terminal sessions. A finite maximum lifetime transfers healthy active work to a fresh native worker. Failed publication stops at that deadline with pending work retained for the next lifecycle or reconcile event, avoiding an endless failure process chain. Check `capture status --repo PROJECT` and durable `capture diagnostics` for failures; logs alone are insufficient evidence. This is native child-process scheduling, with no separately installed service, model recursion or callback network/self-update work.

The ledger remains schema v5 at `<workspaces.dir>/<ticket>/agentic_metrics.json`. Top-level `completed_at` is the observability snapshot finalization boundary, equal to artifact `updated_at`; it does not assert source-session termination or ticket merge. Source invocation completion timestamps remain unchanged. Each owner has its own source-read checkpoint; there is no simultaneous source-read guarantee. Source failures preserve accepted facts and visibly degrade coverage.

`detailed_report` is deprecated and never restores sidecar output. Optional Git warnings require explicit install-check; local commit_warning: false disables output. Installed files/version checks alone do not establish capture health: verify the current manifest's actual callback, explicit binding and source exchange. Changing the shared pin/templates or installing an incompatible/unowned wrapper requires explicit refresh. A verified installer-owned compatible v1 shared-wrapper update preserves other projects' historical installation identity and does not require their reinstallation. Source/harness/platform gaps remain visible.

### Windows lifecycle transport budgets

The native registration/request/binding handoff retains its two-second default. Native Windows Codex hook declarations allow five seconds for PowerShell startup plus that handoff; POSIX declarations retain two seconds. The OpenCode Windows transport owns a five-second deadline, limits each output stream to 64 KiB, closes owned pipes on failure and requests termination of only its spawned Windows process tree through SystemRoot taskkill. Cleanup has its own one-second bound and reports failure explicitly. POSIX OpenCode transport retains its existing 1.5-second bound. Process-transport fixtures establish these measured boundaries; actual Windows Codex/OpenCode callback delivery and generic console cancellation still require separate harness evidence. Existing OpenCode supplies its runtime; no adopting-project Node installation is introduced.

## Optional generate output

Use shared `generation.metrics.skills`, `generation.metrics.delegations` and `generation.metrics.anomalies` booleans. All default to true; explicit false intentionally omits that category, including retained target history. These settings are not machine-local overrides or global redaction. See the canonical [generation configuration reference](../../../../docs/SETUP_MANUAL.md#optional-metrics-during-generation) for precedence, disabled-policy metadata, `--anomalies`, safe replacement and restoration limits.
