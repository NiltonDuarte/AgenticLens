# AL-70: Explicit configuration and ticket identity

AL-70 implements the confirmed [contract](CONTRACT.md) and [inventory verdicts](LEGACY_AND_FALLBACKS.md), including the approved A12 correction in [iteration 03](iterations/03/PLAN.md). CLI/configuration selection and owning-file path origins govern sources and workspaces. Required failures are actionable; optional unavailable hydration preserves recorded history. YAML/YML discovery rejects ambiguity, and linked worktrees use base-local policy only when their own file is absent.

Literal prefix mappings, explicit case policy, overlapping filters and natural sorting preserve recorded IDs, zeros and types. Exact session selectors and explicit opt-in recovery replace heuristic selection. Git metrics measure committed merge-base net contribution through independent production/test rules. Shared configurable summary output, generic artifact items, explicit consumed settings and loopback port 8085 policy reach CLI/API/UI. Retired configuration/CLI features have no compatibility plumbing; explicit data migrations, formatting, supported source families, portable owner history, state guards and exact interrupted-publication recovery remain supported.

Missing model identity now uses an empty canonical string and displays Unknown. Observed names, including literal default and unknown, retain separate raw identities. OpenCode metadata-only extraction retains its zero-token aggregate and session counters; grouping/counting formulas and historical ledger reads remain unchanged. Codex late attribution distinguishes an empty model key from an absent turn and preserves observed names. The [operator representation](../../docs/OWNED_OCCURRENCES.md#missing-model-identity-and-aggregate-rows) documents these boundaries. No schema or migration was added.

Approved shared classification/generation choices and the five-field base-local manual policy remain intact. Capture is false, sources are unchanged and no worktree-local file exists. No installation, hook changes or automatic capture activation occurred.

## Verification and review

Hosted Backend Check at 07129d42 exposed a fixture cleanup race (run 42988883853). CLI tester reproduced it under `go test -race -count=20`, traced late detached reconcile workers and added the existing lease/completion helper to one test. All functional assertions are unchanged. The identical 20-run race check passed (15.093s), as did neighboring advisory/completion controls (6.485s); both gpt-6.1-sol/high and gpt-6-luna/medium test reviewers approved. The CLI tester retains discovery credit in the compact findings. Production code and helper implementations are unchanged. Hosted checks for this correction remain pending.

The checkout began clean at eb7d169c; requested initial main integration reached f63e3db6. Iteration02 candidate 69e17b44 integrated main 3384fa48 without conflicts, and full make passed at 604d6338 (462.902s). Its two stale CLI empty-result expectations were corrected before that green rerun. A12 began clean at 54d58be4. Reviewed implementation f6a2f543 and test correction 07d07a9d remain integrated with current origin/main 3384fa48; follow-up fetch/merge checks, including the CI correction, were already up to date.

Required **make passed at `0311da0d4419241c22307eb605cd935b7049b19b`**, exit 0 in 393.936s. It exercised Go units/race integration, 227 frontend units, 751 frontend integrations, static checks, 14 release-workflow, 10 installation, 23 distribution and 6 orchestration cases, plus five-platform native builds. Both host manager/engine identities were executed separately. Standalone embedded UI/API served 12 synthetic v6 ledgers without sources/Git, preserved their bytes and checked explicit/unused configuration and listener behavior. The 07d07a9d gate also passed (426.197s). Current raw gate: /tmp/al70-ci-correction-make-result.json; unchanged-source standalone evidence: /tmp/al70-a12-standalone-1333hom5/RESULT.json.

Consolidated review is **APPROVED**. Iteration02's 23 findings and every discovery tuple remain in its [compact findings](iterations/02/REVIEW_SUMMARY.json). A12's eight independent behavior/data/architecture/test passes used gpt-6.1-sol/high and gpt-6-luna/medium and initially reported no findings. Integrated make then exposed one stale backfill model expectation, discovered by the coordinator using gpt-6.1-sol/high. Tester preserved all row/cost/replacement/history assertions; focused cases, `go test -race -count=1 ./internal/adapters/backfill` and both test-review profile rechecks passed. The [iteration03 findings](iterations/03/REVIEW_SUMMARY.json) retain discovery credit for the stale expectation and CI worker-lifecycle correction. Source-observed native metadata verifies all reviewer model/effort profiles and parent links; verification adds no discovery credit. Unchanged scopes carry forward because the corrections change only test expectations and fixture lifecycle; production code and frontend inputs are identical. Sol test review also demonstrated a passing control and three intended mutation failures.

## Limits and final source record

Windows evidence remains scoped native fixtures/WSL transport, not a Windows full gate or actual harness trust/callback activation. Ruby is absent, so the earlier prose-only Pages check was inspected locally; the hosted build passed at the previous PR head. An earlier test isolation failure created two fixture-owned cache entries; only those proven entries were removed, absence verified and contained runs passed. Failed runs are not counted as success. Developer stores were not test fixtures.

Final Codex refresh verified all 51 sessions, preserving all 51 prior IDs: the runtime root and 50 native children, including 44 review sessions with matching persona/model/effort profiles. CLI tester identity was also verified from native metadata. Source costs remain unavailable. The root is still open; CLI checkpoint feedback reports 972 provisional seconds and an unverified invoking caller. No source-session termination or caller rollover is inferred. Verification: /tmp/al70-ci-final-telemetry-verification.json.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 90
- **Active Agent Duration:** 17h 32m (recorded duration including 972s provisional)
- **Context Window Health:** BLOATED — Root agent peak 450.9k tokens (100% of 450k healthy ceiling, 6 compactions) | Subagents peak 453.8k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 7 turns) | Planner 5.0, Coder 5.0, Tester 4.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.1, Reviewer-tests 1.4
- **Test-to-Production Code Ratio:** 1.9:1 (+7,958 test lines / +4,259 prod lines)
- **Code Delivered:** +12,217 / -15,268 lines across 245 code files (excluding docs)
- **Reasoning Effort:** 943,125 reasoning tokens (33.0% of generated output)
- **Tool Health & Self-Correction:** 3,976 tool calls (237 self-corrections, 94.0% success rate)

### Anomalies & Health Flags
- **Status:** 27 execution anomalies detected (15 CRITICAL, 12 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.9k tokens) — Context tokens reached 450.9k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (453.8k tokens) — Context tokens reached 453.8k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (6 cycles) — Subagent roles 'planner' and 'coder' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (161 cycles) — Subagent roles 'tester' and 'coder' alternated for 161 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (6 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (17 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (11 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (8 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (12 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 12 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (14 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 14 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (7 cycles) — Subagent roles 'planner' and 'coder' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (8 cycles) — Subagent roles 'tester' and 'coder' alternated for 8 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.5k tokens) — Context tokens reached 250.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.6k tokens) — Context tokens reached 251.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (255.7k tokens) — Context tokens reached 255.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.0k tokens) — Context tokens reached 251.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.1k tokens) — Context tokens reached 251.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.2k tokens) — Context tokens reached 250.2k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-behavior (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-behavior' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> tester (4 cycles) — Subagent roles 'reviewer-tests' and 'tester' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 516,318 | 31,145 | 8,217,088 | 94.1% | 222.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 3 | 3,579,264 | 548,263 | 204,271,744 | 98.3% | 453.8k | BLOATED (101%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 2 | 2,006,018 | 290,249 | 141,394,176 | 98.6% | 450.2k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 1 | 2,779,497 | 733,215 | 256,574,976 | 98.9% | 450.9k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 12 | 1,221,111 | 84,420 | 20,194,176 | 94.3% | 216.6k | HEALTHY | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **reviewer-behavior** | 12 | 1,248,442 | 82,156 | 19,160,832 | 93.9% | 251.6k | ELEVATED | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **reviewer-data** | 10 | 841,025 | 57,225 | 11,411,328 | 93.1% | 159.2k | HEALTHY | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **reviewer-tests** | 10 | 1,112,767 | 91,441 | 24,297,600 | 95.6% | 234.4k | ELEVATED | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **Total** | **51** | **13,304,442** | **1,918,114** | **685,521,920** | **98.1%** | **453.8k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 51 unavailable / 51 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 22 | 1,793,289 | 115,475 | 21,744,768 | 92.4% | Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations) |
| **gpt-6.1-sol** | 29 | 11,511,153 | 1,802,639 | 663,777,152 | 98.3% | Unavailable (0 observed, 0 partial, 29 unavailable / 29 invocations) |
| **Total** | **51** | **13,304,442** | **1,918,114** | **685,521,920** | **98.1%** | **Unavailable (0 observed, 0 partial, 51 unavailable / 51 invocations)** |

Capture observation finalized: 2026-11-11T09:09:54.694309204Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 40; retained older facts: 1; pending: 11; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 51; reserved bytes: 66913395. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1c897-7e94-7753-acd3-f56829bb369c: current; last successful source read 2026-11-11T09:09:52.788358098Z; accepted source checkpoint 2026-11-11T03:14:33.244Z.
Owner codex/01a1c80c-9492-7c40-966a-c9565829476d: current; last successful source read 2026-11-11T09:09:52.901349785Z; accepted source checkpoint 2026-11-11T08:02:48.333Z.
Owner codex/01a1c9a6-eab6-7fb1-8ec6-b281a23e3fc4: current; last successful source read 2026-11-11T09:09:52.927620597Z; accepted source checkpoint 2026-11-11T08:10:43.755Z.
Owner codex/01a1c8d4-7887-7a23-a0a9-7f3c10a334b9: current; last successful source read 2026-11-11T09:09:52.957494782Z; accepted source checkpoint 2026-11-11T04:22:33.639Z.
Owner codex/01a1c9a6-af3d-7220-863a-f6b163fca4d7: current; last successful source read 2026-11-11T09:09:52.984852167Z; accepted source checkpoint 2026-11-11T08:10:37.179Z.
Owner codex/01a1c842-0c0d-7021-8833-ab3fb04e518f: current; last successful source read 2026-11-11T09:09:53.101524962Z; accepted source checkpoint 2026-11-11T08:59:07.751Z.
Owner codex/01a1c8b9-2501-7110-888a-b471fc0f8702: current; last successful source read 2026-11-11T09:09:53.135052663Z; accepted source checkpoint 2026-11-11T03:53:47.414Z.
Owner codex/01a1c9ab-3a42-7103-a72c-4dc93f40534c: current; last successful source read 2026-11-11T09:09:53.15606383Z; accepted source checkpoint 2026-11-11T08:15:06.536Z.
Owner codex/01a1c8e3-be5f-7fd1-a129-c21139bddfb9: current; last successful source read 2026-11-11T09:09:53.177347278Z; accepted source checkpoint 2026-11-11T04:37:20.555Z.
Owner codex/01a1c819-3b83-73a3-a41a-7d6ed666cecf: current; last successful source read 2026-11-11T09:09:53.373042385Z; accepted source checkpoint 2026-11-11T08:29:06.8Z.
Owner codex/01a1c89d-5cf0-7472-8c14-8bbe35b226d4: current; last successful source read 2026-11-11T09:09:53.39805691Z; accepted source checkpoint 2026-11-11T03:20:50.408Z.
Owner codex/01a1c8d1-7236-7711-964d-7f8865f3778b: current; last successful source read 2026-11-11T09:09:53.4226318Z; accepted source checkpoint 2026-11-11T04:16:57.31Z.
Owner codex/01a1c872-515d-7a80-973b-5c5698949af2: current; last successful source read 2026-11-11T09:09:53.475469913Z; accepted source checkpoint 2026-11-11T02:37:55.945Z.
Owner codex/01a1c80c-5c06-7ef0-b1f1-8139fd6f5fc9: current; last successful source read 2026-11-11T09:09:53.616991504Z; accepted source checkpoint 2026-11-11T08:05:28.062Z.
Owner codex/01a1c8b9-4c03-7eb3-aef7-7908151ced86: current; last successful source read 2026-11-11T09:09:53.651714687Z; accepted source checkpoint 2026-11-11T03:56:30.433Z.
Owner codex/01a1c8d6-914e-7171-b8d7-34e18a4fe074: current; last successful source read 2026-11-11T09:09:53.677989402Z; accepted source checkpoint 2026-11-11T04:22:41.071Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->
