# AL-60 capture implementation evidence

Baseline capture source: 7b00cc1ac7fb5cd8afa2d93f7c65931c8ab2fa30. User-approved iteration 02; search source remains unchanged. Coder: gpt-6.1-sol/high. No production edit preceded the confirmed regression below.

## Deterministic confirmation

A new CLI regression provisions a disposable opencode executable by re-executing the test binary. The helper writes a reached marker and blocks indefinitely until the installer's context kills it. Thus probe expiry is intentional, independent of machine speed or installed harness versions. The test expects installation to succeed with unavailable optional probe data, then validates the committed native/wrapper state.

Command from backend: `go test ./cmd/agentic-lens -run '^TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation$' -count=1`.

Baseline red in evidence/INSTALL_RED.txt: reached marker succeeded; failure at 2.01s was `optional probe exhausted required installation validation: exit status 1; Capture: pending installer native version is unavailable or changed`. The command failed in 2.887s including fixture build. This confirms optional probe consumption of the required transaction's context, without claiming a manifest-variability cause.

The first setup attempt did not dispatch the helper because exec.Command uses opencode as argv[0], rather than the resolved executable path; it was corrected before recording the confirmed red and before production edits. The final helper selects the explicit disposable executable by basename with a dedicated environment guard. Later shared fixture isolation is explicitly disabled only for this intentional blocked-probe test.

## Minimal repair

- InstallOptions.RequiredTimeout retains the two-second default and caps required planning/validation and commit phases separately. The optional harness probe retains its separate three-second allowance. All phase contexts derive from the caller context, so explicit cancellation/deadlines still win; no context detachment or unbounded subprocess was introduced.
- CLI install retains a bounded initial setup phase, then hands the cancellation-aware caller context and requested required-phase timeout to the adapter. Ctrl-C cancellation is preserved through that handoff. Other command timeout defaults remain unchanged. Help states the per-required-phase meaning for installation and the separate optional allowance.
- Native version/protocol validation now compares output only after successful execution. Deadline/cancellation and process execution failure have distinct wrapped causes; actual mismatch reports observed and expected output. Initial selection and transactional revalidation both use the same narrow helper. Required pin/protocol checks, wrapper validation, transaction ownership/rollback and recovery remain in place.
- No production worker, domain, schema, metrics or frontend change was made for this repair.

Initial green: the same blocked-probe CLI regression passed in 3.976s, including installed native/wrapper health. Final race-enabled green in evidence/INSTALL_GREEN.txt passed in 5.343s after the independent shared-harness fixture updates: `go test -race ./cmd/agentic-lens -run '^TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation$' -count=1`. A scoped health check was corrected to use the returned coordination directory rather than the unscoped fixture base; the confirmed baseline failure occurs earlier and its red proof is unchanged.

## Focused verification

Native diagnostic subprocess table passed under race detection in 1.062s: matching output, actual mismatch, exit 19 with retained ExitError, missing executable with retained filesystem cause, caller cancellation and expired deadline. Command from backend: `go test -race ./internal/adapters/capture -run '^TestInstallNativeValidationDistinguishesExecutionAndIdentity$' -count=1`.

A separate real-process API regression synchronizes on an optional-probe TCP handshake before canceling the caller context, and asserts cancellation cause plus absence of a committed manifest. The tester's coordinated adapter run passed it together with native-validation and manifest cases in 1.781s. Successful worker publication, deliberately expired budgets and manifest probe determinism are owned by the tester's independent fixture work; no claim about the historical unretained manifest field is made here.

## Review corrections on frozen candidate e43f3b57f5979e59324599d75e279e2abf6d5099caccbfbe8bdcbdd33aa70770

Three further boundary counterexamples were reproduced before their production corrections. ARCH-01 was independently reported by architecture gpt-6.1-sol/high (severity 6); ARCH-02 by architecture gpt-6-luna/medium (severity 4); DATA-01 by data/failure gpt-6-luna/medium (severity 5). Coder is the correction/verification actor, not an additional independent discoverer.

- ARCH-01: a 100ms optional allowance waited 2.0026s for an opencode launcher's inherited output pipes and 2.0043s for a codex features launcher's pipes. Both launchers have a deliberately finite two-second descendant. Red: evidence/CORRECTIONS_ADAPTER_RED.txt. Local installer command ownership now covers optional version/features, required native identity and wrapper resolution. Unix uses a separate process group, kills that owned group on cancellation and completion, and gives pipe draining a finite 100ms grace. The same optional context still covers version and features together. This is local installer subprocess handling, not a worker/process framework or new global timeout.
- ARCH-02: the durable CLI test signals only the install PID with SIGTERM after a disposable blocked probe reports its PID. Baseline CLI exited while its probe child remained alive; fixture cleanup killed the child. Red: evidence/CORRECTIONS_SIGTERM_RED.txt. Installation now subscribes to SIGTERM as well as interrupt and cancels the caller context before exit; the Unix regression verifies the probe PID is gone after the CLI exits.
- DATA-01: an existing real installer writer seam applies the last target then cancels its required context. Baseline returned nil for both source families. Red: evidence/CORRECTIONS_ADAPTER_RED.txt; an earlier read-only overlay also reproduced it in /tmp/al60-corrections/final_write_red.log. The transaction now checks context after every write, including the last, and uses existing exact-byte rollback/recovery. It also checks after final prefix validation before accepting/removing the journal; expiration there returns the wrapped context cause and retains recovery evidence. The durable test verifies cancellation, exact restoration, journal cleanup after completed rollback, and successful uncanceled retry for codex/opencode.

The tester separately owns TEST-01's inherited harness-fixture guard correction and hostile external harness red/green evidence; see TESTING.md. This implementation does not claim discovery credit for that finding or alter its fixtures.

Coordinated post-correction commands from backend:

- `go test -race ./internal/adapters/capture -run 'TestInstallOptionalProbeBoundsInheritedPipes|TestInstallerFinalWriteCancellationRollsBack|TestInstallNativeValidationDistinguishesExecutionAndIdentity|TestInstallCallerCancellationDuringOptionalProbe' -count=1`: pass 1.516s, evidence/CORRECTIONS_ADAPTER_GREEN.txt.
- `go test -race ./cmd/agentic-lens -run 'TestCaptureInstallSIGTERMCancelsProbeChild|TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation' -count=1`: pass 6.449s, evidence/CORRECTIONS_SIGTERM_GREEN.txt. Original evidence/INSTALL_RED.txt and initial slow-probe evidence remain retained.
- `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/al60-corrections/capture-windows.test.exe ./internal/adapters/capture`: pass.
- `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/al60-corrections/agentic-lens-windows.exe ./cmd/agentic-lens`: pass.
- `git diff --check`: pass.

### Windows ownership and verification limits

The installer runner starts a child suspended, creates a kill-on-close Windows job using the existing x/sys/windows dependency, assigns the suspended process before it can launch descendants, then enumerates/resumes its initial thread through Toolhelp/OpenThread/ResumeThread. Cancellation terminates the job and the directly owned process; the latter also handles cancellation before job assignment. Abort paths reap the suspended child. Job/process/thread/snapshot handles close explicitly with errors preserved. Job closure covers descendants after an ordinary launcher exit; 100ms WaitDelay bounds inherited-pipe draining independently. No taskkill race, new dependency or detached worker was introduced.

New Windows-only runtime regression TestInstallWindowsInspectionOwnsLauncherTree uses a disposable native launcher and descendant, synchronizes a stable child-process handle before permitting launcher exit, and checks actual descendant termination after both inherited-pipe/ordinary-exit and caller-cancel paths. The existing Windows CI job now has one additive step running this test; prior Windows checks remain. Linux cross-compilation passed. Parent subsequently used WSL interoperability to run the actual compiled Windows test executable: `/tmp/al60-corrections/capture-windows.test.exe -test.run '^TestInstallWindowsInspectionOwnsLauncherTree$' -test.v -test.count=1 -test.timeout=60s`. Both launcher-exit and launcher-cancel passed in 0.42s (0.27s and 0.16s respectively); source log `/tmp/al60-windows-runtime.log`. This is parent-supplied actual Windows runtime evidence, distinct from the coder's cross-compilation. The entire hosted CI job has not been run locally. Production worker policy and source remain unchanged. Parent owns fresh independent review, full integration checks and delivery.


## DATA-02 completed recovery acceptance correction

Architecture reviewers gpt-6.1-sol/high and gpt-6-luna/medium independently identified the same distinct completed-journal acceptance path during revision-aware review of candidate 9f418c930f64269bc3da5d2f02aa18627d187729c5b71cb840fdbd6c23613ff0 (both severity 4). This shares cancellation requirements with DATA-01 but has a separate execution path and regression assertion. Coder is the correcting/verifying actor, not a third independent reporter.

New portable regression: `backend/internal/adapters/capture/install_recovery_cancellation_test.go`, TestInstallerCompletedRecoveryRetainsExpiredAcceptance. It uses the existing exact journal fixture, fully applies an inactive/removal transaction, then attempts completed recovery with either an already canceled or expired context for codex and opencode. The inactive path intentionally avoids earlier native-process cancellation checks. Red command from backend: `go test ./internal/adapters/capture -run '^TestInstallerCompletedRecoveryRetainsExpiredAcceptance$' -count=1 -v`. All four cases incorrectly returned a non-nil successful result and nil error in 0.104s; evidence/RECOVERY_RED.txt retains the output. No production edit preceded that red.

Minimal production correction: a context checkpoint in the complete applying-journal branch, after exact prefix validation and before journal removal/result acceptance. Failure wraps the caller's cause and retains the complete journal. Partial/canceled rollback cleanup is unchanged. The regression checks byte-identical retained journal and after-files, then verifies an uncanceled retry accepts the same exact outputs and removes the journal.

Green command: `go test -race ./internal/adapters/capture -run 'TestInstallerCompletedRecoveryRetainsExpiredAcceptance|TestInstallerFinalWriteCancellationRollsBack|TestInstallerDurablePrefixRecoversHookPluginAndManifestCheckpoints' -count=1 -v`: passed in 1.561s. evidence/RECOVERY_GREEN.txt includes all four new cases, the canceled final-write rollback cases and all eight original durable-prefix recovery cases. `git diff --check` passed. Only install_transaction.go, the new focused regression and this evidence changed; no worker, search or subprocess production change. Parent owns the final full review matrix and integrated make.

## Static cancellation cleanup correction after integration

The integrated make at 3e5b7a8 stopped in go vet's lostcancel analysis: the original deferred closure captured a cancel variable that was reassigned for the second required phase. Runtime cancellation was covered, but the analyzer could not establish cleanup on all paths.

The correction changes only install.go: defer the first phase's cancel function directly, and defer the second phase's cancel function directly immediately after creating it. First-phase cancellation still happens explicitly before optional probing. At function exit, the second cancel runs at the same point as before; canceling the already-canceled first phase again is idempotent. Early exits and remove paths still cancel their sole phase. No deadline, caller parent, subprocess policy, transaction ordering, worker or search behavior changes.

Coder verification: standalone `go vet ./internal/adapters/capture` exited 0. Focused adapter race guards for caller cancellation, inherited pipes, native error classification, completed recovery and final-write cancellation passed in 1.666s (evidence/PHASE_GUARDS.txt). The CLI blocked optional-probe race regression passed in 6.286s (evidence/PHASE_CLI_GUARD.txt). `git diff --check` passed. Parent owns the observational-equivalence review rationale, final integrated gate and publication; no full suite was run by coder.
