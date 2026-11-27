# GEN-34 behavior and integration review (RB-OPUS, first pass)

**Outcome: APPROVED** for the behavior and integration focus. The native fixture integrates with every way these tests run today, and it answers every probe that the converted tests send. Two low-severity findings (RB-01, RB-02) are suggestions. Neither blocks delivery.

## Run identity

- Run: RB-OPUS, first pass. Requested profile: opus_5_5_medium (claude-opus-5-5, medium effort). Runtime model reported to the agent: claude-opus-5-5. The agent or session ID is not visible to this subagent.
- Reviewed revision: head `66a62c9d` against base `ab42d501` (`git diff ab42d501 66a62c9d -- backend`). The working tree was clean at `66a62c9d` when the probes ran.
- Independence: no other file under `evidence/review/` was read.

## Hypotheses and disposition

| # | Hypothesis | Disposition | Evidence |
|---|---|---|---|
| H1 | `TestMain`, which holds the cleanup, is missing on some platform where the fixture is built. | Refuted for supported platforms. | `TestMain` is in `installation_unix_test.go` (`linux \|\| darwin`). Untagged callers such as `user_install_test.go` and `control_inbox_test.go` would build the fixture elsewhere. But `GOOS=windows go vet ./internal/adapters/capture/` fails in production `internal/adapters/config` (`syscall.O_NOFOLLOW`, `Stat_t`). Those helpers also hard-code `/var/tmp`. CI runs only on `ubuntu-latest`. `GOOS=darwin go vet` passes. |
| H2 | The build depends on the working directory (`../../..`) and breaks under another runner. | Refuted. | `go test` runs each test binary in its package directory. This is the same convention as the existing `installationManagerExecutable`. No capture test calls `Chdir`. |
| H3 | A test changes `PATH`, `HOME`, `XDG_*` or `TMPDIR` before the first fixture call. The once-only build then fails, or lands in a directory that gets deleted, and every later caller fails. | Refuted. | The build environment is captured at package init, before `TestMain` sets the XDG variables. `exec.Command("go")` resolves `go` through the live `PATH`, but the only change before a first call (`installFixture`, line 330) prepends a harness directory without a `go` entry. No test sets `TMPDIR`. Probes run each likely first caller on its own with `-run`: relocation (which sets `XDG_CACHE_HOME`/`HOME` first), `TestIncompatibleUserManager…` (bad-manager variant) and `TestWorkerSuccessor…`. All passed and left no fixture directory behind. |
| H4 | The fixture build fails under sandbox or VCS conditions that the old `os.Executable()` path did not hit, such as missing `git` or `git` exiting 128 with "dubious ownership". | Refuted. | The probe built the fixture with `PATH` holding only `go`, and with a fake `git` that exits 128. Both builds succeeded. The binary carries no VCS stamp. `GOFLAGS` entries that do not apply to `go build` are ignored. `go test` puts `GOROOT/bin` first on the child's `PATH`. |
| H5 | `go vet ./...`, `go list ./...` (`scripts/go-test-packages.sh`) or `go build ./...` pick up the new `package main`. | Refuted. | `./...` patterns skip `testdata`. `go vet` passes on the fixture package explicitly. `gofmt -l` is clean. No repository script walks `testdata`. |
| H6 | The fixture misses an argument set that production sends to the native or manager binary in these tests. | Refuted. | Production exec sites: `config/native.go` (engine/capture protocol, version --json/--short, manager protocol, --manager-version --json), `config/protected_codex.go` (protected-worker-protocol), `capture/worker_successor.go`, `install_process_unix.go`, `manager/run.go`. Every probe there is in the fixture table. The tester's invocation log shows no rejected arguments. Non-probe commands (`capture worker`, `admit-protected-worktree`, dispatch) were never answerable before: the old binary fell through to running the test suite, so no test could rely on them. |
| H7 | The removed `TestMain` branch and `init()` change subprocess re-runs of the test binary (`diagnostics_test.go`, `store_test.go`, `install_native_validation_test.go`, `install_deadline_unix_test.go`). | Refuted. | Both branches fired only when their environment variables were set and the arguments matched a probe. `-test.run` re-runs never matched. The `opencode` blocking probe has its own `init` keyed on `AL_CAPTURE_ADAPTER_BLOCKING_PROBE` and `argv[0]`. The JSON commit/build_date strings for the `installFixture` path changed (`adapter-fixture` became `native-test-fixture`), and no test asserts on them. |
| H8 | The new location or mode (`os.TempDir`, a 0700 directory, a 0555 file) changes how production ownership, privacy or hard-link checks treat the fixture. | Refuted. | Protected checks (`verifyProtectedNames`: regular file, single link, owner, no group/other write) apply to installed copies under the `/var/tmp` home, not to the source. The symlink targets are in the same temporary-directory class as the old `go-build` test binary. Mode 0555 passes the executable-bit checks and blocks accidental writes to the shared fixture. `os.RemoveAll` removes it on Linux and macOS because the parent directory is writable. |
| H9 | The fixture's cleanup is skipped when tests fail or panic. | Confirmed for panics and timeouts. See RB-01. | Ordinary test failures return from `m.Run`, so cleanup runs. |
| H10 | A data race or ordering bug exists on the shared fixture variables. | Refuted. | `sync.Once` provides happens-before. `RemoveNativeTestFixture` runs after `m.Run`. `t.Fatal` sits outside `Do`. |

## Findings

### RB-01: fixture directory leaks when the test process panics or times out

- Location: `backend/internal/adapters/capture/native_fixture_test.go:31` (directory created in `os.TempDir`), cleanup at `installation_unix_test.go:42`.
- Classification: introduced (same pattern as the existing `installationManagerRoot` and `capture-opencode-config-*`). Suggestion.
- Trigger: a test panics, or the run hits `-timeout`. A timeout is exactly the condition GEN-34 investigates.
- Consequence: `TestMain` never regains control, so a 2.4 MB `agenticlens-native-fixture-*` directory stays in `TMPDIR` on every such run. The old test binary lived in Go's `go-build` work directory, which the go command cleans up.
- Evidence: `TMPDIR=<scratch> go test -count=1 -timeout 8s -run '^TestRelocationResumes…$' ./internal/adapters/capture/` produced `panic: test timed out after 8s` and left `agenticlens-native-fixture-2880563861` (2.4M) next to the existing leaks `capture-opencode-config-*` and a `t.TempDir` root.
- Severity: 2/10. Test hygiene only. Optional fix: build under a stable per-user cache path keyed by content, or accept the leak and document it alongside the existing ones.

### RB-02: a failed first build is cached for the whole process

- Location: `native_fixture_test.go:31-57`.
- Classification: introduced (same as `installationManagerExecutable`). Suggestion.
- Trigger: the first `go build` fails transiently, for example from memory pressure or a killed compiler under heavy parallel `make` load.
- Consequence: every later fixture test in that process fails with the same build error rather than retrying. The failure is loud and correctly attributed, never silently passing.
- Severity: 1/10. No change needed unless such flakes are observed.

## Integration notes (no finding)

- Untagged tests that use `compatibleCaptureTestExecutable` now need `go` on `PATH` at test time. Before, they used only `os.Executable()`. The same package already needed `go` through `installationManagerExecutable`, and both `go test` and CI provide it, so this adds no new requirement for the supported lanes.
- The fixture is not built with the race detector, while the code that validates it runs in the race-built test process. This matches the design.
- Documentation, outside this focus: `design.md` Steering still says "Delivery status: approved. Not yet implemented." The parent may want to refresh it at delivery.
- Unrelated observation, not checked against the base: the isolated relocation run left an empty `opencode/` directory in the overridden `TMPDIR`. The fixture never creates it.

## Verification run by this reviewer

- `git diff ab42d501 66a62c9d -- backend` read in full. Production exec call sites, build tags, CI `backend-check`, the Makefile `test-backend-integration`/`vet` targets and `scripts/go-test-packages.sh` traced.
- `gofmt -l` on the two new files: clean. `go vet` on the fixture package: clean. `GOOS=darwin go vet ./internal/adapters/capture/`: clean. `GOOS=windows`: production code does not compile, which predates this change.
- Fixture build probes in scratch: without `git` and with a failing `git`, both rc=0. Normal build took 0.33 s with a warm cache.
- Single-test first-caller runs (no race, `-count=1`, isolated `TMPDIR`): relocation passed in 59.1 s, incompatible manager in 0.6 s, worker successor in 3.6 s. No fixture directory remained.
- Timeout probe: confirmed RB-01.
- All probe temporaries were removed. No repository file other than this report was written.

## Gaps

- The integrated `make` after merging `main` was not run here. The orchestrator owns it.
- macOS was checked only with `vet`. No tests were run on macOS.
- Full race suites were not rerun. The tester's evidence covers them.

## Verdict

**APPROVED.** RB-01 and RB-02 are low-severity suggestions. Fixes, consolidation, telemetry and publication belong to the parent.
