# AL-97: Technical design

## Scope and boundaries

Extend the existing native `capture check` advisory with the [staged snapshot requirements](specs/staged-pending-capture-advisory/spec.md). Use the [AL-96 snapshot and discard contract](../al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/design.md#reversible-discard-of-ad-hoc-capture). Pending capture, discard, restore, and canonical publication keep their existing lifecycle. Git continues to own staging and commits.

## Decisions and ownership

- **Setup.** The coder owns any setup guidance changes in the setup skill and setup manual. Use the existing `install-check --dry-run` and opt-in install or refresh flow, including its foreign-hook manual composition guidance and local `telemetry.capture.commit_warning` switch. Keep the existing stable manager entrypoint and effective `core.hooksPath` handling.
- **Index and staged content.** The coder owns the native CLI and adapter changes. Resolve the pending location through the existing shared configuration resolver. Add a read-only Git path for the hook's effective index, including an inherited `GIT_INDEX_FILE`, while keeping the existing repository-authority helper's environment filtering for other callers. Scope the staged diff to the configured directory, use NUL-delimited records and staged blob IDs, and cover the initial commit where no `HEAD` exists. Limit inspection to added or changed regular snapshot files, including changes introduced at a new path.
- **Classification.** Reuse the pending snapshot format and identity checks on bounded staged blob reads, independent of working-copy bytes. Read exact discard disposition for the current worktree through the existing read-only capture control layer. Only a confirmed matching discard suppresses a warning. Missing or unreadable local capture state leaves the staged evidence usable.
- **Output.** Put a compact summary of both pending categories and their next actions within the existing hook's first-line and byte limits. Run staged inspection before capture-health inspection so slow or unavailable health data does not hide a staged result. Preserve the existing health advisory when no staged warning is due and reuse the hook's fail-open behavior. Bound candidate count, bytes and elapsed time, and make incomplete inspection visible. Emit no snapshot payloads or raw source details.
- **Verification and delivery.** The tester owns self-contained regression fixtures and tests. Focused reviewers inspect the Git index boundary, setup ownership and read-only capture-state use. The orchestrator owns integration and delivery evidence.

## Implementation approval

The user explicitly approved implementation with “$agentic-lens-dev aprovado” on the presented AL-97 proposal, specification, design and tasks at baseline `db4477fa`. Approval covers the staged-snapshot advisory and opt-in setup guidance described here.

## Implementation approach

Keep Git and capture-state reads in adapters and compose the advisory in the native check. Test the installed hook on a real path-limited commit where the default index differs from Git's commit index. Separately test a staged snapshot whose working copy was changed or removed. Check an initial commit, a custom pending directory and linked worktree, and quiet cases for deletions and unchanged snapshots. Verify that a confirmed discard suppresses a stale staged copy while unavailable state still permits classification. Compare index bytes before and after direct hook execution and verify a real commit succeeds with a warning. Reuse existing ownership and capture-health regression coverage and run integrated `make` after implementation.

Select the [configuration and project portability](../../../docs/REGRESSION_PREVENTION.md#configuration-and-project-portability) and [source availability](../../../docs/REGRESSION_PREVENTION.md#cleanup-compatibility-and-source-availability) cases. Behavior review follows the installed hook through manager and CLI. Data review checks staged blob authority, worktree-local discard, bounded reads and preservation of the index. Architecture review checks reuse of the setup, configuration and capture-control boundaries. Test review checks that real Git fixtures distinguish the default index, commit index and working copy.

## Steering

The OpenSpec exploration proposed reusing the existing advisory and warning only for snapshots introduced or updated by the commit. This plan carries that interpretation into the acceptance scenarios. It preserves intentional versioning and tests the effective commit index explicitly, including where Git provides a temporary index.

Independent review confirmed that pre-commit receives the effective index without an amend indicator. Use the current `HEAD` as the staging baseline, including during amend. The advisory concerns changes being staged and leaves replacement-history checks outside this hook. The review also found that a mutable worktree directory, an unavailable private state path, an alternate Git object store, or an unassociated capture registration could change the result. Keep the index path and object view authoritative, and make exact local discard optional. Preserve existing health output when an incomplete staged scan classifies nothing. Verify each boundary with the native hook or the focused adapter oracle.
