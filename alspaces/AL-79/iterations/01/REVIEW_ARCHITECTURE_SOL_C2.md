# AL-79 architecture re-review: AL79-C2

Verdict: **CHANGES_REQUESTED**, scoped to architecture, common dispatch compatibility and installation ownership. The three supplied C1 findings are fixed on the inspected candidate. Three additional introduced defects remain below. This is not consolidated approval or merge authorization.

## Identity and revision

- Run: `review_arch_sol_c2`; phase: revision-aware re-review with additional independent investigation; persona: architecture/maintainability.
- Requested model/effort: `gpt-6.1-sol` / `high`, acknowledged as requested. Exact runtime model/effort metadata is unavailable in this child context; no observed identity, usage or cost is asserted.
- Baseline/HEAD: `4df45967c801824500ac2ea39a8f52fa9cbdcf66`.
- Frozen candidate: `AL79-C2`; staged binary diff excluding `alspaces`, SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de`, matching `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C2.json` and `CANDIDATE_C2.diff`. Recomputed unchanged at close; no unstaged product/test changes were observed.
- Requirements: AL-79 contract and approved iteration 01 plan, AL-79 ticket, PO-20 F1-F6, canonical FCIS, compatibility and portable-ledger rules. The supplied C1 architecture report was read for fix verification. No sibling persona or sibling report was loaded. Product/test source remained read-only; only this assigned report and disposable `/tmp` probes were written.

## Additional findings

### ARCH-SOL-4: Native manual dispatch loses ownership of command cancellation

Severity **6/10**; introduced correctness regression, independently discovered during C2 re-review. It was already present in C1 and is distinct from verification of the three supplied findings.

Location: `backend/cmd/agentic-lens/dispatch.go:94-101`; caller: `scripts/agentic-lens-wrapper.sh:11` and its setup copy. The Windows activation forwarding stage at `dispatch.go:193-197` also creates a process boundary without cancellation forwarding, but the executed counterexample is POSIX.

Trigger: explicitly install a native dispatcher, invoke a long-running selected command through the common wrapper, then cancel that wrapper PID with SIGTERM. The shell execs the dispatcher, but the dispatcher starts the selected executable with `exec.Command(...).Run()` and neither replaces itself nor forwards termination. The dispatcher exits while the selected command continues as an orphan. Automation can report cancellation while a server or generator continues running and retaining resources or writing output.

Supported expectation: the approved common entrypoint must preserve process/exit behavior; the plan explicitly retains argument/stdin/stdout/cwd/exit semantics. The existing distribution test also explicitly preserves exec/signal behavior. Its signal oracle still exercises only the legacy shell-bootstrap branch; the newly added native-dispatch portion checks cwd/relative output and never signals that branch.

Evidence: `/tmp/al79-architecture-sol-c2/signal_probe.py` installs the actual candidate through the shipped downloader, selects a compiled native development reporting fixture and holds its stdin open. It identifies the final command by executable plus `fixture-wait` argv, excluding the short-lived interface probes. Sending SIGTERM to dispatcher PID `1311908` yielded exit `-15`; target PID `1311946` remained sleeping and was reparented to PID `2339`. The probe explicitly killed that fixture child during cleanup. Raw result: `/tmp/al79-architecture-sol-c2/signal_probe.log`.

Required correction: make the process boundary preserve cancellation and child lifetime, with platform-appropriate exec or explicit forwarding/reaping. Extend the existing signal regression through the actual installed native-dispatch branch. Preserve automatic callback routing and Windows stable-launcher constraints.

### ARCH-SOL-5: PowerShell native-dispatch fast path bypasses the existing argument/handle transport

Severity **6/10**; introduced compatibility defect, independently identified during C2 re-review. This source path was already present in C1. Evidence is a precise source-level counterexample; native Windows execution remains unavailable.

Locations: `scripts/agentic-lens-wrapper.ps1:7-9` and `.agents/skills/agentic-lens-setup/scripts/agentic-lens-wrapper.ps1:7-9`; relevant unchanged implementation: wrapper lines 207-225; downstream consumer: `dispatchNative` forwards the argv it actually received.

Trigger: on supported Windows PowerShell 5.1, install a native dispatcher, then invoke the common wrapper with an intentional empty argument or literal embedded quotes. The new early `& $dispatcherPath dispatch @args` uses PowerShell's native argument conversion. It bypasses the existing CRT quoting and `ProcessStartInfo` inherited-handle transport, whose explicit purpose is to preserve empty arguments and avoid native pipeline re-encoding. Once an empty value has disappeared before reaching Go, the native resolver cannot recover it. A call containing `--session-id ''` therefore no longer preserves the same argv after native adoption. The early return also bypasses the existing 1.5-second `--al-offline capture check` wait/termination bound; native dispatch consumes that flag without installing an equivalent bound.

Supported expectation: common CLI transport and supported consumers remain compatible; existing Windows distribution coverage explicitly asserts empty values, quotes, exact stdin and bounded offline advisory behavior. The C2 native Windows cwd scenario supplies no empty argument, while the existing argument/timeout scenarios initially install a legacy fixture without native dispatch. Thus their oracles do not cover the new fast path. All Windows scenarios are additionally unexecuted in this environment.

Required correction: use one lossless native-process launch mechanism for both selected-executable and dispatcher execution, retaining existing bounded advisory behavior. Exercise the existing empty/quote/stream/timeout cases after actual native-dispatch installation on Windows PowerShell. This finding does not claim an executed Windows failure.

### ARCH-SOL-6: Windows repair can publish a launcher whose bytes disagree with its ownership hash

Severity **5/10**; introduced by the C2 stable Windows launcher correction, independently discovered during C2 re-review.

Locations: `backend/internal/adapters/capture/user_install.go:195`, `:231-232`, `:252-257`, `:280`; consumer: `backend/internal/adapters/capture/dispatcher_activation.go:48-55`.

Trigger: install Windows build A, upgrade to B while preserving stable launcher A, then lose/delete only `dispatcher.exe` and explicitly reinstall B from its pinned native binary. The root ownership record and hooks still exist. `next.DispatcherHash` starts with the old stable hash A; the Windows branch changes it only when `owner.Hash` is empty. Because the dispatcher file is absent, the later write recreates it using B's bytes, but the transaction writes hash A into the root record, receipt and checksum. Installation returns success. The next launch calls `ActiveDispatcher`, compares recreated B against recorded A and fails with `stable native dispatcher was modified; explicit repair required`. Repeating explicit setup with B cannot resolve that inconsistent ownership state.

Supported expectation: explicit installation/repair must preserve coherent machine-cache ownership, and a successful repair must produce a usable dispatcher. Keeping a present loaded Windows launcher immutable is valid; a missing launcher must either be recreated with its actual identity or yield a clear error before mutation. The normal A-to-B test covers a present launcher and misses this branch.

Evidence: a disposable Go overlay selects the Windows branches of the unchanged candidate `user_install.go`, `retained.go` and `dispatcher_activation.go`. All other adapter code, filesystem transaction behavior and validation remain actual candidate code. A host-native ELF fixture is used, so this proves the platform-branch ownership algorithm, not native PE, Windows sharing, path or PowerShell execution. The present-launcher A-to-B control successfully activates B. Removing only the stable launcher and reinstalling B succeeds but produces stable SHA `0301637a09fe93822f58705f409e4e9490fbb6f0c36bd86211cacfb986297846` against recorded SHA `e8dd9691b142ace5d0b3b972f53ae0ed60694ff9cf55c5ecdb6bf09ef1590a43`. The actual activation function then rejects it. Raw failing test: `/tmp/al79-architecture-sol-c2/windows-branch-probe.log`; overlay generator and test: `prepare_windows_branch_probe.py`, `windows_branch_probe_test.go` and `windows-branch-overlay.json` in that directory. No shared source was modified.

Required correction: derive the stable ownership hash from the bytes actually written when the launcher is recreated, while retaining the old hash for an existing unchanged stable launcher. Add the missing-launcher repair case alongside the A-to-B ownership regression and retain the normal control.

## Supplied C1 finding verification

- **ARCH-SOL-1, conflicting upgrade ownership: fixed in the inspected POSIX flow.** The downloader delegates dispatcher replacement to `InstallUserCodex`; root `dispatcher-installation.json`, dispatcher bytes, checksum and the current home receipt are advanced by one native transaction under the root installation lease. Another home's historical receipt cannot override that root authority. Independently rerunning the original actual installer A-to-B counterexample now succeeds, receipt SHA equals dispatcher SHA and subsequent uninstall succeeds. Reused C2 distribution evidence additionally covers another effective home, modified dispatcher/hooks and preservation of foreign hooks. ARCH-SOL-6 concerns the additional Windows missing-file branch, not a recurrence of the original POSIX upgrade result.
- **ARCH-SOL-2, legacy manual pins rejected after native adoption: fixed.** Cached manual selection now validates ordinary native release identity, while automatic callbacks and development selections retain the stricter capture/dispatch probes. The original compiled legacy fixture installs and its manual wrapper invocation succeeds with a real dispatcher already installed. No capture-local file or project installation receipt was added to hide the compatibility case.
- **ARCH-SOL-3, manual cwd reset: fixed.** `command.Dir` is now assigned only for automatic dispatch. The original nested-cwd probe reports the exact invocation directory; C2 distribution evidence additionally covers relative output and explicit `--al-project`. Automatic callbacks still route with validated project/source context.

Executed fix-verification probe: `PYTHONDONTWRITEBYTECODE=1 python3 /tmp/al79-architecture-sol-c2/supplied_fixes_probe.py`, exit 0; raw results in `supplied_fixes_probe.log`. This is verification of supplied issues, not new discovery credit.

## Other consequential hypotheses assessed

- **Resolver divergence across CLI/callback/development selection:** inspected common `ResolveEntrypoint`/`NativePath`, selected-local configuration and automatic dispatcher callers. Exact shared pins remain required; present invalid development output cannot silently fall back; runtime selection performs no download. The manual/automatic interface distinction fixes a real compatibility boundary without introducing a separate executable-choice implementation. Wrapper fallback remains necessary for older targets without a dispatcher, though its PowerShell process transport must also be used by the new branch as described above.
- **Preferred-build changes retarget existing worker children:** not substantiated. `LaunchWorker` retains bytes, `captureWorker` uses its admitted fence path and publication children execute that retained path. Fence checks preserve source/state/disablement boundaries without retiring solely for preference/installation changes. Reused the unchanged real-process retention evidence. No automatic retention cleanup is introduced.
- **Root/home transaction and legacy migration:** inspected root authority, lease ordering, recovery journal, reverse rollback, unchanged-byte guards, per-home receipts and explicit legacy group ownership. The sequential POSIX upgrade/removal defect is repaired; unknown/modified groups remain refused, legacy migration preserves state and unrelated callbacks, and rollback refuses unexpected bytes. No generic crash/concurrency proof is claimed: this review did not execute process-interruption or multiple-install-root recovery races. The concrete Windows recreation defect is reported separately.
- **FCIS and typing:** changed domain configuration/health code remains value-oriented, with no introduced disk/network/process/clock dependency or `any`/`interface{}`. Filesystem/process ownership and JSON/YAML decoding stay in the imperative shell. Necessary locks, retained artifacts and journals serve actual runtime state and are not candidates for deletion solely because they add adapter code.
- **Portable history and other consumers:** inspected manual admission/generation compatibility, capture registration/worker callers and history preservation boundaries. C2 changes executable installation/dispatch behavior; it does not add ledger authenticity gates, schema changes or private-state requirements for supported stored facts. Operational executable hashes remain distinct from telemetry ownership. No additional portable-history blocker was substantiated in this focus.

## Verification and remaining limits

- The supplied three-counterexample script passes on the actual C2 installer/dispatcher.
- The POSIX cancellation probe successfully reproduces ARCH-SOL-4, with explicit child cleanup. Its intermediate probe attempts initially observed short-lived interface probes or only one Go thread's child list; the final oracle enumerates all dispatcher task children and requires final-command argv. Only the final recorded result supports the finding.
- `go test -overlay /tmp/al79-architecture-sol-c2/windows-branch-overlay.json ./internal/adapters/capture -run '^TestReviewWindowsMissingStableRepair$' -v -count=1` fails at the activation assertion for ARCH-SOL-6 after its normal control passes. The overlay preparation's initial cwd error was corrected before this run; no executed Windows claim is inferred.
- Reused raw `TEST_EVIDENCE.md` and `DISTRIBUTION_EVIDENCE.md` for targeted/package coverage and C2 installer controls. Reused genuine Linux/Codex 0.167.0 native root/child binding, source-backed publication and prior-worktree preservation evidence in `NATIVE_EVIDENCE.md`: C2 does not change source extraction, automatic callback cwd/strict validation, binding, worker publication or ledger reconciliation. Its changed install ownership is covered by the C2 production-installer probes, not retroactively claimed exercised by that native run.
- No full suite, race suite or integrated `make` was rerun by this reviewer. Those gates remain parent-owned after corrections. Native Windows/PowerShell remains unavailable; cross-compilation cannot validate its process/argument or loaded-image behavior.
- Scope covers the AL-79 architecture and common-runtime boundaries inspected above. The parent owns finding reconciliation, fixes, consolidated review, telemetry and publication.
