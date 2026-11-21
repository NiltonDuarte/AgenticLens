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

## Stable ledger order

Generation, backfill and schema migration publish sessions in ID order regardless of which sessions were refreshed. Newly serialized records use stable ordering for nested collections and keep timestamp order for occurrences. Backfill preserves the recorded sequence of retained evidence, including unchanged group evidence when a child source is missing. It keeps the relative order of surviving migration steps. The first rewrite can normalize the order of an older ledger. [Ordering regression evidence](evidence/TESTS.md#stable-ledger-ordering)

The production diff `93a81a79…bd9438d` passed the affected Go race suites, focused CLI tests, `go vet ./...`, strict OpenSpec validation and the Linux host build. A known preexisting migration status defect remains outside this change. The protected callback worker still uses its installed older engine, while the configured manager selects the rebuilt worktree engine for ordinary commands. The latest manual capture retry stopped at the same malformed Codex record and preserved the accepted root snapshot. [Persistence review findings](evidence/REVIEW_SUMMARY.json)

The final persistence re-checks by gpt-6.2-sol/high and gpt-6-luna/medium passed on the same diff. They verified that a partial refresh preserves recorded group references in order and that a complete refresh removes stale references even when the recomputed summary matches the previous one.

The integrated `make` passed on `d7e64263` against `origin/main` at `83f6f6c1`, using `GORACE=atexit_sleep_ms=0` and `GOFLAGS='-p=3 -timeout=20m'`. It ran the Go race and frontend suites, installation and distribution checks, and all four native builds.

## Ledger timestamps

Generation and backfill preserve the accepted ledger when only the publication timestamps differ. Equivalent JSON object formatting is ignored. Session timestamps, array order and recorded evidence still determine whether facts changed. The summary and capture bookkeeping can finish independently. A failed summary update leaves a ledger that already matches the journal in place. [Timestamp regression evidence](evidence/TESTS.md#stable-ledger-timestamps)

The persistence re-review approved the timestamp and recovery corrections. Integrated `make` passed on `bcb27a08` with `GORACE=atexit_sleep_ms=0` and `GOFLAGS='-p=3 -timeout=20m'`. It covered the Go race suites, frontend checks and tests, installation and distribution checks, and native builds for all four platforms.

## Pull request

### What changed

- Pass caller cancellation through live and historical SQLite and Codex reads, including waiting for the live trace cache.
- Preserve usable historical evidence and validate native Codex relationships when the request is interrupted.
- Stop backfill on cancellation before publication and preserve the interrupted workspace's ledger and summary. Earlier committed workspaces keep their updates.
- Make the capture observation in generated Markdown opt-in. Setup writes `generation.summary.capture_observation: false` explicitly, and shared YAML that omits the key fails validation. Ordinary metrics and CLI diagnostics remain available.
- Keep raw ledger array order stable across partial refreshes, backfill and schema migration while preserving historical evidence.
- Keep the ledger bytes and timestamps stable when generation or backfill produces equivalent facts.

### Verification

- Integrated `make -j4` passed on `7d7b58e3`, including Go race suites and frontend integration.
- Focused regressions cover indexed SQLite execution, stale tree admission, partial historical coverage, exact snapshot cleanup and byte-for-byte backfill preservation.
- Independent review approved the behavior, data, architecture and test scopes. See the linked evidence in the FIX-20 OpenSpec change.
- The explicit capture observation policy passed integrated `make` on `9294cb8a`. Focused tests covered explicit YAML serialization, rejection of a missing key, opt-in and removal of an older observation.
- Integrated `make` passed on `d7e64263`. Ledger and backfill regressions covered stable ordering across partial refreshes, migration and preservation of history.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 91
- **Active Agent Duration:** 21h 36m (recorded duration including 61124s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 191.3k tokens (43% of 450k healthy ceiling, 2 compactions) | Subagents peak 447.2k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 2.0, Tester 2.5
- **Test-to-Production Code Ratio:** 1.4:1 (+2,128 test lines / +1,476 prod lines)
- **Code Delivered:** +3,604 / -493 lines across 64 code files (excluding docs)
- **Reasoning Effort:** 111,499 reasoning tokens (22.3% of generated output)
- **Tool Health & Self-Correction:** 1,464 tool calls (109 self-corrections, 92.6% success rate)

### Anomalies & Health Flags
- **Status:** 15 execution anomalies detected (4 CRITICAL, 11 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+436.8k tokens) — Context tokens surged by 436.8k to 436.8k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+151.3k tokens) — Context tokens surged by 151.3k to 151.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+327.7k tokens) — Context tokens surged by 327.7k to 327.7k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (6 cycles) — Subagent roles 'tester' and 'coder' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.3k tokens) — Context tokens reached 256.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.3k tokens) — Context tokens reached 253.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.5k tokens) — Context tokens reached 253.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.0k tokens) — Context tokens reached 401.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.1k tokens) — Context tokens reached 401.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (436.8k tokens) — Context tokens reached 436.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (441.1k tokens) — Context tokens reached 441.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (319.2k tokens) — Context tokens reached 319.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (327.7k tokens) — Context tokens reached 327.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (323.0k tokens) — Context tokens reached 323.0k, breaching the 250k warning ceiling.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (4 consecutive errors) — Tool 'command' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 3 | 723,180 | 7,557 | 6,617,088 | 90.1% | 341.7k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **coder** | 8 | 1,097,594 | 109,625 | 92,339,072 | 98.8% | 447.2k | ELEVATED | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **tester** | 6 | 601,013 | 108,466 | 99,050,624 | 99.4% | 441.3k | ELEVATED | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Unknown** | 37 | 2,843,440 | 163,698 | 44,763,904 | 94.0% | 191.3k | HEALTHY | Unavailable (0 observed, 0 partial, 37 unavailable / 37 invocations) |
| **Total** | **54** | **5,265,227** | **389,346** | **242,770,688** | **97.9%** | **447.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 54 unavailable / 54 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 18 | 981,973 | 53,723 | 9,482,112 | 90.6% | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **gpt-6.2-sol** | 36 | 4,283,254 | 335,623 | 233,288,576 | 98.2% | Unavailable (0 observed, 0 partial, 36 unavailable / 36 invocations) |
| **Total** | **54** | **5,265,227** | **389,346** | **242,770,688** | **97.9%** | **Unavailable (0 observed, 0 partial, 54 unavailable / 54 invocations)** |

Retained recorded history: 50 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
