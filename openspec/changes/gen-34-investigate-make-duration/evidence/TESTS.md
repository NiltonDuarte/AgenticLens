# GEN-34: Tester verification of the native fixture and scale changes

**Verdict: PASS.** The native fixture is faithful to the probe answers it replaced, production validation still runs against it, and the converted tests still fail for the reasons they assert. The isolated `capture` race run drops from 7:07 to 2:58 under the same conditions. The cost of the scale changes is stated below in successful reads and fixture size. Nothing in this pass blocks delivery. The open points at the end are informative or were already present on the baseline.

## Tested revision

- Baseline: `ab42d501`. Candidate: `ab42d501` plus the uncommitted working tree.
- Candidate test diff (`git diff -- backend`) SHA-256 `4749b2ff…07e4`, plus the untracked `capture/native_fixture_test.go` (`7b411ca5…a2d8`) and `capture/testdata/nativefixture/main.go` (`ccacef41…8c1c`). The hashes were identical at the start and the end of this pass.
- Baseline numbers were measured again in this session from a temporary `git worktree` at `ab42d501`, built outside the repository and removed afterwards. Mutation checks ran in a second temporary worktree with the candidate applied. That worktree was also removed, and no repository file was edited.
- Environment: Go 1.26.0 linux/amd64, 12 cores, WSL2. The machine is shared, and the load is listed with each timing.

## 1. Fidelity

**No hidden use of the test executable.** A search for `os.Executable`, `AL79_CAPTURE_NATIVE_FIXTURE`, `AL_CAPTURE_INSTALL_FIXTURE` and the fixture helpers finds `os.Executable()` only in the call sites the design keeps: `install_native_validation_test.go`, the `opencode` symlink in `install_deadline_unix_test.go` (it uses its own `AL_CAPTURE_ADAPTER_BLOCKING_PROBE` helper), and the `-test.run` subprocess re-runs in `diagnostics_test.go` and `store_test.go`. None of them sends protocol probes to the test binary. `AL_CAPTURE_INSTALL_FIXTURE` no longer appears anywhere, and the protocol-answering `init` has been removed.

**Every fixture invocation was recorded.** A scratch copy of the fixture logged each call during a full `capture` run, and all 184 tests passed ([tester-fixture-invocations.tsv](tester-fixture-invocations.tsv)).

- Every invocation matched an entry in the fixture table. No test sent arguments the fixture rejects.
- With the variant unset, the calls come from the `installFixture` and relocation paths: `engine protocol`, `capture protocol`, `version --short` and `version --json`. The removed `TestMain` branch answered exactly these four commands, so these paths get no new answers.
- The table entries `dispatch protocol` and `manager codex-binding-wake-protocol`, and the `old-interface` variant, are never called, on the baseline or on the candidate. They are carried over from the old `init`.

**Mutation checks** ([tester-fixture-mutations.tsv](tester-fixture-mutations.tsv)). Each mutation changed one fixture answer in the scratch copy and then ran the whole `capture` package:

| Mutation | Failed tests | Rejecting check (production message) |
|---|---:|---|
| `capture protocol` answers 2 instead of 3 | 14 | `unsupported native capture interface` |
| `engine protocol` answers 2 | 14 | `unsupported native engine interface` |
| `version --short` answers v9.8.8 | 10 | `native manager resolution differs from shared version/configuration` |
| `version --json` and `--manager-version --json` report v9.8.8 | 34 | `replacement manager identity v9.8.8 differs from requested v9.8.7` |
| `manager protected-codex-hook-protocol` answers 0 | 22 | `protocol mismatch: observed "0", expected "1"` |
| `capture protected-worker-protocol` answers 1 | 1 | protected worker protocol 2 required for guarded publication |
| `bad-manager` variant answers as compatible | 1 (`TestIncompatibleUserManager…`) | `incompatible manager accepted` |
| Self-update given a wrong SHA-256 for the fixture | 2 of 2 run | `native artifact SHA256 mismatch; installation preserved` |
| Helper returns a shell script with identical answers | 3 of 3 run | `bad magic number`, `selected cache artifact is not a native executable` |

- The reason for the `bad-manager` failure was logged in the scratch copy: `native manager unavailable: native manager interface incompatible: "0"`. That is the intended rejection.
- The **missing native** case (`TestWorkerSuccessor…/missing native`) deletes the installed symlink, so it does not depend on what the fixture contains, and it still passes.
- `TestRetainedAdmissionRejectsSourceChangedAfterValidation` still truncates the source or flips a byte in the fixture copy, and its three cases still pass.

## 2. Package runs (`-count=1`)

| Package | Mode | Result | Package time | Load (1/5/15) at start |
|---|---|---|---:|---|
| `capture` | race (precompiled, isolated) | 184 pass | 2:58 wall | 1.00 1.97 1.40 |
| `capture` | no race | 184 pass | 96.8 s; repeated 90.6 s | 0.73 1.24 1.32; 4.03 11.41 11.72 |
| `internal/adapters` | race | 115 pass | 24.5 s | 2.61 2.00 1.60 |
| `internal/adapters` | no race | 115 pass | 6.4 s | 2.34 1.99 1.61 |

- No tests were skipped.
- `TestLargeStoreFamilyWithinCaptureBudget` with race:
  - divisor 4 passed in 26.6 s: 497 files, 761 MB store, 134 MB budget, about 13.3 MB read per pass;
  - the new default, divisor 8, passed in 14.1 s: 261 files, 389 MB store, 67 MB budget, about 6.9 MB read per pass;
  - the baseline default, divisor 4, took 26.0 s.

  At divisor 8 the store is still 5.8 times the budget and the family read is about a tenth of it, so the test still distinguishes between reading only the family and reading the whole store.

## 3. Successful replacement reads

`TestLoadTicketWorkspaceRejectsReplacementIdentity`, 5 repetitions per cell ([tester-replacement-reads.tsv](tester-replacement-reads.tsv)), load 1.7-2.1:

| Revision | Mode | Total reads | Successful reads | Seconds |
|---|---|---:|---|---|
| Baseline | race | 1500 | 108-131 (mean 123) | 19.9-22.0 |
| Baseline | no race | 1500 | 111-135 (mean 121) | 2.9-3.3 |
| Candidate | race | 400 | 31-39 (mean 35) | 5.3-5.7 |
| Candidate | no race | 400 | 22-44 (mean 29) | 0.7-0.9 |

About 8% of reads succeed under both revisions. Only successful reads exercise the identity assertion, and their number falls by a factor of about 3.7, from about 122 to about 32 per run. The package-run results above (29 and 31) fall in the same range.

## 4. Timing against the baseline

Both revisions were measured in this session from precompiled race binaries, run from the package directory with `go tool test2json` and `/usr/bin/time -v`:

| `capture` race, isolated | Wall | User + sys | CPU | Load (1-min) start to end |
|---|---:|---:|---:|---|
| Report (`be630e36`) | 8:45 | 473 + 236 s | 135% | 3.08 |
| Baseline `ab42d501`, this session | 7:07 | 273 + 148 s | 98% | 1.01 to 1.11 |
| Candidate | **2:58** | 122 + 41 s | 91% | 1.00 to 1.25 |

- The per-test durations sum to 425 s on the baseline and 177 s on the candidate. The race saving is 248 s; without race the sum falls from 106 s to 89 s.
- Per-test data is in [tester-capture-per-test.tsv](tester-capture-per-test.tsv), next to the report's numbers. The report's numbers in [go-slowest-tests.tsv](go-slowest-tests.tsv) were measured under parallel `make` load and are higher than these isolated runs.
- Largest savings with race:
  - user-migration and durable-prefix tests: 20-26 s down to 2-3.4 s each;
  - `TestUserHooksDoNotOwnOrReplaceManager`: 15.0 to 1.0 s;
  - `TestRetainedAdmission…`: 13.3 to 1.0 s;
  - `TestLargeStoreFamily…`: 23.1 to 13.3 s;
  - `TestDiagnosticHeadLookup…`: 12.4 to 4.2 s;
  - `TestInstalledHealth…`: 13.6 to 11.1 s;
  - `TestRelocationResumes…`: 11.7 to 10.3 s.
- `internal/adapters`: the replacement test takes about 5.5 s instead of about 21 s with race.

## 5. Relocation test under 16 CPU burners (informative)

[tester-relocation-loaded.tsv](tester-relocation-loaded.tsv), race binaries, load 20-24:

- Candidate: 2 of 3 runs passed (167-189 s).
- Baseline control: 0 of 1 runs passed (144 s).

The candidate's failure has the same cause as the baseline's: `owned integration refresh failed: required installation planning: context deadline exceeded`, the 2 s `requiredTimeout` in `install.go`. The fixture shortens native probes, but under heavy CPU load the race-built planning code in the test process still exceeds the budget. The fixture does not remove this flake. The report's proposal on wall-clock budgets is still needed.

## Coverage kept and lost

**Kept.** Every converted assertion keeps its oracle, and production checks are unchanged: format and architecture, the executable bit, the protocol and version probes, the self-update checksum and journal, protected-manager ownership, engine installation and executable retention. The mutation checks in section 1 show this. All production code running in the test process is still race-built.

**Lost.**
- About 3.7 times fewer successful replacement reads per run.
- The 256-owner head-lookup case. The constant `HeadLookups == 8` and the query-plan oracle still hold at 32 and 128 owners.
- Large-store coverage at the default scale below divisor 8. The flag still allows 1-16, and divisor 4 passes.
- The copy, journal and retention paths now handle a native file of about 2.4 MB instead of 35 MB.

## Gaps and observations (non-blocking)

1. **The `bad-manager` assertion accepts any error** (`err == nil` is the only check). This was already the case on the baseline. The scratch log shows the intended reason today. Scenario: if the fixture's `bad-manager` variant stopped working so that the manager probe exits non-zero, for example on an unknown-argument path, `InstallUserCodex` would fail with `exit status 2` instead of `interface incompatible`, and the test would still pass. Suggested follow-up: assert that the message contains `interface incompatible`.
2. **The replacement test does not assert `successful > 0`.** This was already the case on the baseline. At the observed success rate of about 8%, the probability of zero successes in 400 reads is below 10^-14, so this is not a practical risk today. If a slower read path pushed the rate to zero, the test would pass without checking anything.
3. **Unused fixture entries:** `dispatch protocol`, `manager codex-binding-wake-protocol` and the `old-interface` variant. They are harmless but are not exercised.
4. **`TestRelocationResumes…` takes about 40 s without race but about 11 s with race**, on both revisions. This was already the case on the baseline and comes from a time-dependent wait. It does not affect the race lane that `make` runs.
5. **The fixture build needs `go` on `PATH` and a writable temporary directory**, as the existing `installationManagerExecutable` helper already does. Cleanup runs only from the Unix `TestMain`. The fixture tests are Unix-only in practice because they use `/var/tmp`. No fixture directory was left behind by any run in this pass.
6. **Not run by the tester:** the integrated `make` after merging `main`. The orchestrator owns that step.
