# AL-68 bounded security re-review

Overall outcome: **INCOMPLETE**. Scoped code verdict: **APPROVED**. CAP-06 and CAP-07 are fixed; no substantiated blocking security defect remains in the assigned correction and trust-interaction scope. Required real Codex human-trust/callback/agent-binding evidence remains unavailable, so this report does not pass the overall delivery gate or claim complete Codex support.

Run: `security-sol-04`, final bounded correction re-review. Requested execution: `codex/gpt-6.1-sol/high`. Observed environment: Codex tools on Linux; actual executing model, effort, usage and cost metadata were not independently exposed. Requested identity is not asserted as observed identity.

Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Reviewed revision: corrected-04, fingerprint `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`, as defined in `CANDIDATE-04.json` by sorted tracked/nonignored regular paths outside `alspaces`, path/NUL/content/NUL. Independently recomputed and matched before building the disposable probe binary. Existing staged/unstaged implementation and ticket artifacts were present at entry and preserved. Only this assigned report was written in the repository.

## Findings and correction dispositions

### CAP-06 / SEC-SOL-01: fixed

Source: `backend/internal/adapters/capture/install.go`, OpenCode rendering block; embedded `plugin.ts`.

The installer JSON-encodes binary/state/store paths and feeds the original template to one `strings.NewReplacer(...).Replace` operation. Replacement text is not rescanned. The plugin continues to use native `execFile` with an argument array; no shell interprets callback or binding-tool input.

Independent native probe: built this exact candidate to `/tmp/al68-security04-7_qh_z1c/native`, copied it to a handler named `native __STATE__ __STORE__`, and installed into a disposable repository using a state path containing a quote and `__BINARY__`, and a store path containing `__STATE__`, quotes, semicolons and dollar text. Actual `capture install --source opencode` exited zero. Parsed all three generated constant declarations and recovered each original path exactly; `node --check` on those emitted JavaScript declarations exited zero. This directly settles the original escaping/replacement failure; no arbitrary-code-execution claim was made for the original defect.

Durable oracle inspected: `TestCaptureInstallerRendersMarkerBearingPathsOnceAndRollsBackMidWrite/literal_marker_paths` in `backend/cmd/agentic-lens/capture_test.go`. Its mutually referring markers prevent a repeated-substitution implementation from passing by a favorable iteration order, and its assertions compare decoded original values rather than expected template text. Initial failing native evidence remains in `REVIEW-security-sol.md`; affected race results are recorded in `TEST_REPORT.md`.

### CAP-07 / SEC-SOL-02: fixed

Source: `backend/internal/adapters/capture/hooks.go:ReadHookContext`; `backend/cmd/agentic-lens/capture.go`, receive dispatch.

Native receive now selects between bounded parsing completion and context cancellation before registering an event. Cancellation closes native stdin and returns the deadline error. The 64 KiB volume limit remains separate from the elapsed-time limit.

Independent native probe: ran two real receive processes with `--timeout 100ms`, retaining their stdin write ends open: one supplied no bytes; one supplied and flushed a complete valid SessionStart JSON object without EOF. Both exited 1 after approximately 0.115 seconds, reported deadline exceeded, and left zero registrations in their actual SQLite registries. Neither required the independent two-second kill deadline. Thus the corrected path handles a pipe that remains open even after complete JSON, rather than depending on EOF or an outer harness timeout.

Durable oracle inspected: `TestCaptureReceiveNativeTimeoutBoundsOpenStdin` writes valid JSON into a real native process, keeps its pipe open and requires the command's 50 ms timeout to return before a separate three-second process deadline; it also verifies that no event persisted. Initial failure beyond 0.801 seconds remains in `REVIEW-security-sol.md`. No unresolved CAP-06/CAP-07 trigger remains.

## Changed trust interactions checked

- Exact identity and challenge: inspected `Store.Change`, `Store.Bind`, `CanonicalStore`, `SameStore`, `CaptureIdentity.Validate` and `CaptureBinding`. A single equivalent legacy alias row retains its challenge and accepted facts inside the write transaction; multiple equivalent rows fail explicitly. Distinct real stores are not equated. Bind still requires the exact stored challenge and rejects conflicting ticket ownership.
- Independent native binding probe: received SessionStart through a real directory symlink while including forged ticket/parent and an extra secret field in stdin. The registration stayed unbound with no invented parent, and the extra fields did not appear in agent feedback or local registry files. Binding through the resolved actual store with the exact challenge succeeded. Rebinding that owner to another ticket through the alias, using the challenge for another store or session, and using a forged challenge all failed. No request tokens or raw payloads were copied into this report.
- Actual agent context: Codex start/prompt responses still include the exact session/store/challenge command; SubagentStart still warns that the callback identifies the parent rather than guessing the child's source ID. OpenCode's tool obtains `sessionID` from runtime `context.sessionID`; only ticket/request come from agent tool arguments. Source inspection and synthetic native probes establish those code boundaries, not a real Codex agent exchange.
- Publication ownership: inspected `AcquirePublication`, both platform lock implementations, its early `runGenerate` caller, and `saveGenerationLedger`. Lock names derive from resolved target paths under the user's cache, not hook-controlled shell text; kernel ownership is retained across atomic ledger replacement and released on process exit. It is acquired before baseline/source reads, including generation without a capture registry. No introduced concrete trust regression was established. Reused the reported overlapping-native-publisher regression; did not repeat its full package suite or claim non-Linux runtime execution.
- Final privacy: inspected `ReportJSON`, `redactedRegistrations`, result/registration JSON tags and static source-error details. Final report projection admits only the requested ticket's rows, clears challenges and source-store paths, and omits transcripts, snapshots, skills/task arrays and refreshed source arrays through their JSON exclusions. Coverage summary renders counts/checkpoints rather than local paths. The existing `TestReportDoesNotPublishLocalPathsOrBindingToken` checks redaction and preservation of local state; the unassigned-root oracle separately checks exclusion of unrelated identities. Existing source-derived AL-62 facts retain their approved preservation contract; no broader arbitrary-source secret-redaction guarantee is inferred.
- Trust remains harness-controlled: canonical setup/operator/manual guidance still requires normal trust/plugin review and actual binding/source verification; it does not auto-enable trust. No active user/global settings, source sessions or trust state were modified by this review.

## Evidence reused and limits

Read the reviewer persona/shared procedure, review skill, technical/regression rules, CONTRACT/PLAN including the explicit actual-agent/no-guess clarification, CORRECTIONS/FINDINGS, initial security Sol/Luna reports, IMPLEMENTATION, TEST_REPORT, SMOKE, SETUP_IMPLEMENTATION and VERIFICATION. Source investigation preceded reconciliation with prior security conclusions. Scope was limited to CAP-06/CAP-07 and changed identity/publication/privacy interactions; no unrelated dependency audit was performed.

Executed: candidate fingerprint calculation; exact-source disposable Go build; actual native installer and JavaScript constant syntax/value probe; two actual open-pipe native timeout probes; native alias/challenge/forged-payload probe. All passed. Disposable evidence is under `/tmp/al68-security04-7_qh_z1c`, with normalized `results.json`; raw local state stays outside Git.

Reused: corrected candidate-04 domain/capture/CLI race passes (1.172s/5.140s/13.835s), unchanged source-package results and fourteen distribution tests documented in TEST_REPORT/SETUP_IMPLEMENTATION, and source03 full `make` in VERIFICATION for unchanged components. These are recorded upstream results, not reruns or independent execution claims. Parent still owns final integrated `make`, consolidation and delivery.

The corrected real OpenCode 1.19.5 Linux binding/source exchange in SMOKE remains applicable to the unchanged agent-feedback protocol; it was not rerun here. Codex 0.166.0 currently has version/features/configuration observations but no actual trusted callback, agent-response and source-verified capture evidence. That required gap remains **unfixed as evidence**, distinct from the fixed code findings. No trust bypass or renewed user question was attempted. Non-Linux execution and other unobserved harness/interface paths remain limited as recorded in SMOKE. No stage, commit, publication, telemetry or child agent was performed.
