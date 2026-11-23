# fix(AL-78): Guide metrics refresh after native completion

The manual checkpoint points to an automatic capture refresh after natural completion and source availability. When automatic capture is unavailable, it points to explicit generation from outside the completed turn. Caller attribution still requires independent native evidence. The checkpoint threshold, duration accounting, and ledger format are unchanged.

## Verification

The integrated `make` passed on `0ed420a1` with `origin/main` at `db4477fa`. The command used `GOMAXPROCS=4` and `GOFLAGS='-p=4 -timeout=25m'`. Go race, frontend typecheck and lint, 231 frontend unit tests, 819 frontend integration tests, installation and distribution checks, and the multi-platform build passed. The native completion evidence for `2ee112cf` predates only the final fixture teardown edits, which passed 15 focused Python tests.

- The targeted Go race tests passed for checkpoint output, threshold and error handling, and retained provisional facts on Codex and both supported OpenCode extraction families. Existing cross-source refresh tests also passed for active, completed, and growing sources. Independent regression checks confirmed exact-owner replacement, idempotence, preservation of copied unavailable owners and other-harness history, and unknown cost. Timed automatic fixture cases that failed under concurrent load passed on a targeted retry.
- The packaged baseline and candidate both passed the real commit, in-turn feedback delivery, native completion, automatic publication, and external manual fallback scenario. The [native evidence](evidence/native-feedback.md) records source and API timings, exact owner and interval, final usage, and the scripted provider limitation.
- The initial `make -j2` run had four frontend timeouts. The final standalone frontend integration run passed all 819 tests. A later full run reached the Go CLI package's default 10-minute limit. The failing discard and interrupted-publication cases passed on focused retry.
- `TestCaptureFirstPublicationRefusesMissingUnsafeAndAmbiguousSelection/missing` failed intermittently on both untouched `db4477fa` and candidate `2ee112cf`. Each revision failed three of five matching repetitions, reporting acknowledgement with the worker lifetime exhausted. This capture-auto path does not render AL-78 checkpoint feedback.
- The earlier integrated run observed 9 of 11 owners in the OpenCode idle-restore test. The test passed twice in isolation on each revision and passed in the final integrated run. The isolated diagnostic review found no evidence that AL-78's manual renderer caused the failure.
- Independent behavior, data, architecture, and test reviews approved the scoped corrections under gpt-6.2-sol/high and gpt-6-luna/medium. gpt-6.2-sol/high found the concrete guidance and fixture defects in the [compact review record](evidence/REVIEW_SUMMARY.json). gpt-6-luna/medium requested portable-data evidence, then verified the targeted preservation results. Consolidated delivery review passed.

## Limits

The separate default release gate passed on Codex. Its OpenCode discard and restore path did not finish cleanly. The first candidate run observed a partially reconciled restore before a complete pass. The isolated retry encountered a status command failure that did not recur on the retained state. The baseline default run stopped during cold bootstrap, so there is no same-failure baseline reproduction. This path uses automatic capture with manual feedback disabled and is outside the AL-78 completion probe.

## Telemetry coverage

Codex extraction verified the root and five native descendants and preserved all 23 previously recorded owners. The ledger contains 23 owners. 2 independently launched reviews remain absent without verified bindings. Invocation costs are unavailable, and the current turn still contributes provisional duration.

## Pull request

- Clarify when checkpoint metrics are provisional and how to refresh them without a metrics-only model turn.
- Align setup, operator, and contributor guidance with the native capture lifecycle and external manual recovery.
- Exercise the checkpoint and the delay from a real code commit to a settled consumer read on Codex and OpenCode.
- Verification passed with integrated `make`, targeted regression checks, and the packaged completion scenario on Codex 0.170.0 and OpenCode 1.19.5. The measurements demonstrate the refresh workflow, with no established latency reduction.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 63
- **Active Agent Duration:** 2h 55m (recorded duration including 609s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 344.0k tokens (76% of 450k healthy ceiling, 0 compactions) | Subagents peak 370.8k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 1.0
- **Test-to-Production Code Ratio:** 3.3:1 (+23 test lines / +7 prod lines)
- **Code Delivered:** +30 / -14 lines across 4 code files (excluding docs)
- **Reasoning Effort:** 75,697 reasoning tokens (24.7% of generated output)
- **Tool Health & Self-Correction:** 818 tool calls (48 self-corrections, 94.1% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (1 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (14 cycles) — Subagent roles 'tester' and 'coder' alternated for 14 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.5k tokens) — Context tokens reached 254.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (332.2k tokens) — Context tokens reached 332.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (284.6k tokens) — Context tokens reached 284.6k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 145,804 | 3,066 | 1,340,032 | 90.2% | 158.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 49,198 | 7,336 | 4,359,680 | 98.9% | 165.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 3 | 650,834 | 77,536 | 52,712,704 | 98.8% | 370.8k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 18 | 1,767,261 | 142,499 | 86,573,824 | 98.0% | 344.0k | ELEVATED | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **Total** | **23** | **2,613,097** | **230,437** | **144,986,240** | **98.2%** | **370.8k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 9 | 446,582 | 26,217 | 2,447,872 | 84.6% | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **gpt-6.2-sol** | 14 | 2,166,515 | 204,220 | 142,538,368 | 98.5% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **23** | **2,613,097** | **230,437** | **144,986,240** | **98.2%** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

Retained recorded history: 17 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
