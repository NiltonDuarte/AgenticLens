# feat(AL-97): warn about staged pending capture snapshots

## Outcome

The optional pre-commit advisory reads changed snapshots from the effective commit index and reports which captures await a ticket or a publication destination. Setup offers a separate opt-in and respects existing hook ownership. Confirmed exact local discards suppress their stale staged copies. The advisory leaves staging under Git's control.

## Verification

The implementation candidate is `2e7465d5` over `db4477fa`. Integration commit `371a9bc1` merged the newer `main`; the roadmap index conflict was resolved by keeping AL-97 In Progress and preserving main's AL-98 Ready entry. Focused race tests passed for real commits with a temporary index, linked worktree and alternate object storage, corrupt or missing discard evidence, and bounded inspection. The final test-only correction uses valid snapshots to assert complete coverage at each limit and partial coverage above it. Removing either limit signal broke its intended test.

The integrated command `GOFLAGS='-p=1 -timeout=35m' GOMAXPROCS=4 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed at `371a9bc1`. It ran the CLI and adapter suites with the race detector, 231 frontend unit tests, 819 frontend integration tests, installation and distribution checks, vet, typecheck, lint, and the multi-platform build. The first broad run had capture-fixture failures and two Go package timeouts. Serializing Go packages and extending the package deadline enabled the full integrated pass. Strict OpenSpec validation and `git diff --check` also passed.

Integration commit `990e1590` merged `main` again, bringing in AL-98 (Claude Code source) and GEN-33. The roadmap index conflict was resolved by listing both AL-98 and AL-97 as In Progress. `make build` passed at that commit. The integrated `make` gate did not complete locally there. The default run hit the 10-minute Go deadline in `cmd/agentic-lens` and `internal/adapters/capture`. Four capture tests in that run failed because their workers were killed when their time ran out, while another worktree was running `make`. One serialized run passed the frontend unit and integration tests and the distribution checks. An outside SIGTERM then stopped its Go integration step. The final run was stopped on request, and hosted CI on the pushed commit is the verification for `990e1590`.

The `publish` check on `5ef7638f` uploaded the Pages artifact successfully but attempted to read it before the run artifact index exposed it. Commit `837d3f06` makes the workflow wait up to 60 seconds for the current run's artifact metadata before downloading it. The workflow YAML and shell syntax parse, and the ticket-pages preview tests pass. The hosted `publish` check still failed on `5baf9eff` because `pull_request_target` runs the workflow from `main`, so this branch's change cannot take effect on its own PR. The same race also fails on `main` and other PRs. Branch protection does not require the check, and GEN-32 ([#166](https://github.com/NiltonDuarte/AgenticLens/pull/166)) owns the main-only Pages fix. This PR leaves the check as it is.

## Review

Behavior and architecture review on GPT-6.2 Sol (high) found that an incomplete staged scan could hide the existing health warning. Data review on the same profile identified working-copy path, Git object store, and discard evidence edge cases. Architecture review also found the SQLite byte-limit issue. Test review on that profile identified the missing limit oracles and worker-dependent fixture. The corrective checks reproduced each changed behavior. GPT-6 Luna (medium) reviewed the same four focuses and independently reproduced the corrected linked-worktree behavior after one unavailable-hook result. The final test-only adjustment was checked with both targeted mutations. No blocking review finding remains. See the [compact findings](evidence/REVIEW_SUMMARY.json).

## Telemetry coverage

The ledger holds the Codex root `01a205de`, its planner, coder and tester children, 21 independent review sessions, and the Claude Code session that merged `main`. One earlier architecture review run lacked a verified binding for its reported session ID and is absent. Costs are unavailable for every session.

After `5baf9eff`, the installed capture worker (`v3.0.0-rc.0+fix20.192bb060`) republished this ledger without the root's three `child_session_id` delegation links. The children stayed in the ledger with the correct `parent_id`; only the links were lost. The worker refreshes each owner separately and reached its 512 MiB source budget after rereading the root but before the children. It then resolves delegation links only against transcripts read in that pass (`finishCaptureAnnotations` in `backend/internal/adapters/capture/reconcile.go`). Codex `spawn_agent` results carry only the agent path, so the root's links were rebuilt as empty and replaced the accepted ones. That rewrite was not committed. Explicit generation with the engine built at the integration commit read all four rollouts of the root tree and restored the links. The capture defect remains on `main` and needs its own ticket.

The final Codex generation reused the literal nonce recorded in that session's last generation command, because a completed Codex session cannot receive a new nonce. The Claude Code session was generated with `--source claude` from the same build. Engines without Claude Code support, including the installed capture worker, reject a ledger containing a `claude` session.

## Pull request

Add an opt-in advisory for pending captures added or updated by the proposed Git commit. The hook distinguishes the two pending states, keeps working when the optional local registry is unavailable, and defers commit and staging decisions to Git and the developer.

The hook uses Git's effective commit index, including path-limited commits, and retains the existing health warning when a staged scan is incomplete. Existing hooks remain under their owners' control. The integrated `make` gate passed at `371a9bc1`; the later merge of `main` relies on hosted CI.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 85
- **Active Agent Duration:** 3h 44m (recorded duration including 37s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 317.7k tokens (71% of 450k healthy ceiling, 1 compactions) | Subagents peak 254.7k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 4.0, Tester 2.0
- **Test-to-Production Code Ratio:** 1.0:1 (+588 test lines / +605 prod lines)
- **Code Delivered:** +1,193 / -33 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 113,359 reasoning tokens (25.5% of generated output)
- **Tool Health & Self-Correction:** 1,018 tool calls (69 self-corrections, 93.2% success rate)

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
| **Unknown** | 24 | 2,613,434 | 250,098 | 110,063,090 | 97.7% | 317.7k | ELEVATED | Unavailable (0 observed, 0 partial, 24 unavailable / 24 invocations) |
| **Total** | **27** | **3,095,208** | **331,318** | **151,899,634** | **98.0%** | **317.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 27 unavailable / 27 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 1 | 168 | 32,765 | 8,542,450 | 100.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **gpt-6-luna** | 12 | 717,647 | 83,112 | 16,249,728 | 95.8% | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **gpt-6.2-sol** | 14 | 2,377,393 | 215,441 | 127,107,456 | 98.2% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **Total** | **27** | **3,095,208** | **331,318** | **151,899,634** | **98.0%** | **Unavailable (0 observed, 0 partial, 27 unavailable / 27 invocations)** |

Retained recorded history: 25 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
