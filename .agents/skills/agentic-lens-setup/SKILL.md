---
name: agentic-lens-setup
description: Autonomous setup skill to bootstrap AgenticLens in OpenCode and Codex repositories, downloading precompiled release binaries, configuring project parameters, explaining telemetry features, and reviewing agent guidelines with mandatory user confirmation.
---

# AgenticLens setup

Install exact published native binaries and the setup/operator bundle in an adopting repository. The native manager handles installation/routing; a separate standalone engine handles capture, telemetry and the embedded dashboard. Neither requires Python or Node in production. OpenCode's existing plugin remains a separate integration until its global-runtime delivery.

## Read and preserve before setup

Read the selected release help and existing shared/local YAML, inspect Git metadata and installed user hooks, and preserve user changes. Never infer required settings from AgenticLens's own repository or ticket names. Explain the concrete proposed file changes and obtain explicit user confirmation before applying adopting-repository setup or editing agent guidelines. Existing explicit authorization for those changes satisfies this gate.

Shared `.agenticlens-config.yaml` owns project/ticket/workspace/metrics rules, mandatory exact `agenticlens.version` and optional `agenticlens.binary_dir`. The native manager selects host output `binary_dir/agentic-lens-engine-GOOS-GOARCH[.exe]` when valid; absent output uses the exact installed cache. Present invalid output errors. Normal execution never downloads, selects latest or changes pins. The explicit installer alone downloads verified published artifacts; local builds need no publication.

Select the whole worktree `.agenticlens-config.local.yaml` when present, otherwise the whole base-local file found through Git. Never merge keys across local files or use base shared YAML for a worktree missing its own shared YAML. Selected local configuration owns explicit `telemetry.capture.enabled`, state, cadence/limits and telemetry source locations. Relative local paths belong to the selected file directory. Invalid or incomplete required choices error without repair/fallback. Disabled capture validates configuration, then performs no capture-only source/state I/O; manual generation remains available.

## Base and worktree boundaries

Worktree-local configuration is user-managed. Do not guide its creation/copying, create it or write base configuration from a linked worktree. Enabled configured worktrees adopt on their first real event; no per-worktree install/enrollment/manifest is needed. Discovery/status stay read-only. Perform base-local setup only from the base checkout.

For base setup retain valid existing values, show retained choices and ask only for missing required choices. Recommend capture enabled; save the explicit answer. Ask for source, its actual source path and state location when missing. Noninteractive setup requires explicit values. Use `init --local-only --capture-enabled true --source codex --codex-home <ACTUAL_HOME> --state-dir <STATE>` to fill missing base-local choices without replacing valid ones. For OpenCode use `--source opencode --db-path <ACTUAL_DATABASE>` instead. Invalid existing values need an explicit correction, not replacement with defaults.

For new shared configuration, use `init --project-name <NAME> --src-folders <ROOTS> --ticket-prefixes <PREFIXES> --agenticlens-version <EXACT_VERSION>` plus the explicit local choices above. An existing valid shared file is retained. `--complete-scaffold` completes only the unchanged installer-owned scaffold. Do not use force to repair absent worktree files.

## Explicit installation and migration

Use bundled `scripts/download.sh` or `scripts/download.ps1` with an exact engine pin, project and explicit destination. A first remote bootstrap requires an exact manager release (`--manager-version`, defaulting to explicit `--version`); offline bootstrap accepts `--local-manager-bin`/`--manager-sha256` and `--local-bin`/`--sha256` for the separate engine. The shell only obtains/checks the bootstrap manager; shared YAML parsing, engine installation and routing live in native Go. Install the manager directly as the user's PATH command; there is no shell launcher, separate dispatcher executable or pointer file. Preserve edited/unowned files and other cache versions. Manager updates use `self-update --version <EXACT_MANAGER_VERSION>` and never change a project engine pin.

Bootstrap provisions the requested project engine. Outside configured projects, engine commands use the manager's exact-version companion from the default user cache. When the manager version differs, explicitly install that companion with `agentic-lens install <MANAGER_VERSION> --install-root <USER_HOME>/.agenticlens`; bootstrap never downloads an unpublished companion implicitly. `--manager-version`, `manager --help`, installation and self-update remain available without an engine.

For Codex preview `capture install-user --repo <PROJECT> --dry-run`, review the effective Codex user home's hooks and installed manager, then apply the approved command. It owns only additive unchanged groups and the home hook receipt; manager executable ownership remains separate. Callbacks contain no project/source/state path. Existing unknown/modified home callbacks, including Python workarounds, are preserved and diagnosed. Installation also checks current and Git-resolved original-base project callbacks; an active legacy callback requires explicit safe migration before global activation. Explicit `--migrate-legacy-repo <ORIGINAL_CHECKOUT>` removes only unchanged groups attested by the old project receipt, preserving state and unrelated hooks. Do not claim duplicate removal without native discovery evidence. Interrupted user installation keeps a recovery journal; subsequent installation acquires every affected current/prior home and legacy resource lease before restoring unchanged transaction-owned bytes, then recomputes the migration plan. Later user edits remain conflicts.

`capture uninstall-user` removes unchanged user groups and retains collection state, the installed manager and its ownership record. Stop automatic capture through selected local `telemetry.capture.enabled: false`; uninstalling a hook is separate from deleting retained records. The legacy project installer remains for the existing OpenCode integration. Optional Git advisory hooks remain separately opt-in through `install-check`/`uninstall-check`; never alter CI/global Git settings or arbitrary hooks.

## Binding, publication and verification

Hooks notify the exact-session worker. If current-task ticket context is explicit, answer its exact binding request directly without another confirmation. Missing/ambiguous context stays unassigned and asks for information; incidental tickets/branches/roles are not ownership. Unassigned source facts remain local. Source-proven same-task children inherit a bound parent unless explicitly bound already; parent callback IDs never become child IDs.

Workers publish during activity/checkpoints with bounded coalescing/retry. Failure preserves facts and pending work for a later trigger. Deleted destinations are not recreated or redirected; all complete/incomplete state remains until explicit cleanup. A running worker and its subprocesses keep retained immutable executable bytes through normal exit; new invocations use the preferred valid build. Status distinguishes active/preferred build, selected configuration origin and pending/failed/disabled states.

Use `capture status`, `diagnostics` and actual native inspection to separate installation from observed capture. Record actual Codex version/interface and hook sourcePath, normal user trust, root binding, source-proven child IDs, source collection and automatic publication. Never bypass/forge trust, manufacture callbacks or introduce per-worktree certification as production activation. Installed files and fixtures are not runtime proof; unavailable coverage remains explicitly incomplete.

Portable ledgers remain valid project history without the producing machine's private state or source. Preserve unavailable owners and other harnesses; complete local snapshots replace only their exact owners. Follow the canonical [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract), [configuration reference](../agentic-lens-operator/references/configuration.md) and [setup manual](../../../docs/SETUP_MANUAL.md). Never commit raw transcripts, request tokens or private paths as evidence.

The selected engine must expose `engine protocol` 1, `capture protocol` 3 and its own version identity. The manager exposes `manager protocol` 1 and `--manager-version`; it refuses the engine role probe. Engines never select or invoke another build. The manager preserves caller cwd and relative arguments even when `--al-project` selects configuration elsewhere. Callback execution uses validated project context.

Windows still has native manager/engine build and transport targets. Self-update may refuse replacement of a loaded manager with explicit manual-update guidance; native Windows runtime/update certification remains unavailable and its CI gate remains removed. Do not reintroduce shell selection or a second POSIX dispatcher to accommodate that limitation.

Engine retention validates a complete stable snapshot, host format/platform, capture interface and build identity before use. Manager installation separately validates manager format, protocol and its own identity. A changing, truncated or incompatible replacement leaves the accepted installation intact. Retained worker children execute the admitted immutable build; normal manual POSIX dispatch replaces its process image so cancellation and signal status remain intact. The Windows bootstrap preserves exact native argument transport, caller working directory and bounded setup diagnostics.

Legacy `capture disable`/`uninstall` stops further automatic passes for that integration while preserving accepted facts. A successfully migrated global user integration remains independent of the retained inactive legacy receipt. Local `enabled: false` remains authoritative for all automatic paths; historical manual source-only compatibility does not supply automatic collection limits.
