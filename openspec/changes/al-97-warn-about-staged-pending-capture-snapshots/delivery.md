# feat(AL-97): warn about staged pending capture snapshots

## Outcome

The optional pre-commit advisory reads changed snapshots from the effective commit index and reports which captures await a ticket or a publication destination. Setup offers a separate opt-in and respects existing hook ownership. Confirmed exact local discards suppress their stale staged copies. The advisory leaves staging under Git's control.

## Verification

The implementation candidate is `2e7465d5` over `db4477fa`. Integration with `origin/main` required no merge changes. Focused race tests passed for real commits with a temporary index, linked worktree and alternate object storage, corrupt or missing discard evidence, and bounded inspection. The final test-only correction uses valid snapshots to assert complete coverage at each limit and partial coverage above it. Removing either limit signal broke its intended test.

The integrated command `GOFLAGS='-p=1 -timeout=35m' GOMAXPROCS=4 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed at `2e7465d5`. It ran the CLI and adapter suites with the race detector, 231 frontend unit tests, 819 frontend integration tests, installation and distribution checks, vet, typecheck, lint, and the multi-platform build. The first broad run had capture-fixture failures and two Go package timeouts. Serializing Go packages and extending the package deadline enabled the full integrated pass. Strict OpenSpec validation and `git diff --check` also passed.

## Review

Behavior and architecture review on GPT-6.2 Sol (high) found that an incomplete staged scan could hide the existing health warning. Data review on the same profile identified working-copy path, Git object store, and discard evidence edge cases. Architecture review also found the SQLite byte-limit issue. Test review on that profile identified the missing limit oracles and worker-dependent fixture. The corrective checks reproduced each changed behavior. GPT-6 Luna (medium) reviewed the same four focuses and independently reproduced the corrected linked-worktree behavior after one unavailable-hook result. The final test-only adjustment was checked with both targeted mutations. No blocking review finding remains. See the [compact findings](evidence/REVIEW_SUMMARY.json).

## Telemetry coverage

Final explicit Codex generation refreshed the root and its three verified descendants. The ledger retained all 21 previously recorded independent Codex sessions. All 25 verified registrations are present with their recorded parent links. One earlier architecture review run lacked a verified binding for its reported session ID and is absent. Automatic reconciliation reached the discovery budget, so this checkpoint does not establish a fresh collection for the independent sessions. Costs are unavailable and the open root interval remains provisional.

## Pull request

Add an opt-in advisory for pending captures added or updated by the proposed Git commit. The hook distinguishes the two pending states, keeps working when the optional local registry is unavailable, and defers commit and staging decisions to Git and the developer.

The hook uses Git's effective commit index, including path-limited commits, and retains the existing health warning when a staged scan is incomplete. Existing hooks remain under their owners' control. The integrated `make` gate passed.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 75
- **Active Agent Duration:** 2h 52m (recorded duration including 1784s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 301.2k tokens (67% of 450k healthy ceiling, 0 compactions) | Subagents peak 254.7k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 4.0, Tester 2.0
- **Test-to-Production Code Ratio:** 1.0:1 (+588 test lines / +605 prod lines)
- **Code Delivered:** +1,193 / -33 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 77,821 reasoning tokens (23.9% of generated output)
- **Tool Health & Self-Correction:** 834 tool calls (58 self-corrections, 93.0% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (1 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (32 cycles) — Subagent roles 'coder' and 'tester' alternated for 32 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (252.0k tokens) — Context tokens reached 252.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (271.9k tokens) — Context tokens reached 271.9k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (4 cycles) — Subagent roles 'planner' and 'coder' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 140,360 | 2,629 | 1,145,600 | 89.1% | 138.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 176,853 | 40,727 | 19,895,808 | 99.1% | 254.7k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 164,561 | 37,864 | 20,795,136 | 99.2% | 241.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 22 | 1,814,584 | 166,876 | 82,631,552 | 97.9% | 301.2k | ELEVATED | Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations) |
| **Total** | **25** | **2,296,358** | **248,096** | **124,468,096** | **98.2%** | **301.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 25 unavailable / 25 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 11 | 551,053 | 39,382 | 4,479,744 | 89.0% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6.2-sol** | 14 | 1,745,305 | 208,714 | 119,988,352 | 98.6% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **25** | **2,296,358** | **248,096** | **124,468,096** | **98.2%** | **Unavailable (0 observed, 0 partial, 25 unavailable / 25 invocations)** |

Retained recorded history: 21 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
