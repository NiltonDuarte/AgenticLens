# chore(GEN-34): explain the slow integrated make and speed up capture test fixtures

## Outcome

GEN-34 delivers two things. Production code, the Makefile and CI are unchanged.

- **Investigation report.** The [report](evidence/REPORT.md), measured on `be630e36`, explains where the integrated `make` spends its time and why it times out, and ranks fixes by gain, risk and coverage. A passing `make` took about 11.7 minutes on a quiet machine and 25.4 minutes with other worktrees running `make`. Two Go packages set the pace: `cmd/agentic-lens` waits on real worker processes, and `internal/adapters/capture` is CPU-bound. Each sits just under Go's default 10-minute test limit.
- **Faster `capture` tests, approved by the user as steering** ([design](design.md#steering)).
  - Capture tests used to install the 35 MB race-built test executable as a fake manager. Production self-update journals and hashes the whole binary, which is very slow under the race detector.
  - The tests now use a 2.4 MB host-native fixture, built once per test process. Every production validation step still runs against it.
  - Three scale cases are smaller: the head-lookup case drops 256 owners and keeps 32 and 128, the workspace-replacement race runs 400 reads instead of 1500, and the large-store default divisor goes from 4 to 8. The design states what coverage each change gives up.
  - **Effect:** the isolated `capture` race run fell from 7:07 to 2:58 under the same load. Inside the integrated `make` it took 290 s, against 421 s in the report's low-load run.
  - The `make` wall time did not drop. `cmd/agentic-lens` (525 s) still sets the length of the Go lane and still uses 87% of the 10-minute limit. The report's other proposals remain open, and each needs its own ticket.

## Verification and review

- **Tested revision.** `3b6c485f`. `main` had no new commits, so the merge was a no-op.
- **Integrated `make`.** Passed in 769 s at load 1.2 to 5.6. Per-recipe and per-package times are in [make-gate-3b6c485f.tsv](evidence/make-gate-3b6c485f.tsv).
- **Tester** ([TESTS.md](evidence/TESTS.md)): PASS.
  - A log of every call to the fixture during a full `capture` run matched the old answer table.
  - Nine mutations of the fixture or its helper were each rejected by production validation.
  - `capture` and `internal/adapters` pass with and without `-race`, and the large store also passes at divisor 4.
  - Successful replacement reads per run fell from about 122 to about 32.
- **Review.** Test-effectiveness and behavior reviewers each ran on Claude Opus 5.5/medium and Claude Sonnet 5.5/high. All four approved.
  - They reported 11 low-severity items, recorded in [REVIEW_SUMMARY.json](evidence/REVIEW_SUMMARY.json).
  - TEST-01 is fixed in `3b6c485f`. The bad-manager test now asserts why the manager is rejected, and a scratch mutation proves the assertion catches the old gap.
  - TEST-06 is fixed in the design text. The rest are accepted limits or pre-existing debt.
  - Reviewer mutation runs show that the 400-read replacement test still catches identity-guard regressions: 8 of 8 race runs failed (Opus), and 25 of 25 failed at both 400 and 1500 reads (Sonnet).
- **Coverage not reviewed separately.** The data-integrity and architecture focuses were not run. Only test files changed, and no persistence, production boundary or module dependency changed.
- **Change checks.** `scripts/openspec.sh validate gen-34-investigate-make-duration --strict` and `git diff --check` passed.

## Limits

- **Timing.** The `make` results come from single runs on a machine shared with other worktrees.
- **Flake still present.** `TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest` still fails under heavy CPU load, because of the 2 s planning deadline in `install.go`. The fixture does not fix this.
- **Fixture build.** The fixture directory is left in `TMPDIR` if the test process panics or times out. The converted tests need `go` on `PATH`.
- **Telemetry.** Generated last with `--source claude` from the root session `d1336214-…`.
  - The ledger contains the root and all seven expected native subagents: planner, coder, tester and four reviewers, Opus 5.5 and Sonnet 5.5 as configured.
  - Cost is unavailable because Claude Code transcripts do not record it.
  - The CRITICAL `SUBAGENT_REWORK_LOOP` flag is a false positive. Each reviewer ran once, all four at the same time, and their interleaved events look like alternation.
  - The commit, push and PR update after generation are not included.

## Pull request

Investigation of the slow integrated `make`, plus the first fix the user approved.

- **[Report](openspec/changes/gen-34-investigate-make-duration/evidence/REPORT.md).** It covers stage and test timings, timeout and flake causes, and fixes ranked with their coverage trade-offs.
  - `cmd/agentic-lens` (8-9 min) waits on real worker processes and runs mostly serially.
  - `internal/adapters/capture` was CPU-bound under `-race`.
  - Both sit near Go's 10-minute default limit.
- **Capture test fixture.** Tests installed the 35 MB race-built test executable as a fake manager, so production self-update journaled and hashed it under the race detector. A small host-native fixture keeps every production validation step and cuts the isolated `capture` race run from 7:07 to 2:58.
- **Smaller scale cases.** Three cases are reduced, and the design records what coverage each gives up.

Integrated `make` passed on `3b6c485f`. No production, Makefile or CI changes. The report's other proposals become separate tickets.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 5h 23m (recorded duration including 84s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 240.3k tokens (53% of 450k healthy ceiling, 1 compactions) | Subagents peak 187.0k (Reviewer-tests)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 2.0, Coder 1.0, Tester 2.0, Reviewer-behavior 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 154:0 (+154 test lines / +0 prod lines)
- **Code Delivered:** +154 / -72 lines across 8 code files (excluding docs)
- **Reasoning Effort:** 149,056 reasoning tokens (28.4% of generated output)
- **Tool Health & Self-Correction:** 524 tool calls (9 self-corrections, 98.3% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 CRITICAL).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (36 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 36 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 100 | 45,045 | 4,846,708 | 100.0% | 161.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 56 | 18,635 | 1,956,539 | 100.0% | 90.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 152 | 51,335 | 7,722,958 | 100.0% | 171.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 436 | 125,198 | 30,236,866 | 100.0% | 240.3k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 2 | 136 | 67,318 | 6,092,544 | 100.0% | 140.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 196 | 67,492 | 11,427,467 | 100.0% | 187.0k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **8** | **1,076** | **375,023** | **62,283,082** | **100.0%** | **240.3k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 6 | 840 | 284,051 | 48,446,685 | 100.0% | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **claude-sonnet-5-5** | 2 | 236 | 90,972 | 13,836,397 | 100.0% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **8** | **1,076** | **375,023** | **62,283,082** | **100.0%** | **Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations)** |
<!-- agentic-lens:metrics:end -->
