# AL-64 work summary

## Initial state

- Initial revision: a84adbe5be46c2ef52ede0987a417562715d38e1.
- Initial Git status: clean.
- Initial branch: AL-64/investigate-fix-active-turn-duration-codex-opencode.
- Work branch: fix/AL-64-active-turn-duration.
- Workspace: alspaces/AL-64, from configured workspaces.dir.
- Scope: diagnosed Codex active-turn duration fix and independent OpenCode v1/v2 investigation with remediation of confirmed defects, compatible persistence and presentation, generation and refresh regressions.

## Iteration 01

Planning completed by the planner (gpt-6.1-sol, high; inherited runtime settings). Proposed [contract](CONTRACT.md) and [plan](iterations/01/PLAN.md) are ready for explicit approval under agentic-lens-dev. No application code has changed.

Grounding confirms Codex counts completed lifecycle pairs, OpenCode v1 counts completed assistant intervals, and v2 falls back to row update time without explicit completion. Available local OpenCode samples contain completed activity only; source-proven unfinished runtime behavior remains unverified. See the plan for the read-only inventory and verification limits.

The AL-55 timing observations in the ticket are user-supplied evidence, not observations independently reproduced by this session. Its live workspace and ledger must remain untouched.

## Verification and environment

- Planner and orchestrator: git diff --check passed for planning artifacts.
- Required application verification has not run during planning.
- agentic-lens --help: command unavailable on PATH. Build/use the repository CLI and inspect supported options before final telemetry extraction.
- Telemetry extraction is deferred until the final delivery step. No ledger has been generated or claimed complete.

## Approval

On 2026-11-04 the user explicitly approved the proposed contract and iteration 01 plan, authorizing implementation within those boundaries.

## Implementation candidate and native evidence

Coder (gpt-6.1-sol/high, inherited runtime) implemented cutoff-aware extraction, compact optional duration evidence, persistence validation, aggregate coverage and CLI/dashboard labels. Application edits are frozen for independent testing. See [investigation](iterations/01/INVESTIGATION.md) and [native OpenCode evidence](iterations/01/RUNTIME_EVIDENCE.md). Required integrated checks and final review remain pending.

The runtime test engineer (gpt-6.1-sol/high, inherited runtime) verified isolated native OpenCode 1.19.5 and 2.2.0 streaming/tool/completion/interruption shapes. Native v1 SIGINT can leave a stored running interval unchanged after process exit. The orchestrator requested clarification: count explicitly open source intervals provisionally with disclosed process-liveness uncertainty, or require extra live-process evidence. This is pending; no final accounting-policy approval or completed delivery is claimed. Existing approval for the rest of iteration 01 remains valid.

Reader-first v6 extension limit: expanded readers accept old valid facts and new optional evidence; pre-change strict readers reject the emitted extension. Upgrade consuming readers before writer rollout. No release or distribution rollout is performed by this ticket.

## Independent verification checkpoint

The independent tester (gpt-6.1-sol/high) passed full Go race tests across all test-bearing packages, 834 frontend tests, typecheck and lint on frozen candidate 01. See [test evidence](iterations/01/TEST_REPORT.md). Native live CLI generation/active recomputation/completion/repeat passed v1 (3/4/21/21 seconds) and v2 (2/3/20/20 seconds); provisional contributions were replaced by completed snapshots. These are pre-correction checks, not final approval.

All four Sol discovery focuses completed with changes requested: invalid Codex completion classification (shared discovery across behavior/data/architecture/test-effectiveness); v1 future-message suppression (behavior); v1 later-update suppression (architecture); fallback recovery losing operation cutoff (data); missing no-tool streaming regression guards (test effectiveness). Every listed discovery used gpt-6.1-sol/high. Detailed reports remain active; Luna discovery and verified corrections are pending. No integrated make, final approval, telemetry generation, commit or PR publication has been claimed.

## Correction checkpoint

Six confirmed issues are corrected. Shared initial Codex invalid-completion discovery: all four Sol personas, gpt-6.1-sol/high. V1 future-message cutoff discovery: behavior gpt-6.1-sol/high and architecture gpt-6-luna/medium. V1 progress-update cutoff discovery: architecture gpt-6.1-sol/high. Fallback cutoff discovery: data gpt-6.1-sol/high. Missing no-tool streaming regression discovery: test effectiveness gpt-6.1-sol/high. Invalid/missing Codex start discovery: architecture gpt-6.1-sol/high during correction review. Coder gpt-6.1-sol/high implemented fixes; tester gpt-6.1-sol/high verified red/green overlays, streaming mutants and affected race suites.

Final candidate03 manifest: f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49. Architecture/test-effectiveness Sol correction reviews approve; behavior/data Sol candidate02 correction approval carries forward with the dependency rationale in REVIEW.md. Luna correction reviews and user liveness clarification remain pending, so active detailed reviews are retained. No consolidated delivery approval or completed PR is claimed.

Native Codex source lifecycle/child ownership records were inspected. Direct unchanged source-copy fresh generation in an isolated project was rejected by the correct source-project ownership boundary. Native Codex generate/refresh remains explicitly unverified; deterministic adapter-to-output regressions pass. No synthetic identity seed, live source edits or early real-ticket telemetry was used to bypass that limit.

All required correction review cells now pass (gpt-6.1-sol/high and gpt-6-luna/medium), with dependency-justified carryforward of Sol behavior/data correction approvals. All six issues are fixed and regression-verified. Parent is committing the candidate for upstream integration and required make. The source-state/liveness clarification still gates final approval/publication; detailed active review evidence is retained until that gate is resolved. Telemetry remains deferred.

## Integrated verification in progress

Implementation commit: 44db8658097e9fb27c7ed1149ee132906903d862. origin/main a84adbe5be46c2ef52ede0987a417562715d38e1 was already integrated, so reviewed product/test evidence remains applicable. Parent started required make; final result and standalone asset verification are not yet recorded.

[Compact findings and discovery credit](iterations/01/REVIEW_SUMMARY.json) records all six corrected findings and all independent reporting persona/model/effort tuples. Detailed reports remain active because the source-state/liveness clarification and final consolidated delivery approval are pending. No review evidence has been deleted. Coder/tester/re-review confirmation is verification credit, distinct from the retained discovery tuples.

## Integrated verification result

Parent verified implementation revision 44db8658097e9fb27c7ed1149ee132906903d862 after origin/main integration (already up to date). Required make passed exit 0: dependency setup, frontend build, Go vet/frontend typecheck/lint, repository release/install/capture checks, all Go race packages, all 834 frontend tests across 37 files, canonical frontend rebuild and all configured platform binaries. Full command log remains outside Git at /tmp/al64-integrated-make.log. No application/test files changed during verification.

Parent verified standalone bin/agentic-lens-linux-amd64 reports commit 44db8658. In an isolated empty-workspace server run with unavailable hydration, HTTP index and both bundled assets matched web/dist and backend/internal/adapters/web/dist byte for byte. JavaScript index-z1wgKOcT.js SHA256 359da8a14fe52b72dd85ce196db33a3eb1d5c85e2fc76427308d97cb9f991869; CSS index-C3WE_otj.css SHA256 da46ce98a0a085703e286d00d417781999b3525de6b168d8a68d2005db46d4ba. The probe server exited and its disposable workspace was removed.

Consolidated status remains INCOMPLETE solely for the pending acceptance-policy clarification on explicitly open native intervals versus live-process proof. Native Codex fresh generation/refresh remains an explicitly unverified runtime path after isolated project-ownership rejection; deterministic regressions and real source lifecycle evidence are retained. All confirmed correctness findings are fixed and scoped reviews pass. No final approval, publication, ticket completion or real-ticket telemetry is claimed. Existing implementation approval is preserved; the pending question only concerns the discovered source-state ambiguity.

## Iteration 02 steering: agent-assisted liveness policy

Source: user asks to expose the decision to the agent, use provisional accounting for its own open session and check other sessions for liveness. This supersedes the earlier binary choice; do not keep asking the user to select unconditional source-state counting versus universal live-process proof. Proposed acceptance is per-session evidence, exact current caller/turn versus independently checked other session, and explicit unknown/unsupported results. Exact proof boundaries and supported probes require a grounded delta plan and scope approval before new application changes.

Preserved guarantees: all six corrected defects, completed measurements, deterministic cutoffs, portable recorded facts, exact-owner replacement and unavailable-source retention. Existing candidate03 reviews, full make and standalone checks remain valid for unchanged code. Superseded decision: one blanket policy for every open session. Verification impact: new caller/turn correlation, read-only probe semantics, races, unknown fallback and per-session agent decisions; re-review only affected scopes with both configured profiles. No session has been poked, woken or messaged.

Planner owns iterations/02/PLAN.md. Existing CONTRACT.md stays immutable; iteration02 records the proposed delta. The candidate is committed and the implementation remains unchanged pending this new scope gate.

Iteration 02 planning completed by planner gpt-6.1-sol/high. [Delta plan](iterations/02/PLAN.md) supports single-command exact-current-caller verification, optional supported read-only other-session status, and explicit unknown/unsupported retention. A live exact-interval observation supports provisional elapsed time, not CPU utilization; unobservable pauses are disclosed rather than requiring unprovable continuous execution. Pokes that start new work remain excluded. Native transport help does not establish exact-turn status support; no universal cross-session probe is promised. Application changes await approval of this changed scope under agentic-lens-dev.

Documentation preference: paraphrase steering and approval decisions without reproducing user messages in plans or contracts. Direct message quotations were removed from both iteration plans and the summary; the contract contains no direct message quotations. This documentation correction does not establish iteration02 implementation approval.

## Main integration requested on 2026-11-05

Pending planning documents were checkpointed at e9ca2676 before merging origin/main 40b16def. Fifteen content conflicts span CLI/capture composition, OpenCode family readers, duration/optional DTOs, API validation, FleetOverview and tests. Coder gpt-6.1-sol/high resolved the contents, preserving upstream optional-category field-presence semantics and shared-session warnings plus AL-64 duration evidence and six fixes. See [integration plan](iterations/01/INTEGRATION.md) and [implementation evidence](iterations/01/INTEGRATION_IMPLEMENTATION.md).

Focused affected Go packages, frontend typecheck/build and387 frontend tests pass. One initial CLI worker-lease deadline failed despite acknowledged publication; the unchanged isolated rerun passed, and no assertion or timeout was weakened. The failed command is retained in integration evidence. Independent interaction tests, integrated make and affected review are pending; the earlier integrated make remains historical evidence for the pre-merge implementation only. Iteration 02 hybrid implementation remains unapproved and untouched.

## Main merge completed and verified

Merge 1378860955deb1acc05924953b8b959b825d3ebc integrates origin/main 40b16def into this feature branch and resolves 15 conflicts. Required make passes on that revision, including all Go race packages and 863 frontend tests across 38 files. The standalone host binary serves exact current canonical/embedded assets (index-BGuN6wTp.js and index-BGH-Pi_l.css). No application/test changes followed verification.

All eight integration persona/profile cells completed: behavior, data/failure, architecture and test-effectiveness on gpt-6.1-sol/high and gpt-6-luna/medium. Seven scoped approvals; data gpt-6-luna/medium reported INT-DATA-01 severity 6. Parent verified the implicated block is byte-identical to upstream main, refuting its introduced-merge classification. The pre-existing delegation-retention concern is not fixed; owner is the capture adapter maintainer, follow-up ticket unassigned. [Original finding](iterations/01/INTEGRATION_DATA_LUNA.md), [compact findings](iterations/01/REVIEW_SUMMARY.json), and [consolidated integration outcome](iterations/01/INTEGRATION.md) preserve the debt, exact reporter tuple and source-comparison evidence. No confirmed merge regression remains.

Consolidated local integration is APPROVED, with that explicit pre-existing follow-up. Independent tester gpt-6.1-sol/high verified combined output-policy/duration behavior; test-effectiveness gpt-6.1-sol/high independently confirmed a duration-evidence-loss mutation is caught. Parent verified final make and standalone assets. Detailed completed no-finding integration reports were compacted; the debt report and still-open ticket reviews remain.

Iteration 02 remains unapproved and unimplemented. The merge request does not approve that scope, ticket completion, PR publication or release. Telemetry remains delayed until the final iteration delivery step; no successful or empty ledger is manufactured for this intermediate merge.

## Draft PR publication checkpoint

The user authorized creating a PR for the current branch. Publish a draft for the reviewed initial implementation and verified main merge; do not treat this as approval to implement iteration02 or mark AL-64 complete. The existing todo specification and in-progress ticket entry remain. PR title/body are prepared from PULL_REQUEST.md; no application changes followed reviewed merge13788609.

Final telemetry for this publication checkpoint uses the built native binary (the installed wrapper is unavailable on PATH, and wrapper-only --al-offline/--al-project options are unsupported by the native parser). Supported source/ticket/nonce/output options were inspected. Use explicit source codex and ticket type fix. The generation is real current-project extraction, not the previously blocked isolated source-copy experiment.

Source inventory before extraction: expected root01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6,18 native child sessions with that authoritative parent, and12 additional gpt-6-luna review sessions whose native metadata has no parent link. Source task records associate those12 with the known review report assignments, but no native parent relationship will be fabricated. Extraction coverage and pricing limitations must be checked before committing the ledger. No pre-existing AL-64 ledger or other-harness facts exist to preserve at this first publication.

The first native generate attempt failed before ledger publication: existing configured capture state has no worktree mapping table and is an unsupported coordinate layout. That state is preserved, not reset or migrated. Retry manual nonce-based extraction using a fresh isolated capture-state directory; no existing AL-64 ledger is present, and no successful registered-cache recovery is claimed. The stale registry is unavailable evidence, distinct from the native source sessions.

The isolated-state retry passed capture admission but stopped at the required first-ledger title check. No ledger was published. The next attempt supplies the explicit AL-64 ticket title together with a fresh literal nonce; configured legacy state remains untouched.

## Publication telemetry coverage

Generation succeeded using explicit codex source, fix type and an isolated manual capture state. The raw ledger contains root 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6 and all 18 authoritative native children: 19 unique Codex sessions/invocations, no duplicate sessions and no missing native child from the source inventory. Provisional root time is explicitly labeled with cutoff 2026-11-05T10:36:46.850981184Z; no source completion timestamp is invented for that open invocation. There was no prior AL-64 or other-harness ledger to overwrite. The generated native-source snapshot is real current-project evidence; it does not claim the separate native-refresh lifecycle experiment was performed.

Coverage is incomplete for the full orchestrated effort: 12 gpt-6-luna review sessions have no authoritative native parent relationship and are absent from this ledger. Their source task assignments match the delegated reviews, but they were not attached through guessed ancestry. Missing source session IDs: 01a1a804-8077-7be0-8ef6-630b932f4717, 01a1a804-8328-74e0-82a6-ab6849d1b9d9, 01a1a804-86b4-7330-b8d7-d118f11cfbaf, 01a1a806-afb8-7072-a03a-884dcdb9b113, 01a1a815-e50b-7061-b99f-4b1ae3727f3e, 01a1a815-e80f-7a71-8bbf-b7bb2eb6bb37, 01a1a815-ec8e-7fd0-a760-769e3ae923af, 01a1a817-98d3-74b1-9db2-3742ca1ee17f, 01a1a944-15e0-7b03-9f1e-65e8c3887135, 01a1a944-17e9-72d3-862b-43cebfba1f61, 01a1a944-1b4f-71c1-a02c-c3a9899f8a89, 01a1a945-dd24-7e73-ac4e-78d0df278ce9. The source inventory and task-to-report matches were inspected outside Git; raw transcripts remain outside Git.

Pricing evidence is unavailable for all 19 recorded invocations; no zero cost or complete total cost is claimed. The generated duration/token/cache tables cover only the 19 recorded native sessions and exclude those 12 delegated review sessions. Configured legacy capture state remains untouched and unavailable; successful extraction used the explicit manual-recovery state. Native read-only parent metadata, rather than the legacy registry, established the 18 child relationships.

Draft publication is authorized at this checkpoint. Iteration 02 implementation and final ticket completion remain pending; a later delivery must refresh telemetry to include subsequent work and preserve existing facts.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 21
- **Active Agent Duration:** 3h 28m (excludes user wait time; includes 350s provisional across 1 invocation(s); snapshot cutoffs: [2026-11-05T10:36:46.850981184Z])
- **Context Window Health:** ELEVATED — Root agent peak 306.1k tokens (68% of 450k healthy ceiling, 0 compactions) | Subagents peak 269.5k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.6 avg turns/run (peak: 5 turns) | Planner 3.0, Coder 5.0, Tester 2.3, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.1:1 (+810 test lines / +752 prod lines)
- **Code Delivered:** +1,562 / -206 lines across 48 code files (excluding docs)
- **Reasoning Effort:** 86,886 reasoning tokens (21.2% of generated output)
- **Tool Health & Self-Correction:** 934 tool calls (66 self-corrections, 92.9% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (4 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (21 cycles) — Subagent roles 'coder' and 'tester' alternated for 21 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (6 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (265.0k tokens) — Context tokens reached 265.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.4k tokens) — Context tokens reached 250.4k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> tester (3 cycles) — Subagent roles 'reviewer-architecture' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 208,076 | 13,965 | 2,110,208 | 91.0% | 127.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 378,928 | 50,715 | 20,779,136 | 98.2% | 269.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 3 | 543,238 | 64,155 | 22,429,440 | 97.6% | 197.5k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 1 | 802,890 | 108,876 | 46,286,848 | 98.3% | 306.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 316,007 | 24,085 | 4,185,088 | 93.0% | 159.4k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 3 | 238,274 | 19,655 | 2,928,896 | 92.5% | 106.7k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-data** | 3 | 257,245 | 20,345 | 3,153,664 | 92.5% | 103.7k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 3 | 240,542 | 20,975 | 3,397,120 | 93.4% | 108.3k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Total** | **19** | **2,985,200** | **322,771** | **105,270,400** | **97.2%** | **306.1k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 19 unavailable / 19 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 19 | 2,985,200 | 322,771 | 105,270,400 | 97.2% | Unavailable (0 observed, 0 partial, 19 unavailable / 19 invocations) |
| **Total** | **19** | **2,985,200** | **322,771** | **105,270,400** | **97.2%** | **Unavailable (0 observed, 0 partial, 19 unavailable / 19 invocations)** |

Capture observation finalized: 2026-11-05T10:36:46.900371127Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
