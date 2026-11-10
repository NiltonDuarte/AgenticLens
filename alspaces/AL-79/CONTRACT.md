# AL-79: Native user-level Codex capture across worktrees

Status: Proposed baseline; implementation awaits explicit user approval.

## Authority and outcome

This contract implements the AL-79 specification in `alspaces/plan/to-do/AL-79 native user level codex capture across worktrees.md` and PO-20 F1-F6 in `alspaces/PO-20/SPEC.md`, including the confirmed decisions in `alspaces/PO-20/CAPTURE_VISION_AND_GAPS.md`. The developer session starts at clean revision `4df45967c801824500ac2ea39a8f52fa9cbdcf66` on `feat/AL-79-native-user-codex-capture`. Workspace root is the configured `alspaces`.

Install once in the effective Codex user home. Configured ordinary checkouts and linked worktrees then capture exact native sessions, bind to explicit tasks, collect authoritative source snapshots and autonomously publish portable ticket ledgers. Deliver reusable native configuration, executable-selection and worker behavior for AL-80. Discovery alone does not establish runtime success.

## Configuration boundary

The validated session context identifies the active checkout, including nested invocations. A callback's different process cwd cannot redirect it. Require that checkout's shared `.agenticlens-config.yaml`; automatic capture silently skips if it is absent. Never fall back to base shared configuration. Select one complete worktree-local `.agenticlens-config.local.yaml` when present, otherwise one complete base-local file resolved from Git metadata. Missing, malformed, unreadable or wrongly typed required configuration fails with its selected path and origin. A present invalid worktree-local file never falls back. A valid worktree-local file needs no accessible base-local file. Relative local paths resolve from the selected file's directory; absolute and home-relative paths keep their defined semantics.

Shared YAML owns project, ticket, workspace, classification, metric and executable-selection rules. Selected local YAML owns capture enablement, source location, state, cadence, limits and machine integration settings. Neither shared capture settings nor base-local keys supply missing keys in a selected worktree-local policy. Documented numerical defaults remain defaults, rather than inheritance. Require an explicit boolean `telemetry.capture.enabled` and explicit required machine paths. Validate configuration before applying false; disabled capture then performs no capture-only source reads, state initialization, readiness checks or worker launch. It preserves manual generation. Existing explicit manual CLI source/path overrides remain explicit and cannot be silently injected by callbacks to mask source disagreement.

CLI and capture share selection and project/metrics interpretation. This ticket does not implement PO-20 F7's independent AL-81 registry-decoupling change; it must not add new capture-installation prerequisites for manual generation or stored-history reads. Coordinate overlapping publication call paths without absorbing AL-81.

## Native installation and adoption

Own additive Codex user hooks, preserving unrelated hooks and normal persisted user trust. Callbacks contain a stable native routing entrypoint and only harness-required glue; they freeze no project, source or state path. Python and per-worktree manifests cease to be runtime requirements. A valid enabled checkout initializes only on its first real event. Status/discovery are read-only and cannot initialize siblings.

Installation, explicit legacy migration and removal must be transactional and preserve pending state. Remove only unchanged owned legacy callbacks with verifiable ownership; preserve modified or unknown hooks and diagnose conflicting duplicate invocation. Never guess ownership of the temporary home Python launcher. Do not silently remove arbitrary home files. Legacy migration must make duplicates visible and resolve verified owned duplicates explicitly.

Hooks signal exact-session refresh to the existing collector/worker. They are not an event-replay queue, another service or a model invocation. Codex SubagentStart/SubagentStop parent callback identity remains the parent. Validate authoritative source/session/worktree context; an established source-location contradiction reports both locations and stops capture. Unflushed/unreadable correct sources remain pending or failed collection, never an established mismatch or successful empty snapshot.

## Executable selection and lifetime

Require exact shared `agenticlens.version` even when development output exists. Add optional shared `agenticlens.binary_dir`, relative to the active shared file. Select the current host executable using a documented layout compatible with normal native build output. Local YAML cannot override version or development selection. Preserve a versioned user cache shared by projects and harnesses.

Prefer completed valid development output and report its real build identity. If absent, use the exact installed pin. If present but invalid by native format/platform, executable or required-interface checks, fail without fallback. Missing required cache produces an exact installation diagnostic. Normal CLI/capture selection never downloads, chooses latest, edits pins or requires publication of development builds. Explicit installation alone provisions published binaries; local build/install supplies unpublished output.

An admitted worker and all later collection/publication children use an immutable retained executable identity until normal exit. Rebuilding or replacing the preferred mutable development output cannot change the executable of that worker's subprocesses or force retirement merely because another build is now preferred. New invocations/workers use the preferred valid selection; scope serialization prevents competing publishers. Status distinguishes actual active build from preferred build. Cache maintenance is separate and cannot remove an executable still needed by active work.

## Binding, publication and recovery

Use the user's explicit current task to answer the exact-session binding request directly. Missing or ambiguous ticket information remains visibly unassigned and asks for that information. No inference from branch, cwd, role, timing or incidental references, no placeholder tickets and no redundant confirmation. Collect locally while unassigned; ticket ledger and summary publication waits for binding. Source-proven same-task children inherit a bound parent unless explicitly bound already, retaining native child ID and actual worktree scope. Independent roots bind independently.

Coalesce publication during activity and at turn/session checkpoints. Preserve provisional/unknown values and never fabricate finalized totals. Use bounded local retry/backoff in the existing worker. On exhaustion preserve facts, published history, pending generation and diagnostics; stop that cycle and resume on a later normal trigger or explicit retry. Configuration errors and authoritative contradictions stop rather than becoming retry success. Serialize publication for each scope and preserve simultaneous bound contributors, unrelated tickets and harnesses. Do not automatically commit generated project artifacts.

After worktree deletion retain complete and incomplete registrations, bindings, snapshots, pending work and diagnostics until explicit cleanup. Publication reports the unavailable destination without recreating it or redirecting to another checkout. Cleanup implementation beyond an already supported explicit facility is not required.

## Portable history and architecture

`TECHNICAL_GUIDELINES.md` is authoritative. Complete successful local snapshots replace only exact refreshed session owners, including smaller or empty snapshots. Failed, unavailable and partial sources preserve prior owners, other developers and other harnesses. Repeated snapshots are idempotent and totals derive from the resulting full ledger. Capture receipts, installations and local databases are operational state, never prerequisites for accepting supported Git-shared historical facts. Preserve existing schema/API/dashboard consumers; no bulk ledger migration or schema expansion is assumed.

Keep pure policy/eligibility, executable-choice, binding and snapshot decisions in `backend/internal/domain`. Filesystem, Git, configuration parsing, source reads, process execution, locking and persistence stay in adapters; CLI composes them. Go remains value-oriented with no domain I/O, clocks or mocks. Zero `any`/`interface{}` and zero emojis. No new dashboard feature or OpenCode global plugin implementation is in scope. Production remains a standalone Go binary with embedded frontend and no Python/Node capture runtime requirement.

## Setup and evidence

Base setup preserves valid settings, reports retained values, recommends capture and asks only for missing choices. Noninteractive setup requires existing or explicitly supplied required choices; it never guesses enablement. Invalid selected configuration errors. Worktree-local configuration is user-managed: no guided creation/copying or base writes from worktree setup. Runtime fallback never creates a local file. Update canonical setup/operator/contributor guidance and their shipped copies consistently.

Acceptance includes self-contained ordinary-checkout/two-worktree/nested Git fixtures, whole-file precedence and origin, disabled no-I/O behavior, exact session binding/child ancestry, concurrent publication, source partiality, bounded retry, deleted destinations and worker build retention across rebuild. Isolate HOME, config, cache, state and output destinations. Fixtures cannot depend on live ticket/PO workspaces.

Native evidence must record actual Codex version/interface, pre-fix discovery with sourcePath, resulting home-hook discovery, normal trust, real root binding, a source-proven child, collection and automatic publication with source session IDs. Installed files, mocked exchanges and old PO observations cannot substitute. Current read-only grounding found Codex 0.167.0 and AgenticLens v3.0.0-rc.131+03f5b05-2-g57a3503e; no live callback/trust proof has been performed. Human trust interaction or native access gaps remain explicitly incomplete. No per-worktree certification, test callbacks or trust bypass is shipped.

## Delivery boundaries

Implementation starts only after explicit scope approval. Parent owns integration, telemetry, commits, final review and PR. Apply focused independent review under both configured model profiles, resolve findings, integrate current main and pass required `make`. Record revision-specific test/runtime evidence and remaining limits. Generate Codex telemetry at the final delivery step, verify actual source/root/child extraction and preservation of existing sessions, and never fabricate an empty successful ledger. Complete the ticket lifecycle and unblock AL-80 only when all AL-79 acceptance passes; publish the PR without merging it.
