# AL-60 iteration 02 focused verification

Baseline: 7b00cc1ac7fb5cd8afa2d93f7c65931c8ab2fa30. Search source at 7a50a5e is untouched. Coder retained deterministic baseline red evidence in INSTALL_RED.log before production repair; tester independently inspected marker, explicit-cancellation/no-manifest and typed native diagnostic oracles.

## Fixture and coverage changes

- CLI installer helper provisions disposable Codex/OpenCode executables using the current test executable, exact optional inspection commands and an explicit subprocess guard. New helper: backend/cmd/agentic-lens/capture_harness_fixture_test.go. Hardlink with copy fallback supports executable fixtures without Windows symlink privileges. Real installed native and wrapper identity checks remain in the existing resolver path. Deliberate probe tests can override PATH after prepareCaptureInstall.
- Unix adapter installFixture uses disposable harness scripts instead of the developer's installed harnesses. The existing manifest stability/pin test keeps its full-byte and pin/wrapper-health assertions.
- Controlled manifest successful probe then exit-42 probe reproduces only HarnessVersion and Capability variation, with all other manifest identity fields unchanged. Observed HarnessVersion 1.19.5 -> unavailable; Capability inspected copy -> Harness unavailable copy. This proves the mechanism under controlled outcomes, not the field responsible for the historical unretained failure.
- Existing exact-project shared-store and multiple-root/periodic-refresh success cases use an explicit bounded success policy (lifetime 10000ms, pass 5000ms) instead of a one-second expiry policy. Their token/source isolation, invocation freshness, handoff and ownership assertions remain intact. The failure-feedback test retains its short failing phase and selects the success policy only for later successful retained-fact publication.
- New real publication-lock expiry/recovery test deliberately uses lifetime 1000ms/pass 100ms while actual ledger publication ownership remains held. It asserts unchanged accepted bytes, retained pending generation and acknowledgement, worker ownership release, then real recovery with updated source tokens and acknowledgement after releasing the lock. No global policy defaults or test framework timeouts changed.

## Focused execution

Commands from backend after canonical assets existed:

1. go test -race ./internal/adapters/capture -run 'TestInstallCallerCancellationDuringOptionalProbe|TestInstallNativeValidationDistinguishesExecutionAndIdentity|TestInstallationManifest(StableAndPinSwitchRequiresRefresh|RecordsChangedOptionalProbeOnly)$' -count=1 -v
   PASS, 1.781s. Includes matching/mismatch/process-failure/missing/canceled/deadline native diagnostic subcases, caller cancellation during blocked optional probe, preserved installer stability/pin behavior and controlled optional-field variation. Log: /tmp/al60-installer-fixture-tests.log.
2. go test -race ./cmd/agentic-lens -run 'TestCapture(PublicationBudgetExpiryRetainsPendingAndRecovers|AutomaticSharedStoreAndTicketStayInExactProjects|NativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate|LaterAgentChannelReportsBoundPublicationFailure|CLIOpenCodePluginOwnershipAndMissingHandler|InstallerUsesSharedLocalAndExplicitStatePolicy)$' -count=1 -v
   PASS, 10.430s. Installer policy subcases shared/local/explicit passed. Log: /tmp/al60-worker-fixture-tests.log.
3. git diff --check: PASS.

An initial tester development run exposed an introduced fixture mistake: writing deprecated detailed_report:false emitted advisory stderr into CombinedOutput JSON. The helper now omits that deprecated key except the existing deliberate legacy true case; the corrected focused run above passed. No product assertion was weakened to hide this fixture error.

Parent owns full package/integration gate and independent review. These focused results do not establish passing full make, real-harness compatibility, hosted CI or final publication approval. No tests silently skipped; no capture production worker behavior changed by tester.

## Review correction: inherited harness dispatch flag

Canonical TEST-01 was independently reported by test-effectiveness reviewers gpt-6.1-sol/high and gpt-6-luna/medium. The tester verified and corrected it; this is verification/fix credit, not new discovery. The previous setup trusted inherited AL_CAPTURE_HARNESS_FIXTURE=1 as evidence that disposable PATH provisioning had already occurred. The guard is child dispatch intent and does not prove executable ownership.

The helper now always provisions fresh disposable executables and prepends their directory for each installer setup. It retains the child-only dispatch behavior, Windows hardlink/copy fallback, real installed native/wrapper checks and intentional blocked-probe overrides applied after prepareCaptureInstall.

Added TestCaptureHarnessFixtureOwnsInheritedDispatchGuard. For both Codex and OpenCode and inherited guard values 0 and 1, it first proves a hostile disposable external PATH executable records invocation and exits unsuccessfully. Actual installer execution must then use the test-owned harness, leave that hostile marker absent and report the controlled supported version. The hostile helper identifies its executable directory through os.Executable/filesystem identity, rather than argv[0], because actual PATH execution can retain a bare command name.

Command from backend: go test -race ./cmd/agentic-lens -run '^TestCaptureHarnessFixtureOwnsInheritedDispatchGuard$' -count=1 -v

- Red with the original setup early-return retained: FAIL, 1.239s. Both guard=0 controls passed; both guard=1 cases failed with `inherited dispatch guard bypassed disposable harness provisioning` because the hostile marker existed. Log: /tmp/al60-harness-guard-red.log.
- Green after removing the inherited-flag setup shortcut: PASS, 2.537s, all four cases. Log: /tmp/al60-harness-guard-green.log.
- git diff --check passed. No product source edits or full-suite reruns by tester during this correction.

The broader installer cancellation/descendant-process corrections and final integration verification remain with coder/parent; these narrow results do not establish their completion.
