# AL-85: Publish telemetry safely across change moves

## Goal

Telemetry refresh and automatic capture follow an existing OpenSpec change into its archive location without recreating an active folder or changing session ownership.

## Context

The automatic worker currently constructs its output from root plus ticket ID. Changing discovery to multiple roots leaves that behavior intact. Late callbacks after archive could recreate a second active folder and split history. This ticket completes the write-side lifecycle in [PO-24](../../PO-24/SPEC.md).

## Acceptance criteria

- In managed existing-change mode, automatic publication and normal manual destination selection use the same unique resolved change as readers. AgenticLens can initialize its ledger inside a known existing change but never create a missing change directory. No target or multiple targets produces an actionable non-success without fallback publication.
- Preserve existing explicit manual output/initialization behavior for legacy single-root projects. Explicit destinations in managed OpenSpec roots must agree with the resolved change and cannot bypass containment, identity or missing-parent checks.
- A late/manual refresh of an archived change updates that existing archived ledger. Do not reactivate the change, create an active sibling or require renaming its recorded ticket. Concurrent archive/move versus publication yields one valid update or a recoverable explicit retry, never lost history, duplicate outputs or recreated parents.
- Exact native source/session bindings remain attached to canonical ticket identity within their current project/worktree. Changing configured root or folder location needs no rebinding; multiple roots do not permit another worktree/session to claim ownership.
- Refresh complete local source snapshots while preserving unavailable/partial owners, other harnesses and developers. Repeated refresh is idempotent, and a receiving machine can retain supported history without private capture receipts.
- Missing, ambiguous or interrupted destinations appear consistently in CLI/capture diagnostics and dashboard recovery context. Demonstrate active-to-archive reads and publication with an existing ledger, a first ledger, unavailable sources and an archive race.

## Implementation notes

Use the read-side identity/location boundary established for this initiative. AL-81 separately removes manual generation's capture-registry dependency; do not reintroduce it or weaken automatic ownership guards. The archive-race oracle must exercise a real filesystem boundary, not just a path-string assertion. Compressed archives and cross-repo stores remain outside scope.
