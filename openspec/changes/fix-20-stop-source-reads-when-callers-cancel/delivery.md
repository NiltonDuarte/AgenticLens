# fix(FIX-20): stop source reads when callers cancel

Live and historical reads now propagate cancellation through SQLite and Codex source work. Interrupted live requests release the cache lock without admitting an incomplete snapshot. Historical responses preserve usable evidence with explicit incomplete coverage. Backfill keeps a cancelled workspace's ledger and summary intact when it observes cancellation before publication and keeps earlier committed workspaces.

## Verification

The orchestrator integrated `7d7b58e3` with `origin/main` at `83f6f6c1`. The merge was already up to date. `make -j4` passed on the integrated revision, including Go race suites, 231 frontend unit tests, 819 frontend integration tests, static checks, release and installation checks, and native builds for four platforms. The configured manager selected a local engine built from `7d7b58e3`. The OpenSpec change passed strict validation. See [regression evidence](evidence/TESTS.md) for controlled query, file and publication cases.

The preliminary run on a superseded candidate hit worker lifetime limits in three capture tests. Those tests passed in the integrated run. The same three cases also passed independently under the race detector on the base revision.

## Independent review

Behavior, data integrity, architecture and test effectiveness were reviewed with both configured profiles, gpt-6.2-sol/high and gpt-6-luna/medium. Initial reviews requested corrections to Codex evidence and child links, SQLite interruption and cache ordering, and native cancellation oracles. A test review also reported incomplete evidence before the additional regression cases. Focused follow-up reviews approved the affected code and tests. The architecture verdict carries forward because the final relationship correction leaves SQLite and cache ownership unchanged. [Findings and dispositions](evidence/REVIEW_SUMMARY.json)

SQLite materialization can add temporary storage for large reads. Synchronous filesystem operations may take time to return. The verification establishes cancellation at the tested native boundaries without a universal latency guarantee.

## Telemetry

The final nonce-based generation rejected the root Codex rollout at malformed record 943 and preserved its accepted snapshot. Reconciliation attempted 38 owners and reported 34 pending. The ledger preserves the 34 sessions from the preceding branch revision and includes both reviewers of the explicit configuration policy with verified FIX-20 bindings. Source freshness remains partial. Cost stays unavailable where the source lacks price evidence.

## Optional capture observation

New shared configurations write `generation.summary.capture_observation: false` explicitly. The loader requires a boolean at that key even when summary publication is disabled. Add the key to older configurations before loading them. Set it to `true` to include capture observations. A successful publication with the option off removes the previous observation and preserves ordinary metrics and authored text. Capture diagnostics remain available through the CLI.

Independent configuration and publication re-reviews by gpt-6.2-sol/high and gpt-6-luna/medium approved the change at `9294cb8a` without findings. Focused config and CLI tests covered explicit serialization, missing and invalid values, opting in and removing a previous report when switching back to `false`. Disposable mutations of the loader and YAML tag reproduced the omission and serialization regressions. The CLI suite, affected adapter race tests and installation, distribution and release checks passed on the same code.

The integrated `make` passed on `9294cb8a` with `GORACE=atexit_sleep_ms=0` and `GOFLAGS='-p=3 -timeout=20m'` against `origin/main` at `83f6f6c1`. It ran the Go race suites, 231 frontend unit tests, 819 frontend integration tests, distribution checks and four native builds. Strict OpenSpec validation passed. The configured manager selected the engine built from `9294cb8a`.

The active protected callback worker retains an older engine and can write its legacy observation until a worker starts with the updated build.

## Pull request

### What changed

- Pass caller cancellation through live and historical SQLite and Codex reads, including waiting for the live trace cache.
- Preserve usable historical evidence and validate native Codex relationships when the request is interrupted.
- Stop backfill on cancellation before publication and preserve the interrupted workspace's ledger and summary. Earlier committed workspaces keep their updates.
- Make the capture observation in generated Markdown opt-in. Setup writes `generation.summary.capture_observation: false` explicitly, and shared YAML that omits the key fails validation. Ordinary metrics and CLI diagnostics remain available.

### Verification

- Integrated `make -j4` passed on `7d7b58e3`, including Go race suites and frontend integration.
- Focused regressions cover indexed SQLite execution, stale tree admission, partial historical coverage, exact snapshot cleanup and byte-for-byte backfill preservation.
- Independent review approved the behavior, data, architecture and test scopes. See the linked evidence in the FIX-20 OpenSpec change.
- The explicit capture observation policy passed integrated `make` on `9294cb8a`. Focused tests covered explicit YAML serialization, rejection of a missing key, opt-in and removal of an older observation.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 71
- **Active Agent Duration:** 16h 58m (recorded duration including 47686s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 191.3k tokens (43% of 450k healthy ceiling, 2 compactions) | Subagents peak 447.2k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.9 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 2.0, Tester 2.0
- **Test-to-Production Code Ratio:** 1.5:1 (+1,469 test lines / +998 prod lines)
- **Code Delivered:** +2,467 / -402 lines across 49 code files (excluding docs)
- **Reasoning Effort:** 81,144 reasoning tokens (22.6% of generated output)
- **Tool Health & Self-Correction:** 1,085 tool calls (77 self-corrections, 92.9% success rate)

### Anomalies & Health Flags
- **Status:** 10 execution anomalies detected (2 CRITICAL, 8 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+436.8k tokens) — Context tokens surged by 436.8k to 436.8k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (6 cycles) — Subagent roles 'tester' and 'coder' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.3k tokens) — Context tokens reached 256.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.3k tokens) — Context tokens reached 253.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.5k tokens) — Context tokens reached 253.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.0k tokens) — Context tokens reached 401.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.1k tokens) — Context tokens reached 401.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (436.8k tokens) — Context tokens reached 436.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (441.1k tokens) — Context tokens reached 441.1k, breaching the 250k warning ceiling.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (4 consecutive errors) — Tool 'command' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 194,561 | 2,321 | 1,275,904 | 86.8% | 192.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 6 | 922,547 | 77,351 | 64,791,680 | 98.6% | 447.2k | ELEVATED | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **tester** | 4 | 420,397 | 72,846 | 63,504,000 | 99.3% | 441.3k | ELEVATED | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Unknown** | 27 | 2,216,953 | 125,599 | 39,562,496 | 94.7% | 191.3k | HEALTHY | Unavailable (0 observed, 0 partial, 27 unavailable / 27 invocations) |
| **Total** | **38** | **3,754,458** | **278,117** | **169,134,080** | **97.8%** | **447.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 38 unavailable / 38 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 13 | 756,748 | 41,458 | 8,235,136 | 91.6% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **gpt-6.2-sol** | 25 | 2,997,710 | 236,659 | 160,898,944 | 98.2% | Unavailable (0 observed, 0 partial, 25 unavailable / 25 invocations) |
| **Total** | **38** | **3,754,458** | **278,117** | **169,134,080** | **97.8%** | **Unavailable (0 observed, 0 partial, 38 unavailable / 38 invocations)** |

Retained recorded history: 34 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
