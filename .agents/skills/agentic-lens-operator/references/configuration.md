# Configuration Reference and Editing

AgenticLens uses one shared project file and one selected machine-local file. Read them before proposing changes; preserve comments, unrelated values and the repository's tracking policy. Explain each proposed value and obtain approval before writing it. Existing explicit authorization satisfies that gate. Never store credentials or raw transcripts in configuration.

## File selection and path origins

Automatic shared discovery accepts `.agenticlens-config.yaml` or `.agenticlens-config.yml`, walking upward within the project/Git boundary. Both names in one directory are an error naming both files; neither has priority. Explicit supported YAML inputs accept either suffix. Init output uses a canonical shared basename and puts `.agenticlens-config.local.yaml` beside it.

Select the whole worktree-local file when present. Only if absent, use the base checkout's local file identified by the same repository's Git worktree metadata. A present invalid/unreadable file never triggers base selection. Never merge keys between local files, copy them automatically or borrow another project's policy. Guided setup runs only from the base checkout; worktree-local files are manually managed.

Configured relative paths belong to their selected file's directory; CLI-relative paths belong to the invocation directory. Absolute paths remain absolute. An explicit `~/` expands to the executing user's home or fails clearly. Resolve symlinks before later `..` components. Missing, invalid or unavailable selected paths never select another location.

Source-free ledger reads use required shared policy without requiring local sources or capture state. Optional hydration reports unavailable while preserving supported recorded facts. Operations requiring local policy fail with its exact selected path and setup guidance.

## Shared Release Authority and User Entrypoint

Shared `agenticlens.version` is an exact release pin. Preserve an existing valid pin. Fresh setup requires an approved exact value; a known exact running release may be suggested, while dev/dirty/unknown builds cannot supply a pin. Noninteractive setup must receive the approved value explicitly. Installation is a separate explicit action; ordinary dispatch never downloads, builds or probes other versions.

Configured `agenticlens.binary_dir` requires that directory's valid host-native `agentic-lens-engine-GOOS-GOARCH[.exe]`. Blank, inaccessible, absent or invalid selected output fails. When the setting is absent, select the exact pinned installed engine deliberately. Local files cannot override the shared pin or development mode. Workers retain their admitted immutable engine bytes through normal exit.

Installation-cache root order is selected-local `agenticlens.install_root`, shared `agenticlens.install_root`, then the fixed `~/.agenticlens` root. Supplied roots must be absolute or explicit `~/` paths, resolve physically and fail without another root. This cache holds engines, independently of source/session/capture state. Exact installed leaves are `bin/<version>/<os-arch>/<artifact>`.

The bootstrap reads a standard YAML mapping and validates its `agenticlens` fields. Preserve supported quoted scalars and unrelated content/comments. Duplicate mappings, malformed YAML, invalid values and oversized/non-regular input fail; never evaluate file contents. Manager and engine follow the same YAML format contract. The manager preserves caller cwd and relative arguments even with `--al-project` selecting another project.

## Sources, workspaces and identity

Generation source order is explicit `--source`, then selected-local `telemetry.default_source`, otherwise an actionable error without extraction output. Invalid explicit input fails. OpenCode path order is `--db-path`, then local `telemetry.sources.opencode.database_path`; Codex uses supported `--codex-home`, then local `telemetry.sources.codex.home`. There is no automatic source/store search or switch after failure. Set actual paths explicitly; a sanitized local example is a candidate to review, not machine discovery.

Direct OpenCode lookup uses exact `--session-id` roots or literal `--nonce` matches grouped by root. SQL wildcard characters are ordinary text. Both supplied selectors must resolve to the same root; unmatched/ambiguous selectors fail without output or newest-session selection. Codex also requires an explicit selector. Ticket text alone cannot choose a session. Explicit backfill heuristic recovery is separate and requires `--allow-fallbacks`; it must not replace known owners, choose ambiguous roots or hide incomplete exact recovery.

Workspace root order is `--workspaces`, then shared `workspaces.dir`; missing/invalid/unavailable selected roots fail. Foreign targets use their own required project policy.

`ticket.prefixes` is a required list of mappings with nonblank literal `prefix` and `description`, plus optional `ticket_type`. Trim surrounding whitespace only. Edge separators and digits/internal separators are valid; never append or strip a separator. With prefixes `T`, `T-` and `T_`, filtering `T` includes `T01`, `T-01` and `T_01`; the narrower filters include their respective forms independently. `ticket.case_insensitive` is an explicit boolean. Preserve recorded spelling and leading zeros.

Sort recognized numeric suffixes naturally; equal numeric values use the full recorded text, then stable input order for complete ties. Unmatched IDs remain visible as text after recognized IDs ascending. Type order is explicit CLI, recorded type, then agreeing applicable configured mappings. Conflicting mappings require `--ticket-type`; unavailable type remains unknown with a diagnostic, or fails under `--strict-type`.

## Shared policy

All listed containers/booleans must be explicit where required. Missing/null/wrong-type values fail; zero/false/empty containers are accepted only with the meanings below. Setup explains and obtains approval before writing values, preserving valid choices. Generation/live/source settings are validated when used, so source-free reads do not demand unused controls.

| Setting | Meaning and validity |
| --- | --- |
| `project_name` | Project label and capture namespace; nonblank. |
| `workspaces.dir` | Required workspace root. |
| `workspaces.artifacts` | Explicit list; `[]` means none. Safe nonblank ticket-relative paths; optional title/description. Nonblank title wins, otherwise basename including extension. Missing files remain unavailable content. API/UI uses required `artifacts.items`; `[]` is a genuine empty state. |
| `database.busy_timeout_ms` | Required when opening OpenCode source DBs: integer >=0. Zero means no lock waiting. All pooled source connections remain read-only and use the exact budget. |
| `live.active_threshold_seconds` | Positive integer when detecting recent activity. |
| `live.poll_interval_ms` | Explicit positive cadence for server live polling and dashboard REST fallback polling. Missing policy fails before polling; zero is invalid. Unused live policy is not required for persisted-ledger views. |
| `generation.metrics.skills/delegations/anomalies` | Three explicit booleans before generation; local config cannot override them. Respect recorded-history preservation and optional-category presence. |
| `generation.feedback.provisional_threshold_seconds` | Nonnegative integer for manual checkpoint feedback; zero disables human stderr diagnostics. |
| `generation.summary.enabled` | Explicit boolean for shared manual/capture summary policy. |
| `generation.summary.filename` | Required safe basename when configured writing is enabled; reject ledger collisions. |
| `role_mappings` | Explicit mapping; `{}` means no aliases. Nonblank string names/values; presentation/aggregation only, without changing source identity. |
| `role_order` | Explicit string list; `[]` orders observed roles alphabetically. Listed observed roles first, unlisted observed roles alphabetically; no invented rows. |

### Git contribution and classification

Require nonblank `code_metrics.base_branch`; resolve it only for requested fresh Git calculations, without branch probing or automatic fetch. Normal generation measures committed worktree HEAD against its merge base with that configured base: net branch contribution, including remaining reversions and excluding upstream-only merged changes. Staged/uncommitted work is excluded. Historical backfill targets remain explicit. Refresh/integrate the base through the delivery workflow before final delivery measurements.

Require all four lists: `code_metrics.production.include/exclude` and `code_metrics.tests.include/exclude`. `[]` includes no files or excludes none for its category. Each category independently matches included paths not excluded by its own rules. Exactly one category counts; neither is uncounted; both fail classification before affected publication. No priority or double counting. Blank/invalid glob entries fail. Patterns with `/` are repository-relative; `**` matches zero or more directories; patterns without `/` match basenames. No implicit extension, workspace, lockfile or dotfile classification.

A failed requested fresh Git calculation preserves existing ledger/summary bytes; it never publishes cached or fabricated-zero metrics. Valid batch siblings can still succeed. Read-only recorded metrics need no available Git base/history.

### Anomaly thresholds

Top-level shared `anomaly_thresholds` supplies the current detector policy. Required when the corresponding calculation is used; no substituted constants or guessed model capacity.

| Fields | Validity |
| --- | --- |
| `error_rate_min_calls` | Positive integer. |
| `error_rate_threshold`, `error_rate_critical_threshold` | Finite ratios >0 and <=1; warning <= critical. |
| `context_ceiling_warning_tokens`, `context_ceiling_critical_tokens` | Positive integers; warning <= critical. Derived context saturation needs the explicit denominator; absent policy yields unavailable read projections, preserving recorded facts. |
| `consecutive_errors_streak_warning`, `consecutive_errors_streak_critical` | Positive integers; warning <= critical. |
| `tool_thrashing_min_failures`, `tool_thrashing_critical_failures` | Positive integers; minimum <= critical. |
| `turn_usage_warning_tokens`, `turn_usage_critical_tokens` | Positive absolute token counts; warning <= critical. |
| `context_growth_warning_tokens`, `context_growth_critical_tokens` | Positive increases between successive context observations; warning <= critical. “Token usage surge” is a count rule, without a time-rate implication. |
| `stall_threshold_seconds` | Positive integer for live status classification; independent of activity window, with no clamping. |

### Summary output

Manual `--append <path>` explicitly requests a summary and overrides configured enablement/destination. Otherwise enabled shared policy writes its filename beside `--json-output`; without a ledger path, use the resolved configured ticket workspace. Disabled policy writes none absent `--append`. Resolve/validate one effective destination before authorization, locking and journaling. Preserve user content, idempotent updates, atomic recovery and deleted-destination safeguards; do not delete existing summaries.

### Optional generate output

Optional category fields preserve three states: omitted means not recorded, `[]` means a complete supported inspection found no rows, values mean recorded facts. Explicit null is invalid. Deliberately disabled output follows the established generation projection contract; it does not authorize deleting unavailable owners or inventing successful empty observations. Missing/partial/other-harness owners retain their known facts and presence. Readers cannot infer a historical omission's reason from today's settings. Recorded alert summaries are not recomputed with invented thresholds.

Ordinary current-schema reads reject former `generation_policy` and session `anomaly_availability` metadata without modifying files. Existing explicit migrate/backfill preparation can validate/remove those historical metadata fields, preserving contradiction checks and absent/empty distinctions; no new schema or automatic read repair is involved.

## Local capture policy

Capture enablement, installation and manual generation are separate choices. `enabled: false` disables automatic collection while permitting manual generation with required explicit state policy. Never silently enable capture, install hooks or alter Git settings.

| `telemetry.capture` setting | When used and validity |
| --- | --- |
| `enabled`, `state_dir` | Explicit boolean and nonblank selected-local state path for capture-state-dependent generation. No shared inheritance. |
| `refresh_ms` | Worker publication cadence, integer >=100. |
| `debounce_ms` | Worker burst coalescing, integer 0..refresh_ms; zero means no debounce delay. |
| `idle_ms`, `max_lifetime_ms` | Worker lifecycle: refresh <= idle <= lifetime <=86400000 ms. |
| `pass_timeout_ms` | Bounded attempts, integer 100..300000; also bounded by worker lifetime. |
| `max_sessions`, `max_source_bytes` | Positive budgets where consumed; supported explicit CLI overrides precede local values. Limits retain pending work and honest continuation/incomplete status. |
| `commit_warning` | Explicit boolean when advisory checks are used. False suppresses work/output; hook installation is separate and never automatic. |

Worker-free operations need no unused worker timing fields. Base-only setup preserves valid choices and fills missing approved values from `init --local-settings`; worktrees can share the base file. Review dry runs and validate both candidates before writes. `--local-only` completes local policy without replacing shared settings. `--update-capture-config --capture-state-dir PATH` is an explicit selected-local path update subject to retained-state guards; no state location is invented.

State is `state_dir/projects/<configured-project-name>/registry.db`, with exact-name metadata and separate worktree coordination. Unsafe names use the supported reversible representation, never a checkout-derived namespace. A shared project name alone grants no history access. State aliases must preserve the physical project namespace; unknown/escaping layouts fail.

Retain unsupported/unscoped-registry guards and explicit `capture relocate-state` validation, leases/fencing and data safeguards. Configuration changes do not auto-import, discard, overwrite or abandon populated state. Supported relocation is the existing protocol1-to-schema2 workflow; populated schema2 moves/renames remain unsupported. Never copy a live registry or delete it to repair setup. State survives worktree deletion without recreating deleted destinations. Pending work remains for later triggers after bounded failures/lifetime expiry.

## Installation, trust and verification

Use exact approved release artifacts from one bundle. Manager installation/update, engine installation, capture enablement, hook installation and optional Git advisory installation each have distinct effects. Preserve modified/unowned files and other projects. Preview `capture install-user --repo PROJECT --dry-run`; apply only authorized changes. Worktree discovery/status is read-only and requires no per-worktree installation.

Codex user hooks contain no frozen project/source/state paths. Only an explicit session response or source-proven parent establishes ticket binding; missing/ambiguous ownership remains unassigned. Parent links determine hierarchy; generic recorded role strings remain intact. OpenCode's project plugin remains a separate integration. Never change normal harness trust or forge callbacks for verification. Existing owned project-hook migration and interrupted installation recovery retain exact receipt/path/lease guards; user edits remain conflicts.

Installation files/version checks are not proof of collection. Record actual platform/interface/version, callbacks, binding exchange, source-proven children, successful source reads and publication. The previously observed Linux coverage is Codex 0.166.0 exec/resume and OpenCode 1.19.5 CLI/plugin; that evidence does not certify other interfaces, running/interrupted children or native Windows execution. Status/diagnostics distinguish missing binding, pending flush, parse/access failure and unresolved parent evidence; none means measured zero. Engine pin/template changes require explicit installation refresh and actual runtime verification.

`capture install-check`/`uninstall-check` remain opt-in advisory integration honoring hooksPath and preserving arbitrary hooks. Capture diagnostics retain reconciliation/publication evidence. New publication writes ledger plus the selected summary, while exact validated interrupted transactions retain their authorized recovery members. Standalone report files are ignored and untouched.

Windows keeps native manager/engine build and transport targets; loaded-manager replacement may require explicit manual update. Native Windows runtime/update certification remains unavailable. Existing bounded transports do not imply verified callback delivery: two-second registration handoff, five-second Windows startup transports, 64-KiB output limits and owned-process-tree cleanup are transport boundaries, independent of actual harness evidence. No production Python/Node dependency is introduced.
