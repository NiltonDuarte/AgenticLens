# AL-68 iteration 02: automatic publication and project isolation

Status: Approved for implementation by the user's explicit instruction: '$agentic-lens-dev approved, follow with implementation'. The earlier request authorized planning only; this later instruction satisfies the implementation approval gate. This plan is the proposed scope delta to the frozen [CONTRACT.md](../../CONTRACT.md); it does not rewrite that historical contract or imply approval from its iteration number. Iteration 01 implementation/review evidence remains available under [iteration 01](../01/REVIEW.md).

Baseline: `73e9ecb90baad909764b7d21c161bf31a7568eea`. Initial Git status: clean. Configured workspace root: `alspaces`. The current feature branch and PR retain AL-68 ownership. Parent owns integration, summary, telemetry, commits and PR publication. This planning assignment owns only this file; no application, runtime, configuration or tests are changed.

## User steering and acceptance delta

| Source in the conversation | Required behavior | Superseded assumption | Verification impact |
| --- | --- | --- | --- |
| “The plugin [is] a substitution for the generate command” and the agent interacts with the plugin | Automatically extract and publish the ticket ledger without an agent executing generate; agent interaction supplies explicit ticket/context | Registration-only integration plus mandatory agent-run final generation | Real plugin/hook-only publication, including a long active session and restart |
| Multiple roots in one directory must not corrupt one file | Collect all explicitly bound contributors in that directory; serialize complete read/refresh/merge/write operations | Separate plugin instances may publish only their own session or use stale pre-lock baselines | Concurrent independent roots, repeated callbacks, manual/automatic overlap and in-flight events |
| “Two worktrees should not interfere … even … same ticket”; worktree handling is out of scope | Isolate by exact configured project location, source store/session and ticket; different directories publish independently | Repository-wide/common-Git-directory/shared-ticket collection and cross-checkout merging | Two directories, same ticket and same harness store, cannot contaminate each other |
| “We still must be sure we do not update worktree A session into worktree B ledger” | Validate session project ownership, not just destination write locking | Ticket match, local state location or caller-supplied project flag proves ownership | Forged/mismatched context, cross-directory child, legacy registration and selector tests |
| “State-dir … configuration in the YAML?” | One configuration resolver serves installation, callbacks, worker and manual commands | Repeated matching path flags are the only setup mechanism | Shared/local/CLI precedence, defaults and non-default installed-handler round trip |
| The last publication tail is expected; losing forty minutes is not | Refresh available authoritative source data; distinguish retained/stale facts from fresh facts | An old cached snapshot can be presented as a current publication | Source already contains forty more minutes versus genuinely post-cutoff activity |
| completed timestamp is observability completion; PR merge freezes the ledger | Top-level completed_at records snapshot finalization, independent of ticket/PR completion and active-root shutdown | Clearing completed_at merely because the publishing root is active | Snapshot completion, source session timestamps and artifact update semantics tested separately |
| Project checks are not AgenticLens observability; integrations must not be intrusive | No mandatory Git hook, branch protection, CI readiness check, contributor roster or external finalizer | A required capture merge gate is necessary for delivery | Installation leaves project CI/Git hooks unchanged; diagnostics remain honest |
| User questions permanent state and committed capture_report.json | Distinguish accepted requirements from storage/retention/report recommendations below | A permanent home database and a committed detailed report have already been approved | Validate whichever policy is approved; no hidden retention or commit behavior |

Preserved guarantees: explicit session-specific agent binding without guessing; authoritative source identity/model/usage/cost/parents; unknowns stay unknown; independent roots stay roots; additive ownership-aware setup/removal; no trust bypass; no recursive model calls or network/update work in callbacks; exact-source preservation on partial failure; schema v5; bounded processing and privacy. Existing CAP-01 through CAP-12 protections must carry forward, with affected review and regression evidence refreshed after implementation.

The canonical filename remains `agentic_metrics.json`. The user's `agenticlens_metrics.json` spelling describes the existing ledger responsibility and does not request a filename/schema migration.

## Grounded current behavior and seams

- `backend/internal/adapters/capture/plugin.ts` handles OpenCode lifecycle events with receive, requests binding in the system transform and exposes agenticlens_bind. It currently does not publish automatically. Each harness session/process can own its own plugin instance; coordination cannot depend on one shared JavaScript instance.
- `capture/hooks.go` implements bounded Codex stdin handoff and session-specific feedback in ReceiveCodex/BindingPrompt. HookInput currently lacks a project identity field; a new caller flag alone cannot close that ownership gap.
- `domain/capture.go` defines CaptureIdentity as source/store/session, registration/binding and snapshot precedence. `capture/store.go` persists those registrations and snapshots. Project scope must become an explicit local registration boundary without silently changing canonical v5 session facts.
- `capture/identity.go` canonicalizes source-store aliases. Project and output identity need equally explicit normalization, with missing paths and ambiguous aliases producing diagnostics rather than guessed equality.
- `capture/reconcile.go` and `children.go` own bounded source refresh, parent recovery, pending queues and durable cursors. Scope must constrain all these paths, not only final output filtering.
- `codex/rollout.go` has source session metadata including CWD; `codex/discovery.go` performs explicit source selection and native-tree discovery. `opencode/capture_snapshot.go` owns exact-session source extraction. Inspect supported OpenCode schema-family project fields before claiming project verification; do not introduce title/directory heuristics to guess a session ID.
- `cmd/agentic-lens/main.go` runGenerate already acquires publication ownership before baseline reads, reconciles capture and merges source facts. `capture.go` contains selection validation, precedence filtering and staged ledger writes. Extract reusable native publication orchestration rather than making the plugin implement a separate calculator or shell a recursive model/tool command.
- `capture/publication.go` locks the canonical output path, with native platform implementations. It does not coordinate different files, nor should this iteration introduce cross-directory coordination.
- `domain/config.go`, `adapters/config/loader.go`, `writer.go` and `source.go` own configuration defaults, shared/local resolution and generation. The current local overlay is telemetry-specific and must be deliberately extended and validated.

## Proposed runtime flow

1. Installation resolves an exact project location, source store, output mapping and capture settings. Existing review/trust controls apply. The source project context must be checked against this installation context before accepting a session into its publication scope.
2. A callback records a bounded normalized event. Unknown ticket context produces the existing explicit agent request. The agent supplies the ticket and any necessary explicitly known context through the harness-supported interaction. It never estimates source usage or selects another project's sessions.
3. A successful binding/event schedules native publication work for that exact project/ticket. Multiple instances enqueue/coalesce work rather than each writing a private partial ledger. A generation/request counter or equivalent acknowledged-work token must distinguish events received before and during a pass.
4. A native background worker or equivalent harness-supported scheduler refreshes periodically during activity and on proven lifecycle checkpoints. It owns bounded work, cancellation, retries and restart recovery. Callback handoff remains short; shutdown handlers do not parse full transcripts. No additional installed daemon/service/runtime dependency is introduced. The concrete launch/lifetime mechanism must be verified for each claimed harness/platform before adoption; an unavailable mechanism cannot be represented as working background capture.
5. For each destination, acquire publication ownership before baseline/source reads; reread accepted facts after acquiring it. Reconcile eligible sessions for the project/ticket, merge cumulative snapshots by exact source owner and stage replacement. Manual generate and automatic publication use this same orchestration and lock. Do not add cumulative totals across repeated reads, select values by numerical maximum, or let older source evidence overwrite newer accepted facts.
6. Publish the ledger automatically. Render the existing summary at a bounded cadence/checkpoint so tool bursts do not rewrite Markdown continuously. If an event arrives during publication, retain another pending pass; acknowledge only work actually processed. A failed worker must not consume unseen work or erase the last accepted publication.
7. CLI generate remains supported for explicit refresh, inspection and recovery. Normal agent operation must not need to invoke it. Automatic output writes never run Git commit/push or alter project CI.

The worker design must specify maximum work per pass, changed-data detection, refresh/debounce cadence, retry/backoff, lifetime/idle shutdown, ownership takeover after crashes and all process-resource cleanup. Unchanged ticks must not reconstruct every historical source. Source failures preserve accepted facts with visible freshness/failure status. A worker restart, ordinary resume or an event arriving at shutdown must not strand work indefinitely while the supported harness remains active.

## Project and destination isolation contract

Project identity means the canonical configured project directory, not a branch, ticket, remote URL, repository name, Git common directory or worktree relationship. AgenticLens has no worktree discovery/merging feature in this iteration.

A registration is eligible only when its proven project location and explicit ticket match the publisher's scope, in addition to exact source/store/session identity. Harness project/CWD metadata must be interpreted against the configured root; a source-reported working/project directory within that root may normalize to it, but ambiguous nested projects or conflicting roots must remain unresolved. Do not use a string prefix or a matching directory name as path containment evidence. Source/configuration disagreement is an error or visible degraded state, not permission to relabel a session.

Installation context and an agent's request are not substitutes for source identity validation. Validate source metadata when registering or before first inclusion; an explicit selector cannot bypass scope. An agent cannot move another directory's session by passing a different project argument. The planner does not assume unsupported source fields or callback payloads; inspect real source records and supported harness fields first.

Scope registrations, binding requests, child discovery, queue/cursor state, caches, source selection and baseline replacement. Proving a native parent does not automatically authorize a child whose project context conflicts. Preserve independent roots within one project and include them all when bound to the same ticket. Different tickets remain independent inside a project.

Normalize existing symlink aliases consistently for source, project and output. Aliases of the same actual project/output must coordinate; two distinct directories must remain distinct even with the same ticket/source store. Reject destinations that escape the approved project/output mapping through aliases or traversal. A shared home state directory cannot be an authorization boundary by itself; namespace and validate each project independently. Do not import another project's sessions from an existing ledger merely because IDs or tickets match. An ambiguous legacy baseline requires explicit diagnosis/repair rather than silent acceptance or destructive deletion.

Two independent project directories may each publish their own AL-68 ledger. Subsequent Git conflicts between those files are outside AgenticLens responsibility. There is no repository-wide aggregation or cross-directory merge reconciliation.

## Configuration, state and report policy

Required configuration behavior: command override precedes machine-local YAML, then shared portable configuration/defaults. Validate types/ranges and resolve relative paths against the owning project configuration, not an arbitrary callback process CWD. One resolver feeds install, captured native paths, automatic publication, status, reconcile and manual generate. A configuration change or incompatible installed path must be diagnosed; do not silently split collection and publication across registries.

Proposed YAML surface: a capture section under telemetry for automatic-publication enablement, cadence/debounce/work bounds, report/summary behavior and state/retention preferences. Machine-specific state/source/native paths belong in `.agenticlens-config.local.yaml`; shared YAML contains portable preferences. Exact key spelling/default values are implementation-design choices to freeze before coder dispatch, not currently supported settings. Preserve existing source-path precedence outside the new capture settings.

A small persistent association store is useful when an independent reviewer exits before another root publishes, or when a process restarts. Source snapshots are rebuildable caches while original harness records remain accessible; they are not a permanent second archive. In-memory state alone needs a proven alternative for retaining explicit session-ticket associations between independent processes. No design may reconstruct those associations through ticket-text, branch, role or timing guesses.

Recommendation pending approval: keep minimal recoverable state outside disposable project directories, for example under `~/.agenticlens/capture/<project-location-key>/`, with explicit configurable retention and cleanup. Project location, not common repository or ticket, determines isolation. A configured local project-state mode may remain available with its deletion tradeoff documented. A permanent database, mandatory home default and indefinite snapshot retention have not been approved. SQLite may be reused, but it is not a product requirement. Binding/pending work and published/rebuildable caches need distinct retention policies; no automatic cleanup may discard outstanding work while claiming success.

Legacy state has no verified project ownership. Migration must verify each row's source context before assigning it, preserve unknown rows for explicit recovery and never bulk-import it into whichever project happens to launch first. Copying state or losing a project directory is not permission to associate its old sessions with a new directory.

The full capture_report.json remains a diagnostic option. Proposed policy for approval: retain compact coverage/freshness in the summary and make detailed report publication/commit expectations configurable; committing the report is not necessary to drive automatic collection. Do not remove the current sidecar or introduce a v5 coverage field until the policy is approved. When a report is produced, preserve its runtime ledger pairing and avoid links to absent reports. Do not reintroduce review hash bookkeeping.

## Snapshot freshness and completion semantics

The user accepts activity after the published snapshot being omitted. An active root alone must not block automatic publication. No external finalizer is needed merely to wait for that root's last actions.

Top-level completed_at is observability snapshot finalization, not ticket completion, PR merge time or an assertion that every source session ended. It must be set from the actual native publication boundary according to one consistent documented clock rule. Per-session invocation timestamps remain authoritative source facts; updated_at describes artifact refresh. Do not imply a common simultaneous source read: preserve each owner's successful-read/freshness evidence and state the overall pass boundary separately. Subsequent automatic refresh advances observability completion when it produces another accepted snapshot.

Example acceptance: publication at 11:00 includes source records available through 10:59:58, while commit/push afterward is an accepted tail. Reusing a 10:20 cache while the authoritative source already contains forty more minutes must never be labelled fresh. A source that has genuinely not flushed is visibly partial/stale; no timestamp or zero-usage guess fills that gap. Previous accepted facts survive failure, with their actual age/checkpoint visible.

Earlier AL-68 artifacts set top-level completed_at to null because the root was active. That interpretation is superseded by the user's clarification. Correct current artifacts through a fresh extraction/actual publication timestamp and update active guidance; preserve historical archived reports and source invocation facts. Do not invent retrospective completion dates or rewrite historical review outcomes.

## Implementation phases and writable ownership

1. Planner/parent finalize the approval decisions below and record the approved boundaries here. Freeze the original contract and iteration01 history. No coder dispatch before explicit scope approval.
2. Coder owns domain capture/config decisions and adapter/CLI implementation: `backend/internal/domain/capture.go`, `config.go`; `backend/internal/adapters/config/{loader,writer,source}.go`; `backend/internal/adapters/capture/{identity,store,hooks,reconcile,children,publication,install}.go`, `plugin.ts` and any responsibility-named worker/publication module; `backend/cmd/agentic-lens/{main,capture}.go`; relevant Codex discovery/rollout and OpenCode snapshot/family extraction seams. Record exact new paths before edits. Keep pure selection/state decisions in domain and clocks/files/processes/locks/source validation in adapters/CLI. No UI feature or ledger schema migration.
3. Coder owns matching setup/distribution/guidance changes: both shipped downloader/wrapper copies, setup/operator references and `docs/SETUP_MANUAL.md`; root contributor guidance only where it conflicts with the revised automatic-publication/completion boundary. Preserve release-only distribution, native handler compatibility, additive rollback and trust. Do not install hooks/CI controls into adopting projects without explicit opt-in.
4. Tester owns assigned tests after inspecting existing coverage: capture domain tests; capture store/reconcile/hooks/installer tests; config loader/writer/source tests; CLI capture/main Codex tests; exact-source Codex/OpenCode family fixtures; `scripts/tests/test_capture_distribution.py` when distribution behavior changes. One writable owner per file; shared coder/tester files require sequential handoff. Return product defects to coder.
5. Parent freezes a readable candidate commit/ID, coordinates focused reviews, integrates, runs required verification and records the current verdict. Children share the checkout and preserve others' edits. Reviewers are read-only for code/tests and use uniquely assigned private temporary reports; parent alone consolidates findings, evidence summaries and attribution.
6. Parent owns summary, PR description, lifecycle synchronization, final telemetry, commits and PR delivery. Preserve historical AL-68 facts from both harnesses and distinguish iteration02 planning/implementation evidence. No remote merge or release authorization is implied.

## Acceptance and regression evidence

Use self-contained fixtures, never the live ticket workspace. Reuse unchanged valid evidence, add only missing regression oracles and run affected checks after changes.

| Case | Observable acceptance | Existing coverage owner/seam |
| --- | --- | --- |
| Two independent roots, same directory/ticket | Automatic ledger includes both exact owners; neither replaces the other's facts; no agent generate call | CLI capture tests, store/reconcile tests, real plugin smoke |
| Two directories, same ticket and source store | Each ledger excludes the other's registrations, metrics and children, through both automatic and explicit selector paths | Capture identity/reconcile/children, CLI capture tests, Codex/OpenCode source fixtures |
| Wrong caller project or destination | Source-context mismatch, path traversal or escaped output is rejected/degraded; no reassignment by --repo or state location | Hooks/store/identity/CLI installer tests |
| Source-proven parent with conflicting child project | Parent linkage does not bypass project ownership; child remains excluded/unresolved | Children tests and native parent source fixtures |
| Legacy rows and copied baseline/state | Unknown project attribution is not silently imported; accepted facts are not erased; recovery is explicit | Store migration/CLI persisted-ledger fixtures |
| Symlink/store/project/output aliases | True aliases coordinate; distinct directories never collapse; output escape is rejected | Identity tests and CLI publication/alias fixtures |
| Repeated callbacks and resumed cumulative usage | Same snapshot is idempotent; later source facts replace the exact owner, preserving other roots/harnesses/auxiliary facts | Domain capture and CLI precedence/repeat tests |
| Parallel automatic/manual publishers | Lock precedes baseline/source reads, later owner rereads baseline, staged publication cannot regress facts | CLI overlapping-publisher tests and publication adapters |
| Event during an active pass | A subsequent pass includes it; first pass cannot acknowledge unseen work or consume a pending request | New worker tests at native scheduling boundary, reused real registry fixtures |
| Worker crash, process exit and resume | Lock releases; last accepted files remain readable; pending work survives according to approved state policy and resumes without duplicate totals | Publication/store/cursor tests and native process probes |
| Forty-minute stale cache versus accepted tail | Available newer source usage is captured; failed refresh is visibly stale; post-snapshot actions are simply outside that snapshot | CLI Codex/source refresh tests, source comparison smoke |
| Unflushed/missing/invalid source and budget exhaustion | Preserve accepted facts; distinguish pending/failure from current; bounded continuation is fair | Reconcile/CAP-05/CAP-12 tests |
| Snapshot timestamp with active root | Non-null observability completion matches finalization boundary; source timestamps/unknowns remain intact; PR merge irrelevant | CLI generation timestamp fixtures, coverage rendering tests |
| YAML, overrides and installed path | Non-default settings round-trip loader/writer/install/worker/manual commands; invalid settings fail; no split state paths | Config loader/writer/source and capture installer tests |
| Optional detailed report | Approved mode preserves compact warnings and freshness; absent report has no dangling summary link; produced report matches ledger | CLI report/publication tests and Markdown rendering coverage |
| Native distribution and nonintrusion | Compatible binary paths/rollback work; no added runtime dependency or unapproved Git/CI change | Distribution tests and native install/upgrade smoke |

Real acceptance must run the actual Codex and OpenCode integrations with normal human-controlled trust/plugin loading, two independent roots and a native child where supported. Demonstrate automatic ledger creation and subsequent updates without any contributing agent executing generate. Include two project directories sharing ticket/source store, a long active interval/checkpoint, restart/resume and overlap. Compare IDs, project context, model/usage/known-cost evidence and native parents with owning source records. Fixtures do not establish runtime support. Codex SubagentStart remained unobserved on iteration01's exec path; running/interrupted-child coverage stays degraded until a new real probe establishes it. No trust changes/bypasses, model substitution or invented API closes a missing runtime gate.

Prepare canonical embedded assets before affected Go checks; run affected domain/config/capture/CLI/source packages with race checks and the relevant distribution probes. Required implementation integration gate is `make` on the integrated candidate, including frontend stages/native builds even though no dashboard feature is proposed. Report unsupported platforms/interfaces and skipped stages explicitly. Old code approvals do not approve changed scheduler, identity, configuration or timestamp behavior.

## Independent review roster after implementation

Every selected focus runs both configured profiles: `sol_6_1_high` (gpt-6.1-sol/high) and `luna_6_medium` (gpt-6-luna/medium). These are future dispatch requirements, not completed reviews. Load the current canonical model policy before launch; preserve requested/acknowledged/observed settings separately. Use capacity-limited waves, equal frozen candidate evidence and independent first passes.

| Focus | Boundary | Proposed run-ID stems |
| --- | --- | --- |
| Behavior | Agent binding through actual automatic lifecycle publication, resume and partial outcomes | iter02-behavior-sol; iter02-behavior-luna |
| Data | Cross-project isolation, child selection, cumulative replacement, freshness, timestamps and concurrency | iter02-data-sol; iter02-data-luna |
| Architecture | FCIS, worker ownership/restart, shared resolver and bounded resource work | iter02-architecture-sol; iter02-architecture-luna |
| Tests | Regression oracle strength, genuine overlap/crash probes, source-vs-fixture evidence | iter02-tests-sol; iter02-tests-luna |
| Security | Source/project proof, path/alias/output escape, request binding and payload privacy | iter02-security-sol; iter02-security-luna |
| Release/setup | Native scheduling lifecycle, compatibility/rollback, packaged copies and nonintrusive installation | iter02-release-sol; iter02-release-luna |

Parent assigns each run a unique private temporary report path outside the tracked workspace, retains compact reviewer/model/finding outcomes in REVIEW.md, and maintains consolidated REVIEW.md only when produced. No placeholders, per-run committed reports or persistent review hashes. Parent validates relevant counterexamples and obtains revision-aware affected re-review under both profiles; missing required runtime or review evidence is INCOMPLETE.

## Decisions to resolve for implementation approval

Accepted requirements above are independent of these still-unapproved design recommendations. Approval must state the chosen policy rather than treating earlier assistant suggestions as user authorization:

- Exact supported native worker lifetime/launch strategy and configured cadence/work bounds, demonstrated against actual harness constraints; no mandatory external service.
- Minimal persistence implementation, default location (home storage is only a recommendation), retention/cleanup and how explicit association recovery works when local project state disappears.
- Detailed capture report production/commit defaults and summary cadence; no removal of current runtime evidence solely on the assumption that the user questioned its value.
- Exact YAML key names/defaults and portable/shared versus local fields; no claimed current support for proposed keys.

Scope approval does not approve harness trust or auto-install Git hooks, CI gates, branch protection, release publication or merge. If authoritative project context or an automatic scheduling mechanism cannot be proven for a claimed path, record the blocked requirement and revise the supported boundary before implementation approval/delivery; never silently revert to agent-run generate as the normal workflow.

## Implementation authorization and conservative defaults

Source: the user explicitly approved this iteration and invoked agentic-lens-dev to proceed with implementation. No renewed scope question is required. Initial implementation baseline is the committed plan revision de31cb1; Git status was clean before recording approval. The frozen original contract and historical reviews remain unchanged.

Within this approval, retain minimal project-local recoverable state as the compatibility default, make its location configurable through machine-local YAML and CLI, and support an explicitly selected external/home location with project-directory isolation. Do not introduce mandatory permanent home storage or implicit repository-wide sharing. Keep the existing detailed report as the compatibility default, add an explicit configuration policy for optional report production, and retain compact truthful coverage in summaries. Never delete an existing committed report automatically. No Git/CI integration is installed; optional Git hook work is deferred.

Implement automatic publication by a bounded native worker started through the installed native handler, sharing durable per-project request/progress state and per-output publication locking. It is not a separately installed service and adds no external runtime. The coder must record exact YAML keys, refresh/debounce/work/idle settings and lifecycle behavior in this plan or returned implementation design before integration, verify the real launch/restart behavior, and surface unsupported paths. Detailed settings are routine implementation choices within the approved automatic-publication boundary, not permission to invent a harness API. Manual generate remains supported. Real harness trust remains human-controlled.

Parent owns this plan/summary/review records, integration and delivery. Coder owns application/config/setup implementation paths listed above; tester receives explicit test-file ownership after code handoff. All contributors share the checkout and must preserve others' edits. Telemetry remains deferred until implementation/review/integration finalization.

## As-built policy and final correction scope

Implementation approval settled the historical design questions above through the conservative defaults already recorded. Current `telemetry.capture` defaults are enabled=true, state_dir=.agenticlens-capture, refresh_ms=15000, debounce_ms=500, idle_ms=120000, max_lifetime_ms=1800000, pass_timeout_ms=30000, max_sessions=256, max_source_bytes=67108864, detailed_report=true and summary=true. Durations are milliseconds. Local YAML overrides shared/default policy; explicit CLI state selection wins with installed-state drift diagnostics. Physical project namespaces isolate a selected state base.

Review corrections preserve the approved automatic-publication boundaries: source/store/project/ticket proof, periodic progress despite callback traffic, finite shutdown/restart, status before first request, selected-project configuration, strict caller-authorized recovery and consistent physical aliases across actual reads/writes/locks. AUTO01-21 findings and refutations are review corrections, not new scope or new approval requests. The final candidate keeps the separate optional report and schema v5; iteration03 remains unapproved. Candidate07 changes only source installer/prefix consistency after the candidate06 complete core reviews, so paired release/test re-review and explicit unchanged-dependency carry-forward apply. Final telemetry stays last.
