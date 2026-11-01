# AL-68 release/setup specialist re-review

Run: `release-luna-04`
Requested profile: Codex / `gpt-6-luna` / medium. This records the requested profile only; this bounded review did not independently inspect its source-session metadata and does not infer runtime identity from the request.
Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`
Frozen candidate: `CANDIDATE-04.json`; declared source fingerprint `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`. I independently recomputed the sorted path/NUL/content/NUL hash over the manifest's declared source paths and obtained the same fingerprint. The manifest file's own SHA256 is `f2362ccd35d970fb2a5c3e2fbba2746a1beadaa57cf91f33bea5af8538fc550c`; that is a separate value and is not the source fingerprint. Worktree `HEAD` remains the baseline; this is a frozen content candidate, not a commit.

## Scope and verdict

Re-reviewed CAP-08 and CAP-09 only: both shipped downloaders and wrappers, native handler protocol checks, replacement/rollback paths, foreign binaries, public direct Unix/PowerShell setup directions and the distribution fixture oracles. Read the shared reviewer context/procedure, contract, plan and hook-agent exchange requirement, corrections/findings, prior release-luna focus report, and implementation, setup, smoke, test and verification evidence. Did not read the partner Sol04 closing report. Product/tests stayed read-only; only this report is assigned for writing.

**Scoped code verdict: APPROVED.** CAP-08 and CAP-09 are fixed on this candidate. I found no supported concrete regression in the inspected release/setup paths and no remaining CAP-08/09 finding.
**Overall acceptance evidence: INCOMPLETE.** Codex's concrete hook configuration still awaits human trust and `/hooks` review, followed by actual callback, agent-binding and source-record comparison. No trust action or bypass is claimed. This runtime acceptance gap is independent of the scoped code verdict.

## Findings

No blocking findings.

- **CAP-08 — fixed.** Both downloaders stage replacements for local and release-download routes, including explicit `--force --version TAG`, and probe a native candidate before activating it when capture is requested or an existing native handler reports a capture protocol. A mismatch exits before replacing the handler or touching hook configuration. The prior native handler and wrapper remain backed up through native setup; failure restores their exact prior bytes. The native installer independently rolls back its owned configuration writes. Candidate protocol probing invokes the staged candidate itself: if a wrapper is mistakenly supplied, its capture route resolves only beside the staged candidate and fails because no native executable is there; it cannot borrow the working installation's protocol. Wrapper `update --version TAG` stages candidates, checks the host candidate against the active native handler's protocol, and swaps only after successful preflight. Existing foreign-platform binaries in capture-aware installations are retained and diagnosed for native-host verification.
- **CAP-09 — fixed.** The manual clearly separates binary-only setup from lifecycle readiness, links the actual lifecycle flow and gives native Unix and PowerShell install/status invocations using `capture install --source --store --repo --state-dir [--dry-run]` and `capture status --state-dir`. These match `capture.go` parsing and the canonical CLI instructions. It carries through explicit source/session binding, user-controlled trust, owned removal/recovery, and the instruction not to overwrite an active handler with the initial-download example. PowerShell commands use the Windows native executable and do not claim POSIX wrapper behavior.

## Oracles and verification evidence

The corrected `scripts/tests/test_capture_distribution.py` fixtures execute both downloader entrypoints and both wrappers in disposable repositories. The reported 14-test run contains concrete regression oracles for incompatible local and downloaded candidates, with and without capture setup; rollback after a later native setup failure; wrapper update mismatch and success; injected activation rename failure; and preservation/diagnosis of a foreign-platform executable. The wrapper-candidate test specifically confirms a staged wrapper cannot probe successfully through the working native executable. These tests exercise the production scripts with local transports and fixture executables; they do not simulate a real harness callback or establish native Windows/macOS execution.

Reused unchanged evidence: `SETUP_IMPLEMENTATION.md` records `make test-capture-distribution` PASS (14 tests), shell syntax checks and a disposable probe through the built Linux native handler for each downloader. The probe distinguishes restored old native bytes with an appended marker and verifies edited hook bytes survive failed setup. `TEST_REPORT.md` records all affected Go race packages passing on its noted source revision. `VERIFICATION.md` records full `make` PASS on candidate 03 and notes the component scopes are unchanged by candidate 04; the parent owns the final integrated make gate. I did not rerun suites or perform release network actions.

## Runtime evidence and limits

The smoke report contains real OpenCode 1.19.5 Linux root and child agent interactions, explicit binding responses, exact source IDs/parent relationship and source reconciliation. Corrected installation evidence also records a real agent response on the corrected candidate. This supports the observed OpenCode interface only; no broader version/platform claim follows.

Codex 0.166.0 has a concrete disposable hook configuration awaiting human review/trust. There is no claimed Codex callback, agent response, child association, or source comparison. Windows and macOS native execution also remain unverified. These are reported as evidence limits, not defects in the release/setup code reviewed here.

## Closeout

Fixed IDs: CAP-08, CAP-09.
Unfixed IDs in assigned focus: none.
Review scope confidence: high for static release-path tracing and the cited Linux production-script fixtures/probe; bounded by reused test runs and the explicitly unverified runtime/platform evidence above. No PR, publish, trust, commit, staging or telemetry action was performed.
