# fix(telemetry): preserve source-owned skill and delegation facts (AL-62)

Codex recorded skill loading and delegation evidence was discarded before reaching skill analytics. Preserve source occurrence identities under their owning sessions through generation, persistence, refresh and presentation.

Migrate legacy history without guessing ownership, retain unavailable owners during partial refresh, distinguish loading evidence from unknown execution outcomes, and derive offline timeline evidence from canonical facts. Publication receipt authorization uses owner-derived projections so canonical facts cannot bypass source proof. Bounded destination-scoped active/pending receipt generations retain trusted proof across file rollback and receipt-finalization interruption, with explicit compatible metadata migration.

## Draft checkpoint: pending delivery gates

This draft publishes the reviewed implementation checkpoint; it is not merge ready. Integration remains on hold until MGT-2 explicitly notifies this thread of the actual AL-60, AL-61 and AL-76 merges. Preserve their capture repairs, historical timestamps, lossless serialization and publication locks during reconciliation.

After that notification: integrate predecessors, run the required full make/race/build and standalone embedded-asset verification, resolve integration findings with affected re-review, finalize telemetry and review retention, and complete ticket lifecycle updates. MGT-2 owns merging.

Original extraction/ownership checkpoint: 29ee4a802000c1d3302c515e001f7692d3d7fe8f. Subsequent CI and bounded receipt corrections are documented in the summary. Scoped approvals preserve the pending final integration gates above.

[Contract](alspaces/AL-62/CONTRACT.md). [Summary and verification evidence](alspaces/AL-62/SUMMARY.md).
