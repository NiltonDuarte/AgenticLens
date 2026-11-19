# fix(REV-8): preserve ledger selection and recover stalled live streams

Generation, backfill, and migration now hold their physical workspace selection through publication admission. A directory replacement returns a retry diagnostic before recovery or publication. Automatic capture resolves the workspace in the process doing the write, and maintenance continues with independently valid batch targets.

SSE overflow closes the affected stream so its existing reconnect path receives a fresh snapshot. A blocked write or flush has the server's 15-second response budget. Idle waits remain open. The shared repository configuration selects the checksum-verified published engine `v3.0.0-rc.156+9308ac7` for clean-checkout dispatch.

These changes address the four items in the original [seven-PR audit handoff](evidence/iterations/01/REVIEW.md). The audit retains the other existing observations. [Regression evidence](evidence/iterations/02/TESTS.md) · [Review findings and resolution](evidence/iterations/02/REVIEW_SUMMARY.json).

Four independent review focuses approved the corrected production revision `8de66f95` under the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. The Sol architecture and test reviewers found the blocked-socket gap in the first candidate and verified the correction. Both data reviews carry forward from `35013bda` because the later changes affect stream transport and test cleanup. The final fixture-only change at `e39c1071` preserves the test assertions and was verified with the full HTTP race suite and repeated focused runs.

Iteration 02 source-backed generation contained all 21 registered Codex sessions, including the audit root and all 13 previously accepted sessions. It refreshed 15 sessions and retained six earlier snapshots because their source rollouts were unavailable. Costs were unavailable. The pinned release was installed and dispatched on this machine. Its normal CLI works with the local manager, and the machine's guarded capture worker still requires a newer worker protocol. The ledger was refreshed directly with the source-built candidate.

### Archived-rollout recovery

Capture reconciliation now repairs a missing rollout path for an already-bound Codex session when exact discovery finds one authorized source in the same store. It saves the repaired path with the accepted snapshot. An absent, ambiguous, invalid or superseded read preserves accepted facts and keeps unaffected owners eligible for refresh. The change covers the worker, `capture reconcile`, and `generate --capture-auto`.

The [iteration 03 test evidence](evidence/iterations/03/TESTS.md) covers both CLI paths and the registry interleavings. [Review dispositions](evidence/iterations/03/REVIEW_SUMMARY.json) record the ancestry, stale publication and healthy retry corrections. The four reviews examined behavior, data integrity, architecture and tests under the configured `gpt-6.1-sol/high` and `gpt-6-luna/medium` profiles. The reported discovery-session budget defect was refuted against the baseline and CLI configuration path. The six previously identified real REV-8 archives were kept outside the regression fixtures.

Iteration 03 telemetry initially used the exact Codex root family. A later authorized live run of `generate --capture-auto` with the source-built engine refreshed all 21 registered owners, including the six archived sessions. Their registry refresh timestamps advanced and their recorded metrics were unchanged. The ledger preserved all 28 previously recorded sessions. The background worker still reports degraded status on this machine.

The final coverage check matched eight additional delegated reviews to their exact archived Codex sources and added them to the ledger. It contains 36 sessions and preserves all 28 previous session identities. The capture-aware pass refreshed all 21 registered owners with no pending or failed reads. Cost evidence remains unavailable.

The AgenticLens skills now ask the launcher to track expected runs and pass the ticket explicitly. Each separate session answers its own binding request. The launcher verifies the accepted registration and ledger entry independently. The review skill reuses its existing run roster, including retries. [Shared procedure](../../../.agents/skills/agentic-lens-operator/references/cli.md#independent-session-binding-and-ledger-coverage).

The main integration refresh preserves all 36 earlier sessions and records all four new merge sessions. Native Codex metadata links three to the root. The separate review was recovered by exact native ID. Its automatic binding was unverified in this worktree.

## Pull request

Supersedes #152 after renaming the branch to use the REV-8 ticket ID. The earlier PR retains its review history.

### What changed

- Carry the selected directory into write admission for generation, capture, backfill, and migration. Preserve valid sibling work when a batch target moves.
- Close saturated SSE streams and bound socket writes and flushes. Reconnection reads a fresh snapshot even when recovery data is unchanged.
- Select the exact published engine release in a clean checkout. Direct source builds can exercise the application fixes before the next release.
- Recover already-bound Codex sessions when their rollout moves to the archive. Verify the native owner, save the path with accepted facts, and exclude superseded source evidence and dependent links from publication.
- Teach the AgenticLens skills to track independent same-ticket runs, accept each session's own binding, and verify both registry acceptance and ledger coverage.
- Integrate `main` at `ccd48107`. Protected Codex publication checks both the admitted private state and the retained workspace directory. Inbox binding and publication scheduling commit together.

### Verification
- The integrated `VITEST_MAX_WORKERS=2 make -j 3` gate passed at `5f16d31c` after merging `main`. It ran the Go race suites, 1,050 frontend tests, installer and distribution checks, and builds for four platforms. The protected publication regression rejected a same-ticket directory replacement, preserved both ledgers, and admitted a fresh selection with protected state.
- Focused integration reviews on Sol 6.2/high and Luna 6/medium approved merge commit `c6f2ca11`. The review covered protected publication, atomic binding and scheduling, and compatibility with archived refresh and ticket metadata.
- The integrated `VITEST_MAX_WORKERS=2 make -j 3` gate passed at `060847c8` on the renamed branch. It ran the Go race suites, 1,050 frontend tests, installer and distribution checks, and builds for four platforms.
- The iteration 03 integrated `make -j 3` passed at `c1b87c6e`, including the Go race suites, 1,050 frontend tests and four platform builds. The native archived-rollout and SQLite supersession regressions pass.

- The integrated `make -j 3` gate passed at `8de66f95`, with Vitest concurrency limited to two workers. It ran the Go race suites, 1,050 frontend tests, static and distribution checks, and built all four native platform targets. The complete HTTP race suite passed again at `e39c1071`.
- Deterministic ledger replacement and real TCP regressions reproduced the old failures. Focused Go race checks verified directory movement, portable ledgers, native capture and interrupted publication recovery.
- The published Linux manager and engine matched their SHA-256 records. Isolated installation and offline dispatch reported the exact release and listed a fixture ledger.
- A source-built Linux engine served the embedded dashboard, workspace API and unavailable-live-source diagnostic. Native macOS execution and browser-native EventSource execution were outside the verification scope.
- Skill validation, cross-reference checks and an isolated setup bundle install passed. A read-only REV-8 check confirmed that the manually recovered reviews appear in the ledger while their explicit bindings remain unverified in this worktree.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 69
- **Active Agent Duration:** 4h 41m (recorded duration including 796s provisional)
- **Context Window Health:** BLOATED — Root agent peak 451.1k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 446.2k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.7 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 1.8, Tester 2.3, Reviewer 1.0, Reviewer-architecture 2.0, Reviewer-behavior 1.7, Reviewer-data 1.7, Reviewer-tests 1.7
- **Test-to-Production Code Ratio:** 1.5:1 (+1,337 test lines / +917 prod lines)
- **Code Delivered:** +2,254 / -312 lines across 29 code files (excluding docs)
- **Reasoning Effort:** 198,823 reasoning tokens (28.7% of generated output)
- **Tool Health & Self-Correction:** 1,582 tool calls (110 self-corrections, 93.0% success rate)

### Anomalies & Health Flags
- **Status:** 16 execution anomalies detected (6 CRITICAL, 10 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.7k tokens) — Context tokens reached 450.7k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (17 cycles) — Subagent roles 'coder' and 'tester' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
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
| **coder** | 4 | 348,429 | 80,840 | 51,813,504 | 99.3% | 445.9k | ELEVATED | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **tester** | 3 | 304,345 | 62,665 | 63,444,864 | 99.5% | 446.2k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer** | 3 | 403,848 | 12,923 | 3,212,800 | 88.8% | 250.7k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 16 | 2,763,161 | 232,790 | 179,021,696 | 98.5% | 451.1k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 16 unavailable / 16 invocations) |
| **reviewer-architecture** | 3 | 356,044 | 26,474 | 7,977,472 | 95.7% | 166.2k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-behavior** | 3 | 592,278 | 27,172 | 10,101,248 | 94.5% | 188.8k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-data** | 3 | 341,300 | 29,707 | 7,744,128 | 95.8% | 161.3k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 3 | 284,190 | 17,686 | 5,134,848 | 94.8% | 125.1k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Total** | **40** | **5,765,459** | **494,768** | **331,905,024** | **98.3%** | **451.1k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 40 unavailable / 40 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 15 | 1,280,626 | 58,790 | 9,102,336 | 87.7% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **gpt-6.1-sol** | 14 | 1,733,561 | 110,927 | 32,542,592 | 94.9% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **gpt-6.2-sol** | 11 | 2,751,272 | 325,051 | 290,260,096 | 99.1% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **Total** | **40** | **5,765,459** | **494,768** | **331,905,024** | **98.3%** | **Unavailable (0 observed, 0 partial, 40 unavailable / 40 invocations)** |

Retained recorded history: 15 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
