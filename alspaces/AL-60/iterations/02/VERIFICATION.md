# AL-60 retained verification

Integrated tested revision: 057d410253aecd734551995ff1bedb38e94a1f19. Source/test/CI fingerprint against checkpoint 7b00cc1: 3251b53bc2bfdbd9c7e6ea47d5b8e793fc22f4f9e6103ee268d9fbc622a005db. Final review approval, the preceding reviewed fingerprint and mechanical explicit-defer carry-forward rationale are retained in [SUMMARY.md](../../SUMMARY.md). [Compact findings](REVIEW_SUMMARY.json) retain discovery credit and dispositions; this document consolidates the former implementation/testing notes.

Coder and tester verification/fix actors used inherited gpt-6.1-sol/high; the parent integrator used gpt-6.1-sol/high. These actors do not acquire independent discovery credit by reproducing supplied issues. Source/application acceptance is unchanged by the pre-merge artifact cleanup.

## Deterministic confirmation and regression oracles

Commands below run from `backend/` after canonical web assets exist. Each retained red preceded its production/fixture correction. Initial fixture-development mistakes (unreached helper, unscoped health path and deprecated false report key) were corrected before the cited valid proofs; they are not claimed as diagnosed failures.

| Boundary | Meaningful oracle and outcome | Retained output |
| --- | --- | --- |
| Optional probe consumes required validation | Disposable blocked harness writes a reached marker. Baseline failed at required validation after 2.01s; repaired installation records unavailable optional inspection and verifies real installed native/wrapper health. Final race green 5.343s. | [Red](evidence/INSTALL_RED.txt), [green](evidence/INSTALL_GREEN.txt) |
| ARCH-01 inherited output pipes | Actual probe, finite two-second descendant, 100ms allowance: baseline waited 2.0026s/2.0043s for version/features. Corrected group/job ownership and draining return within the guard. Unix timing is not a direct descendant-death/reachability oracle; TEST-02 remains accepted coverage debt. | [Adapter red](evidence/CORRECTIONS_ADAPTER_RED.txt), [green](evidence/CORRECTIONS_ADAPTER_GREEN.txt) |
| ARCH-02 SIGTERM | Wait for a real blocked probe PID, signal only the CLI PID, require non-success and probe death. Baseline left the probe alive; test cleanup killed it. Repaired guard passes under race. | [Signal red](evidence/CORRECTIONS_SIGTERM_RED.txt), [green](evidence/CORRECTIONS_SIGTERM_GREEN.txt) |
| DATA-01 final write | Real transaction writer applies the final target then cancels synchronously. Both baseline source families returned nil; repaired tests require context cause, exact rollback/journal disposition and uncanceled retry. | [Adapter red](evidence/CORRECTIONS_ADAPTER_RED.txt), [green](evidence/CORRECTIONS_ADAPTER_GREEN.txt) |
| TEST-01 inherited fixture guard | Codex/OpenCode × inherited flag 0/1. A positive hostile PATH control records invocation; actual installation must leave its marker absent and report controlled version. Guard=1 failed red; all four cases pass green, 2.537s. | [Guard red](evidence/HARNESS_GUARD_RED.txt), [green](evidence/HARNESS_GUARD_GREEN.txt) |
| DATA-02 complete recovery | Fully applied inactive/removal journals with canceled/expired contexts for both sources. Baseline accepted all four; repaired recovery retains identical journal/after-files, exposes cause, then accepts/removes journal on uncanceled retry. Green includes final-write guards and eight existing prefix recovery cases, 1.561s. | [Recovery red](evidence/RECOVERY_RED.txt), [green](evidence/RECOVERY_GREEN.txt) |

Focused commands associated with these outputs:

```sh
go test ./cmd/agentic-lens -run '^TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation$' -count=1
go test -race ./cmd/agentic-lens -run '^TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation$' -count=1
go test -race ./internal/adapters/capture -run 'TestInstallOptionalProbeBoundsInheritedPipes|TestInstallerFinalWriteCancellationRollsBack|TestInstallNativeValidationDistinguishesExecutionAndIdentity|TestInstallCallerCancellationDuringOptionalProbe' -count=1
go test -race ./cmd/agentic-lens -run 'TestCaptureInstallSIGTERMCancelsProbeChild|TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation' -count=1
go test -race ./cmd/agentic-lens -run '^TestCaptureHarnessFixtureOwnsInheritedDispatchGuard$' -count=1 -v
go test ./internal/adapters/capture -run '^TestInstallerCompletedRecoveryRetainsExpiredAcceptance$' -count=1 -v
go test -race ./internal/adapters/capture -run 'TestInstallerCompletedRecoveryRetainsExpiredAcceptance|TestInstallerFinalWriteCancellationRollsBack|TestInstallerDurablePrefixRecoversHookPluginAndManifestCheckpoints' -count=1 -v
```

## Fixture fidelity and preserved behavior

CLI fixtures always provision disposable harness executables, using hardlink/copy fallback on Windows. The child-dispatch flag is not setup/ownership evidence; intentional blocked-probe overrides apply after preparation. Helpers use executable/filesystem identity where argv[0] can be bare. Unix adapter fixtures use disposable scripts. Native selection, pin/protocol validation and wrapper verification remain real.

The native diagnostic table checks matching output, observed mismatch, exit 19 with retained ExitError, missing-executable filesystem cause, cancellation and deadline causes (race pass 1.062s). The API cancellation case synchronizes on actual optional-probe TCP contact and requires cancellation plus no committed manifest.

Controlled manifest success then exit-42 changes only HarnessVersion and Capability, comparing every remaining identity field. Full-byte stability and pin/wrapper guards remain. This reproduces the optional-field mechanism, not the specific field behind the historical unretained failure. The coordinated adapter diagnostic/cancellation/manifest run passed under race in 1.781s:

```sh
go test -race ./internal/adapters/capture -run 'TestInstallCallerCancellationDuringOptionalProbe|TestInstallNativeValidationDistinguishesExecutionAndIdentity|TestInstallationManifest(StableAndPinSwitchRequiresRefresh|RecordsChangedOptionalProbeOnly)$' -count=1 -v
```

Successful publication cases use a distinct bounded 10000ms lifetime/5000ms pass policy; global defaults and worker production code are unchanged. Short failure-feedback budgets remain until its later retained-fact success phase. The intentional publication-lock case uses 1000ms/100ms, preserving accepted bytes, pending generation/unchanged ACK and released worker ownership, then requires real updated tokens, ACK and idle release after unlocking. Shared-store/project/source multiplicity, multi-root periodic refresh and invocation freshness assertions remain. Six coordinated CLI cases passed under race in 10.430s:

```sh
go test -race ./cmd/agentic-lens -run 'TestCapture(PublicationBudgetExpiryRetainsPendingAndRecovers|AutomaticSharedStoreAndTicketStayInExactProjects|NativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate|LaterAgentChannelReportsBoundPublicationFailure|CLIOpenCodePluginOwnershipAndMissingHandler|InstallerUsesSharedLocalAndExplicitStatePolicy)$' -count=1 -v
```

## Integration, failed gates and final acceptance

Historical failures are retained verbatim: [initial make](../01/evidence/MAKE_INITIAL.txt), [bounded make](../01/evidence/MAKE_BOUNDED.txt), and [integrated vet failure](evidence/MAKE_INTEGRATED_VET_FAILURE.txt). The [historical diagnosis](../01/CAPTURE_DIAGNOSIS.md) distinguishes local evidence from reported AL-61/MGT-2 baseline failures.

Vet could not prove cleanup through a closure capturing reassigned cancel functions. Explicit first/second defers preserve contexts, ordering and final cancellation; the extra deferred first cancel is idempotent because it already ran before probing. Standalone vet exited 0 with no stdout; no empty log is needed as evidence. Focused [phase race guards](evidence/PHASE_GUARDS.txt) passed 1.666s and the [CLI race guard](evidence/PHASE_CLI_GUARD.txt) passed 6.286s.

Full `make` on 057d410 passed with four-CPU affinity, GOMAXPROCS=2 and GOFLAGS=-p=1: install, vet, frontend typecheck/lint, script checks, every Go package with race detection, 786 frontend tests in 36 files, canonical frontend build/sync and five platform binaries. [Verbatim full gate](evidence/MAKE_VERIFIED.txt). Main integration added AL-68 workspace prose only. No test silently skipped and no assertion was weakened.

Parent compared every generated/frontend embed file byte-for-byte, then served the fresh host binary against a disposable workspace. HTTP index/referenced assets matched the embed tree and included search/no-match copy. The server was terminated; no Node runtime served production assets. Ignored build assets remain outside Git. Search source and AL-55 ledger remain preserved.

## Actual Windows runtime

Cross-compiled capture package tests and CLI passed. The parent used WSLInterop to execute the Windows PE binary on the Windows host, not merely infer runtime behavior from compilation. Reproduction uses a temporary output executable:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o <temporary-executable.exe> ./internal/adapters/capture
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o <temporary-cli.exe> ./cmd/agentic-lens
<temporary-executable.exe> -test.run '^(TestInstallWindowsInspectionOwnsLauncherTree|TestInstallerCompletedRecoveryRetainsExpiredAcceptance)$' -test.v -test.count=1 -test.timeout=60s
```

The integrated test passed launcher-exit/cancel ownership (0.37s) and both sources' canceled/deadline recovery cases (0.31s). Its [retained output](evidence/WINDOWS_INTEGRATED.txt) supersedes earlier Windows success snapshots. The test synchronizes a stable descendant handle before launcher release and requires actual termination. Suspended launch/job assignment precedes child execution; job close/cancellation owns descendants and abort paths reap the direct child. Snapshot/thread/process/job handles close explicitly. The Windows CI step remains intact.

## Limits and cleanup verification

Windows race instrumentation, real external-harness callback compatibility and browser geometry were not exercised locally. Hosted CI state belongs to its exact PR head: Windows, Backend, Frontend and Build Binary/Version all passed on pre-cleanup 2a99b422bd24c674e3c4d5e787738b5238c2e3d7 (GitHub run 41702443672). Refreshed cleanup-head checks are handed to MGT-2 separately. The accepted TEST-02 Unix descendant-oracle recommendation remains in the compact findings and summary with its owner; no current process leak is claimed or concealed.

Pre-merge cleanup changes artifacts only. JSON/finding-credit preservation, retained evidence hashes, local references and source equality against tested 057d410 are checked before commit/push. Existing application review, full make and runtime results carry forward because application/test/CI bytes are unchanged; no fresh product review or full-suite rerun is necessary. Merge remains held for MGT-2's cleanup handoff and refreshed exact-head checks.
