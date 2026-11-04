# REV-6: measured test performance and verification boundaries

Status: active review; measurements and independent reviews are still being reconciled. No implementation is included.

## Scope and method

Source revision: a84adbe5be46c2ef52ede0987a417562715d38e1, matching origin/main at review start and including AL-62 merged PR #109 (86bea597). Closed PR #103 is not delivered work. AL-74 remains the owner of its separate reliability/oracle/compatibility remediation.

One measurement owner runs commands sequentially. Existing shared Go build/module caches are retained; test-result reuse is disabled with -count. Initial compile-cache contents are unknown, so initial runs are not called cold. No global cache is cleared. Frontend dependency installation and canonical embedded-asset preparation are measured separately. Parent/subtest intervals overlap and are never summed into suite wall time. A -count=3 cohort supplies three test repetitions within one package process; it does not repeat TestMain three times. A no-match command still executes TestMain.

Local environment: WSL2 Linux x86_64 on Intel i5-11400F, 12 logical CPUs, approximately 32 GiB memory, ext4 storage. Go 1.26.0, Node 24.22.1, npm 11.19.1, Git 2.43.0, CGO enabled. GOMAXPROCS/GOFLAGS/GORACE were unset at launch, and no -p/-parallel overrides were supplied; package/test parallelism follows the toolchain default. Fixture-local GORACE settings remain active. CI instead configures Go 1.23, Node 20, Ubuntu/Windows hosted runners. Root cgroup v2 limits were unavailable at the probed paths. Host contention was observed from unrelated OpenCode workloads; REV-6 neither stopped them nor ran competing profiling agents. Ratios are descriptive, not isolated causal estimates or hosted-runner predictions.

GNU time records command wall/user/system seconds and maximum RSS in KiB. CPU seconds include child work as reported by the tool; maximum RSS is not aggregate simultaneous process-tree memory, a cloud bill or a harness inference cost. Exact command metadata, sanitized environment and timings will be retained in MEASUREMENTS.json. Raw JSON test streams and logs stay outside Git.

## Boundary classification and preserved coverage

| Existing owner | Actual boundary | Recommended lane and limit |
| --- | --- | --- |
| backend/internal/domain and domain/migration tests | Pure decisions, migration values, calculations, codecs | Initial fast Go gate; not native/storage proof. |
| web/src/domain tests | Pure TypeScript transformations; current default config uses jsdom | Fast frontend gate; validate Node environment selector separately. Does not replace DOM/component tests. |
| Capture adapter TestCodexInstalledHookWireBudgets | Pure generated hook JSON, including Windows fields | Unit candidate in a mixed adapter package. Windows wire values do not exercise PowerShell. No broad adapter package is classified pure. |
| Capture hooks/store/project/diagnostics/reconcile/relocate tests | Real files, SQLite/WAL, Git, namespace reservation and concurrency; some reexec processes | Integration, even if a particular test is fast. Preserve transactional and physical-state fixtures. |
| CLI TestMain and native capture tests | Once-per-process go build, real application processes, wrappers, source fixtures, signals and detached workers | Native integration. A -run filter still pays the package build; outer -race does not automatically instrument nested go build. |
| Capture adapter installation fixture | Reexecuted test binary answering protocol/version plus real wrapper/filesystem | Boundary integration, with explicit protocol stub; different fidelity from the built CLI. |
| CLI disposable codex/opencode probes | Reexecuted test binary for inspection protocol; real selected application/wrapper remains | Harness-inspection fixture, not a real external harness lifecycle. Preserve real-harness compatibility separately. |
| scripts/tests distribution/installation/release checks | Real Git/shell/installer/build and workflow block/guard assertions | Boundary checks, not all cheap static checks. Text/block assertions do not prove hosted publishing. |
| Windows watchdog/launcher/transport suites | Native Windows process-tree, PowerShell and plugin transport | Hosted Windows integration after fast gates; Linux results do not certify them. |
| scripts/opencode-compat.sh | Actual OpenCode v1/v2 runtimes with local provider | Explicit external compatibility lane, not a unit test; AL-74 owns historical fixture/fixed-port coupling. Not executed in REV-6. |

Independent proof to preserve in follow-ups:

- Pure binding/current-installation rules: domain capture_test.go and capture_diagnostics_test.go. Native verification additionally owns producer/selector identity through TestCaptureVerifyNativeExchangeBelongsToCurrentInstallationAndSelector.
- Pending allocation: TestCaptureInstallReportsPendingAllocationUntilCommittedManifest checks repeated dry-runs do not create a registry, real install commits a nonplaceholder manifest, and reopening/preview preserves its scope.
- Generation ACK: adapter TestPublicationGenerationAcknowledgesOnlyObservedWork and TestPublicationControlEmptyRegistryAndTicketRotation exercise real persisted requests. Native crash, late-ticket and release-switch tests separately prove caller ordering and recovery. Lease availability is weaker than completion of every inherited fixture writer.
- Failed publication: TestCaptureSourceAcceptanceSurvivesFailedPublicationWithoutClaimingSuccess preserves previous artifact bytes while retaining accepted source/checkpoint state, then verifies recovery. WAL status keeps the writer open while a separate native process reads committed data and explicit continuation.
- Fencing/relocation: keep already-open old writers, destination namespace conflict, peer state and exact successor installation/state checks. A reopened-only/in-memory fixture is weaker.
- Automatic capture: ordinary captureCLI disables automatic publication when no policy exists. Tests with explicit enabled policy and real native workers remain independent; they cannot be replaced by ordinary installer/binding success.
- Portable Git history: preserve copied/merged committed ledgers and unavailable other-source owners without requiring the original private capture database.

No deletion is authorized. A low-priority optional duplicate ACK block has an existing recipient in TestPublicationControlEmptyRegistryAndTicketRotation; project_test.go's separate two-project namespace-isolation assertions must remain. Any later consolidation needs explicit assertion transfer and regression proof, and currently has no demonstrated material saving.

## Recommended execution dependencies

Proposed implementation follow-up, not an existing Make target:

1. Run the complete Go pure-core gate from backend with `go test -count=1 ./internal/domain/...`. It does not need embedded assets. Run the frontend domain gate after npm ci; use `npm --prefix web test -- src/domain --environment=node` only if the measured validation below passes. Typecheck/lint can join this frontend gate. These are explicit initial fast subsets, not a claim that all unit tests already reside there.
2. Require success of both fast gates before starting storage/native integration, the full frontend DOM/component suite, Python installation/distribution/release-block checks, or Windows native transport/watchdog integration. Hosted jobs need explicit `needs` on both fast jobs; cancellation of superseded PR runs remains useful but does not establish ordering.
3. Prepare canonical web assets once per consuming job before Go packages that embed them. Preserve the current full `go test -race ./...` and `npm test` as later coverage owners initially, accepting cheap pure-test repetition rather than introducing incomplete test-name filters. Keep Python check suites and Windows jobs. A later exclusion-based split must reconcile the complete package/test inventory and new-test routing before replacing the full guard.
4. Require all applicable boundary/platform jobs before native build verification or release actions. Existing build-binary already needs backend-check/frontend-check; extend prerequisites for Windows if claiming all-platform success. Never move release publishing before its checks. Real OpenCode compatibility is an explicit separately provisioned lane, with AL-74 fixture corrections and honest unavailable-environment reporting.

Local use can run the same stages in sequence and stop on the first failing command. Dependency installation, toolchain setup and cache restoration are separate startup costs. No hosted time or billing saving has been measured, and no timeout is changed by this review.

## Prioritized bounded follow-ups

| Priority | Proposed responsibility | Required preservation and acceptance |
| --- | --- | --- |
| P1 | Introduce explicit fast gate and CI dependencies | Use the measured full domain selectors; boundary/native/Windows work starts only after both fast jobs pass. Keep all existing later suites, race, canonical assets and platform checks. A deliberate pure failure must prevent integration launch. |
| P1 | Attribute diagnostic-history fixture versus backfill/read cost before optimizing | Keep real SQLite, historical owner rotation, bounded HeadLookups and indexed EXPLAIN query-plan assertions. Measure setup, backfill and query independently; separate history-length and owner-width scale axes only with equivalent regression proof. No automatic reduction of the 256-case or claim that setup is the entire cost. |
| P2 | Evaluate sharing immutable versioned native fixtures within a package run | Release-switch scenarios build the same two versions independently. Reuse only current-source executables keyed by revision, version metadata, target, flags and embedded assets; isolate every mutable repository/registry and keep both pending/acknowledged handoff scenarios. Measure net saving before adding infrastructure. |
| P2 | Review long quiet-window tests as lifecycle contracts | Live-disable includes three modes with 300 ms settling plus 2100 ms no-write windows: at least 7.2 seconds of deliberate waits across those cases. Keep beyond-owner-lifetime no-write/lease-release proof. A handshake or controlled clock must prove that same condition before replacing waits; merely lowering constants weakens coverage. Git watchdog and optional-probe deadline tests remain true integration. |
| Existing AL-74 | Listener/readiness, configuration/Git fixture isolation, oracle repairs, compatibility fixture/fixed-port decoupling | Reuse AL-74 scope and source findings; no duplicate implementation. Do not infer current remediation from an old report. |
| P3, optional | Consolidate only the duplicated method-level ACK block | Existing publication-control test receives the same captured-generation assertion; retain project namespace isolation and native worker caller proof. No material runtime saving demonstrated, so it is not a performance priority. |

These are recommended implementation slices, not newly allocated or approved implementation tickets. No unrelated concrete application defect has yet been established by this audit requiring a new FIX ticket.

## Measured costs and attribution

All commands below exited 0. Full capture/CLI Go cohorts reported zero runtime skips. Frontend full execution passed 824/824 cases; the domain selector passed the same 158 cases in jsdom and each Node run, with no pending cases. Expanded cases are execution evidence, not a coverage quota. Package elapsed includes package setup; wall also includes compilation/driver overhead.

| Cohort | Command wall s | Package elapsed s | User + system CPU s | Max RSS MiB |
| --- | ---: | --- | ---: | ---: |
| capture-norace | 48.33 | 40.139 | 29.38 | 179.9 |
| capture-race | 112.53 | 110.342 | 75.00 | 294.9 |
| cli-norace | 105.86 | 105.092 | 75.79 | 248.4 |
| cli-race | 115.25 | 113.764 | 83.76 | 307.5 |
| domain-repeat | 0.48 | 0.088, 0.134 | 1.31 | 88.5 |
| domain-race | 5.35 | 1.238, 1.318 | 7.75 | 123.4 |
| frontend-all | 24.19 | not separately collected | 179.01 | 469.4 |
| frontend-domain-jsdom | 2.24 | not separately collected | 14.85 | 159.7 |
| frontend-domain-node-1 | 1.85 | not separately collected | 10.15 | 154.5 |
| frontend-domain-node-2 | 2.17 | not separately collected | 11.85 | 156.4 |
| frontend-domain-node-3 | 1.85 | not separately collected | 10.44 | 159.4 |

The pure Go command was also executed in three separate warm processes with -count=1 (see dataset); these differ from the three repetitions inside domain-repeat. Canonical assets took 12.96 s; npm ci took 21.56 s. These startup/preparation costs are excluded from test-body claims.

Warm no-match adapter command: 0.53 s without race, 1.74 s with race. Warm no-match CLI command: 1.49 s without race, 2.76 s with race. The first CLI no-match run cost 4.94 s and first race no-match cost 24.24 s under initially unknown cache state. Neither is a cold-build benchmark. A fast test-name filter therefore does not remove native TestMain setup, but this setup is not the entire slow-suite cost.

| Selected top-level test | Non-race repetitions s | Median s | Race repetitions s | Median s |
| --- | --- | ---: | --- | ---: |
| TestGitCheckAdvisoryFailuresOutputAndDeadline | 4.14, 4.08, 4.1 | 4.1 | not repeated | not repeated |
| TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest | 2.82, 3.42, 3.87 | 3.42 | not repeated | not repeated |
| TestDiagnosticHeadLookupWorkIndependentOfRotatingOwnerHistory | 2.2, 3.8, 3.75 | 3.75 | 12.35, 9.62, 15.32 | 12.35 |
| TestCaptureLiveDisableStopsWritesAndManualRecoverySurvives | 8.92, 8.68, 8.85 | 8.85 | not repeated | not repeated |
| TestCaptureWorkerReleaseSwitchFinishesCapturedPassAndHandsOffPending | 5.71, 6.69, 5.53 | 5.71 | not repeated | not repeated |
| TestCaptureGenerationOverlappingPublishersCannotRegressFinalSnapshot | 4.55, 4.51, 4.5 | 4.51 | not repeated | not repeated |

The reported near-43-second duration cannot be uniquely identified from the original report. This run found a 40.139 s non-race capture package and a 38.65 s single diagnostic-history test in the first race package cohort. Its subsequent race repetitions were 9.62–15.32 s; a single number is not a stable test SLA. The CLI release-switch case likewise varied from 16 s initially to 5.53–6.69 s after repeated builds. Cache priming and observed host contention are confounders.

The diagnostic fixture combines n historical rows with n owners per row at n=32/128/256: 82,944 owner occurrences across those cases. It exercises real legacy insertion, index backfill, bounded current-head reads and EXPLAIN index-seek assertions. A separate race CPU profile took 9.72 s with 8.34 s sampled CPU. BackfillDiagnosticIndex had 3.97 s cumulative samples (47.60%); top flat symbols included __tsan_read (0.90 s), runtime.checkptrBase (0.90 s) and racecall (0.87 s). Cumulative stacks overlap. This supports investigating backfill/SQLite and instrumentation as well as fixture creation; it does not establish that the final indexed lookup itself took tens of seconds or that setup alone explains it. A broad production optimization needs its own performance contract.

Git watchdog repeats stayed at 4.08–4.14 s, consistent with intentional deadline/escalation behavior. Live-disable repeats stayed at 8.68–8.92 s and include at least 7.2 s of explicit quiet-window sleeps across three modes. Cancellation fixtures using hour-long sleeping children terminate via a reached handshake and cancellation; those literals do not imply hour-long test cost. No timeout inflation, arbitrary shorter sleeps or boundary removal is justified.

## Provisional budgets and limits

For this warm local environment, investigate regressions if the Go pure command exceeds 2 s or the frontend domain Node command exceeds 5 s; allow a combined local test-execution feedback budget of 10 s after dependencies are ready. These are proposed review alerts with headroom, not new test failures/timeouts or a hosted-runner SLA. The current Go command is subsecond and frontend Node repetitions are about 2 s; the modest Node/jsdom difference is not grounds for a claimed dramatic frontend speedup. The larger separation is full frontend 24.19 s versus the explicit pure subset, which does not replace the other 666 cases.

Track setup, race and non-race integration separately. Capture package first observations were 48.33 s command wall without race and 112.53 s with race; CLI was 105.86/115.25 s. Use these as baseline observations, not hard ceilings from one noisy sample. A practical initial warm local integration planning allowance is about 5 minutes for these two race packages run sequentially, with overruns triggering diagnosis rather than timeout expansion. It does not budget the rest of make, downloads, hosted startup or Windows. Hosted budgets require measurements on the configured Go 1.23/Node 20 runners and preserved platform fixtures before adoption.

No cold-cache, hosted-runner, Windows-native or real OpenCode compatibility timings were collected. No private source/session data was used as a test fixture. The review does not prove mutation equivalence, whole-repository correctness, complete test-body audit or a monetary CI saving. Native child go build commands were not passed -race, so the outer race suite alone does not certify race-instrumented application subprocesses. Full later suites remain necessary.

Evidence: [MEASUREMENTS.json](MEASUREMENTS.json) contains commands, resource basis, repeats, package events, top test timings and the CPU-profile summary; [INVENTORY.md](INVENTORY.md) records syntax-derived file inventory and build constraints. Independent review/dispositions and final integrated verification are recorded in the ticket SUMMARY.md and compact REVIEW_SUMMARY.json at closure.
