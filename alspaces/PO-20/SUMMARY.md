# PO-20: Machine-wide automatic capture across Git worktrees

## Reclassification and boundaries

The user explicitly reclassified the same GEN-22 initiative as Project Owner work, invoked agentic-lens-po, requested a PO workspace/new branch/updated PR, and asked for as few development tickets as possible. This is a relocation/reclassification of the active initiative, not concurrent work on an unrelated ticket.

Initial status for this PO phase was clean at `c487364e` on `chore/GEN-22-capture-worktree-findings`. Checked current workspaces, roadmap, completed ticket paths, origin/main and open PR names: PO-20, AL-79 and AL-80 were unused. Created `docs/PO-20-user-level-worktree-capture` and moved `alspaces/GEN-22/` to `alspaces/PO-20/`; no alias/tombstone or developer workspace was created.

No application code, tests, runtime scripts, local hook files or trust settings are modified in this PO phase. The pre-existing user-approved shared development pin is retained unchanged. No CONTRACT.md or implementation plan is authored. Implementation must start in separate agentic-lens-dev sessions.

## Product deliverables

- [SPEC.md](SPEC.md): problem, personas/stories, complete functional requirements F1-F6, architectural/portable-history constraints, acceptance and minimal delivery sequencing.
- [CAPTURE_VISION_AND_GAPS.md](CAPTURE_VISION_AND_GAPS.md): consolidated decision record, preserving the closed user review and separating current local observations from intended behavior.
- [AL-79](../plan/to-do/AL-79%20native%20user%20level%20codex%20capture%20across%20worktrees.md): Ready; delivers the shared policies/native runtime and a functional Codex user-level integration end to end.
- [AL-80](../plan/to-do/AL-80%20global%20opencode%20capture%20with%20shared%20worktree%20runtime.md): Blocked by AL-79; delivers the global TypeScript OpenCode adapter and native parity on that runtime.
- [Roadmap](../plan/TICKETS.md): adds reciprocal dependency metadata and preserves all unrelated entries/statuses.
- [PULL_REQUEST.md](PULL_REQUEST.md): PO draft description replacing the old GEN draft.

Two tickets are the smallest practical batch that preserves bounded vertical delivery: common behavior is delivered with real Codex ingestion/CLI/publication, then OpenCode supplies its distinct plugin/runtime binding path. No separate structs/configuration/UI/test-only tickets or unrequested dashboard feature are created. AL-70, AL-73, AL-78 and AL-63 remain coordination points rather than artificial prerequisites or duplicated audit scope.

## Confirmed behavior carried forward

- Require active shared YAML or silently skip capture. Select one whole worktree-local file, otherwise base-local; missing/invalid required settings error. Relative local values retain file origin.
- Capture runtime settings and explicit enablement are local; project/metrics and binary-selection rules are shared and common to CLI/plugin. Worktree-local files are user-managed; base setup preserves valid values and asks only for missing choices.
- Install once in harness user configuration; use native session-context routing with no per-worktree manifest/enrollment or Python requirement. Hooks notify; the worker collects real source snapshots. Known source-home mismatch errors; transient correct-source unreadability remains pending.
- Always require shared version; optional development output is preferred when valid, absent output uses the exact pinned cache, invalid output errors without fallback. Provisioning is explicit. Active workers and their subprocesses retain their selected build until normal exit.
- Bind from explicit task context without redundant confirmation. Collect unassigned facts locally, publish after binding, and inherit only from proven bound parents while preserving child identity/worktree scope.
- Publish during activity/checkpoints, coalesce notifications, retry temporary failures with bounded backoff, preserve unresolved facts, and retain all state after worktree deletion until explicit cleanup. No checkout recreation or redirected publication.
- Root/child verification belongs in development tests and evidence, not production activation gates. Valid portable ledgers remain usable without private source/capture receipts from the producing machine.

## Investigation evidence and limits

The original checkout started clean at `57a3503e`. A local Linux amd64 build, protocol 3 and checksum-verified offline installation were demonstrated. Original Codex worktree hook discovery returned zero; isolated sourcePath controls established original-checkout lookup. The exact six groups worked in an ordinary repository. The parent and planning investigator used gpt-6.1-sol with high reasoning effort; these were troubleshooting actors, not a formal implementation review.

A guarded original-checkout workaround was created and removed at the user's request. Six Codex-home hooks and a local Python launcher were subsequently installed outside Git. They were discoverable but untrusted at the last observation; no live callback, binding, extraction/publication round trip was established. Worktree-local legacy hook/manifest files remain locally excluded because current validation depends on them. The target specification removes that dependency; this PO session does not perform the migration.

Earlier user-authorized legacy database/state resets retained the binary cache and did not delete harness session logs. These are historical operations, not cleanup authority for future installations. The prior full make gate passed at `1ce7cc25`; no application source changed afterward. Prior GEN records remain in Git history, rather than creating a second active workspace.

## PR delivery and session telemetry

The former draft is [#124](https://github.com/NiltonDuarte/AgenticLens/pull/124). GitHub cannot retarget it to the new branch, and renaming an open PR head closes that PR; delivery uses a replacement PO draft and marks #124 superseded without deleting its branch/history.

Attempted native regeneration under PO-20 with conventional type docs before final commit. The existing binding to GEN-22 blocks a fresh extraction; see the explicit incomplete status below. The known Codex root is `01a1b268-f31e-7503-a954-4f4fd77370b2` (T3 Code origin, source class vscode); authoritative metadata proves planner child `01a1b291-2ff1-7830-81d8-7d0dff387ee3`. The original GEN ledger is retained in prior Git revisions. Verify the final PO ledger preserves these source identities and parent link; raw transcripts stay outside Git. Telemetry is a source snapshot, not live capture verification.

PO validation and final delivery results are recorded below. This session does not implement the generated development tickets.

PO artifact validation passed: local Markdown links resolve; AL-79/AL-80 each have one specification and registry entry; status and reciprocal dependency metadata agree. No GEN-22 workspace, developer ticket directories, CONTRACT.md or iteration PLAN.md remain. Comparison with c487364e confirms all PO-phase changes are restricted to alspaces documentation/ledger relocation, preserving the earlier approved configuration and every application/test/runtime path. The previously passing full make result remains applicable to unchanged code; no application tests were added or rerun solely for artifact relocation. This is author validation, not an independent implementation review.

### Telemetry limitation and preserved baseline

The final native generate attempt for PO-20/type docs failed with `Capture ownership conflict: session 01a1b268-f31e-7503-a954-4f4fd77370b2 in selected source store is bound to ticket GEN-22, refusing PO-20`. Inspected the CLI and found no supported reticket operation. Did not reset bindings, mutate SQLite, change source IDs or use independent state to bypass the ownership guard.

Retained the last natively generated ledger from c487364e:alspaces/GEN-22/agentic_metrics.json and applied only the user's explicit ticket reclassification metadata: ticket_id PO-20, ticket_type docs and the PO title. Session/source facts, code snapshot and timestamps are unchanged. Verified all non-ticket metadata values are identical. TELEMETRY_STATUS.json marks fresh extraction incomplete and identifies the retained revision/timestamps; later PO-phase activity is not included. The native renderer may render these retained facts, but rendering is not a new extraction or live-capture proof. This limitation does not authorize adding a reticket feature in this PO session.

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
