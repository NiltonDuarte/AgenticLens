h3. Title: AL-71: Keep dashboard requests and live data reliable

*GOAL*

Dashboard selections own their results, time windows behave consistently, and live data survives transport errors and recovers accurately.

*CONTEXT*

The current selection must own asynchronous results, time windows must have one meaning, and malformed or recovering live transports must preserve valid data and truthful status.

*ACCEPTANCE CRITERIA*

h4. Own filter requests and progression date boundaries consistently

Late filter responses overwrite current selections. Progression presets switch from latest-ledger time to server-current time when a type/prefix is selected. Duplicated state ownership lets the views diverge.

* Fence data, diagnostics, loading and errors by the active request, including stale rejections after newer successes. Preserve existing useLiveTrace and TicketDeepDive ownership semantics.
* Choose one documented progression time-window meaning and use it with and without filters; test a clock far beyond the newest ledger while toggling a type/prefix that includes all shown tickets.
* Extract only shared request/filter/date invariants that remove concrete duplication, and keep pure calculations out of views. Do not add a generic caching framework.
* Use deferred A/B response and rejection permutations and a real server-date contract fixture, not only immediately resolved mocked requests. FE-D2 benchmark arithmetic belongs to AL-72.

h4. Validate live payloads and recover polling state

Malformed nested live payloads pass unchecked casts and crash consumers; fallback polling remains in error after successful recovery and launches overlapping requests.

* Validate required nested nodes/messages/parts/anomalies at REST and SSE ingress using shared decoders. Normalize only truly optional collections and retain the last valid snapshot on malformed input.
* Restore fallback-polling status on success after error, serialize/coalesce pending polls, and prevent obsolete or post-teardown results from mutating state.
* Test null/primitive/wrong-scalar/nested-malformed payloads, valid optional omissions, failure-threshold recovery, deferred overlapping requests and teardown. Do not misrepresent malformed data as a valid empty snapshot.

*Status*: Ready
*Blocked by*: None
*Work type*: fix
*Priority*: P2

*IMPLEMENTATION NOTES*

*Related*: AL-69 owns source diagnostics and anomaly semantics; AL-70 owns configured ticket parsing/filter identity; AL-72 owns rendered UI behavior and benchmark arithmetic.

*Findings*: FE-02, FE-03, FE-04, FE-05, FE-D2

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain request ownership, date boundaries and live validation/recovery. Current TicketComparisons, MetricProgression and live hooks remain the consumers. AL-60 exclusion search is shipped and must survive these corrections.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
