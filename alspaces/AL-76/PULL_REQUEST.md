# feat(ledger): compact records and add selective format command (AL-76)

Persist ledger records with a shared readable JSON layout and provide an opt-in `format` command for all, one, or selected ticket ledgers. Preserve original facts, unknown fields and JSON literals while coordinating safe filesystem publication.

Support dry-run and check modes, explicit or configured workspace roots, deterministic target reports and source-independent operation. The formatting policy belongs to the invoked release; no hooks or release-pin updates are installed.

Draft checkpoint only; this PR is not merge-ready. Publication is explicitly authorized ahead of the final delivery gate.

Pending gates:
- Wait for MGT-2's actual AL-60/AL-61 merge notification, then integrate predecessors once and preserve their ticket-index updates.
- Run the required integrated full `make` gate and assess any affected review evidence.
- Complete consolidated delivery approval, final telemetry and ticket lifecycle artifacts before readiness.

MGT-2 owns merging. Keep this PR draft until these gates are satisfied.

Contract: [CONTRACT.md](alspaces/AL-76/CONTRACT.md). Existing review and verification evidence: [SUMMARY.md](alspaces/AL-76/SUMMARY.md).
