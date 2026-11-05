h3. Title: FIX-19: Reject ambiguous Codex owner snapshots

*TARGET*: AgenticLens (Codex Discovery / Generate / Backfill / Historical Hydration)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: AL-67 architecture assessment; user-requested follow-up ticketing.
*Findings*: F04

*CONTEXT*

Codex family discovery can return two files with the same native session ID, such as an active JSONL and its archived gzip copy. Generation/backfill combine both cumulative snapshots: domain grouping produces one owner but appends their invocation rows and sums tools. Hydration instead accepts the first readable copy. Unique session-ID validation therefore does not catch the duplicated facts.

Example: session A has one tool call in sessions/A.jsonl and the same complete content in archived_sessions/A.jsonl.gz. Counting both files can report two calls for A. Two distinct real tool calls inside one authoritative transcript must still count twice.

*GOAL*

Every exact Codex owner is refreshed from an unambiguous native snapshot, with conflicting source copies diagnosed rather than summed or chosen arbitrarily.

*VERTICAL SLICE SCOPE*

* Adapter: give family discovery and exact-owner consumption consistent ambiguity handling before normalization.
* Consumers: cover generate, source-backed backfill and historical event hydration, preserving other usable owners.
* Quality: use real temporary active/archive JSONL and gzip fixtures, then verify saved/reloaded facts and API coverage.

*ACCEPTANCE CRITERIA*

* Identical or divergent readable files for one owner produce an explicit ambiguity/incomplete result before they can be combined or certified complete. Do not choose by filename, timestamp or directory priority without a separately approved authority rule.
* Generate cannot publish doubled owner facts; failed refresh/backfill preserves that owner's accepted ledger history and reports the reason. Historical hydration does not select an arbitrary copy; unaffected owners remain usable with truthful coverage.
* Genuine repeated invocations, skills and delegations inside one selected transcript survive unchanged. Do not repair the bug with broad row deduplication.
* Verify active/archive and compressed/uncompressed duplicates, warning-bearing/conflicting copies, child relationships, unrelated owners and repeated unchanged runs through actual consuming paths.
* Exact discovery's existing ambiguity rejection remains effective. Demonstrate the counterexample before the fix and correct counts/preservation afterward.

*SCOPE LIMITS*

No broad repository redesign, invented source facts or cleanup of real developer logs. Unavailable same-harness and other-harness history must survive; this ticket does not require source receipts for Git-shared ledgers.

*DEFINITION OF DONE*

* Revalidate the AL-67 source evidence at the implementation/review revision. Record actual checks, remaining limits and disposition; do not claim a production incident from a static counterexample.
* Work in a separate dedicated ticket session under the appropriate development/review workflow. Preserve portable ledger history, exact source/session/parent identity, legitimate repeated occurrences and explicit unavailable values. No application implementation is authorized by this planning document alone.
* Use self-contained fixtures; keep I/O outside the pure domain and follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. Implementation requires affected checks, independent review and the required make gate; frontend changes also require canonical embedded-asset verification. Review-only work records evidence and recommendations without application/test edits.
* Publish the scoped deliverables and source-grounded telemetry. Do not mark completion from passing tests or a proposed design alone.

*SOURCE EVIDENCE*

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F04; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Still open: AL-67 F04 documents duplicate active/archive owner snapshots through generation/backfill versus first-readable hydration. No merged corrective delivery was found; retain exact-owner ambiguity semantics.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.
