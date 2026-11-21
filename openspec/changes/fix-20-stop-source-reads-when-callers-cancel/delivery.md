# fix(FIX-20): stop source reads when callers cancel

Live and historical reads now propagate cancellation through SQLite and Codex source work. Interrupted live requests release the cache lock without admitting an incomplete snapshot. Historical responses preserve usable evidence with explicit incomplete coverage. Backfill keeps a cancelled workspace's ledger and summary intact when it observes cancellation before publication and keeps earlier committed workspaces.

## Verification

The orchestrator integrated `7d7b58e3` with `origin/main` at `83f6f6c1`. The merge was already up to date. `make -j4` passed on the integrated revision, including Go race suites, 231 frontend unit tests, 819 frontend integration tests, static checks, release and installation checks, and native builds for four platforms. The configured manager selected a local engine built from `7d7b58e3`. The OpenSpec change passed strict validation. See [regression evidence](evidence/TESTS.md) for controlled query, file and publication cases.

The preliminary run on a superseded candidate hit worker lifetime limits in three capture tests. Those tests passed in the integrated run. The same three cases also passed independently under the race detector on the base revision.

## Independent review

Behavior, data integrity, architecture and test effectiveness were reviewed with both configured profiles, gpt-6.2-sol/high and gpt-6-luna/medium. Initial reviews requested corrections to Codex evidence and child links, SQLite interruption and cache ordering, and native cancellation oracles. A test review also reported incomplete evidence before the additional regression cases. Focused follow-up reviews approved the affected code and tests. The architecture verdict carries forward because the final relationship correction leaves SQLite and cache ownership unchanged. [Findings and dispositions](evidence/REVIEW_SUMMARY.json)

SQLite materialization can add temporary storage for large reads. Synchronous filesystem operations may take time to return. The verification establishes cancellation at the tested native boundaries without a universal latency guarantee.

## Telemetry

The final nonce-based generation rejected the root Codex rollout because record 943 is malformed and the lifecycle interval was incomplete. A read-only check found the malformed record inside the rollout. The accepted ledger and rendered metrics are retained. The ledger contains the root, seven direct native children and 28 Codex owners in total, including all six independently reviewed session IDs available for comparison. Prior accepted owners remain present. A scoped reconciliation attempted all 28 registered owners and left 24 pending at the time of that pass. Latest freshness and full contributor coverage remain unverified. Cost stays unavailable where the source lacks price evidence. A later automatic publication included both Luna review sessions after their explicit FIX-20 bindings were accepted.

## Optional capture observation

Automatic Markdown publication omits the capture observation by default. Set `generation.summary.capture_observation: true` in shared YAML to include it. A successful publication with the option off replaces the previous generated observation and preserves ordinary metrics and authored text. Capture diagnostics remain available through the CLI.

The narrow follow-up received independent behavior review from gpt-6.2-sol/high and gpt-6-luna/medium with no findings. The focused config and CLI race tests covered default omission, opt-in, removal of the prior block, type validation and preservation of runtime diagnostics. An existing cache recovery fixture now enables the option explicitly to keep its coverage of incomplete-source reporting.

The integrated `make` passed on `f05ece30` with `GORACE=atexit_sleep_ms=0` and `GOFLAGS='-p=3 -timeout=20m'`. It ran all Go race suites, 231 frontend unit tests, 819 frontend integration tests, distribution checks and the native builds. The earlier parallel runs exposed the cache fixture assumption and hit CLI and frontend timeouts. The fixture was corrected, the previously failing CLI cases passed in isolation, and the serial run passed without an application or test timeout.

The follow-up nonce-based extraction rejected the same malformed root record. Automatic generation retained all 30 previously accepted sessions and the local ledger now contains 34 Codex sessions. The configured source budget left some reads incomplete. This worktree's protected callback worker still runs an older installed engine, so a subsequent callback may write its legacy observation block until that worker uses the updated build.

## Pull request

### What changed

- Pass caller cancellation through live and historical SQLite and Codex reads, including waiting for the live trace cache.
- Preserve usable historical evidence and validate native Codex relationships when the request is interrupted.
- Stop backfill on cancellation before publication and preserve the interrupted workspace's ledger and summary. Earlier committed workspaces keep their updates.
- Make the capture observation in generated Markdown opt-in while continuing to publish ordinary metrics and keep CLI diagnostics available.

### Verification

- Integrated `make -j4` passed on `7d7b58e3`, including Go race suites and frontend integration.
- Focused regressions cover indexed SQLite execution, stale tree admission, partial historical coverage, exact snapshot cleanup and byte-for-byte backfill preservation.
- Independent review approved the behavior, data, architecture and test scopes. See the linked evidence in the FIX-20 OpenSpec change.
- The capture observation follow-up passed integrated `make` on `f05ece30` with bounded Go package concurrency and an extended Go test timeout.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 67
- **Active Agent Duration:** 15h 5m (recorded duration including 43023s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 191.3k tokens (43% of 450k healthy ceiling, 0 compactions) | Subagents peak 416.7k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 2.0, Tester 2.3
- **Test-to-Production Code Ratio:** 1.4:1 (+1,381 test lines / +994 prod lines)
- **Code Delivered:** +2,375 / -366 lines across 34 code files (excluding docs)
- **Reasoning Effort:** 73,975 reasoning tokens (23.1% of generated output)
- **Tool Health & Self-Correction:** 896 tool calls (67 self-corrections, 92.5% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (1 CRITICAL, 6 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (6 cycles) — Subagent roles 'tester' and 'coder' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.3k tokens) — Context tokens reached 256.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.3k tokens) — Context tokens reached 253.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.5k tokens) — Context tokens reached 253.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.0k tokens) — Context tokens reached 401.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (401.1k tokens) — Context tokens reached 401.1k, breaching the 250k warning ceiling.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (4 consecutive errors) — Tool 'command' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 194,561 | 2,321 | 1,275,904 | 86.8% | 192.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 5 | 817,609 | 70,485 | 57,711,872 | 98.6% | 416.7k | ELEVATED | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **tester** | 3 | 282,084 | 54,634 | 50,479,872 | 99.4% | 415.8k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 25 | 2,039,332 | 119,358 | 38,761,728 | 95.0% | 191.3k | HEALTHY | Unavailable (0 observed, 0 partial, 25 unavailable / 25 invocations) |
| **Total** | **34** | **3,333,586** | **246,798** | **148,229,376** | **97.8%** | **416.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 12 | 696,568 | 38,330 | 7,791,616 | 91.8% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **gpt-6.2-sol** | 22 | 2,637,018 | 208,468 | 140,437,760 | 98.2% | Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations) |
| **Total** | **34** | **3,333,586** | **246,798** | **148,229,376** | **97.8%** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |

Capture observation finalized: 2026-11-21T11:36:50.539658166Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 4; retained older facts: 30; pending: 30; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 34; reserved bytes: 536870912. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1fb74-11fd-7f01-8db2-a744d804ef08: pending; last successful source read 2026-11-21T11:36:05.029149785Z; accepted source checkpoint 2026-11-21T00:24:12.128Z.
Owner codex/01a1fb81-6242-7281-8e7e-d85d1827b939: pending; last successful source read 2026-11-21T11:36:06.093006095Z; accepted source checkpoint 2026-11-21T00:32:10.446Z.
Owner codex/01a1fb8c-ca4a-7c12-a904-d25cdbcff659: pending; last successful source read 2026-11-21T11:36:07.590525158Z; accepted source checkpoint 2026-11-21T00:58:16.842Z.
Owner codex/01a1fb8d-fe92-76e0-86aa-5f2d5036cea4: pending; last successful source read 2026-11-21T11:36:15.729702582Z; accepted source checkpoint 2026-11-21T00:43:44.69Z.
Owner codex/01a1fb9e-5adc-7152-9275-85740179281e: pending; last successful source read 2026-11-21T11:36:17.182492225Z; accepted source checkpoint 2026-11-21T01:01:12.072Z.
Owner codex/01a1fb8c-e2fb-76a3-914e-310dd9374d12: pending; last successful source read 2026-11-21T11:36:18.585167578Z; accepted source checkpoint 2026-11-21T00:42:19.848Z.
Owner codex/01a1fb94-ca62-78e2-ad49-93da432a6368: pending; last successful source read 2026-11-21T11:36:20.02984062Z; accepted source checkpoint 2026-11-21T00:51:38.753Z.
Owner codex/01a1fb70-ce81-7a92-ba46-b1ce3ea7f5f4: pending; last successful source read 2026-11-21T11:36:24.360709097Z; accepted source checkpoint 2026-11-21T00:14:12.55Z.
Owner codex/01a1fb84-907f-7510-98f9-722d8688237b: pending; last successful source read 2026-11-21T11:36:25.6957439Z; accepted source checkpoint 2026-11-21T00:38:53.39Z.
Owner codex/01a1fb63-4432-7eb2-a6cb-7cfa79c495b8: pending; last successful source read 2026-11-21T11:36:27.088921376Z; accepted source checkpoint 2026-11-21T01:00:10.68Z.
Owner codex/01a1fb63-1995-7ee3-a42b-a3d502928956: pending; last successful source read 2026-11-21T11:36:28.36555922Z; accepted source checkpoint 2026-11-21T00:40:44.531Z.
Owner codex/01a1fb74-b75f-72a3-ba3d-e4879d8b2791: current; last successful source read 2026-11-21T11:36:46.661805326Z; accepted source checkpoint 2026-11-21T00:16:00.324Z.
Owner codex/01a1fb70-f9fe-7423-80f5-c4f208684e3b: current; last successful source read 2026-11-21T11:36:47.62234521Z; accepted source checkpoint 2026-11-21T00:13:54.953Z.
Owner codex/01a1fb63-2af9-7832-8d29-c5ae35508ac5: current; last successful source read 2026-11-21T11:36:48.645430556Z; accepted source checkpoint 2026-11-21T00:41:02.753Z.
Owner codex/01a1fb8f-4a23-7ff3-863b-17c00f321f58: current; last successful source read 2026-11-21T11:36:49.424437438Z; accepted source checkpoint 2026-11-21T00:57:18.751Z.
Owner codex/01a1fb96-87f4-7162-ad90-f9dc7f3fcda0: pending; last successful source read 2026-11-21T11:34:01.85885356Z; accepted source checkpoint 2026-11-21T00:53:50.752Z.
Additional owner freshness details remain in local capture diagnostics.

Retained recorded history: 30 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
