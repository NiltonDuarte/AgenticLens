h3. Title: AL-82: Validate live payloads and recover polling state

*TARGET*: AgenticLens (Frontend Live API REST SSE Polling / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-04, FE-05
*Related*: None

*CONTEXT*

Malformed nested live payloads pass unchecked casts and crash consumers; fallback polling remains in error after successful recovery and launches overlapping requests.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to validate live payloads and recover polling state, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Frontend Live API REST SSE Polling / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Validate required nested nodes/messages/parts/anomalies at REST and SSE ingress using shared decoders. Normalize only truly optional collections and retain the last valid snapshot on malformed input.
* Restore fallback-polling status on success after error, serialize/coalesce pending polls, and prevent obsolete or post-teardown results from mutating state.
* Test null/primitive/wrong-scalar/nested-malformed payloads, valid optional omissions, failure-threshold recovery, deferred overlapping requests and teardown. Do not misrepresent malformed data as a valid empty snapshot.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
