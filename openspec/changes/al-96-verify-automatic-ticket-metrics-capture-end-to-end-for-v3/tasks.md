# AL-96: Execution tasks

Use the [design](design.md) for approval and the [delta spec](specs/automatic-ticket-publication/spec.md) for acceptance.

## Implementation and focused verification

- [x] **Planner** · Confirm the operator entry point, private selection lifecycle and protected writer authority. Assign exact writable paths and record scope approval before coding.
- [x] **Coder** · Implement explicit first publication and recovery in the common worker and publisher. Reuse destination admission and locking, preserve ticket binding and portable history, and update the supported setup instructions.
- [x] **Tester** · Verify missing and conflicting destinations, concurrent first publication, interrupted commit and retry, relocation or deletion, and preservation of existing owners. Include protection against reusing the consumed initial selection.
- [x] **Tester** · Build the repeatable packaged-candidate gate using isolated projects and real native clients. Compare root, implementation and verification children, nested descendants where supported, and the excluded unrelated session with independent native observations. Verify first publication, completion and the subsequent normal turn against the spec.
- [x] **Coder and tester** · Reproduce and fix any supported-path failures, including avoidable completion-to-commit delay. Rerun the affected checks and the full gate on the corrected package.

## Integration and delivery

- [x] **Independent reviewers** · Check selection authority, concurrent publication and recovery, exact source coverage, timing evidence and whether the gate can pass while bypassing installed lifecycle callbacks.
- [x] **Orchestrator** · Run integrated `make` and affected checks. Record package identity, actual native results and support limits in delivery evidence. Verify AL-96 session telemetry through automatic publication and record any missing facts. Derive and hand off the PR from delivery with the gate result.

## Workflow follow-up

Archive only after verified delivery. Reconcile roadmap status separately.
