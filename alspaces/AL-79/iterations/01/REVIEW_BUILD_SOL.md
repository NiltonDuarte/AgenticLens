# AL-79 C2 build and distribution integrity review

Verdict: **CHANGES_REQUESTED**, scoped to build/distribution integrity. Two introduced correctness defects were reproduced through the actual installed native dispatcher. This is not a whole-repository approval or merge authorization.

## Identity and revision

- Run: review_build_sol_c2; independent discovery, build/distribution specialist using the generic reviewer and shared review context. No sibling personas, reports or active consolidated REVIEW.md were consulted for discovery.
- Requested and acknowledged profile: gpt-6.1-sol / high. Actual execution model/effort source metadata was unavailable to this reviewer; no source-observed model, usage or cost is claimed.
- Baseline: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen candidate: AL79-C2; `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C2.diff`, SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de`. The staged source diff excluding alspaces independently matched before and after investigation.
- Actual probe executable: parent canonical C2 host build, copied into the isolated fixture; SHA-256 `897b372af09d959273956f26953c6e51b0c15f9be416f603159e8cdce0ff0450`. Observed identity: version `v3.0.0-rc.133+57a3503-1-g4df45967-dirty`, commit `4df45967`, build_date `2026-11-08T16:44:25Z`. Build provenance is the parent's canonical-build notification plus candidate fingerprint; the embedded commit alone does not identify dirty source.
- Requirements: AL-79 CONTRACT.md and approved iteration PLAN.md; AL-79 ticket and PO-20 F1-F6; applicable executable, process, fixture and evidence boundaries in TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md.

## Findings

### BUILD-01 — Native dispatcher breaks signal and process lifetime semantics

Severity: **6/10**. Classification: introduced process-control regression; required CLI compatibility.

Location: `backend/cmd/agentic-lens/dispatch.go:94` through `101`, with signal-exit conversion at `212` through `217`. The additional Windows active-dispatcher stage at `193` through `200` uses the same supervisor pattern; Windows runtime behavior was not exercised.

Expectation: the installed common entrypoint preserves selected CLI process termination and signal semantics. The approved plan explicitly includes signals alongside args/stdin/stdout/cwd/exit behavior.

Trigger and call path: installed POSIX wrapper execs the real native dispatcher; native selection launches the chosen executable with `exec.Command(...).Run()`. A process supervisor sends SIGTERM to the exposed wrapper/dispatcher PID. There is no signal forwarding, process replacement or child cleanup in this path.

Observed consequence: the dispatcher exits from SIGTERM while the selected command continues running. A canceled command can keep writing outputs or serving after its caller considers it stopped. This is separate from deliberately detached capture workers. A second trigger, SIGTERM delivered to the selected command itself, becomes normal wrapper exit 255 because ExitCode() is -1 for signal termination and is passed to os.Exit.

Evidence: `/tmp/al79-build-sol-c2-qxfnskuh/probe.py` installed the actual C2 application, copied the shipped wrapper and selected a compiled compatible native probe using shared binary_dir. The probe wrote its PID and blocked on a held stdin pipe; it did not implement selection policy. `/proc/<wrapper>/exe` identified the installed real native dispatcher, and the selected process initially named it as parent. Terminating only the dispatcher returned -15. A confirming repetition recorded the selected child as `S (sleeping)` and reparented to the container subreaper, proving live orphaning rather than merely a zombie. Raw results: `results.json` and `signal-parent-result.json` in that temporary directory. Selected-child SIGTERM produced wrapper result 255 and stderr `AgenticLens dispatcher: signal: terminated`, recorded in `signal-child-result.json`.

Test-oracle gap: `scripts/tests/test_capture_distribution.py:198` has a signal assertion, but that assertion runs before the actual native dispatcher is installed and explicitly waits for /proc/exe to equal the selected native. Its later real-dispatcher section checks args/stdin/cwd/relative I/O/exit only. Thus the green test does not refute this native-branch defect.

Required correction: preserve native process/signal lifetime and observable termination status through the actual installed dispatcher. Add the regression at that branch, retaining the existing bootstrap contract checks.

### BUILD-02 — Explicit user installation can replace the shared dispatcher with an incompatible native executable

Severity: **7/10**. Classification: introduced activation-validation defect affecting the shared machine installation.

Location: `backend/internal/adapters/capture/user_install.go:211` through `218`; unchanged `capture/entrypoint.go:76` implements the called four-byte-header check. Reachable CLI caller: `backend/cmd/agentic-lens/capture.go:120` through `136`.

Expectation: explicit provisioning must activate a usable host executable implementing the required dispatcher/capture protocols and reporting build identity. Existing owned working bytes must remain usable when a supplied replacement is invalid. Native executable selection already has these format/platform/interface checks; explicit activation bypasses them.

Trigger and call path: begin with a valid owned user installation; run the actual candidate `capture install-user --repo <fixture> --binary /usr/bin/true`. InstallUserCodex checks only the native magic header, hashes the supplied bytes and commits those bytes and the canonical owner through the normal installation transaction. Neither the CLI caller nor this transaction verifies dispatch/capture interfaces.

Observed consequence: installation exits zero without diagnostics and replaces the shared dispatcher and root ownership record with /usr/bin/true. Subsequent `dispatcher dispatch protocol` exits zero with empty output. The installed callbacks and common wrapper on this root can silently do no capture/CLI work, and all effective Codex homes sharing that root are affected. The valid transactional ownership record now attests to an unusable handler. This requires no unowned-file overwrite or race.

Evidence: `/tmp/al79-build-sol-c2-qxfnskuh/results.json`, fields wrong_native_install and post_wrong_native_protocol, records exit zero, dispatcher_replaced=true, owner_changed=true, and empty protocol output. The same real C2 binary first installed valid owned hooks successfully. All destinations were disposable. The script subsequently restored a valid owned dispatcher for the independent signal-status follow-up.

Test-oracle gap: user_install_test.go's basic and upgrade tests pass os.Executable(), the Go test runner, as an installer binary. They verify bytes/ownership transactions but do not establish that the activated handler implements native dispatcher/capture interfaces. The actual downloader performs a dispatch protocol probe before calling install-dispatcher, which protects that caller but does not protect the supported direct install-user/install-dispatcher CLI path.

Required correction: validate the replacement's host format, executable status, required native protocols and build identity before activation, preserving existing artifacts on rejection. Keep transactional and cross-home ownership guarantees. Add a direct CLI/adapter invalid-handler case that exercises the production admission boundary rather than accepting a test runner as a usable capture handler.

## Consequential hypotheses and dispositions

| Hypothesis | Disposition and basis |
| --- | --- |
| Native invocation loses signals or leaves a live child | Confirmed BUILD-01 by actual installed-dispatcher probes. |
| Explicit installation bypasses native-interface admission | Confirmed BUILD-02 by replacement of a valid owned handler with /usr/bin/true. |
| Development output layout disagrees with canonical build | Refuted: Makefile build emits agentic-lens-GOOS-GOARCH[.exe], matching NativePath and setup documentation. A direct-Go example using another name does not describe the canonical Makefile output. |
| Present invalid development output falls back to a cache pin | Refuted by Lstat/ValidateNativeExecutable path; native format/platform and bounded capture/dispatch/version probes reject it before fallback. Reused actual distribution evidence exercises this branch. |
| Missing development output invokes a network/latest fallback | Refuted by NativePath/CachedNativePath/dispatch trace: only exact selected cache path is returned; validation failure diagnoses installation. Network clients remain in explicit downloaders. |
| New preference redirects children of an already admitted worker | No defect found: LaunchWorker retains bytes; newCaptureWorkerFence records retained identity; collection/publication use fence.binary. Preference no longer retires that owner. Inspected the real Linux build-switch test, which gates actual subprocess identity before/after replacement. |
| One Codex home's stale hook receipt blocks a shared-root upgrade/removal | No new defect found: root canonical owner validates dispatcher bytes; per-home historical observations do not replace that authority. Reused actual A/B/C and two-home distribution evidence; independently traced root locking/journal, unchanged ownership guards and uninstall preservation. |
| Windows upgrade replaces a loaded stable dispatcher | Static path addresses this with retained active_native/active_hash and root activation, while preserving the stable launcher. ActiveDispatcher checks root-owned paths and hashes. No native Windows runtime proof is asserted. |
| Shipped wrapper copies diverge | Refuted by byte-for-byte cmp for both POSIX and PowerShell source/setup copies. |

## Verification, scope and limits

Executed independently:

- Read-only requirements/source/diff tracing across shell/PowerShell bootstrap, explicit downloaders, Go dispatch, config executable resolution and validation, user installation/transaction/recovery/migration, retained executable/worker/fence lifetime and relevant existing tests/callers.
- Source fingerprint and raw frozen diff SHA checks: matched C2 before and after probes.
- cmp of scripts/agentic-lens-wrapper.sh and .ps1 against their setup-distributed copies: both equal.
- Isolated actual C2 installation/selection/process-control and invalid replacement probe: two defects reproduced, with parent termination and selected-child termination separately exercised. The sole fixture compilation was a standalone native reporting/blocking probe, not another application build.

Reused with source-impact limits, not rerun as independent native evidence:

- DISTRIBUTION_EVIDENCE.md: affected POSIX capture suite 18 passed and installation suite 10 passed after canonical root-owner correction; actual native development/cache, A/B/C upgrade, cross-home, legacy manual, cwd/relative I/O and edited-owner behavior. Their passing status does not cover the demonstrated signal/invalid-handler admission gaps.
- TEST_EVIDENCE.md: affected Go adapter/CLI and dispatcher ownership/worker retention evidence for unchanged relevant paths. Full integrated make/race remained parent-owned at dispatch; this review did not rerun them.
- NATIVE_EVIDENCE.md: normal-trust Linux Codex root/source-proven-child automatic publication and prior-owner preservation on the callback-fixed native candidate. C2 changes preserve that callback routing; no fresh Codex native run, source observation or human trust action was performed by this reviewer. SessionEnd's retained-run visibility limitation remains explicit in that evidence.
- Windows cross-compilation supplied by the coordinator and algorithm/source examination support compilation/static claims only. Native Windows PowerShell/distribution/process/loaded-image behavior remains unexecuted and unavailable here. No release publication, package download or external delivery was performed.

All probe HOME, USERPROFILE, CODEX_HOME, XDG config/cache/state/data, install, capture/output and temporary/build-cache destinations were explicitly children of `/tmp/al79-build-sol-c2-qxfnskuh`. No authentication or private home source was read or modified. Live probe children were explicitly terminated in cleanup; report is the sole task-owned repository write. Initial staged product/test/documentation work and other agents' artifacts were preserved. Parent owns fixes, consolidation, full gates and delivery.
