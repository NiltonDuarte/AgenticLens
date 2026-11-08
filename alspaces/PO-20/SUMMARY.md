# PO-20: Machine-wide automatic capture across Git worktrees

## Scope and outcome

Define automatic AgenticLens capture that works from a base checkout or Git worktree after one user-level installation in Codex or OpenCode. This Project Owner ticket records the completed decision review, product requirements and the smallest practical development backlog. Implementation belongs in separate agentic-lens-dev sessions.

No application code, tests, runtime scripts or machine-local integrations are changed by this specification work. The shared development pin selects the locally built version. No developer workspaces, technical contracts or implementation plans are created.

## Deliverables

- [SPEC.md](SPEC.md): problem, personas, user stories, requirements F1-F6, architectural constraints and acceptance criteria.
- [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md): closed decision record and observed integration gaps.
- [AL-79](../plan/to-do/AL-79%20native%20user%20level%20codex%20capture%20across%20worktrees.md): Ready; shared native runtime and end-to-end user-level Codex capture.
- [AL-80](../plan/to-do/AL-80%20global%20opencode%20capture%20with%20shared%20worktree%20runtime.md): Blocked by AL-79; global OpenCode integration using that runtime.
- [AL-81](../plan/to-do/AL-81%20decouple%20manual%20generate%20from%20capture%20registry.md): Ready; manual generation independent of capture state.
- [Roadmap](../plan/TICKETS.md): ticket statuses and reciprocal dependency metadata.
- [Draft PR #125](https://github.com/NiltonDuarte/AgenticLens/pull/125): specification and backlog delivery on `docs/PO-20-user-level-worktree-capture`.

Two tickets keep delivery vertical: the shared runtime ships with a usable Codex integration, then OpenCode adds its distinct plugin binding path. Configuration, routing, binding and publication policies are implemented once. No separate horizontal configuration, UI or testing tickets are introduced. Existing roadmap tickets remain unchanged.

## Manual-generation clarification

The user requested AL-81 to make manual generate independent of the AgenticLens capture registry. The correction is Ready with no prerequisites and is specified in F7. Harness source access, source validation, portable ledger rules and publication serialization remain required; automatic capture continues to enforce registry bindings. This adds one focused correction alongside the two capture integration deliveries, without application changes.

## Agreed behavior

- Require shared YAML in the active checkout or silently skip capture. Select the whole worktree-local file when present, otherwise the base-local file. Missing or invalid required settings report errors. Relative paths resolve from the selected file.
- Keep capture enablement and machine-specific settings local; share project, metrics and executable-selection rules between manual CLI and automatic capture. Worktree-local configuration is user-managed. Base setup preserves valid choices and asks only for missing settings.
- Install once in harness user configuration. Route through validated native session context without per-worktree installation or enrollment. Hooks notify; the worker collects authoritative source snapshots. Report established source-home mismatches and retain pending collection during temporary unreadability.
- Require a shared version pin and optionally prefer valid local build output. Absent local output uses the exact pinned cache; invalid local output errors without fallback. Provision explicitly. Running workers and subprocesses retain their selected build until normal exit.
- Bind using explicit task context without redundant confirmation. Collect unassigned facts locally and publish only after binding. Proven children may inherit a bound parent's ticket while retaining their own identities and worktree scope.
- Publish during activity and checkpoints, coalesce notifications, retry temporary failures within bounded budgets and retain unresolved facts. Keep all state after worktree deletion until explicit cleanup.
- Test root/child capture during development without production certification gates. Preserve portable ledger facts when private source or capture state is unavailable.

## Validation and limits

Local Markdown links resolve. Both development tickets have unique specifications and roadmap entries, with matching statuses and reciprocal dependencies. Documentation checks pass. Application, test and runtime paths are unchanged by the PO work; the prior full `make` result at `1ce7cc25` is reused. This is author validation, not an independent implementation review.

Local investigation demonstrated a Linux amd64 build and checksum-verified offline installation. Codex worktree hook discovery exposed a dependency on original-checkout hook files. User-home hooks were discoverable but untrusted at the last observation; a live callback, binding and publication round trip remains unverified. The specification addresses these gaps; this ticket does not implement the migration.

## Session telemetry

Native `generate --source codex` refreshed PO-20 from the authoritative root and planner-child rollouts after the user requested deletion of the AgenticLens capture registry. A verified SQLite backup was retained outside Git; only the registry and its SQLite sidecars were deleted. Harness source stores and existing ledger history were preserved. The CLI recreated its registry and recorded PO-20 manual selections.

The first post-reset attempt reported stale installation coordinates. Explicitly selecting the configured state base with `--capture-dir /home/niltonduarte/.agenticlens/state` allowed generation using the same configured project namespace. The retained installation metadata still reports coordinate drift; automatic integration has not been repaired. No application code, configuration or source identity was changed. Live automatic hooks remain unverified; successful manual generation does not establish live capture coverage.

The refreshed ledger contains Codex root `01a1b268-f31e-7503-a954-4f4fd77370b2` and source-proven planner child `01a1b291-2ff1-7830-81d8-7d0dff387ee3`, with the parent link preserved. [TELEMETRY_STATUS.json](TELEMETRY_STATUS.json) records the extraction cutoff and remaining limits. It is a current source snapshot, not final accounting for activity after extraction. Raw transcripts remain outside Git.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 126
- **Active Agent Duration:** 2h 4m (recorded duration including 190s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 450.0k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 208.5k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 98,921 reasoning tokens (29.9% of generated output)
- **Tool Health & Self-Correction:** 347 tool calls (18 self-corrections, 94.8% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+163.4k tokens) — Context tokens surged by 163.4k to 163.4k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.0k tokens) — Context tokens reached 254.0k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 388,552 | 16,674 | 5,996,160 | 93.9% | 208.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 938,212 | 215,400 | 115,524,224 | 99.2% | 450.0k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **2** | **1,326,764** | **232,074** | **121,520,384** | **98.9%** | **450.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 2 | 1,326,764 | 232,074 | 121,520,384 | 98.9% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **2** | **1,326,764** | **232,074** | **121,520,384** | **98.9%** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

Capture observation finalized: 2026-11-08T14:34:07.904911895Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 2; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 2; reserved bytes: 10949775. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1b291-2ff1-7830-81d8-7d0dff387ee3: current; last successful source read 2026-11-08T14:34:07.690601615Z; accepted source checkpoint 2026-11-06T20:41:25.096Z.
Owner codex/01a1b268-f31e-7503-a954-4f4fd77370b2: current; last successful source read 2026-11-08T14:34:07.885465142Z; accepted source checkpoint 2026-11-08T14:34:07.557Z.
