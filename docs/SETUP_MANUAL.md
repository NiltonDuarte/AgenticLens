# AgenticLens Platform Setup Manual and Telemetry Integration Guide

AgenticLens is a developer telemetry and observability platform for OpenCode and Codex workflows. It provides real-time monitoring for OpenCode and offline telemetry ingestion from either harness.

Workspace ledgers are durable project records designed for Git-based sharing across developers, machines and harnesses. A new installation's local sources enable collection and enrichment; missing another developer's private logs/capture database must not invalidate supported recorded history. See the [portable ledger contract](../TECHNICAL_GUIDELINES.md#portable-ledger-contract) for the distinction between preservation and source refresh. This is a product requirement; check the installed release's documented limits rather than assuming every compatibility path is already verified.

OpenCode traces are extracted read-only from its explicitly selected local SQLite database. The supported evidence window is OpenCode v1.19.5 (`session`/`message`/`part`) and npm `@opencode/cli` v2.2.0 (`session_v2`/`session_message`), whose `latest` tag was resolved on 2026-10-29. AgenticLens validates the complete required table and column set and automatically selects exactly one schema family. Partial, mixed, malformed, and unknown schemas fail with diagnostics; schema detection does not identify the exact runtime release. `--db-path` overrides the selected-local database path. Missing or unavailable selected paths fail without another location.

Compatibility workflows are `make opencode-compat-v1`, `make opencode-compat-v2`, and `make opencode-compat`. They require the corresponding runtime (`OPENCODE_V1_BIN` / `OPENCODE_V2_BIN` can select explicit binaries) and a local deterministic provider already running. The workflows create isolated temporary HOME, XDG data/config/state roots, project roots, and absolute `OPENCODE_DB` paths, then remove them at exit. Ordinary adapter tests use deterministic SQLite fixtures and need no OpenCode installation. Evidence fixtures are sanitized runtime-produced records, not synthesized runtime integration runs. Nonzero native-provider pricing, imported-history extraction after migration, partial migration states, duplicate handling, and releases outside v1.19.5/v2.2.0 are not verified.
Codex traces are read from rollout files beneath its explicitly selected home directory. OpenCode and Codex are independent sources; selecting one never falls back to the other.

Use this guide to install, configure, and operate AgenticLens in an adopting repository.

---

## 1. Quick Start and Setup Pathways

Use the canonical [setup skill](../.agents/skills/agentic-lens-setup/SKILL.md) for complete bundle installation, approval boundaries, full configuration, normal harness trust and actual callback/binding/source verification. Install official precompiled releases only; adopting projects need no Go/Node/Python/yq runtime.

Retain an existing approved exact release pin. For fresh setup, obtain approval for an exact published release before retrieving its bundle or writing configuration. A valid running release may be suggested; dev, dirty or unknown builds cannot select a pin. Retrieve manifests, scripts and binaries from that exact tag and report missing bundle files without mixing revisions.

Shared `.agenticlens-config.yaml` or `.agenticlens-config.yml` owns `agenticlens.version`; the native manager detects the executing platform. Installation-cache placement is separate from telemetry state. Ordinary commands and integrations never install a missing release.
The commands below require a selected release containing the complete setup bundle. Review and retrieve its installer before running it.

```sh
bash .agents/skills/agentic-lens-setup/scripts/install.sh --version <EXACT_TAG> --dest <PROJECT>
bash .agents/skills/agentic-lens-setup/scripts/download.sh --project <PROJECT> --manager-version <EXACT_MANAGER_TAG> --dry-run
bash .agents/skills/agentic-lens-setup/scripts/download.sh --project <PROJECT> --manager-version <EXACT_MANAGER_TAG>
agentic-lens --al-project <PROJECT> version --short
```

The first command bootstraps both setup/operator skills from the selected tag. Restore missing setup scripts from that tag after reviewing them. Existing configurations/files are preserved; do not fetch moving-main wrappers or moving latest artifact URLs. For a new project use explicit --init-config --version only as a bootstrap scaffold; complete reviewed shared/local candidates through native init before claiming publication readiness. Never overwrite arbitrary project/local config using --force.

AgenticLens supports macOS and Linux on amd64 and arm64. The native manager defaults to `~/.local/bin/agentic-lens`. Windows is unsupported.

Exact engines coexist in `~/.agenticlens/bin/<version>/<os-arch>/agentic-lens-engine-<os>-<arch>`. Manager and engine local artifacts require separate expected SHA256 values. The bootstrap provisions the project engine and installs the manager directly; no shell runtime launcher or dispatcher copy is installed. Native installation verifies staged bytes and preserves modified/unowned files and other versions. Installers report the required PATH directory without changing profiles.

Manager and engine versions are independent. Inspect the manager with `agentic-lens --manager-version --json`, inspect selected engine with `agentic-lens --al-project <PROJECT> version --json`, install engines using `agentic-lens install <EXACT_ENGINE_TAG>` and explicitly update the manager using `agentic-lens self-update --version <EXACT_MANAGER_TAG>`. Updating the manager never edits project pins. Outside configured projects, utility commands require the manager's same-version companion in the default user cache; install it explicitly with `agentic-lens install <MANAGER_VERSION> --install-root <USER_HOME>/.agenticlens` when needed. Native maintenance help remains available without an engine.

### Reviewed lifecycle installation and verification

```sh
agentic-lens --al-offline --al-project <PROJECT> capture install-user --repo <PROJECT> --dry-run
```

Review exact additive hooks/plugin/state-ignore changes, then repeat without --dry-run under existing applicable approval. Every installed lifecycle/Git callback uses the stable native manager with explicit project context. Internal workers retain their selected native executable/version for a pass. Capture install, version output and files prove installation only: verify actual normal harness trust, callback, exact session/request, explicit agent response and available source extraction through status/diagnostics/verify. Independent roots require independent bindings; unknown ownership remains unassigned.

After a shared pin change, explicitly install the release, preview owned setup refresh with the original store/state and repeat live verification. Old in-flight workers may finish; other cache versions remain intact. User hooks invoke the stable native manager; ordinary CLI and callbacks use the same offline resolver. No automatic trust changes or repository updater is supported. The optional install-check/uninstall-check integration is separately opt-in and advisory; it honors hooksPath, derives the committing worktree and refuses arbitrary-hook overwrites. No adopting-project CI/global Git settings are changed.

Publication writes the ledger and the summary selected by shared generation.summary policy, preserving facts and explaining coverage/freshness/failure. Reconciliation/publication evidence stays in local diagnostics. Unowned standalone report files are ignored and untouched; exact interrupted transactions retain authorized recovery. Local state deletion can lose bindings and pending observations; the native home cache does not move the capture database.

See [CLI operations](../.agents/skills/agentic-lens-operator/references/cli.md) and [release/configuration profile](../.agents/skills/agentic-lens-operator/references/configuration.md#shared-release-authority-and-user-entrypoint).

## 2. Agnostic Adoption for OpenCode Projects

Live SSE uses OpenCode; offline generation supports OpenCode and Codex. Sources are independent and never silently substituted after a failure.

### 2.1 Complete Methodology Neutrality within OpenCode

Adopters choose their own single-agent, human-assisted, hierarchical, or CI workflows, review gates, and personas. They do not inherit this repository's maintainer skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).

### 2.2 Custom Conventions and Taxonomy

Configure ticket prefixes as equal peers, any workspace root, and custom persona mappings. Trunk-based, GitFlow, GitHub flow, and short-lived branches are supported; unmapped personas can remain custom roles.

## 3. Configuration Reference

The [configuration reference](../.agents/skills/agentic-lens-operator/references/configuration.md) owns current fields, bounds, selection and editing rules. Shared discovery accepts either canonical YAML suffix, fails if both names exist in one directory and respects project/Git boundaries. Local configuration remains `.agenticlens-config.local.yaml`; keep it out of Git and never store credentials.

### Telemetry Source and Session Selection

Generation selects `--source` first, otherwise the selected-local `telemetry.default_source`, otherwise fails without output. OpenCode uses `--db-path` before local `telemetry.sources.opencode.database_path`; Codex uses supported `--codex-home` before local `telemetry.sources.codex.home`. No automatic store/harness search follows a selected failure.

Choose a root with `--session-id` or supply a unique literal `--nonce`. Direct OpenCode validates exact roots, searches nonce text literally, requires one matching root and requires both supplied selectors to agree. Ticket text alone cannot choose a session. Codex uses explicit selectors too. Root selection can discover source-proven children; it cannot invent parent identities or trust a UI task identifier as a native session ID.

Select one whole worktree-local file; only when absent, use the same repository's Git-established base-local file. A present invalid file never triggers base selection. Relative configured source paths belong to the selected file, CLI paths to invocation cwd; explicit `~/` expands or fails. Preserve physical symlink/`..` traversal. Guided setup is base-only; no local copy or per-worktree enrollment is required when the base file is shared.

Source-free recorded-ledger reads do not require local sources or capture state. Optional hydration reports unavailable and preserves accepted facts. Required extraction/Git failures prevent affected publication. Current runtime ingestion requires supported schema v6; use explicit migration for historical data. See [owned occurrences](OWNED_OCCURRENCES.md) and [cost evidence](COST_SEMANTICS.md).

### Code Classification and Interpretation

Git metrics measure the committed net contribution from merge-base(configured `code_metrics.base_branch`, target) to target. Normal generation targets worktree HEAD; historical backfill targets are explicit. Incoming upstream-only merges do not count, reverted edits contribute only their remaining net effect, and staged/uncommitted work is excluded. Refresh/integrate the base through the delivery workflow before final delivery measurements; measurement never fetches.

Production and tests each have explicit include/exclude lists. A category matches included paths not excluded by its own list. Evaluate independently: one match counts, neither is uncounted, both fail before affected publication. No priority or double counting. `[]` deliberately includes none or excludes none; missing/invalid lists fail. Patterns with `/` are repository-relative, `**` matches zero or more directories and basename patterns match basenames. No implicit roots, extensions, workspace exclusions, dotfiles or lockfile rules are added.

Ratios use the configured categories and changed-line basis; whole-tree counts answer a different question. There is no universal ratio target. See [test metrics guidance](../TECHNICAL_GUIDELINES.md#interpreting-test-metrics).

### 3.1 Reviewed Configuration Candidates

These examples illustrate syntax, not presets. Review and approve each value for the adopting project, including empty lists/mappings. Preserve valid existing values. Setup must explain manual telemetry, capture enablement and installation separately; it never silently activates or installs an integration.

```yaml
project_name: PaymentGateway
agenticlens:
  version: v1.2.3 # replace with the approved exact published release
ticket:
  prefixes:
    - prefix: PAY-
      description: Payment service development
      ticket_type: feat
  case_insensitive: true
workspaces:
  dir: tickets
  artifacts: []
code_metrics:
  base_branch: origin/main
  production:
    include: ["service/**/*.go"]
    exclude: ["**/*_test.go"]
  tests:
    include: ["service/**/*_test.go"]
    exclude: []
database:
  busy_timeout_ms: 1000
live:
  active_threshold_seconds: 60
  poll_interval_ms: 1000
anomaly_thresholds:
  stall_threshold_seconds: 30
  error_rate_min_calls: 5
  error_rate_threshold: 0.35
  error_rate_critical_threshold: 0.7
  consecutive_errors_streak_warning: 4
  consecutive_errors_streak_critical: 8
  context_ceiling_warning_tokens: 100000
  context_ceiling_critical_tokens: 200000
  turn_usage_warning_tokens: 30000
  turn_usage_critical_tokens: 60000
  context_growth_warning_tokens: 40000
  context_growth_critical_tokens: 80000
  tool_thrashing_min_failures: 3
  tool_thrashing_critical_failures: 5
generation:
  metrics:
    skills: true
    delegations: true
    anomalies: true
  feedback:
    provisional_threshold_seconds: 0
  summary:
    enabled: true
    filename: METRICS.md
role_mappings: {}
role_order: []
```

An illustrative local candidate for manual-only telemetry follows. Replace its paths with approved actual locations. Disabled automatic capture permits manual generation; worker-only timing fields are not needed until workers are enabled.

```yaml
telemetry:
  default_source: codex
  sources:
    codex:
      home: /srv/codex-home
  capture:
    enabled: false
    state_dir: ~/.agenticlens/state
    pass_timeout_ms: 20000
    max_sessions: 32
    max_source_bytes: 8388608
```

### 3.2 Applying Approved Choices

From the base checkout, preview `init --settings <SHARED_CANDIDATE.yaml> --local-settings <LOCAL_CANDIDATE.yaml> --workspaces <ROOT> --dry-run`. New shared configuration requires nonblank explicit `--workspaces`; output defaults to `.agenticlens-config.yaml`, or an explicit destination with either canonical basename. The local file is beside it. Validate both candidates before writes. Existing local completion uses `--local-only` and preserves valid values. An exact unchanged installer scaffold can use `--complete-scaffold`; arbitrary configs are not scaffolds.

Prefix mappings preserve literal separators and descriptions; scalar shortcuts are not part of the current schema. Explicit case policy applies consistently. Generation type order is CLI, existing recorded type, then agreeing configured mappings. Conflicts require an explicit type; no available type remains unknown or fails with `--strict-type`. Summary output follows shared policy, with manual `--append` overriding enablement and destination. The full [field reference](../.agents/skills/agentic-lens-operator/references/configuration.md#shared-policy) explains numerical requirements and intentional false/zero/empty values.

## 4. OpenCode Agent Workflow Integration Guide

Integrating AgenticLens into your repository requires zero changes to agent reasoning loops, model selection, or coding styles. Installed native lifecycle capture publishes automatically after explicit session-ticket binding. Agents answer the binding request; manual generation below is an explicit telemetry operation and recovery option.

### 4.1 Automatic publication and manual recovery

Verify automatic publication and visible coverage before staging task-owned deliverables and telemetry. If native publication is unavailable, recover with manual generation after implementation, testing and review, close to the final commit so the snapshot includes the available work.

### 4.2 Nonce Generation and Ingestion Command (`generate`)

For manual recovery, select the actual harness and exact root session ID or pass a unique execution nonce directly to `generate`:

```bash
agentic-lens generate \
  --source "<opencode-or-codex>" \
  --nonce "<GENERATED_NONCE>" \
  --ticket-id "<TICKET_ID>" \
  --ticket-type "<CONVENTIONAL_TYPE>" \
  --json-output "<WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json" \
  --append "<WORKSPACES_DIR>/<TICKET_ID>/<SUMMARY_FILENAME>"
```

#### Literal Nonce Mandate (FIX-5)
Generate a unique 8–16 character alphanumeric or hex nonce and pass it literally. Never use `$NONCE` or `$(...)`: OpenCode stores tool inputs before shell expansion, so searching for the expanded value cannot find that input. See [CLI flags](#agentic-lens-generate) for output and source options.

### 4.3 Conventional Commit Type Classification

AgenticLens requires explicit commit type classification to generate accurate engineering velocity rollups:
- `feat`: New product capabilities, user-facing enhancements, or API additions.
- `fix`: Defect remediation, bug fixes, and patch corrections.
- `docs`: Documentation updates, specifications, or architectural manuals.
- `chore`: Maintenance tasks, dependency updates, configuration changes, or build scripts.
- `refactor`: Code reorganization without functional behavior changes.
- `test`: Adding, updating, or fixing automated tests.
- `ci`: Continuous integration and delivery pipeline modifications.
- `perf`: Performance optimizations and latency reductions.
- `build`: Build toolchain modifications or compiler flag updates.

### 4.4 Custom Persona Role Mapping

Configure `role_mappings` as shown in the [configuration example](#31-reviewed-configuration-candidates). Tokens, costs, and durations aggregate under mapped roles while raw ledgers preserve persona identifiers.

### 4.5 Standard Repository Agent Guidelines Policy Block

Below is the standard policy block that can be added to your repository's OpenCode agent instructions file (`AGENTS.md` or `.agents/agent/*.md`):

```markdown
## AgenticLens Telemetry Guidelines

1. Ticket Tracking Mandate:
   - Every task requires an explicit ticket ID and workspace under the configured `workspaces.dir`. Missing configuration requires setup, without inventing a root.
   - Use an existing ticket when one is provided. For ad hoc work without an ID, obtain an identifier using the configured literal prefixes; do not create a planning specification or index entry under `plan/` solely for it.
   - Agents work on one ticket at a time. All turns belong to the active ticket.

2. Plan Lifecycle:
   - When a matching specification exists under `plan/to-do/`, move it to `plan/done/` when its ticket is completed.
   - Remove the ticket entry from `plan/TICKETS.md` if present. Do not create a specification or index entry just to complete this step; ad hoc tickets still have an ID and workspace.

3. Dedicated Workspace Allocation:
   - Store task documentation and metrics in `<WORKSPACES_DIR>/<TICKET_ID>/`.
   - Maintain `CONTRACT.md` (scope boundaries) and `SUMMARY.md` (current outcome, verification and limits) as required by the project's artifact list.

4. Automatic Telemetry Publication:
   - Answer the native lifecycle session-specific binding request with the explicitly supplied ticket and title; never infer ticket ownership.
   - Verify automatic ledger/summary publication and coverage before committing. Manual `generate` is available independently, using the selected harness and an exact session ID or literal nonce.
   - Verify configured-summary coverage and local capture diagnostics. Existing standalone reports are left untouched; interrupted publication recovery is explicit.
   - Stage only reviewed task-owned deliverables and telemetry together. Keep raw transcripts and local capture state out of Git.

```

---

## 5. CLI Commands and Embedded Web Dashboard Reference

AgenticLens provides a unified command-line interface and an embedded web dashboard for observability.

### 5.1 CLI Subcommands Reference

Use the [CLI reference](../.agents/skills/agentic-lens-operator/references/cli.md) and the selected release's help for exact arguments. Ordinary native-manager dispatch never installs a missing engine. Installation and self-update are explicit commands with exact approved versions.

| Command | Operation |
| --- | --- |
| `init` | Preview/apply explicit reviewed YAML candidates; preserve valid completion values. |
| `generate` | Exact source extraction and committed Git metrics; requested failures preserve existing publications. |
| root scan | Read current supported ledgers, report invalid siblings, preserve recorded history without optional sources. |
| `server --port 8085` | Embedded UI/API on 127.0.0.1; explicit port overrides deliberate default 8085. Invalid/in-use ports fail without probing. |
| `backfill` | Explicit exact-owner recovery/recomputation; heuristics require opt-in, recompute-Git requires a target. |
| `migrate --backup` | Existing audited/resumable schema transforms and supported explicit metadata cleanup. |
| `format` | Representation-only historical/current JSON formatting, without source hydration or schema changes. |
| `capture status` / `diagnostics` | Read-only installation/state/coverage evidence and bounded continuation. |

#### `agentic-lens generate`

Manual `--append` explicitly selects a summary path. Otherwise enabled shared generation.summary writes its filename beside the requested ledger, or in the resolved configured ticket workspace if no ledger path is supplied. Disabled policy writes no summary absent `--append`. Resolve one destination before publication authorization/locks/journals. Preserve user content and idempotent atomic updates; never delete existing summaries automatically.

### 5.2 Embedded Web Dashboard Views

When running `agentic-lens server --port 8080`, navigate to `http://localhost:8080` to access the interactive dashboard.

#### 1. Fleet Overview (`/`)
- Cross-ticket aggregate rollups showing total tokens consumed, total financial cost, total development wall-clock time, and average iterations per ticket.
- Context window health distribution gauge displaying percentages across HEALTHY, ELEVATED, BLOATED, and CRITICAL states.
- Interactive ticket inventory with search filtering by ticket ID, title, status, and prefix tags.

#### 2. Live Observability (`/live`)
- Real-time active agent monitoring powered by Server-Sent Events (SSE).
- Live session cards displaying active subagent personas, model in use, current turn count, token accumulation, and elapsed wall-clock duration.
- Real-time trace viewer streaming tool calls, parameter payloads, and output logs.
- Active anomaly alert banners highlighting detected stalls, elevated error rates, and context exhaustion in real time.

#### 3. Ticket Detail (`/ticket/:id`)
- Deep-dive analytics for an individual ticket workspace.
- Multi-iteration timeline breaking down SDLC passes, subagent roles, and turn progression.
- Tool invocation distribution chart detailing tool call frequencies and failure rates.
- Model usage and cost breakdown (e.g. GPT-4o, Claude 3.5 Sonnet, Gemini Flash).
- Interactive Artifact Viewer rendering configured items and titles as Markdown with raw view toggles; an empty items list shows an empty state.

#### 4. Ticket Comparisons (`/comparisons`)
- Side-by-side comparative analysis of two or more tickets.
- Differential metrics evaluating token efficiency, cost variations, tool reliability, and code diff volume per iteration.

#### 5. Metric Progression (`/progression`)
- Longitudinal trend graphs tracking agent efficiency over time.
- Rolling averages for cost-per-ticket, turn velocity, test-to-production code ratios, and context window health evolution.

---

## 6. Operational Troubleshooting and Best Practices

### 6.1 SQLite Concurrency and Busy Timeouts

OpenCode source connections are always read-only. Choose explicit nonnegative shared `database.busy_timeout_ms`: zero means no lock waiting, positive values are the lock-wait budget in milliseconds on every pooled connection. Internal writable capture state is separate. A missing/invalid required source setting fails; unavailable optional hydration preserves persisted ledgers.

### 6.2 Explicit Paths and Home Expansion

Set the actual database/home path in the selected local file or a supported CLI override. Explicit `~/` expands to the current user's home; ordinary environment-variable strings are not path discovery. Configured relative paths belong to their owning file; CLI paths to invocation cwd. Resolve symlinks before later `..`. Missing/invalid/unavailable selection never tries a conventional home path or another harness. Quote paths with spaces.

### 6.3 Executable Permissions and macOS Quarantine Flags

When installing precompiled binaries manually on Unix systems (Linux and macOS):
- Ensure the executable bit is set: `chmod +x agentic-lens`.
- If Gatekeeper blocks execution on macOS, clear quarantine flags:
  ```bash
  xattr -d com.apple.quarantine agentic-lens
  ```
- Store the binary in the selected installation directory (such as `./bin/` or `.agents/skills/agentic-lens-operator/bin/`). The repository owner decides what to track or ignore; installation/repair preserves that policy.

### 6.4 Explicit Git Base and Calculation Failures

Require shared `code_metrics.base_branch`, chosen and approved during setup. It must resolve in the target repository for fresh calculation; measurement does not fetch or try other branches. Fetch/integrate deliberately in the project's delivery workflow. A requested failed calculation preserves existing ledger/summary bytes and does not publish cached or zero metrics. Recorded-ledger reads need no Git history availability.

### 6.5 Workspace Directory Path Resolution

Explicit `--workspaces` overrides shared `workspaces.dir`; absent/invalid/unavailable required roots fail without conventional directory guesses. Relative CLI paths belong to invocation cwd and configured paths to their shared file. A foreign target supplies its own required project policy. Explicit portable formatting can operate on its selected workspace without unrelated source/capture policy. Keep agent instructions aligned with the actual configured root.

### 6.6 Worktrees and Container/Host Platforms

The stable user wrapper is shared; each physical configured project/worktree selects its own pin. No repository executable is required. Missing files in another worktree do not justify init --force. Resolve the nearest shared YAML without crossing .git and use explicit --al-project when cwd differs from native --repo. Install missing leaves explicitly; ordinary dispatch/hooks never fetch. Containers and hosts select their executing platform and may cache separate artifacts. Existing repository binaries/wrappers are preserved during owned migration. Git tracking and ignore policy stay with the owner.

## Appendix: Platform Architecture and Core Concepts

### A.1 Functional Core, Imperative Shell (FCIS) Architecture

See [architecture and subsystem responsibilities](../ARCHITECTURE.md) and the mandatory [technical boundaries](../TECHNICAL_GUIDELINES.md#2-functional-core-imperative-shell-fcis).

### A.2 Native Manager and Self-Contained Engine

Go 1.23+ builds with `CGO_ENABLED=0`; `modernc.org/sqlite` needs no C compiler or shared C libraries. Vite compiles the TypeScript dashboard into `web/dist`, embedded via `embed.FS`. The standalone engine provides CLI and embedded web UI without external runtimes; the separate native manager handles installation/routing.

### A.3 Local Immutable Telemetry Fact Ledgers (`agentic_metrics.json`)

Rather than transmitting proprietary code or execution telemetry to external cloud services, AgenticLens operates entirely locally:
- Telemetry facts are captured per ticket or task in isolated workspace directories: `workspaces/<TICKET_ID>/agentic_metrics.json`.
- Schema-v4 and schema-v5 sessions own invocation records, per-session tool summaries, session-specific anomaly summaries, and sparse direct-evidence events. Cross-session anomaly summaries (including high tool failure rates) live at ticket level without session IDs or event rows. Metrics are derived from summaries and tool totals, not by rerunning sequence detectors over sparse evidence. When OpenCode hydration is unavailable, the timeline shows committed anomaly evidence only; full Codex event hydration is not provided. Migration reports and removes legacy aggregate tool rows without reliable session attribution; re-extract from OpenCode/Codex source logs to recover those counts when available.
- Each ledger is an immutable JSON document recording session identifiers, wall-clock duration, token volume (input, output, cache-read, and cache-write tokens), financial cost evidence where known, preserving unknown cost explicitly, turn counts, tool invocation distributions, error rates, and git commit diff volumes (insertions, deletions, and affected files).
- These ledgers are committed directly to version control alongside code deliverables, establishing an auditable, repository-native history of agent engineering metrics with zero cloud egress.

#### Recovering unscoped tool totals

Preserve the old ledger first. Run `generate` for each known root session ID with `--json-output` pointing to a new, initially nonexistent file. Reuse that same new output path for later root sessions; generation merges the session facts by session and tool name. For Codex, selecting a root thread also discovers its child rollouts. For OpenCode, generate once for each root session. Do not point `--json-output` at the old ledger: its unscoped tool rows fail validation before generation can merge new facts.

```bash
cp workspaces/<TICKET_ID>/agentic_metrics.json workspaces/<TICKET_ID>/agentic_metrics.unscoped-backup.json
agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_1> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json
agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_2> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Compare the IDs and per-session tool rows with the sessions you intended to recover.
jq -r '.sessions[].id' workspaces/<TICKET_ID>/agentic_metrics.recovered.json | sort -u
jq '[.sessions[].tools[]] | length' workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Replace the active ledger only after verifying all expected sessions and counts.
mv workspaces/<TICKET_ID>/agentic_metrics.recovered.json workspaces/<TICKET_ID>/agentic_metrics.json
```

Use `--source opencode` and repeat with each OpenCode root session ID for OpenCode recovery. Verify all expected session IDs (including Codex child rollouts) and tool counts before replacing the original. An unscoped aggregate cannot be accurately divided among sessions; if the source logs are unavailable, those old aggregate counts cannot be recovered. Keep the backup and do not invent IDs or redistribute the totals.

### A.4 Real-Time Live Observability and SSE Event Streaming

In addition to post-turn historical ledgers, AgenticLens provides live observability of active agent sessions:
- The embedded server polls the local OpenCode SQLite database at the explicit positive `live.poll_interval_ms` cadence. Dashboard REST fallback polling uses the same configured interval; it does not select another cadence when policy is missing or invalid.
- Active sessions are identified based on recent turn activity within a sliding time window (`live.active_threshold_seconds`, explicit positive seconds).
- Telemetry events, tool calls, and status updates are broadcast to connected browsers via Server-Sent Events (SSE) through the `/api/live/stream` endpoint.
- Live traces use a source-revision cache to reuse unchanged snapshots, providing visibility into subagent handoffs, prompt dispatches, and tool outputs.

### A.5 Automated Anomaly Detection

AgenticLens continuously analyzes active sessions and historical ledgers against configurable statistical thresholds to detect engineering inefficiencies:
- **Session Stalls**: Sessions that remain open with no new turn activity beyond a configurable duration (`anomaly_thresholds.stall_threshold_seconds`, explicit positive seconds).
- **High Tool Error Rates**: The configuration records a minimum tool-call count, warning ratio (`error_rate_threshold`) and critical ratio (`error_rate_critical_threshold`). Ratios are explicit values between zero and one; configuration generation supplies initial values, while runtime loading requires the fields to be present.
- **Context Ceilings and Token Usage Surge**: Explicit absolute token ceilings and turn-usage/context-growth thresholds drive existing detectors; no model capacity or time rate is guessed.

### A.6 Context Window Health States

Context saturation severely impacts model reasoning quality, instruction adherence, and cost. AgenticLens classifies context health into display states based on utilization against the explicit configured denominator. Without a denominator, the projection is **UNAVAILABLE** and preserves recorded facts. These bands do not infer model capacity:
- **HEALTHY** (0% to 50% utilization): Optimal operating envelope. The model retains complete instruction adherence and context fidelity with minimal compaction pressure.
- **ELEVATED** (50% to 75% utilization): Normal working range for complex implementations. The agent operates effectively, but monitoring is recommended.
- **BLOATED** (75% to 85% utilization): High context saturation. Model reasoning degradation and prompt compaction risks increase significantly.
- **CRITICAL** (85% to 100% utilization): Severe exhaustion. Immediate session compaction or subagent delegation is advised to avoid catastrophic forgetting or dropped instructions.


### Automatic publication policy

Installed lifecycle callbacks hand off to a bounded native worker; contributing agents only answer the explicit session-ticket binding request, including the user-supplied human-readable title (`--ticket-title`) and optional conventional type (`--ticket-type`). Missing titles remain pending because the current schema requires them. No contributing agent needs to run `generate`. The command remains available for manual recovery. Automatic publication uses the same canonical-output lock as manual generation, before reading baseline or source facts.

Automatic capture selects one whole `.agenticlens-config.local.yaml`: worktree-local if present, otherwise base-local through Git metadata. It never inherits missing worktree-local keys from base-local/shared capture settings. Explicit boolean `telemetry.capture.enabled` and valid `state_dir` are required; consumed timings/budgets require explicit approved values with the bounds in the configuration reference. Relative state/source paths resolve from the selected local file. Shared YAML must exist in the active checkout. Native Codex user hooks route validated session context through the installed native manager and freeze no project/source/state path. Disabled callbacks validate configuration, then avoid capture-only I/O. No CI settings or harness trust changes are installed.

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
```

The preceding numeric/path values are example choices requiring approval, not presets. Base-local setup requires explicit capture enable/disable, source/source path, state and consumed timing/budget choices. `init --local-only` fills only missing supplied choices and preserves valid existing values. Worktree-local files remain user-managed; setup from a linked worktree cannot create/copy them or write base settings. Missing/invalid automatic policy errors before mutation. Discovery does not create state. Relative paths belong to the selected local file directory; `~` expands using the executing user home.

The shared database is `state_dir/projects/<configured-project-name>/registry.db`. The namespace comes from required `project_name`; unsafe literal names use the supported reversible representation, never a checkout-derived project hash. Exact name metadata rejects filesystem aliases/collisions. Worktrees sharing that name concurrently use separate exact-worktree records, bindings, source ownership, publication requests/cursors and worker coordination. For the same ticket, publication may retain unchanged committed history from a verified worktree of the same actual Git repository. New capture and workers remain isolated and never acknowledge another worktree's pending work. A shared project name alone grants no history access.

Aliases for the configured state base are supported. An alias inside its project namespace must stay within that physical base and preserve the `projects/<configured-project-name>` suffix; escaping or differently shaped targets are rejected. Configure the intended base explicitly when using another layout.

Home state survives worktree deletion, but missing sources or publication destinations still require explicit recovery. No deleted checkout is recreated or rebound by name. Existing explicit project-local/external paths remain unchanged on narrow configuration update; relocation is explicit, preserves pending/failed work and retained original data, and refuses unsupported layouts or active owned workers. Copying a live database or changing a YAML path alone is not a migration. No automatic cleanup removes registrations or pending observations.

The relocation command supports only the inspected legacy registry protocol1 family into schema2 named state. Moving or renaming populated schema2 state is currently unsupported and fails without retiring the source or modifying configuration; preserve it and report the limitation rather than inventing a copy/rename repair.

The worker uses durable request generations and acknowledges only the generation observed before its pass. Kernel process locks release on crashes. Later callbacks restart workers; workers periodically refresh active sessions, debounce bursts, back off after failed publication, and exit after idle terminal sessions. A finite maximum lifetime transfers active/pending work to a fresh native worker. Check `capture status --repo PROJECT` and local `worker.log` for failures. This is native child-process scheduling, with no separately installed service, model recursion or callback network/self-update work.

The ledger uses schema v6 at `<workspaces.dir>/<ticket>/agentic_metrics.json`. Top-level `completed_at` is the observability snapshot finalization boundary, equal to artifact `updated_at`; it does not assert source-session termination or ticket merge. Source invocation completion timestamps remain unchanged. Each owner has its own source-read checkpoint; there is no simultaneous source-read guarantee. Source failures preserve accepted facts and visibly degrade coverage.

Publication uses shared generation.summary policy. Durable local diagnostics retain reconciliation/publication evidence; standalone report files are ignored and preserved.

### Explicit installation consistency and completion scope

Validated byte-identical native cache leaves retain their filesystem identity, including force-identical installation and in-place permission repair, so another project's admitted worker is not retired. A changed-content repair remains explicit. Pre-apply validation/activation failures roll back owned changes. Once external capture apply may have committed, an incomplete result, output-limit or timeout failure retains validated native/wrapper files and recoverable prior backups; it reports incomplete setup and directs status/owned reinstall recovery rather than claiming complete rollback or removing handlers referenced by hooks. Resolve those diagnostics before declaring setup healthy.

Printed scaffold completion runs inside the selected project and restores the caller's location. The native manager's --al-project selects release authority; it does not change ordinary native working-directory semantics. Keep the printed scoped command when supplying project facts, so shared and local configuration writes cannot target another same-version scaffold in the caller's directory.

### Optional metrics during generation

Shared project settings independently control optional `generate` output:

```yaml
generation:
  metrics:
    skills: true
    delegations: true
    anomalies: true
```

All three members must be explicitly configured before generation; missing/null/wrong-type members fail. Explicit false disables that category; values must be actual booleans, not null, strings or numbers. The public key is `delegations`; there is no `task_links` alias. Invalid settings fail before ledger or summary replacement.

Shared YAML discovery respects project boundaries and rejects two canonical suffixes in one directory; values are explicit, without member substitutions. `.agenticlens-config.local.yaml` retains its machine-source/capture override whitelist and cannot override these generation settings. No environment or CLI override exists for them.

Disabled categories do not authorize deleting unavailable, partial or other-harness owners' recorded facts. Complete successful owner refresh remains the authority boundary; no new analysis is performed for deliberately disabled categories. New ledger/API output contains no `generation_policy`, `anomaly_availability` or replacement persisted reason marker. Field presence carries the contract for skills, delegations and anomalies at every canonical owner: omitted means not recorded, `[]` means a complete supported inspection recorded no rows, and values mean recorded facts. Explicit null arrays are invalid. Former generation-policy/availability metadata is rejected by ordinary reads and cleaned only through existing explicit migrate/backfill preparation, preserving contradiction checks and absent/empty categories. There is no schema version bump or bulk rewrite.

Ordinary execution, tool errors, child identities, tokens, costs and code metrics remain available. The managed Markdown section is replaced without stale disabled values or changes to user-authored prose. Current `generate --anomalies` reports disabled when `generation.metrics.anomalies` is false; it never re-enables analysis or persistence. The running command may explain its effective settings. Later readers use field presence and say not recorded/partial; they cannot distinguish disabled from unavailable or unsupported evidence and do not infer today's configuration for historical records.

Re-enabling can recover source-backed observations where the original source remains available. Unavailable recorded history remains preserved; omission never proves an empty successful inspection. Complete supported snapshots establish empty arrays or values only for their exact owners. Missing, partial or other-source owners keep their existing field presence and known facts. Root/group anomaly analysis requires complete coverage of its required owners, independently of session-local evidence; cached known group findings survive unavailable children. Reads and non-source writes preserve absence rather than deriving successful zero or healthy analysis from ordinary tool counts. Partial known findings remain visible, and incomplete population rates are omitted.

These settings apply to manual and automatic generate only: they do not rewrite other ledgers, suppress live trace collection, redact raw text, globally filter the dashboard, or change explicit backfill defaults. Explicit backfill may replace successfully complete supported owners, including recorded-empty results, while preserving missing/partial/other-harness owners. Fleet tool reliability remains visible independently of anomaly recording.

### Provisional checkpoint feedback

Explicit `generate` commands may write human checkpoint diagnostics to stderr after successful publication. Configure their threshold in the shared project file:

```yaml
generation:
  feedback:
    provisional_threshold_seconds: 180
```

The explicit nonnegative integer is required when manual feedback is used: zero disables diagnostics, and positive values set the threshold. Negative, fractional, quoted, null and unrepresentable values are invalid. Configuration discovery and the local override whitelist follow the generation settings above; there is no CLI or environment override. This threshold is independent of stall and live activity settings.

Fresh native interval age is compared before rounding: 179999ms does not qualify, and 180000ms does. Individually known short intervals produce no feedback when only their sum reaches the threshold. Retained or unavailable native evidence may instead produce qualified invocation-subtotal guidance. Each diagnostic identifies the owning source/session/invocation, and source-open evidence does not establish that the invoking caller owns the interval or that a process remains executing.

Installed generic command transports do not currently verify the caller's exact native interval. The command therefore uses qualified source-open or unavailable-evidence wording. It never ends a turn, schedules continuation or resumes itself. After the owner completes naturally, a user or supported harness action must initiate the next turn before refreshing telemetry. Diagnostics never enter ledgers, summaries, APIs or the dashboard, and optional feedback failures do not invalidate successful publication. Automatic lifecycle generation, backfill and ordinary reads do not emit these checkpoint diagnostics.

### Native Codex user adoption and development output

Codex uses `capture install-user` in the effective user home; `uninstall-user` removes only unchanged owned groups and retains state. Installation is additive and journaled. Explicit `--migrate-legacy-repo <ORIGINAL_CHECKOUT>` removes unchanged owned project callbacks while preserving state; unknown/modified callbacks error and remain untouched. No per-worktree installation manifest is required by this path. Legacy OpenCode project-plugin commands remain separate until its global integration is delivered.

Shared `agenticlens.version` is an exact required pin. Configured `agenticlens.binary_dir` requires valid host output `agentic-lens-engine-GOOS-GOARCH`; missing/invalid selected output fails. Absent development mode deliberately selects the exact pinned installed engine. Local files cannot override these selectors. Workers retain immutable executable bytes for later children through normal exit; new invocations select the current preferred build. Status exposes local configuration origin and active/preferred identity.

The native manager owns installation and routing. `agentic-lens install <EXACT_ENGINE_VERSION>` provisions an engine; `agentic-lens self-update --version <EXACT_MANAGER_VERSION>` explicitly updates the installed manager without editing any project pin. Both verify staged artifacts, preserve modified/unowned destinations and recover interrupted replacements under a destination lease. Codex hook installation only writes owned hooks and their receipt; it does not copy an engine into a dispatcher path. Hook removal retains the installed manager and engine cache.

Ordinary engine commands preserve caller cwd and relative-argument meaning even with `--al-project`. The manager validates the selected engine and forwards ordinary argv. Engines contain no dispatch mode and workers execute their retained engine directly.

## Ticket tags

Add optional ticket metadata in shared `.agenticlens-config.yaml`:

```yaml
tags:
  team: core
  environment: dev
```

Private `.agenticlens-config.local.yaml` can contain only a `tags` mapping, for
example `tags: {environment: staging}`. Invocation overrides use repeatable
`generate --tag environment=prod --tag experiment=baseline`. Precedence is per
key: CLI overrides local, which overrides shared; unrelated lower-layer keys
survive. Repeated CLI keys use the last value and split at the first `=`, so a
value may itself contain `=`. Keys and values must be nonblank strings. Outer
whitespace is trimmed; internal text and case remain unchanged. Quote YAML
values such as `"true"`, `"12"` and `"null"` to keep them strings. Duplicate keys
including whitespace-normalized collisions are errors.

A successful generation merges effective tags into the ticket ledger. Missing
inputs preserve recorded keys; there is no deletion syntax. Automatic publication
uses shared/local defaults. CLI tags apply only to that invocation and do not
change configuration or capture registrations. Backfill and dashboard reads
preserve historical tags without applying the current reader's defaults.

Local configuration remains private, but its **effective tags become shared
ledger data**, including after committing the ledger to Git. Do not put secrets
in tags. Dashboard tag filters select whole tickets (OR values within a key,
AND across keys). Missing identifies absence separately from a literal value
named `Missing`. Comparison coloring by tag changes visual identity while
preserving ticket metrics and point coordinates.
