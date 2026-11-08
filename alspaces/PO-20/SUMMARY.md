# PO-20: Machine-wide automatic capture across Git worktrees

## Scope and outcome

Define automatic AgenticLens capture that works from a base checkout or Git worktree after one user-level installation in Codex or OpenCode. This Project Owner ticket records the completed decision review, product requirements and the smallest practical development backlog. Implementation belongs in separate agentic-lens-dev sessions.

No application code, tests, runtime scripts or machine-local integrations are changed by this specification work. The shared development pin selects the locally built version. No developer workspaces, technical contracts or implementation plans are created.

## Deliverables

- [SPEC.md](SPEC.md): problem, personas, user stories, requirements F1-F6, architectural constraints and acceptance criteria.
- [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md): closed decision record and observed integration gaps.
- [AL-79](../plan/to-do/AL-79%20native%20user%20level%20codex%20capture%20across%20worktrees.md): Ready; shared native runtime and end-to-end user-level Codex capture.
- [AL-80](../plan/to-do/AL-80%20global%20opencode%20capture%20with%20shared%20worktree%20runtime.md): Blocked by AL-79; global OpenCode integration using that runtime.
- [Roadmap](../plan/TICKETS.md): ticket statuses and reciprocal dependency metadata.
- [Draft PR #125](https://github.com/NiltonDuarte/AgenticLens/pull/125): specification and backlog delivery on `docs/PO-20-user-level-worktree-capture`.

Two tickets keep delivery vertical: the shared runtime ships with a usable Codex integration, then OpenCode adds its distinct plugin binding path. Configuration, routing, binding and publication policies are implemented once. No separate horizontal configuration, UI or testing tickets are introduced. Existing roadmap tickets remain unchanged.

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

Fresh native extraction for PO-20 is incomplete because the source session has an existing conflicting ticket binding and the CLI offers no supported reticket operation. Capture bindings and source identities were not altered to bypass that guard.

The ledger retains the last native snapshot with PO-20 ticket metadata; source facts and timestamps are unchanged. [TELEMETRY_STATUS.json](TELEMETRY_STATUS.json) records its provenance and incomplete refresh. Later PO activity is not included. Rendering retained facts does not establish fresh extraction or working automatic capture.

The retained Codex root is `01a1b268-f31e-7503-a954-4f4fd77370b2`; its source-proven planner child is `01a1b291-2ff1-7830-81d8-7d0dff387ee3`. Their identities and parent link are preserved. Raw transcripts remain outside Git.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 115
- **Active Agent Duration:** 1h 41m (recorded duration including 111s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 385.6k tokens (86% of 450k healthy ceiling, 0 compactions) | Subagents peak 208.5k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 81,821 reasoning tokens (30.3% of generated output)
- **Tool Health & Self-Correction:** 302 tool calls (11 self-corrections, 96.4% success rate)

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
| **Unknown** | 1 | 799,940 | 171,350 | 103,065,600 | 99.2% | 385.6k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **2** | **1,188,492** | **188,024** | **109,061,760** | **98.9%** | **385.6k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 2 | 1,188,492 | 188,024 | 109,061,760 | 98.9% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **2** | **1,188,492** | **188,024** | **109,061,760** | **98.9%** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |
