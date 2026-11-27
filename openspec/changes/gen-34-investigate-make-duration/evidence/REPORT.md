# GEN-34: Why the integrated `make` is slow and times out

## Finding

The integrated `make` spends most of its time in one recipe, the Go integration lane (`go test -race -count=1` over every non-domain package). Within that lane, two packages set the pace:

- **`cmd/agentic-lens`** takes 8 to 9 minutes. Its time is mostly real waiting, not computation: 240 of its 262 top-level tests run serially, and the 132 capture tests that start real worker subprocesses account for 79% of the time.
- **`internal/adapters/capture`** takes about 7 to 9 minutes under the race detector. It is CPU-bound and has no parallel tests.

On a quiet machine each package finishes just under Go's default 10-minute per-package limit, and the Makefile does not raise that limit. When other worktrees run `make` at the same time, both packages exceed the limit. Load also makes tests with fixed wall-clock budgets fail. The reported symptoms are confirmed: the timeouts in both packages and the capture worker failures under load. The frontend 5-second timeout was not reproduced, but its margin is small.

## How it was measured

- **Machine.** Intel i5-11400F (12 logical CPUs), 31 GB RAM, WSL2 (Linux 6.6). Go 1.26.0, Node 24.22.1, Python 3.12.3. The Go build cache was warm and is shared with the other worktrees.
- **Revision.** Main `be630e36` with this change's documents only. No source, test, Makefile or CI file was modified.
- **Instrumentation** stayed outside the repository's behavior:
  - [`timed-sh.sh`](timed-sh.sh), passed as `make SHELL=`, timestamps each top-level recipe line.
  - A 10-second load sampler.
  - `go test -json` and precompiled test binaries run under `/usr/bin/time -v`.
  - The Vitest JSON reporter.
- **Run 1.** `make -k`, fresh worktree, with two to three other worktrees running their own `make` (load 14 to 31). Measured once.
- **Run 2.** Plain `make`, existing `node_modules`, no other `make` in its critical part (load 1 to 5). Measured once.
- **Package and test profiles.**
  - Race profile: load average 6.4, with other worktrees' race lanes running for part of it.
  - No-race profile: load about 5.
  - Isolated per-package runs: load 2 to 3.
  - Frontend integration: two runs at load about 4.
  - Cross-builds: two repetitions.
- **Flaky-test reproduction.** Each test ran 3 times idle and 3 times under 16 busy-loop CPU burners (load 16 to 21).

Single runs on a shared machine carry noise. The ratios and orders of magnitude below are stable across runs; the exact seconds are not.

## 1. Wall time per stage

Seconds per recipe. `make` with no `-j` runs recipes serially. Raw data: [run 1](make-run1-recipes.tsv), [run 2](make-run2-recipes.tsv).

| Stage | Run 1 (load 14-31) | Run 2 (load 1-5) |
|---|---:|---:|
| `go mod download` | 0.3 | 0.1 |
| `npm ci` (only when the lockfile is newer than `node_modules`) | 148.7 | skipped |
| Go domain units (`go test -count=1 ./internal/domain/...`) | 5.3 | 0.7 |
| Frontend units (`npm run test:unit`) | 4.5 | 2.7 |
| Web assets (`npm run build`: `tsc --noEmit` and `vite build`) | 26.6 | 14.5 |
| `go vet ./...` | 21.5 | 0.5 |
| Frontend typecheck and lint | 41.7 | 16.3 |
| `test_release_revision.py` | 5.6 | 2.7 |
| `test_installation.py` (two nested `make verify-build`) | 61.9 | 18.3 |
| `test_platform_support.py` (nested `make release-*`) | 24.2 | 4.2 |
| `test_capture_distribution.py` | 161.0 | 30.8 |
| `test_test_orchestration.py` (about 30 nested `make` and `npm` runs) | 249.6 | 59.2 |
| Go domain race (`go test -race ./internal/domain/...`) | 16.1 | 2.0 |
| **Go integration race lane** | **667.4, failed (two package timeouts)** | **497.7, failed (one flaky test)** |
| Frontend integration (Vitest, 27 files, 829 tests) | 55.7 | not reached; 33.9 and 34.8 when run separately |
| Engine and manager builds, 4 platforms | 35.0 | not reached; 13 on relink, under 2 when fully cached |
| **Total** | **1526 (25 min 26 s), exit 2** | **650 (10 min 50 s), exit 2** |

At low load a passing `make` would take about **700 s (11.7 min)**. That is run 2 plus the frontend integration and build stages measured separately. The Go integration lane is 71% of it, and the Python checks are 16%. Under concurrent load the whole run more than doubles: the Python checks went from 115 s to 502 s.

### Go packages in the integration lane

Seconds per package. Raw data: [package table](go-integration-packages.tsv).

| Package | Run 1, race | Run 2, race | Race profile | No-race profile |
|---|---:|---:|---:|---:|
| `cmd/agentic-lens` | 618, timeout | 495, flaky fail | 1086, flaky fails | 499 |
| `internal/adapters/capture` | 601, timeout | 421 | 880 | 187 |
| `internal/adapters/backfill` | 194 | 57 | 128 | 37 |
| `internal/adapters` | 155 | 59 | 109 | 27 |
| `internal/adapters/manager` | 81 | 26 | 51 | 22 |
| `internal/adapters/live` | 82 | 20 | 53 | 16 |
| `internal/adapters/opencode` | 60 | 23 | 41 | 8 |
| `internal/adapters/http` | 34 | 16 | 25 | 7 |
| The other six packages | at most 17 each | at most 9 each | at most 12 each | at most 4 each |

Go runs up to 12 packages at once, so the lane takes as long as its slowest package. Making the small packages faster would not shorten it.

### Isolated runs with CPU accounting

Each package ran alone, from a precompiled binary, at load 2 to 3. The CPU column is a percentage of one core. Raw data: [CPU table](go-isolated-cpu.tsv).

| Package | Mode | Wall | User + system CPU | CPU |
|---|---|---:|---:|---:|
| `capture` | race | 8:45 | 473 s + 236 s | 135% |
| `capture` | no race | 2:20 | 216 s + 115 s | 236% |
| `cmd/agentic-lens` | race | 8:55 | 251 s + 184 s | 81% |
| `cmd/agentic-lens` | no race | 6:11 | 142 s + 130 s | 73% |

Two different problems follow from this:

- **`capture` is CPU-bound and serial.** It uses about 1.3 of 12 cores. The race detector makes it 3.75 times slower, because pure-Go `modernc` SQLite and file I/O are fully instrumented.
- **`cmd/agentic-lens` waits on the clock.** It uses less than one core even under the race detector, and `-race` adds only about 44%. Its 6-minute floor without race comes from worker lifetimes, polling intervals, sleeps and subprocess startup.

## 2. Largest consumers and their causes

Raw data: [slowest tests with and without race](go-slowest-tests.tsv).

| Consumer | Race (s) | No race (s) | Cause |
|---|---:|---:|---|
| The 132 capture-style tests in `cmd/agentic-lens` | 435 of 549 in isolation | 305 of 366 in isolation | Real `capture worker` subprocesses with refresh, idle, lifetime and debounce windows, polling loops (every 10-50 ms up to 8 s deadlines) and absence checks (`time.Sleep(2100ms)`). They are serial because the shared fixture calls `t.Setenv` (`CODEX_HOME`, `OPENCODE_DB`, `XDG_*`; 55 calls in 15 files), and Go forbids `t.Parallel` there. |
| `capture` package (184 tests, 0 parallel, 22 `Setenv`) | 525 isolated | 140 | CPU in SQLite and file I/O, multiplied by `-race`. Heavy single tests listed below. |
| `TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest` | 56-76 | 58 | Two sources, each repeating full install and relocation with native probe subprocesses (the race-built test binary re-executed). The same duration without race points to waiting and subprocess cost, not CPU. |
| `TestDiagnosticHeadLookupWorkIndependentOfRotatingOwnerHistory` | 17-41 | 13 | Writes N diagnostics with N owners each, for N = 32, 128 and 256. That is about 83,000 owner records serialized into SQLite under race. |
| `TestLargeStoreFamilyWithinCaptureBudget` | 28-40 | 6 | An incident-scale Codex store fixture (`-fix21.large-store-divisor`). Generating and parsing it is CPU work that `-race` multiplies by about 7. |
| `TestUserMigration*` and `TestUserInstallationDurablePrefixes*` (about 6 tests) | 14-37 each | 3-4 each | Crash-prefix enumeration, repeating install and migration for each prefix. CPU and I/O work that `-race` multiplies by about 8. |
| `TestCaptureStaleProvenance…`, `…TerminalTicketSweeps…`, `…OverlappingPublishers…` | 25-54 | 20-23 | Several worker lifetimes run in sequence. These are real-time waits. |
| `test_test_orchestration.py` | 59-250 | — | About 30 nested `make` runs over a fixture. Each starts real `npm`, Vitest and `go test`. Node and Vitest startup dominates and scales badly under load. |
| `test_capture_distribution.py` | 31-161 | — | Installs and runs real binaries and launchers. |
| Frontend integration | 34-56 | — | jsdom with `user-event`. The largest files are `TicketDeepDive` (27 s), `TraceWaterfall` (19 s) and `TicketComparisons` (15 s) of 135 s of file time spread over workers. See the [per-file table](frontend-integration-files.tsv). |

**Duplicated work.** None of these items is a major cost:

- The domain suite runs twice, without and with `-race` (3 s at low load).
- `tsc --noEmit` runs in both `web-assets` and `lint` (about 5 s).
- `go vet` compiles all packages (0.5 s when cached).
- `make` builds 4 platforms and 2 binaries each time. `BUILD_DATE` changes every run, so every binary relinks (about 13 s).

**`-count=1`.** It disables Go's test cache, but the expensive packages create temporary files and subprocesses, so caching would rarely apply to them. The gain was not measured.

**Not causes.**

- Compilation: the race test binaries build in 2 s from the warm cache.
- The Go race lane's `-p` default of 12: it is already parallel across packages.

## 3. Tests that fail because of timing

Each test ran 3 times idle and 3 times under 16 CPU burners. Raw data: [reproduction results](flaky-reproduction.txt). "Observed" lists where the failure appeared in the `make` runs and profiles.

| Test | Observed | Reproduction | Probable cause |
|---|---|---|---|
| `TestCaptureFirstPublicationRefusesMissingUnsafeAndAmbiguousSelection` | Race profile | Idle 3/3 pass. **Loaded 0/3 pass** (deterministic) | The test lowers `max_lifetime_ms` to 1400. Under load the worker is killed ("native worker lifetime exhausted … signal: killed") before it reaches the state the subtest asserts (`missing` and `outside_roots` see an acknowledged generation). Needs owner analysis to confirm whether only the test expectation depends on timing. |
| `TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest` | Run 1 | Idle 3/3 pass (34 s). **Loaded 1/3 pass** (103-120 s) | Production probe deadlines: `requiredTimeout` 2 s and a 3 s probe in `install.go`, applied to the race-built test binary used as the native handler. Errors: "native capture handler protocol timed out" and "required installation planning: context deadline exceeded". |
| `TestCaptureLiveDisableStopsWritesAndManualRecoverySurvives`, `TestCaptureOpenCodeIdleRestoreContinuesBoundedDescendantDiscovery`, `TestCaptureNativeSelectionRechecksConflictsBeforeFirstCommit`, `TestCapturePendingRootsRetainTheirMetadataWhenTicketTypesConflict`, `TestCaptureNativeDestinationPromotesPendingInSameDirectory` | Race profile, with three worktrees' race lanes running together | 3/3 pass both idle and loaded, at 2 to 5 times the idle duration | Worker lifetime and pass budgets (100 ms to 10 s), `runCaptureNative`'s 8 s limit and polling deadlines. Real contention (several race lanes plus I/O) exceeded them; synthetic CPU-only load did not. |
| `TestCapturePendingCompletionWakesWorkerDuringDestinationBackoff` | Run 2, at low load | 3/3 pass both idle and loaded | `TempDir RemoveAll cleanup: … directory not empty`. The test cancels and waits for its `capture worker`, but a process the worker started can still write in the state directory while cleanup runs. This is an inference: grandchildren are not in a killed process group. |
| `TestCapturePublicationBudgetExpiryRetainsPendingAndRecovers` (reported by FIX-21) | Not seen in these runs | 3/3 pass both idle and loaded | The recovery step uses `pass_timeout_ms: 5000` and `max_lifetime_ms: 10000` inside an 8 s exec limit. Heavy contention can exceed it. |
| SQLite "database is locked" (reported) | Not seen | — | The capture store uses `busy_timeout` of 250 ms for read-only and 1000 ms for writes. A slow, race-instrumented writer holding the lock longer surfaces as `SQLITE_BUSY`. This is an inference from the source. |
| `TicketComparisons` 5 s (reported) | Not seen | 9/9 pass | The slowest `TicketComparisons` and `FleetOverview` tests take 3.0 to 3.2 s at load 4, which is 60 to 64% of Vitest's default 5 s `testTimeout`. A full suite under a concurrent race lane can push them over. |

The common cause is that real wall-clock budgets are chosen for an idle machine, and the tests depend on them. Production timeouts (probe deadlines, busy timeouts) and test policies (worker lifetime and pass budgets, exec limits) both apply to processes that, under `-race` and contention, run 2 to 5 times slower. Retrying succeeds because the load has changed, not because anything was fixed.

## 4. Proposals

Ordered by estimated gain for the time and risk involved. None was implemented; each needs its own ticket. Estimates marked "inference" were not prototyped.

1. **Set an explicit timeout for the Go integration lane.** For example, `-timeout 30m` on that recipe only.
   - **Gain:** removes the false "timed out after 10m0s" failures seen in run 1 and in other threads. Duration does not change.
   - **Keeps:** all coverage.
   - **Loses:** a real hang takes 30 minutes to detect instead of 10.
   - **Risk:** very low.
2. **Make timing-dependent tests independent of machine speed.**
   - Success-path tests use generous budgets.
   - Expiry tests use deterministic blockers. `TestCapturePublicationBudgetExpiry…` already does this by holding the publication lock; extend that to the 1.4 s and 2 s lifetime cases.
   - Probe and handler deadlines become injectable for tests, or scale with a test-only multiplier.
   - Worker tests kill and wait for the whole process group before `TempDir` cleanup.
   - The frontend integration config raises `testTimeout`, for example to 15 s.

   **Gain:** removes the retry loop, which is currently the main source of lost time, because a flaky failure reruns about 11 minutes of `make`. **Keeps:** all coverage, provided each expiry assertion is still triggered on purpose. **Risk:** low to medium. A careless change could turn an expiry test into one that never expires.
3. **Serialize the heavy lane across worktrees on one machine.** For example, take a machine-wide `flock` around the Go integration lane, or set a documented `GOFLAGS=-p=N` for agent sessions.
   - **Gain:** concurrency made the same `make` 2.3 times slower (1526 s against 650 s) and caused both timeouts. Queueing should keep each run near the 11-minute idle time (inference).
   - **Keeps:** all coverage.
   - **Loses:** time spent waiting in the queue.
   - **Risk:** low. This can also be agent guidance instead of Makefile code.
4. **Make the `cmd/agentic-lens` capture tests parallel-safe.**
   - Pass `CODEX_HOME`, `OPENCODE_DB` and the `XDG_*` variables through each subprocess's `exec.Cmd.Env` instead of `t.Setenv`.
   - Then mark independent tests `t.Parallel()`.

   **Gain:** the package spends most of its time waiting on one core, so this has the largest duration gain. Estimate: from 8-9 to about 2-3 minutes (inference; needs a prototype). **Keeps:** all coverage. **Risk:** medium. Hidden shared state (the process environment set in `TestMain`, shared `HOME` caches) must be audited, and step 2 must come first so that extra concurrency does not create new flakes.
5. **Parallelize the `capture` package.** It is CPU-bound and uses 1.3 of 12 cores.
   - **Gain:** estimated from 7-9 to about 2-3 minutes under race (inference).
   - **Keeps:** all coverage.
   - **Risk:** medium. There are 22 environment mutations to remove first, and the global `heldPublications` lock registry and process-level locks must allow independent directories.
   - Separately, reduce fixture size where a test asserts an algorithmic bound, not scale: `DiagnosticHeadLookup` with N=256, and the large-store divisor. That loses the largest-scale case unless it moves to an opt-in lane.
6. **Run `make -j`.** The Makefile's dependency graph already supports parallel goals, and `test_test_orchestration.py` checks it.
   - **Gain:** overlaps the Python checks, lint and frontend integration (about 2.5 to 3 minutes at low load) with the Go lane. Estimated about 9 minutes instead of 11.7 at low load (inference).
   - **Keeps:** all coverage.
   - **Risk:** more CPU contention, which makes timing failures worse until step 2 lands.
7. **Split the targets into a fast loop and the complete gate.**
   - Keep `make` as the complete gate for CI and release.
   - Add a faster local target: units, vet, lint, the Go integration lane without `-race`, and frontend integration. Run the Python distribution and orchestration checks only when `Makefile`, `scripts/` or `.github/` change.
   - **Gain:** at low load, about 7 instead of 11.7 minutes. It is bounded by `cmd/agentic-lens` without race (6:11), so the gain is much larger after step 4.
   - **Loses locally:** race detection in adapters and the CLI, and the Makefile, distribution and orchestration contracts when those files are unchanged. CI's `backend-check`, `distribution-check` and `build-binary` jobs still run them before merge.
   - **Risk:** policy. AGENTS.md currently requires the integrated `make` after merging main, so this is the user's decision.
8. **Make the orchestration Python test cheaper.**
   - Reuse one fixture run for several assertions, or use the existing `npm` and `go` probe for failure cases where real Vitest adds no evidence.
   - **Gain:** 40 to 200 seconds depending on load.
   - **Keeps:** all coverage, if each observed ordering assertion stays.
   - **Risk:** medium, because this test guards the gate itself.
9. **Small items.**
   - Drop the second `tsc` run.
   - Build only the host platform in the default local `make`. CI and release still cross-compile, but local runs would no longer catch darwin-only compile errors.
   - **Gain:** about 15 to 20 seconds.
   - **Risk:** low.

**Not recommended now.** Removing `-count=1`: the gain is unmeasured and probably small for the heavy packages, and it risks reusing stale results from tests that depend on subprocesses and the environment.

**Suggested order.** Steps 1 and 3 now, because they are cheap and stop the failures. Then step 2. Then steps 4 and 5, the real duration fix. Then reconsider steps 6 and 7.

## Limits

- Each full `make` was measured once under each condition, and no run had zero background load. Isolated package runs were the closest to idle (load 2 to 3).
- Synthetic CPU burners reproduced two of the failures. The I/O and memory contention of several real race lanes was observed in the profiles, not replayed.
- The SQLite lock and frontend 5-second failures were not reproduced. Their causes are inferred from the source and the measured margins.
- Gains for steps 4 to 7 are estimates, not prototypes.
