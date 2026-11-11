# GEN-25: Remove Windows support

AgenticLens targets Linux and macOS on amd64 and arm64. Windows native adapters, PE/executable naming, PowerShell setup, hook/plugin transports, build/release targets and execution suites were removed. Unsupported selections fail before downloads/install writes. Release checksums and all three upload workflows select only supported binaries/archives, leaving unrelated stale files untouched.

The user clarified that Windows has never been installed and no compatibility with older Windows configurations is required. That explicit instruction supersedes retaining `commandWindows` inspection and recognizing historical multi-platform Git hook templates. Both mechanisms, the retired script fixture, historical-template-only tests and Windows profile environment overrides in test scaffolding are now removed. Only the current generated hook is automatically owned; edited or unknown hooks remain protected. Existing supported POSIX callback detection remains. The approved steering and deletion rationale are recorded in [iteration 02](iterations/02/PLAN.md).

Historical audit records, portable ledger facts, path/security assertions and negative tests rejecting Windows targets/artifacts remain evidence rather than compatibility implementations. No schemas or shared dependency lockfiles were changed. Current exact pin selection, separate manager/engine identity, cancellation, quoting, atomicity and ownership are preserved.

## Verification and review

Initial status was clean at `4e5c4583`; this configured GEN-25 workspace uses `alspaces` and branch `chore/GEN-25-remove-windows-support`. Iteration 02 started clean at `183c5fd9`. No matching roadmap specification/index entry requires lifecycle changes.

Iteration 01's integrated `make`, four-target manager/engine builds, real release packaging/12 checksums, standalone Linux identities and disposable absent/repeated asset builds passed. Those results establish the earlier release boundary, not the newly changed compatibility behavior. Its independent four-focus/two-model reviews found stale release inclusion and ZIP guidance (Sol/high behavior and architecture), and the coder found old-template refusal. All attribution and original dispositions remain in [iteration 01 review findings](iterations/01/REVIEW_SUMMARY.json); the old-template fix is intentionally superseded by the user's new compatibility boundary. The prior test-effectiveness Luna sentinel suggestion/refutation remains historical evidence.

Iteration 02 uses direct parent implementation/testing/review under the user's no-subagent instruction. Existing `TestGitCheckOwnershipPreviewUpgradeAndRemoval` and `TestGitCheckPreservesForeignEditedAndSymlinkHooks` retain current hook preview/update/removal, literal paths, idempotence, edited/foreign refusal and symlink protection. Deleted historical-template tests asserted the explicitly retired upgrade contract, so their assertions were not transferred. Windows profile overrides were unused on maintained hosts; HOME and other supported isolation variables remain.

Focused capture hook/user migration/conflict checks passed (10.494s). Exact source searches found no remaining `commandWindows`, USERPROFILE, retired signature or Windows watchdog implementation/fixture in backend/scripts. Direct diff review confirms the parser handles only the supported command field and the hook ownership check compares only the current renderer. Integrated verification of iteration 02 is pending; no new subagent sessions were created.

## Limits

Native macOS execution, live harness callbacks and hosted release publication are untested. Cross-compilation does not imply native certification. Existing dependency audit findings are outside scope. Final telemetry will refresh the existing ledger while preserving the root and 19 prior children; source-reported unknown pricing, role and open-turn/provisional facts remain authoritative, with publication calls outside the extraction cutoff.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 1h 13m (recorded duration including 127s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 185.6k tokens (41% of 450k healthy ceiling, 0 compactions) | Subagents peak 164.2k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 2.0, Tester 3.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 6.1:1 (+288 test lines / +47 prod lines)
- **Code Delivered:** +335 / -1,149 lines across 51 code files (excluding docs)
- **Reasoning Effort:** 27,966 reasoning tokens (18.2% of generated output)
- **Tool Health & Self-Correction:** 522 tool calls (20 self-corrections, 96.2% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (3 CRITICAL, 2 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (10 cycles) — Subagent roles 'coder' and 'tester' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (5 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 5 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 71,539 | 2,276 | 383,744 | 84.3% | 72.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 134,244 | 16,109 | 6,157,696 | 97.9% | 164.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 64,985 | 13,497 | 4,539,904 | 98.6% | 133.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 177,501 | 41,286 | 23,745,664 | 99.3% | 185.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 219,891 | 14,252 | 1,749,248 | 88.8% | 91.9k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 217,841 | 13,316 | 1,834,112 | 89.4% | 86.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 4 | 213,544 | 11,960 | 1,724,160 | 89.0% | 91.8k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 4 | 183,530 | 12,800 | 1,908,864 | 91.2% | 82.2k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **20** | **1,283,075** | **125,496** | **42,043,392** | **97.0%** | **185.6k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 8 | 317,650 | 18,473 | 2,412,928 | 88.4% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.1-sol** | 12 | 965,425 | 107,023 | 39,630,464 | 97.6% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **Total** | **20** | **1,283,075** | **125,496** | **42,043,392** | **97.0%** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

Capture observation finalized: 2026-11-11T19:44:50.252938638Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 20; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 20; reserved bytes: 17150085. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1cbdf-16db-77e2-8e29-6145af952823: current; last successful source read 2026-11-11T19:44:49.648373611Z; accepted source checkpoint 2026-11-11T18:30:34.712Z.
Owner codex/01a1cbda-fbe5-7ac1-825b-883fadf1a963: current; last successful source read 2026-11-11T19:44:49.664367036Z; accepted source checkpoint 2026-11-11T18:25:27.116Z.
Owner codex/01a1cbde-c6fc-71b0-9fbe-58af8b115261: current; last successful source read 2026-11-11T19:44:49.686022956Z; accepted source checkpoint 2026-11-11T18:30:45.633Z.
Owner codex/01a1cbd1-b939-7491-b7a3-c34c6ec104b9: current; last successful source read 2026-11-11T19:44:49.722038955Z; accepted source checkpoint 2026-11-11T18:27:37.954Z.
Owner codex/01a1cbda-307f-7a30-b331-c276151d5e29: current; last successful source read 2026-11-11T19:44:49.738658131Z; accepted source checkpoint 2026-11-11T18:24:40.122Z.
Owner codex/01a1cbcb-f6a3-71b3-9317-282cc89a7991: current; last successful source read 2026-11-11T19:44:49.816311237Z; accepted source checkpoint 2026-11-11T19:44:49.363Z.
Owner codex/01a1cbe0-8adf-7792-9aa7-e094a49d9d30: current; last successful source read 2026-11-11T19:44:49.831717482Z; accepted source checkpoint 2026-11-11T18:31:44.964Z.
Owner codex/01a1cbda-59d6-7061-98b9-59e99ceaea88: current; last successful source read 2026-11-11T19:44:49.850244599Z; accepted source checkpoint 2026-11-11T18:24:44.976Z.
Owner codex/01a1cbe1-7f5d-7721-a28d-abb1887ed272: current; last successful source read 2026-11-11T19:44:49.876275961Z; accepted source checkpoint 2026-11-11T18:33:41.154Z.
Owner codex/01a1cbcc-827d-75d2-afd7-b87155d38539: current; last successful source read 2026-11-11T19:44:49.89207898Z; accepted source checkpoint 2026-11-11T18:10:10.878Z.
Owner codex/01a1cbe0-c548-7782-94e2-d080ed9dc088: current; last successful source read 2026-11-11T19:44:49.907193652Z; accepted source checkpoint 2026-11-11T18:31:52.619Z.
Owner codex/01a1cbd8-33d4-7030-8e05-d1f4ba05fb4e: current; last successful source read 2026-11-11T19:44:49.934424925Z; accepted source checkpoint 2026-11-11T18:23:54.009Z.
Owner codex/01a1cbd2-2aa7-7800-82a9-4f06cb0aba42: current; last successful source read 2026-11-11T19:44:49.962383343Z; accepted source checkpoint 2026-11-11T18:28:51.095Z.
Owner codex/01a1cbd9-18c4-7ac3-bef2-94c584f37e1b: current; last successful source read 2026-11-11T19:44:49.990390648Z; accepted source checkpoint 2026-11-11T18:25:24.241Z.
Owner codex/01a1cbd7-1c32-7f90-8e6e-6e0989fb15a2: current; last successful source read 2026-11-11T19:44:50.017529419Z; accepted source checkpoint 2026-11-11T18:22:40.863Z.
Owner codex/01a1cbde-f432-7b62-ac66-67e50dfe68fb: current; last successful source read 2026-11-11T19:44:50.038141431Z; accepted source checkpoint 2026-11-11T18:30:58.718Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->
