# AL-79 behavior and integration review — behavior-luna-01

## Assessment

- **Candidate:** AL79-C2; base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`; source/guidance/test diff SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de` (workspace metadata excluded). The working tree `HEAD` remains the base; review was against the supplied frozen diff and corresponding candidate files.
- **Identity:** requested `gpt-6-luna`, medium effort. Runtime-observed model/effort and cost are unavailable and are not asserted.
- **Requirements:** approved AL-79 contract and iteration plan, AL-79 specification and PO-20 F1-F6, plus the canonical portable-history and regression boundaries. Review scope was behavior/integration across session context, configuration selection, native dispatch, worker/publication, status, and manual/disabled/legacy compatibility.
- **Initial status:** task source and evidence files were already staged/modified, with multiple untracked review reports. No unrelated files were changed. Only this assigned report was written.

## Findings

No substantiated introduced behavior finding in the inspected Linux/native-Codex paths (severity 0).

## Consequential hypotheses investigated

1. **Callback cwd could redirect selection away from the real session checkout.** Refuted by `nativeCodexHook`: it requires an absolute callback cwd, finds the active shared YAML there, and routes from that project. The dispatcher regression evidence also exercises an unrelated process cwd while preserving absolute session context.
2. **A worktree could inherit machine policy piecemeal or silently fall back from an invalid local policy.** Refuted by `SelectConfig` and `readSelectedLocal`: the worktree file wins as a whole when present, otherwise Git worktree metadata selects the base local file; parsing/validation failure is returned without fallback. The raw config evidence covers linked/nested worktrees, custom shared values, missing values and invalid selected files.
3. **Disabled capture could still inspect source, allocate state, check executable readiness, or launch a worker.** In `nativeCodexHook`, selected local configuration is validated first, then `Enabled == false` returns before Codex source checks, session normalization, or receiver/worker dispatch. The dispatcher test evidence exercises disabled mode with unavailable source and mismatched runtime home. `ResolveEntrypoint` is consulted before disablement for required shared selector configuration; no selected binary is validated or executed on this branch.
4. **Codex SubagentStart could bind/collect the parent as the child or reject the actual native event shape.** Refuted by `NormalizeCodexSubagentStart` and `ReceiveCodexUser`: the callback's parent ID remains the registration identity, the child path is queued separately, and child metadata must prove the parent/project relationship. Actual normal-trust native evidence records the corrected callback and distinct source-proven child ID; child binding inherits the parent's explicit task.
5. **Manual commands, OpenCode consumers, or old pinned binaries could be broken by capture-only requirements.** Manual `generate` resolves through `ResolveGenerationCapture`, which preserves shared-only/manual policy rather than requiring automatic selected-local enablement. `dispatchNative` preserves caller cwd for ordinary commands (`command.Dir` is set only for automatic callbacks). The supplied distribution evidence exercised an actual dispatcher with an installed legacy exact pin lacking capture interfaces and checked nested-cwd relative file behavior. Shared capture APIs remain used by OpenCode paths; no concrete incompatibility was found in the inspected caller changes.
6. **Preferred-build changes could retire an active worker or redirect its later publication children.** The runtime test evidence describes a real process probe that replaces preferred mutable output while an admitted worker is active and verifies retained executable identity for later children and pending work. Native root/child publication also observed the admitted build identity and preserved another worktree's owners.
7. **Status could report installation as proof of capture, or fail to expose recorded activity.** Raw native evidence records the initial observed status defect and its correction. The corrected status reports only recorded callback/binding/source/publication stages, separately from harness trust and unobserved sessions. This appropriately bounds the status claim.

## Verification reused and limits

- Reused `TEST_EVIDENCE.md`: focused Go config/capture/dispatcher and native callback tests; repeated corrected CLI boundary tests; retained-worker real-process test. These are scoped results, not a full integrated gate.
- Reused `DISTRIBUTION_EVIDENCE.md`: POSIX distribution/setup tests, actual dispatcher A-to-B ownership transaction, legacy manual compatibility and nested-cwd relative-argument checks.
- Reused `NATIVE_EVIDENCE.md`: Codex 0.167.0 normal-trust root/child capture across two linked worktrees, explicit root binding, source-proven child inheritance and automatic publication. The record notes SessionEnd visibility limits and that no separate human source session was asserted.
- **Not run by this reviewer:** no duplicate suites or native session probe. Parent's integrated `make`/race gate is pending. Native Windows execution is unavailable; only cross-compilation/static fixture evidence exists. OpenCode runtime was not separately exercised in this review.

## Verdict

**INCOMPLETE** for this behavior assignment. I found no blocking Linux behavior defect, but native Windows behavior remains a material supported-platform gap for this change, and the integrated gate is still pending. This scoped result does not approve the full candidate or certify Windows behavior. Review confidence is moderate for the inspected Linux/Codex and manual-dispatch paths, low for Windows runtime behavior.
