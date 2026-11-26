# chore(GEN-34): explain why the integrated make is slow and times out

## Outcome

The [investigation report](evidence/REPORT.md) measures the integrated `make` on this machine, explains where its time goes and why it times out, and proposes ranked fixes. The raw tables are in `evidence/`. No Makefile, test, production or CI file changed. Each proposal the user accepts becomes its own ticket.

- **Duration.** A passing `make` takes about 11.7 minutes on a quiet machine. With two or three other worktrees running `make`, it took 25.4 minutes. The Go integration race lane is 71% of the quiet run, and the Python distribution and orchestration checks are 16%.
- **Two packages set the pace.** `cmd/agentic-lens` takes 8-9 minutes and spends it waiting: it uses less than one core, and 240 of its 262 tests run serially because shared fixtures call `t.Setenv`. `internal/adapters/capture` takes 7-9 minutes; it is CPU-bound with no parallel tests, and `-race` makes it 3.75 times slower. Each package is just under Go's default 10-minute limit when the machine is quiet. Under concurrent load both exceeded it in run 1, which confirms the reported timeouts.
- **Flaky tests depend on wall-clock budgets** in tests and production: worker lifetimes of 1.4-10 s, probe deadlines of 2-3 s, exec limits of 8 s and SQLite busy timeouts of 250-1000 ms. Under load these fail and then pass on retry. Two failures were reproduced under synthetic CPU load:
  - `TestCaptureFirstPublicationRefusesMissingUnsafeAndAmbiguousSelection` failed 3 of 3 times.
  - `TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest` failed 2 of 3 times.

  One cleanup race, a worker subprocess outliving its temporary directory, appeared even at low load.
- **Proposals, in suggested order.**
  1. Set an explicit timeout for the Go integration lane.
  2. Make the timing-dependent tests independent of machine speed.
  3. Serialize the heavy lane across worktrees on one machine.
  4. Make the `cmd/agentic-lens` and `capture` tests parallel-safe.
  5. Parallelize the `capture` package.
  6. Run `make -j`.
  7. Add a fast/complete target split; this one is a policy decision.
  8. Make the orchestration Python test cheaper.
  9. Small cleanups.

  Each proposal in the report states what coverage it keeps and what it loses.

## Verification and review

- **Revision measured.** Main `be630e36`, unmodified apart from this change's documents.
- **Full `make` runs.**
  - Run 1 (`make -k`, load 14-31): 1526 s. `cmd/agentic-lens` and `capture` timed out after 10 minutes, and `TestRelocationResumes…` failed.
  - Run 2 (plain `make`, load 1-5): 650 s. It failed on `TestCapturePendingCompletionWakesWorkerDuringDestinationBackoff` (`TempDir` cleanup).
- **Profiles.**
  - `go test -json` over the integration lane, with race (60 min limit) and without.
  - Isolated test binaries under `/usr/bin/time -v`, with and without race.
  - Two frontend integration runs with the JSON reporter.
  - Two repetitions of the cross-builds.
- **Flaky reproduction.** Each test ran 3 times idle and 3 times with 16 CPU burners.
- **Change checks.** `scripts/openspec.sh validate gen-34-investigate-make-duration --strict` and `git diff --check` passed.
- **Integrated `make`.** It was not rerun as a delivery gate, because this change touches only OpenSpec documents. The two measured runs above are the evidence for this ticket.
- **Review.** No independent review ran; this is an investigation report without code changes.

## Limits

- Each full `make` ran once per condition, on a machine shared with other worktrees. Isolated runs were the closest to idle (load 2-3).
- The SQLite "database is locked" and frontend `TicketComparisons` 5 s failures were not reproduced. Their causes are inferred from the source and the measured margins (the slowest frontend tests take 3.0-3.2 s against the 5 s default).
- Gains for parallelization, `make -j` and the target split are estimates, not prototypes.
- **Telemetry.** Generated last, with `--source claude` and the installed `v3.0.0-rc.169+6c41095e-2-gbe630e36`, from this session (`d1336214-…`). The ledger has one complete root session and no subagents.
  - Because generation ran inside the still-open final turn, 1835 s of the recorded duration is provisional.
  - Cost is unavailable because Claude Code transcripts do not record it.
  - The commit and PR steps after generation are not included.

## Pull request

Investigation only: why the integrated `make` takes about 12 minutes on a quiet machine and over 25 minutes under concurrent load, and why it times out.

- The Go integration race lane dominates. `cmd/agentic-lens` (8-9 min) waits on real worker subprocesses and runs 240 of its 262 tests serially because of `t.Setenv`. `internal/adapters/capture` (7-9 min) is CPU-bound with no parallel tests, and `-race` makes it 3.75 times slower. Each package is just under Go's default 10-minute limit when quiet, and both exceed it when worktrees run `make` concurrently.
- Flaky failures come from wall-clock budgets in tests and production that assume an idle machine. Two were reproduced under CPU load.
- The report ranks fixes, each with its coverage trade-off. First an explicit lane timeout, load-independent timing tests and serializing the heavy lane across worktrees. Then parallel-safe CLI and capture tests, `make -j`, and an optional fast/complete target split.

No Makefile, test, production or CI changes. The proposals become separate tickets after the user decides.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 1h 47m (recorded duration including 1835s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 234.5k tokens (52% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 21,370 reasoning tokens (23.2% of generated output)
- **Tool Health & Self-Correction:** 141 tool calls (3 self-corrections, 97.9% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 288 | 70,800 | 20,778,370 | 100.0% | 234.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **288** | **70,800** | **20,778,370** | **100.0%** | **234.5k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 1 | 288 | 70,800 | 20,778,370 | 100.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **288** | **70,800** | **20,778,370** | **100.0%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |
<!-- agentic-lens:metrics:end -->
