# fix(REV-8): preserve ledger selection and recover stalled live streams

Generation, backfill, and migration now hold their physical workspace selection through publication admission. A directory replacement returns a retry diagnostic before recovery or publication. Automatic capture resolves the workspace in the process doing the write, and maintenance continues with independently valid batch targets.

SSE overflow closes the affected stream so its existing reconnect path receives a fresh snapshot. A blocked write or flush has the server's 15-second response budget. Idle waits remain open. The shared repository configuration selects the checksum-verified published engine `v3.0.0-rc.156+9308ac7` for clean-checkout dispatch.

These changes address the four items in the original [seven-PR audit handoff](evidence/iterations/01/REVIEW.md). The audit retains the other existing observations. [Regression evidence](evidence/iterations/02/TESTS.md) · [Review findings and resolution](evidence/iterations/02/REVIEW_SUMMARY.json).

Four independent review focuses approved the corrected production revision `8de66f95` under the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. The Sol architecture and test reviewers found the blocked-socket gap in the first candidate and verified the correction. Both data reviews carry forward from `35013bda` because the later changes affect stream transport and test cleanup. The final fixture-only change at `e39c1071` preserves the test assertions and was verified with the full HTTP race suite and repeated focused runs.

Iteration 02 source-backed generation contained all 21 registered Codex sessions, including the audit root and all 13 previously accepted sessions. It refreshed 15 sessions and retained six earlier snapshots because their source rollouts were unavailable. Costs were unavailable. The pinned release was installed and dispatched on this machine. Its normal CLI works with the local manager, and the machine's guarded capture worker still requires a newer worker protocol. The ledger was refreshed directly with the source-built candidate.

### Archived-rollout recovery

Capture reconciliation now repairs a missing rollout path for an already-bound Codex session when exact discovery finds one authorized source in the same store. It saves the repaired path with the accepted snapshot. An absent, ambiguous, invalid or superseded read preserves accepted facts and keeps unaffected owners eligible for refresh. The change covers the worker, `capture reconcile`, and `generate --capture-auto`.

The [iteration 03 test evidence](evidence/iterations/03/TESTS.md) covers both CLI paths and the registry interleavings. [Review dispositions](evidence/iterations/03/REVIEW_SUMMARY.json) record the ancestry, stale publication and healthy retry corrections. The four reviews examined behavior, data integrity, architecture and tests under the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. The reported discovery-session budget defect was refuted against the baseline and CLI configuration path. The six previously identified real REV-8 archives were kept outside the regression fixtures.

Iteration 03 telemetry was generated directly from the exact Codex root family using the source-built engine. The ledger contains 28 sessions and retains all 21 previously published owners. The six archived snapshots remain unchanged. A live automatic refresh of those six records remains unverified.

## Pull request

### What changed

- Carry the selected directory into write admission for generation, capture, backfill, and migration. Preserve valid sibling work when a batch target moves.
- Close saturated SSE streams and bound socket writes and flushes. Reconnection reads a fresh snapshot even when recovery data is unchanged.
- Select the exact published engine release in a clean checkout. Direct source builds can exercise the application fixes before the next release.
- Recover already-bound Codex sessions when their rollout moves to the archive. Verify the native owner, save the path with accepted facts, and exclude superseded source evidence and dependent links from publication.

### Verification
- The iteration 03 integrated `make -j 3` passed at `c1b87c6e`, including the Go race suites, 1,050 frontend tests and four platform builds. The native archived-rollout and SQLite supersession regressions pass.

- The integrated `make -j 3` gate passed at `8de66f95`, with Vitest concurrency limited to two workers. It ran the Go race suites, 1,050 frontend tests, static and distribution checks, and built all four native platform targets. The complete HTTP race suite passed again at `e39c1071`.
- Deterministic ledger replacement and real TCP regressions reproduced the old failures. Focused Go race checks verified directory movement, portable ledgers, native capture and interrupted publication recovery.
- The published Linux manager and engine matched their SHA-256 records. Isolated installation and offline dispatch reported the exact release and listed a fixture ledger.
- A source-built Linux engine served the embedded dashboard, workspace API and unavailable-live-source diagnostic. Native macOS execution and browser-native EventSource execution were outside the verification scope.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 42
- **Active Agent Duration:** 3h 40m (recorded duration including 1917s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 427.9k tokens (95% of 450k healthy ceiling, 0 compactions) | Subagents peak 446.2k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.8 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 2.0, Tester 3.0, Reviewer 1.0, Reviewer-architecture 2.0, Reviewer-behavior 1.7, Reviewer-data 1.7, Reviewer-tests 1.7
- **Test-to-Production Code Ratio:** 1.4:1 (+1,240 test lines / +909 prod lines)
- **Code Delivered:** +2,149 / -295 lines across 27 code files (excluding docs)
- **Reasoning Effort:** 166,062 reasoning tokens (29.3% of generated output)
- **Tool Health & Self-Correction:** 1,242 tool calls (81 self-corrections, 93.5% success rate)

### Anomalies & Health Flags
- **Status:** 14 execution anomalies detected (4 CRITICAL, 10 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (17 cycles) — Subagent roles 'coder' and 'tester' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.8k tokens) — Context tokens reached 253.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.2k tokens) — Context tokens reached 251.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (333.2k tokens) — Context tokens reached 333.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (345.3k tokens) — Context tokens reached 345.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (345.5k tokens) — Context tokens reached 345.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.1k tokens) — Context tokens reached 251.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.8k tokens) — Context tokens reached 251.8k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 371,864 | 4,511 | 3,454,464 | 90.3% | 346.6k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 3 | 278,716 | 72,495 | 46,573,568 | 99.4% | 445.9k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 2 | 274,610 | 58,712 | 60,730,112 | 99.5% | 446.2k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer** | 2 | 159,749 | 9,888 | 1,584,896 | 90.8% | 100.3k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 7 | 1,339,073 | 153,996 | 132,581,888 | 99.0% | 427.9k | ELEVATED | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-architecture** | 3 | 356,044 | 26,474 | 7,977,472 | 95.7% | 166.2k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-behavior** | 3 | 592,278 | 27,172 | 10,101,248 | 94.5% | 188.8k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-data** | 3 | 341,300 | 29,707 | 7,744,128 | 95.8% | 161.3k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 3 | 284,190 | 17,686 | 5,134,848 | 94.8% | 125.1k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Total** | **28** | **3,997,824** | **400,641** | **275,882,624** | **98.6%** | **446.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 6 | 566,267 | 34,423 | 5,550,464 | 90.7% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **gpt-6.1-sol** | 14 | 1,733,561 | 110,927 | 32,542,592 | 94.9% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **gpt-6.2-sol** | 8 | 1,697,996 | 255,291 | 237,789,568 | 99.3% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **28** | **3,997,824** | **400,641** | **275,882,624** | **98.6%** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

Retained recorded history: 6 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
