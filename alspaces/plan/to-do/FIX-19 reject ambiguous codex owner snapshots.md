h3. Title: FIX-19: Reject ambiguous Codex owner snapshots

*GOAL*

Every exact Codex owner is refreshed from an unambiguous native snapshot, with conflicting source copies diagnosed rather than summed or chosen arbitrarily.

*CONTEXT*

Codex family discovery can return two files with the same native session ID, such as an active JSONL and its archived gzip copy. Generation/backfill combine both cumulative snapshots: domain grouping produces one owner but appends their invocation rows and sums tools. Hydration instead accepts the first readable copy. Unique session-ID validation therefore does not catch the duplicated facts.

Example: session A has one tool call in sessions/A.jsonl and the same complete content in archived_sessions/A.jsonl.gz. Counting both files can report two calls for A. Two distinct real tool calls inside one authoritative transcript must still count twice.

*ACCEPTANCE CRITERIA*

* Identical or divergent readable files for one owner produce an explicit ambiguity/incomplete result before they can be combined or certified complete. Do not choose by filename, timestamp or directory priority without a separately approved authority rule.
* Generate cannot publish doubled owner facts; failed refresh/backfill preserves that owner's accepted ledger history and reports the reason. Historical hydration does not select an arbitrary copy; unaffected owners remain usable with truthful coverage.
* Genuine repeated invocations, skills and delegations inside one selected transcript survive unchanged. Do not repair the bug with broad row deduplication.
* Verify active/archive and compressed/uncompressed duplicates, warning-bearing/conflicting copies, child relationships, unrelated owners and repeated unchanged runs through actual consuming paths.
* Exact discovery's existing ambiguity rejection remains effective. Demonstrate the counterexample before the fix and correct counts/preservation afterward.

*Status*: Ready
*Blocked by*: None
*Work type*: fix
*Priority*: P2

*IMPLEMENTATION NOTES*

*Findings*: F04

No broad repository redesign, invented source facts or cleanup of real developer logs. Unavailable same-harness and other-harness history must survive; this ticket does not require source receipts for Git-shared ledgers.

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F04; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Still open: AL-67 F04 documents duplicate active/archive owner snapshots through generation/backfill versus first-readable hydration. No merged corrective delivery was found; retain exact-owner ambiguity semantics.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
