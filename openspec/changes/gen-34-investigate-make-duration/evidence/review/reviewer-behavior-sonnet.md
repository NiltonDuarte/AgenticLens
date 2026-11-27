# GEN-34 behavior and integration review (RB-SONNET, first pass)

- Run: RB-SONNET, first pass, requested profile sonnet_5_5_high (claude-sonnet-5-5, high effort). Runtime model metadata not independently verifiable from inside the session.
- Reviewed revision: head `66a62c9d` against base `ab42d501` (`git diff ab42d501 66a62c9d -- backend`), clean working tree apart from the `evidence/review/` directory. Production code untouched.
- Scope: integration of the native fixture build with how tests run, and fidelity of the converted fixture to every probe the capture installation code performs. Other reviewers' files under `evidence/review/` were not read.
- Outcome: **APPROVED** (scoped). No finding above severity 3; three low-severity items for the parent to accept or fix.

## Hypotheses and disposition

| ID | Hypothesis | Disposition |
|---|---|---|
| H1 | The fixture answers fewer or different probes than the removed `init` / `TestMain` branch, so a production probe now fails or passes for a different reason. | Refuted. Production probes against native/manager binaries (`config/native.go`, `config/protected_codex.go:369`, `capture/install.go:303,306`, `install_transaction.go:87,90`, `user_install.go:99,102`, `worker_successor.go:82,89`, `manager/run.go:449`) were enumerated. The fixture answers all of `engine protocol`, `capture protocol`, `capture protected-worker-protocol`, `manager protocol`, `manager codex-binding-wake-protocol`, `manager protected-codex-hook-protocol`, `version --short`, `version --json`, `--manager-version --json`, with the old table's bytes. Not answered by either old or new: `manager opencode-hook-protocol` (`user_install.go:99`) and `capture admit-protected-worktree` (`manager/run.go:449`). The old design fell through to re-running the whole test binary there; the new one exits 2 with a message. No current test reaches them (tester invocation log confirms). |
| H2 | The `AL_CAPTURE_INSTALL_FIXTURE` path (installFixture, relocation) now gets extra answers (`manager protocol`, `capture protected-worker-protocol`) it did not get before, which could mask a rejection a test relied on. | Refuted for current tests. Invocation log shows only the four old answers are requested on those paths. Old behavior for the extra commands was a hang/timeout of a recursive test run, not a deliberate rejection, and no test asserts it. The `version --json` commit/build_date strings changed (`adapter-fixture`/`fixture` to `native-test-fixture`/`test`); no test or golden file references either string (grep over backend tests). |
| H3 | Symlink/copy/hardlink of the fixture in `os.TempDir()` is treated differently by production ownership/privacy checks. | Refuted. `verifyProtectedNames` (`config/protected_codex.go`) checks the manager copy created by self-update (fresh 0700 file, owner, `Nlink == 1`), not the source; `manager/storage.go` writes the target with its own mode (0700), so the 0555 source mode is not inherited. `retained.go` copies bytes and only requires an exec bit and nonzero size. Production never calls `os.Link` on native executables (the only `os.Link` uses are the Codex config and git hook stages). The fixture directory is 0700 from `MkdirTemp` under the same temp parent as the old go-build test binary, and mode 0555 is stricter than before. In installFixture and the relocation test the fixture is only reached through a symlink in a temp cache, as the test binary was. |
| H4 | Mutation through the symlink or a test writing the shared fixture corrupts later tests. | Not reproducible. Mode 0555 makes a write through the symlink fail loudly for non-root users; `TestRetainedAdmission...` mutates its own copy. As root, 0555 does not block writes, but no test writes through the fixture symlink or path. |
| H5 | Build environment varies: GOFLAGS, HOME/XDG replaced by earlier tests, TMPDIR, PATH, coverage, shuffle, GOCACHE. | Mostly refuted. `nativeFixtureBuildEnv` is captured at package init, so later `t.Setenv("HOME")`/`XDG_CACHE_HOME` (relocate_test.go, compatibleCaptureTestExecutable) cannot redirect the nested build's cache (stronger than the existing `installationManagerExecutable`, which uses the live env). No test sets `TMPDIR`, `GOFLAGS` or an exclusive `PATH`. Runs that passed: TMPDIR redirected, `-cover -coverpkg=./...`, `GOFLAGS='-trimpath -mod=readonly -buildvcs=false' GOTOOLCHAIN=local`, `-race -shuffle=on`, race on five fixture tests, full non-race package (184 tests, 116 s under concurrent load). Cold build cache: the fixture builds in 9.1 s (0.43 s warm), 2.4 MB. |
| H6 | Fixture cleanup does not run on every platform or on failure. | Partly confirmed, see RB-01. |
| H7 | TestMain exists only on linux/darwin, so other platforms have no cleanup or fail to compile. | Confirmed as a gap, not a regression: `TestMain` is in `installation_unix_test.go` (`linux || darwin`). `native_fixture_test.go` has no tag and compiles on all platforms where the package builds. `GOOS=darwin go vet ./internal/adapters/capture` passes. Windows and freebsd do not build the package at all (`internal/adapters/config` lacks `syscall.Stat_t`, `O_NOFOLLOW`, capture lock functions), and `user_install_test.go` already hard-codes `/var/tmp`, so no supported platform lacks cleanup. |
| H8 | Removing the `init()` table and the `TestMain` branch changes tests that still exec the test binary. | Refuted. `os.Executable()` remains only in `install_native_validation_test.go` (own `AL_CAPTURE_VALIDATION_HELPER` init), `install_deadline_unix_test.go` (own `AL_CAPTURE_ADAPTER_BLOCKING_PROBE` init, keyed on `filepath.Base(os.Args[0]) == "opencode"`), `diagnostics_test.go:256` and `store_test.go:161` (`-test.run` subprocess re-runs). None used the removed branches (`len(os.Args) == 3` guards with unrelated commands, empty `AL79` guard). Child test processes get their own empty `nativeFixtureRoot`, so their `TestMain` cleanup is a no-op and cannot delete the parent's fixture. |
| H9 | Nested `go build` interplay: PWD/cwd, `go` lookup, VCS stamping, race instrumentation, concurrency. | Refuted for `go test` flows. `go test` runs with the package dir as cwd so `../../..` is the module root; `go test` puts GOROOT/bin on PATH; stale `PWD` is ignored by cmd/go; the fixture imports only the standard library so it is static and needs no cgo. Parallel packages each build into their own `MkdirTemp`. VCS stamping and arch selection follow the same environment as the existing manager build helpers. |
| H10 | `go vet ./...` / `make` do not see the fixture source. | Confirmed, harmless. The `testdata` directory is excluded from `./...` (`go list ./... | grep nativefixture` is empty). I vetted it directly (`go vet` and `gofmt -l` clean). It is compiled only when tests run, and a failure surfaces as a clear `build native fixture` test failure. |
| H11 | CI `backend-check` / integrated `make` flow. | No issue found. `make backend-check` runs `vet` then `go test -race -count=1` over the list from `scripts/go-test-packages.sh integration` (the capture package is selected, 1 match); setup-go provides `go` and a cache. No workflow, script or release gate references the removed env variables (repo-wide grep). |

## Findings

### RB-01 Fixture directory leaks when the test process dies without returning from `m.Run`
- Location: `backend/internal/adapters/capture/native_fixture_test.go` (cleanup design) and `installation_unix_test.go` `TestMain` lines 39-43.
- Class: introduced instance of a pre-existing pattern (the same `TestMain` already leaks `installationManagerRoot` and `capture-opencode-config-*`).
- Trigger: a test panic, a `go test -timeout` expiry (the failure mode this ticket investigates), `os.Exit` in a test, or SIGKILL. `m.Run` does not return, so `RemoveNativeTestFixture` is skipped.
- Consequence: about 2.4 MB per aborted run left in the temp directory under `agenticlens-native-fixture-*`. No functional effect on later runs (unique name per run).
- Evidence: disposable worktree probe with a test that built the fixture and then called `panic`, slept past `-timeout 8s`, or called `os.Exit(3)`. All three left `agenticlens-native-fixture-*` in the redirected `TMPDIR`; `t.Fatal` and normal completion left none. A full normal run of the package left no fixture directory (only the pre-existing `capture-opencode-config-*` leftovers from killed subprocess re-runs).
- Severity: 2/10. Suggestion: optionally call the cleanup from a deferred function in `TestMain` so panics on the main test goroutine are covered, accepting that timeouts and `os.Exit` still leak.

### RB-02 Converted tests now require `go` on PATH and a module-root-relative cwd
- Location: `native_fixture_test.go` `nativeTestFixture` (`command.Dir = ../../..`), used by `compatibleCaptureTestExecutable` (24 tests via 6 files), `installFixture` and the relocation test.
- Class: introduced, accepted by design (design.md says it follows `installationManagerExecutable`).
- Trigger: running a precompiled capture test binary (`go test -c`) from a directory other than the package directory, or in an environment without `go` on PATH.
- Consequence: the user-install family fails with `build native fixture: ...go.mod file not found...` or `exec: "go": executable file not found in $PATH`. At baseline `ab42d501` the same precompiled binary run from another cwd with PATH=`/usr/bin:/bin` passed `TestUserHooksDoNotOwnOrReplaceManager`, `TestIncompatibleUserManager...` and `TestRetainedAdmission...`. The error text is clear and no supported flow (`go test`, `make`, CI) hits it.
- Severity: 2/10. Suggestion: none required; consider noting the dependency in the helper comment, and resolving the module root from `runtime.Caller` if precompiled binaries are ever used for timing.

### RB-03 Stale `GORACE` settings and header comment after conversion
- Location: `user_install_test.go` `compatibleCaptureTestExecutable`, `installFixture`, relocation test (`t.Setenv("GORACE", "atexit_sleep_ms=0")`).
- Class: suggestion.
- Consequence: the fixture is not race-built, so the variable no longer affects any child that these helpers launch. It is harmless; design.md deliberately keeps it. Mention only so a later reader does not infer the fixture is race-instrumented.
- Severity: 1/10.

No issue with fixture fidelity was found. The tester's mutation table (changing each answer and observing production rejections) is consistent with my own reading of the production call sites.

## Items outside focus, noted
- `gofmt -l` reports `internal/adapters/failure_rate_policy_test.go`; it is untouched by this diff (pre-existing).
- The `bad-manager` test only asserts `err != nil` (tester gap 1). With the fixture now exiting 2 on unknown arguments, a mistaken fixture table would still pass that assertion. Pre-existing weakness, slightly more relevant now. Not raised as a finding for this focus.
- The nested build takes arch/OS from the environment rather than `runtime.GOOS/GOARCH`. Consistent for `GOARCH=... go test` because the env var reaches both; the same holds for the existing manager helpers.

## Verification run (all against head `66a62c9d`, Go 1.26.0 linux/amd64)
- `go vet ./...` (clean), `go build ./...`, `go vet ./internal/adapters/capture/testdata/nativefixture`, `gofmt -l` on capture package and testdata, `GOOS=darwin go vet ./internal/adapters/capture`; Windows/freebsd vet fails in unchanged dependencies before reaching the package.
- `go test -race -count=1 -run 'TestUserHooksDoNotOwnOrReplaceManager|TestIncompatibleUserManager|TestInstalledStateDriftAliasRemovalAndDryRun|TestRetainedAdmission|TestInstallCallerCancellationDuringOptionalProbe'` with redirected `TMPDIR`: ok, no fixture left behind.
- `-cover -coverpkg=./...`, restricted `GOFLAGS`/`GOTOOLCHAIN=local`, `-race -shuffle=on` on five fixture-dependent tests: ok.
- Full `go test -count=1 ./internal/adapters/capture` (no race, redirected `TMPDIR`): ok, 116.7 s, no fixture directory left.
- Failure-mode probes in a disposable worktree: `t.Fatal`, `panic`, `-timeout`, `os.Exit` (RB-01); precompiled binary from another cwd and without `go` on PATH, compared with a baseline worktree (RB-02). Cold-cache fixture build timing.
- `scripts/go-test-packages.sh integration` selects the capture package.
- Temporary worktrees and probe directories were removed; no repository file was edited by this review besides this report.

## Gaps
- Not run: integrated `make` after merging main, the race full-package run, and macOS/Windows execution (darwin only vetted, not executed). Tester evidence covers the race package runs.
- Did not test the root-user and noexec-TMPDIR cases; both would affect the old test-binary approach equally.
- Agent/session ID: not available to me.
