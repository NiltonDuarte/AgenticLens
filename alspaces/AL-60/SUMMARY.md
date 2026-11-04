# AL-60 Delivery Summary

Delivered the approved search slice and the explicitly authorized shared capture repair. Initial checkout was clean at 122e983f8c365a2aa248b1b6d492252555e922ae. Iteration 02 resumed checkpoint 7b00cc1 after the user approved “Authorize AL-60 shared capture repair”, superseding the historical read-only hold. PR #101 merged at 6244585dc5cc41ef012cd3e1fb3b964ccd910a95. Iteration 03 is a separate test-only cleanup repair before AL-61 PR #104; MGT-2 owns its merge.

## Delivered behavior

Presentation: Search tickets matches IDs/titles case-insensitively inside comparison exclusion choices. Query changes preserve exclusions, counts, chart points/geometry and hidden selections; clear/no-match recovery and keyboard interactions remain available. Search source is unchanged from reviewed application revision 7a50a5e.

Adapters/CLI: optional harness inspection has its own allowance, separate from required installation phases. Explicit caller cancellation remains authoritative. Native timeout/cancellation and execution failures are distinguished from observed identity mismatches. Installer subprocesses own Unix groups or Windows jobs, bound pipe draining, and clean descendants on supported cancellation/completion paths. SIGTERM cancels installer work. Final-write and completed-recovery cancellation preserve rollback/recoverability instead of reporting success.

Tests: disposable harnesses replace host-dependent inspection, including hostile inherited-flag controls. Successful publication fixtures use a distinct bounded policy; deliberate expiry retains pending generation, accepted output, acknowledgement and ownership/recovery assertions. Production worker policy, domain/API/ledger schemas and global timeout defaults are unchanged. AL-55 telemetry is preserved.

## Review and attribution

Iterations 01/02 consolidated result: APPROVED, with one explicitly accepted nonblocking test-coverage recommendation below. Hosted CI is not claimed as locally executed.

Iteration 01: all four responsibilities under gpt-6.1-sol/high and gpt-6-luna/medium approved the unchanged search source. Its original integrated gate failures remained incomplete until the authorized capture repair passed integration. [Search compact review](iterations/01/REVIEW_SUMMARY.json).

Iteration 02: 19 executed review runs across first discovery, an intermediate correction wave and final review. Five queued intermediate cells were superseded before dispatch after a new recovery finding; no task or approval was fabricated. Every final responsibility/profile cell completed APPROVED. Dispatcher settings were acknowledged; no source-observed identity or cost is inferred from prose. Behavior/integration, data/failure, architecture/maintainability and test-effectiveness each ran under both configured profiles. [Capture compact findings](iterations/02/REVIEW_SUMMARY.json).

| Finding | Severity | Discovering responsibility/model/effort | Disposition and fixing actor |
| --- | --- | --- | --- |
| ARCH-01: inherited-pipe wait | 6 | Architecture, gpt-6.1-sol/high | Fixed owned subprocess cleanup/draining; coder gpt-6.1-sol/high. |
| ARCH-02: SIGTERM child lifetime | 4 | Architecture, gpt-6-luna/medium | Fixed installer signal cancellation; coder gpt-6.1-sol/high. |
| DATA-01: final-write cancellation | 5 | Data/failure, gpt-6-luna/medium | Fixed rollback/acceptance checkpoints; coder gpt-6.1-sol/high. |
| TEST-01: inherited fixture guard | 4 | Test effectiveness, gpt-6.1-sol/high and gpt-6-luna/medium | Fixed disposable setup and hostile controls; tester gpt-6.1-sol/high. |
| DATA-02: completed recovery cancellation | 4 | Architecture, gpt-6.1-sol/high and gpt-6-luna/medium, new discovery during re-review | Fixed retained-journal checkpoint/retry guard; coder gpt-6.1-sol/high. |
| TEST-02: Unix descendant oracle | 2 | Test effectiveness, gpt-6.1-sol/high, final re-review | Accepted nonblocking coverage recommendation; AL-60 maintainer owns follow-up synchronized descendant-death/reachability fixture. No current leak demonstrated. |

Discovery and verification are distinct: test-effectiveness gpt-6.1-sol/high reproduced ARCH-01 after seeing the active-plan mention, so that contribution is retained as verification rather than blind discovery. The tester and parent integrator, both gpt-6.1-sol/high, verified the focused oracles and actual platform behavior. Initial manifest variability was a hypothesis; controlled fixtures reproduced only optional HarnessVersion/Capability differences, not the specific field behind the unretained historical failure.

Reviewed capture fingerprint: 5e332cad1bfb0d1b25af2d9a8def5bfbdcd8e2b8364c3023131733041d4c8493 against checkpoint 7b00cc1. The integrated gate exposed a vet limitation around a closure capturing reassigned cancel functions. Coder gpt-6.1-sol/high replaced it with explicit first/second defers; parent gpt-6.1-sol/high verified observational equivalence: same contexts, timers, handoff and final cancellation; the first cancellation already occurs before the probe and the extra deferred call is idempotent. All scoped approvals carry through that mechanical change, with vet and focused race guards passing. Final integrated source fingerprint: 3251b53bc2bfdbd9c7e6ea47d5b8e793fc22f4f9e6103ee268d9fbc622a005db. Main integration added only AL-68 workspace documentation, with no application dependency delta.

## Verification

The merged iteration 02 application/test/CI source matched tested revision `057d410253aecd734551995ff1bedb38e94a1f19`. Full make, Go race checks, 786 frontend tests, five platform builds, standalone assets and actual Windows ownership/recovery checks passed. Meaningful commands, outcomes, historical failures and limits are consolidated in [VERIFICATION.md](iterations/02/VERIFICATION.md).

## Evidence cleanup

User direction supersedes retaining raw outputs in the final tree. From clean head eb3d515d85493372aab099d7cda82390e3c63fb3, removed 17 per-command outputs, two evidence indexes and the superseded standalone diagnosis: 30 ticket files become 10. Only the verification summary remains for evidence. Contracts, PR description, ledger and both compact findings remain unchanged; discovery credit, verification distinctions and accepted TEST-02 debt/owner are preserved. No application/test/CI change, history rewrite or force push. That cleanup shipped in merged PR #101; it does not deliver iteration 03.

## Iteration 03 follow-up

The user authorized the minimal shared capture isolation-test teardown correction after PR #104's Backend job failed only with TempDir RemoveAll directory-not-empty. Backend and CI at its reported head 64d1e3d were byte-identical to main. A controlled native blocked at configuration admission passes the old ACK/free-lease oracle, then recreates worker.lock after release. This proves a reachable late writer; it does not identify the exact original CI file.

The test-only barrier observes inherited worker.log handles through startup, generation, successor handoff and registry close. Automatic publication, exact session/token/project isolation, pending/ACK and exclusive ownership assertions remain. No production, CI or timeout policy change. Final consolidated review: APPROVED for source `7c97677dacbdb71466bb0d041704c2cc2b2619e2`, diff fingerprint eebf9384cb246dac0bff4783a8ba23f5fb2c9a5973e3ac0cf7984f791ade9dcb against e19a1983. Fifteen executed runs cover all four responsibilities under gpt-6.1-sol/high and gpt-6-luna/medium; behavior approvals carry for the unchanged normal path, and both profiles re-reviewed the corrected data/failure, architecture and test boundaries. Dispatcher identities are acknowledged, not independently source-observed. [Compact findings](iterations/03/REVIEW_SUMMARY.json).

DATA-03 (severity 5) was discovered by data/failure gpt-6-luna/medium and fixed by coder gpt-6.1-sol/high: preserve the deadline error while draining completion, retaining fixtures if the observer fails. TEST-03 was a severity-5 test-effectiveness claim from gpt-6-luna/medium, reconciled to severity 0 and refuted: the same wrapper receives Background on retry and signals fallback. Test-effectiveness gpt-6.1-sol/high independently traced that transition; gpt-6-luna/medium executed the focused race case (1.946s) and approved the refutation. Verification contributions are distinct from discovery.

Tester and integrator gpt-6.1-sol/high verified the controlled failure path and full make. Full make passed in 217.7s after main synchronization, including Go race, 786 frontend tests, workflow/distribution checks and five platform builds. No source changed after the freeze. Hosted final-head checks remain external merge gates. Windows runtime evidence is helper-level, not full capture integration; Fatal/Goexit and permanently broken-observer behavior were inspected rather than executed. The earlier accepted TEST-02 debt remains owned by the AL-60 maintainer. [Plan](iterations/03/PLAN.md) and [verification](iterations/03/VERIFICATION.md).

## Telemetry and delivery

Final source-grounded refresh follows verification/review and preserves 54 exact session IDs: the original root, ten structurally linked native children, 42 delegated review roots across three iterations, and one unparented host continuation containing replayed historical context. Parent/role and steering interpretation of that replay remain source limitations. Delegate source records have no parent/role linkage; exact selection preserves them as independent Codex roots. All36 baseline IDs survived every refresh; the three new native children's parent/source/role match actual metadata. No other-harness baseline existed for AL-60. Four unchanged source checkpoints do not establish newer facts, so accepted facts remain explicitly retained/pending rather than fabricated as current. Pricing remains unavailable, and owner reads do not form a common as-of snapshot. Raw transcripts stay outside Git. Metrics below are generated by the built native CLI without unsupported wrapper flags.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 107
- **Active Agent Duration:** 4h 41m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 450.6k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 425.8k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.8 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 2.3, Tester 2.3, Explorer 1.0
- **Test-to-Production Code Ratio:** 361:0 (+361 test lines / +0 prod lines)
- **Code Delivered:** +361 / -1 lines across 4 code files (excluding docs)
- **Reasoning Effort:** 235,029 reasoning tokens (30.7% of generated output)
- **Tool Health & Self-Correction:** 1,096 tool calls (87 self-corrections, 92.1% success rate)

### Anomalies & Health Flags
- **Status:** 16 execution anomalies detected (9 CRITICAL, 7 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+155.8k tokens) — Context tokens surged by 155.8k to 155.8k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+156.4k tokens) — Context tokens surged by 156.4k to 156.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (403.8k tokens) — Context tokens reached 403.8k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.4k tokens) — Context tokens reached 200.4k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (203.6k tokens) — Context tokens reached 203.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.2k tokens) — Context tokens reached 200.2k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.6k tokens) — Context tokens reached 450.6k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (30 cycles) — Subagent roles 'coder' and 'tester' alternated for 30 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (18 cycles) — Subagent roles 'coder' and 'tester' alternated for 18 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.6k tokens) — Context tokens reached 162.6k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.1k tokens) — Context tokens reached 163.1k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.7k tokens) — Context tokens reached 250.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (425.8k tokens) — Context tokens reached 425.8k, breaching the 250k warning ceiling.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 3 | 521,628 | 12,387 | 4,792,064 | 90.2% | 425.8k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **coder** | 3 | 271,419 | 74,859 | 25,208,064 | 98.9% | 250.7k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 3 | 198,476 | 40,277 | 14,986,240 | 98.7% | 224.8k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 44 | 3,312,046 | 400,355 | 107,562,624 | 97.0% | 450.6k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 44 unavailable / 44 invocations) |
| **explorer** | 1 | 115,406 | 2,546 | 1,165,568 | 91.0% | 139.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **54** | **4,418,975** | **530,424** | **153,714,560** | **97.2%** | **450.6k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 54 unavailable / 54 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 22 | 1,050,743 | 89,401 | 7,281,920 | 87.4% | Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations) |
| **gpt-6.1-sol** | 32 | 3,368,232 | 441,023 | 146,432,640 | 97.8% | Unavailable (0 observed, 0 partial, 32 unavailable / 32 invocations) |
| **Total** | **54** | **4,418,975** | **530,424** | **153,714,560** | **97.2%** | **Unavailable (0 observed, 0 partial, 54 unavailable / 54 invocations)** |
