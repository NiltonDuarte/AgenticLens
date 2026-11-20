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

## Pending capture steering

- [x] **Planner** · Define the pending snapshot contract, shared directory setting, delayed binding and agent destination admission. Assign writable-path owners and record implementation approval for the revised scope.
- [x] **Coder and tester** · Implement and verify pending snapshots, delayed binding, native completion, destination validation and retry-safe publication. Verify custom paths, the explicit setup default, preservation on update and rejection of missing or invalid values at read time. Preserve canonical discovery and Git staging behavior.
- [x] **Tester and reviewers** · Extend the packaged native gate for both pending states, completion before binding, agent-supplied destination, promotion and subsequent refresh. Review admission and recovery independently.
- [x] **Orchestrator** · Verify the revised candidate and record the results alongside the previous delivery.
- [x] **Planning** · Record the optional, advisory Git pre-commit hook and its installation choice in [AL-97](../../../alspaces/plan/to-do/AL-97%20warn%20about%20staged%20pending%20capture%20snapshots.md).

## Reversible discard and restore

- [x] **Planner** · Define CLI selection, registry state and worker coordination contracts. Assign writable-path owners and record the approved implementation scope.
- [x] **Coder and tester** · Implement discard, discoverable status and restore with full native history. Verify unbound and bound captures, repeat operations, cleanup retry, concurrent worker passes, delayed descendants, and preservation of Git and canonical ledgers.
- [ ] **Tester and reviewers** · Exercise native work before, during and after discard. Verify suppression after restart or resume and exact recovery without double counting. Cover missing source history and review the discard status contract for AL-97.
- [ ] **Orchestrator** · Verify the integrated candidate and record results alongside the earlier delivery.

## Workflow follow-up

Archive only after verified delivery. Reconcile roadmap status separately.
