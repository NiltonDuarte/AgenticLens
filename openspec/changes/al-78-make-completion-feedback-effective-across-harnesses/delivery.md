# fix(AL-78): Guide metrics refresh after native completion

The manual checkpoint points to an automatic capture refresh after natural completion and source availability. When automatic capture is unavailable, it points to explicit generation from outside the completed turn. Caller attribution still requires independent native evidence. The checkpoint threshold, duration accounting, and ledger format are unchanged.

## Verification

Candidate `2ee112cf` includes `origin/main` at `db4477fa`. The final fixture teardown changes are covered by 15 focused Python tests. The native completion evidence predates only those teardown edits. The host build and strict OpenSpec validation passed.

- The targeted Go race tests passed for checkpoint output, threshold and error handling, and retained provisional facts on Codex and both supported OpenCode extraction families. Existing cross-source refresh tests also passed for active, completed, and growing sources. Independent regression checks confirmed exact-owner replacement, idempotence, preservation of copied unavailable owners and other-harness history, and unknown cost. Timed automatic fixture cases that failed under concurrent load passed on a targeted retry.
- The packaged baseline and candidate both passed the real commit, in-turn feedback delivery, native completion, automatic publication, and external manual fallback scenario. The [native evidence](evidence/native-feedback.md) records source and API timings, exact owner and interval, final usage, and the scripted provider limitation.
- The initial `make -j2` run had four frontend timeouts. The final standalone frontend integration run passed all 819 tests. A later full run reached the Go CLI package's default 10-minute limit. The failing discard and interrupted-publication cases passed on focused retry.
- `TestCaptureFirstPublicationRefusesMissingUnsafeAndAmbiguousSelection/missing` failed intermittently on both untouched `db4477fa` and candidate `2ee112cf`. Each revision failed three of five matching repetitions, reporting acknowledgement with the worker lifetime exhausted. This capture-auto path does not render AL-78 checkpoint feedback.
- The integrated `make` run on `2ee112cf` used four Go scheduler threads and a 25-minute package timeout. It completed the Go suite and failed on `TestCaptureOpenCodeIdleRestoreContinuesBoundedDescendantDiscovery`, which observed 9 of 11 expected owners. The exact case passed both isolated repetitions on both revisions. The integrated gate remains incomplete.
- Independent review approved the scoped renderer and fixture corrections. The [review record](evidence/REVIEW.md) attributes findings and records the status of integrated verification.

## Limits

The separate default release gate passed on Codex. Its OpenCode discard and restore path did not finish cleanly. The first candidate run observed a partially reconciled restore before a complete pass. The isolated retry encountered a status command failure that did not recur on the retained state. The baseline default run stopped during cold bootstrap, so there is no same-failure baseline reproduction. This path uses automatic capture with manual feedback disabled and is outside the AL-78 completion probe.

The product changes and verification are local. The integrated verification blocker keeps the ticket in progress and the PR unpublished.

The evidence checkpoint includes 22 recorded Codex owners. The extraction verified this root and its four native descendants and preserved all 21 previously recorded owners. Two separately run reviews lacked verified bindings and remain absent. All invocation costs are unavailable. The current turn was still open at extraction and contributes provisional duration.

## Pull request

- Clarify when checkpoint metrics are provisional and how to refresh them without a metrics-only model turn.
- Align setup, operator, and contributor guidance with the native capture lifecycle and external manual recovery.
- Exercise the checkpoint and the delay from a real code commit to a settled consumer read on Codex and OpenCode.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 61
- **Active Agent Duration:** 2h 42m (recorded duration including 3755s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 321.7k tokens (72% of 450k healthy ceiling, 0 compactions) | Subagents peak 283.3k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 1.0
- **Test-to-Production Code Ratio:** 3.3:1 (+23 test lines / +7 prod lines)
- **Code Delivered:** +30 / -14 lines across 4 code files (excluding docs)
- **Reasoning Effort:** 74,024 reasoning tokens (25.1% of generated output)
- **Tool Health & Self-Correction:** 781 tool calls (48 self-corrections, 93.9% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (1 CRITICAL, 2 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (14 cycles) — Subagent roles 'tester' and 'coder' alternated for 14 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.5k tokens) — Context tokens reached 254.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (284.6k tokens) — Context tokens reached 284.6k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 145,804 | 3,066 | 1,340,032 | 90.2% | 158.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 49,198 | 7,336 | 4,359,680 | 98.9% | 165.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 2 | 276,940 | 74,457 | 47,743,488 | 99.4% | 283.3k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 18 | 1,412,916 | 136,588 | 75,868,416 | 98.2% | 321.7k | ELEVATED | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **Total** | **22** | **1,884,858** | **221,447** | **129,311,616** | **98.6%** | **321.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 9 | 446,582 | 26,217 | 2,447,872 | 84.6% | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **gpt-6.2-sol** | 13 | 1,438,276 | 195,230 | 126,863,744 | 98.9% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **Total** | **22** | **1,884,858** | **221,447** | **129,311,616** | **98.6%** | **Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations)** |

Retained recorded history: 17 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
