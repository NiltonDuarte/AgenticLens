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

On 2026-11-04 the user explicitly approved the proposed contract and iteration 01 plan: "Approve scope and begin implementation." Implementation is authorized within these boundaries.

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
