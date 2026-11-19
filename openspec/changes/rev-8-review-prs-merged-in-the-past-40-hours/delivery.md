# fix(REV-8): preserve ledger selection and recover stalled live streams

Generation, backfill, and migration now hold their physical workspace selection through publication admission. A directory replacement returns a retry diagnostic before recovery or publication. Automatic capture resolves the workspace in the process doing the write, and maintenance continues with independently valid batch targets.

SSE overflow closes the affected stream so its existing reconnect path receives a fresh snapshot. A blocked write or flush has the server's 15-second response budget. Idle waits remain open. The shared repository configuration selects the checksum-verified published engine `v3.0.0-rc.156+9308ac7` for clean-checkout dispatch.

These changes address the four items in the original [seven-PR audit handoff](evidence/iterations/01/REVIEW.md). The audit retains the other existing observations. [Regression evidence](evidence/iterations/02/TESTS.md) · [Review findings and resolution](evidence/iterations/02/REVIEW_SUMMARY.json).

Four independent review focuses approved the corrected production revision `8de66f95` under the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. The Sol architecture and test reviewers found the blocked-socket gap in the first candidate and verified the correction. Both data reviews carry forward from `35013bda` because the later changes affect stream transport and test cleanup. The final fixture-only change at `e39c1071` preserves the test assertions and was verified with the full HTTP race suite and repeated focused runs.

Final source-backed generation contains all 21 registered Codex sessions, including the audit root and all 13 previously accepted sessions. It refreshed 15 sessions and retained six earlier snapshots because their source rollouts were unavailable. Costs were unavailable. The pinned release was installed and dispatched on this machine. Its normal CLI works with the local manager, and the machine's guarded capture worker still requires a newer worker protocol. The ledger was refreshed directly with the source-built candidate.

## Pull request

### What changed

- Carry the selected directory into write admission for generation, capture, backfill, and migration. Preserve valid sibling work when a batch target moves.
- Close saturated SSE streams and bound socket writes and flushes. Reconnection reads a fresh snapshot even when recovery data is unchanged.
- Select the exact published engine release in a clean checkout. Direct source builds can exercise the application fixes before the next release.

### Verification

- The integrated `make -j 3` gate passed at `8de66f95`, with Vitest concurrency limited to two workers. It ran the Go race suites, 1,050 frontend tests, static and distribution checks, and built all four native platform targets. The complete HTTP race suite passed again at `e39c1071`.
- Deterministic ledger replacement and real TCP regressions reproduced the old failures. Focused Go race checks verified directory movement, portable ledgers, native capture and interrupted publication recovery.
- The published Linux manager and engine matched their SHA-256 records. Isolated installation and offline dispatch reported the exact release and listed a fixture ledger.
- A source-built Linux engine served the embedded dashboard, workspace API and unavailable-live-source diagnostic. Native macOS execution and browser-native EventSource execution were outside the verification scope.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 28
- **Active Agent Duration:** 2h 9m (recorded duration including 1673s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 315.3k tokens (70% of 450k healthy ceiling, 0 compactions) | Subagents peak 333.3k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.4 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 1.5, Tester 3.0, Reviewer 1.0, Reviewer-architecture 1.5, Reviewer-behavior 1.5, Reviewer-data 1.0, Reviewer-tests 1.5
- **Test-to-Production Code Ratio:** 1.9:1 (+639 test lines / +337 prod lines)
- **Code Delivered:** +976 / -145 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 101,177 reasoning tokens (28.3% of generated output)
- **Tool Health & Self-Correction:** 850 tool calls (56 self-corrections, 93.4% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (1 CRITICAL, 6 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (17 cycles) — Subagent roles 'coder' and 'tester' alternated for 17 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.8k tokens) — Context tokens reached 253.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.2k tokens) — Context tokens reached 251.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.1k tokens) — Context tokens reached 251.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.8k tokens) — Context tokens reached 251.8k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 228,744 | 3,032 | 1,543,680 | 87.1% | 227.3k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 143,076 | 27,943 | 17,615,616 | 99.2% | 282.1k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 148,067 | 28,206 | 23,617,792 | 99.4% | 333.3k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 2 | 159,749 | 9,888 | 1,584,896 | 90.8% | 100.3k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 7 | 1,186,498 | 118,246 | 64,686,592 | 98.2% | 315.3k | ELEVATED | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-architecture** | 2 | 247,805 | 18,942 | 5,513,856 | 95.7% | 166.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 417,013 | 20,440 | 7,873,792 | 95.0% | 188.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 232,325 | 16,233 | 4,461,312 | 95.1% | 161.3k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 191,957 | 13,494 | 4,059,392 | 95.5% | 125.1k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **21** | **2,955,234** | **256,424** | **130,956,928** | **97.8%** | **333.3k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 6 | 566,267 | 34,423 | 5,550,464 | 90.7% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **gpt-6.1-sol** | 10 | 1,248,849 | 78,997 | 23,493,248 | 95.0% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.2-sol** | 5 | 1,140,118 | 143,004 | 101,913,216 | 98.9% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **21** | **2,955,234** | **256,424** | **130,956,928** | **97.8%** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

Capture observation finalized: 2026-11-19T00:38:21.109841185Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 15; retained older facts: 6; pending: 0; failed: 6; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 21; reserved bytes: 1618221505. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1f118-cfc2-7100-853d-1b8c77d29b1e: current; last successful source read 2026-11-19T00:38:13.379818751Z; accepted source checkpoint 2026-11-19T00:00:19.824Z.
Owner codex/01a1f125-c610-7792-a527-0a785c95960e: current; last successful source read 2026-11-19T00:38:13.965068815Z; accepted source checkpoint 2026-11-19T00:21:33.109Z.
Owner codex/01a1f10f-f74f-7500-8b9d-4d974775424f: current; last successful source read 2026-11-19T00:38:14.646523277Z; accepted source checkpoint 2026-11-18T23:53:43.821Z.
Owner codex/01a1f10f-c913-7240-b553-9940a4958f74: current; last successful source read 2026-11-19T00:38:15.26826785Z; accepted source checkpoint 2026-11-18T23:53:51.09Z.
Owner codex/01a1f116-b916-73e3-b3bc-8ef87f8ada87: failed; last successful source read 2026-11-19T00:11:08.045191864Z; accepted source checkpoint 2026-11-18T23:57:46.849Z.
Owner codex/01a1f10f-a932-7a33-9729-506111a6a1b5: current; last successful source read 2026-11-19T00:38:15.816537533Z; accepted source checkpoint 2026-11-18T23:54:33.055Z.
Owner codex/01a1f131-1a82-79c1-b9fe-fdd199e971ed: current; last successful source read 2026-11-19T00:38:16.296544711Z; accepted source checkpoint 2026-11-19T00:29:58.758Z.
Owner codex/01a1f125-e44e-7b23-9993-f90b730316e2: current; last successful source read 2026-11-19T00:38:16.86499262Z; accepted source checkpoint 2026-11-19T00:27:19.254Z.
Owner codex/01a1f12e-0394-7d32-8ee6-216c8dee8e63: current; last successful source read 2026-11-19T00:38:17.447839236Z; accepted source checkpoint 2026-11-19T00:31:17.907Z.
Owner codex/01a1f11a-c14c-7412-bf59-e8745b3c9385: failed; last successful source read 2026-11-19T00:11:09.606039432Z; accepted source checkpoint 2026-11-19T00:02:17.694Z.
Owner codex/01a1f115-131d-7101-baff-d924907769b0: failed; last successful source read 2026-11-19T00:11:10.501769499Z; accepted source checkpoint 2026-11-18T23:55:47.264Z.
Owner codex/01a1f119-3169-74d0-92c0-2d7068d2476c: failed; last successful source read 2026-11-19T00:11:12.120810165Z; accepted source checkpoint 2026-11-19T00:00:14.536Z.
Owner codex/01a1f10e-6c1a-73c3-903b-f4d7fab3c9dd: current; last successful source read 2026-11-19T00:38:18.105701578Z; accepted source checkpoint 2026-11-19T00:38:14.947Z.
Owner codex/01a1f126-011f-7802-bb6e-3ee9b1fcf6c4: current; last successful source read 2026-11-19T00:38:18.640406521Z; accepted source checkpoint 2026-11-19T00:33:16.725Z.
Owner codex/01a1f115-982d-7203-b081-296e401f92ab: failed; last successful source read 2026-11-19T00:11:14.343351899Z; accepted source checkpoint 2026-11-18T23:58:29.618Z.
Owner codex/01a1f118-839a-7381-b0ed-d83cac4de16f: failed; last successful source read 2026-11-19T00:11:15.651005997Z; accepted source checkpoint 2026-11-19T00:00:33.767Z.
Additional owner freshness details remain in local capture diagnostics.

Retained recorded history: 6 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
