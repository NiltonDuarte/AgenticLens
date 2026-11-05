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

Iteration02 planning completed by planner gpt-6.1-sol/high. [Delta plan](iterations/02/PLAN.md) supports single-command exact-current-caller verification, optional supported read-only other-session status, and explicit unknown/unsupported retention. A live exact-interval observation supports provisional elapsed time, not CPU utilization; unobservable pauses are disclosed rather than requiring unprovable continuous execution. Pokes that start new work remain excluded. Native transport help does not establish exact-turn status support; no universal cross-session probe is promised. Application changes await approval of this changed scope under agentic-lens-dev.

Documentation preference: paraphrase steering and approval decisions without reproducing user messages in plans or contracts. Direct message quotations were removed from both iteration plans and the summary; the contract contains no direct message quotations. This documentation correction does not establish iteration02 implementation approval.

## Main integration requested on 2026-11-05

Pending planning documents were checkpointed at e9ca2676 before merging origin/main40b16def. Fifteen content conflicts span CLI/capture composition, OpenCode family readers, duration/optional DTOs, API validation, FleetOverview and tests. Coder gpt-6.1-sol/high resolved the contents, preserving upstream optional-category field-presence semantics and shared-session warnings plus AL-64 duration evidence and six fixes. See [integration plan](iterations/01/INTEGRATION.md) and [implementation evidence](iterations/01/INTEGRATION_IMPLEMENTATION.md).

Focused affected Go packages, frontend typecheck/build and387 frontend tests pass. One initial CLI worker-lease deadline failed despite acknowledged publication; the unchanged isolated rerun passed, and no assertion or timeout was weakened. The failed command is retained in integration evidence. Independent interaction tests, integrated make and affected review are pending; the earlier integrated make remains historical evidence for the pre-merge implementation only. Iteration02 hybrid implementation remains unapproved and untouched.
