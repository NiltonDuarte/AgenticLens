h3. Title: AL-71: Make frontend requests and live data handling reliable

*TARGET*: AgenticLens (Frontend Fleet Comparisons Progression / Live REST SSE Polling)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: FE-02, FE-03, FE-04, FE-05, FE-D2
*Related*: AL-69 owns source diagnostics and anomaly semantics; AL-70 owns configured ticket parsing/filter identity; AL-72 owns rendered UI behavior and benchmark arithmetic.

*CONTEXT*

The current selection must own asynchronous results, time windows must have one meaning, and malformed or recovering live transports must preserve valid data and truthful status.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Frontend Fleet Comparisons Progression / Live REST SSE Polling, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

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

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
