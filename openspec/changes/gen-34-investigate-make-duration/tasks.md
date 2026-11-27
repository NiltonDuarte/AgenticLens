# GEN-34: Execution tasks

Reference design.md for approved scope and steering. This change has no delta specs because product behavior does not change.

## Investigation (delivered)
- [x] Time every `make` recipe with full `make` runs under concurrent load and at low load.
- [x] Profile the Go integration lane per package and per test, with and without `-race`, in parallel and in isolation, with CPU accounting.
- [x] Profile the frontend unit and integration suites per file.
- [x] Reproduce timing-sensitive failures under controlled load and identify their causes.
- [x] Rank the improvement proposals by estimated gain and risk, stating the coverage each one keeps and loses.
- [x] Write the report and delivery summary; validate the change.
- [x] Generate final telemetry, commit, push and open the `chore(GEN-34)` PR without merging.

## Native test fixture (coder)
- [ ] Add `backend/internal/adapters/capture/testdata/nativefixture/main.go`. It imports only the standard library and answers exactly the current `init` table in `user_install_test.go`, keeping the `AL79_CAPTURE_NATIVE_FIXTURE` variants `old-interface` and `bad-manager`. The capture protocol comes from a `-X` variable, and the fixture exits non-zero with a message on any other arguments.
- [ ] Add `backend/internal/adapters/capture/native_fixture_test.go` (package `capture`):
  - a `sync.Once` helper that runs `go build` from the backend root into a temporary directory, passes `HandlerProtocolVersion` with `-ldflags`, makes the result read-only and executable, and fails the calling test on build error;
  - an exported alias for external tests and a cleanup function that `TestMain` calls.
- [ ] In `user_install_test.go`, make `compatibleCaptureTestExecutable` install the fixture instead of `os.Executable()`, and remove the protocol-answering `init`.
- [ ] In `installation_unix_test.go`, point the `installFixture` native symlink at the fixture. Remove the `AL_CAPTURE_INSTALL_FIXTURE` branch of `TestMain` and its `Setenv`, and add the fixture cleanup next to `installationManagerRoot`.
- [ ] In `relocate_install_unix_test.go`, point the native symlink at the fixture and drop the `AL_CAPTURE_INSTALL_FIXTURE` `Setenv`.
- [ ] Leave `install_native_validation_test.go`, the `opencode` symlink in `install_deadline_unix_test.go`, `diagnostics_test.go` and `store_test.go` on `os.Executable()`. Keep the existing `GORACE` settings.

## Scale cases (coder)
- [ ] `capture/diagnostic_index_test.go`: `TestDiagnosticHeadLookupWorkIndependentOfRotatingOwnerHistory` runs counts 32 and 128.
- [ ] `internal/adapters/workspace_replacement_linux_test.go`: `TestLoadTicketWorkspaceRejectsReplacementIdentity` uses one named constant of 400 reads for both the loop and its log line.
- [ ] `capture/reconcile_large_store_test.go`: the default of `-fix21.large-store-divisor` becomes 8, the 1 to 16 range stays, and the comment states the default.

## Verification (tester)
- [ ] Confirm that no remaining capture test runs the test executable for protocol answers, and that converted negative cases (`bad-manager`, mismatched checksum, missing native) still fail for the asserted reason.
- [ ] Run `internal/adapters/capture` and `internal/adapters` with and without `-race` (`-count=1`). Run `TestLargeStoreFamilyWithinCaptureBudget` once more with `-fix21.large-store-divisor=4`, and record the logged count of successful replacement reads.
- [ ] Time an isolated precompiled `capture` race run, with load, and the converted and scale tests through `go test -json`. Compare them with the 8:45 baseline and `go-slowest-tests.tsv` in `evidence/`.
- [ ] Optionally repeat the relocation flake reproduction under 16 CPU burners. The result is informative, not an acceptance condition.

## Review and delivery (orchestrator)
- [ ] Run focused test-effectiveness and behavior reviews of fixture fidelity and coverage, then the consolidated review; retain findings under `evidence/`.
- [ ] Commit the candidate, merge `main`, run the integrated `make` and record the tested revision.
- [ ] Update `delivery.md` and the PR title and body, validate the change, regenerate final telemetry, then commit, push and update PR #171 without merging.

## Workflow follow-up
Archive only after verified delivery. GEN-34 has no roadmap entry to reconcile.
