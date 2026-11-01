h3. Title: AL-75: Remove discarded core work and preserve migration input values

*TARGET*: AgenticLens (Domain Migration / Aggregation / Tests)

*Status*: Ready
*Work type*: refactor
*Priority*: P3
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-09, CC-10, CC-11
*Related*: None

*CONTEXT*

V1 migration mutates its input through a shared session slice. Core aggregation builds unread mapped-event structures and duplicates model grouping; some tests construct and discard their purported sorting fixtures.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to remove discarded core work and preserve migration input values, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Domain Migration / Aggregation / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Clone mutable migration containers before transformation; prove input nonmutation and alias independence for nil/non-nil events and appendable metadata.
* Remove unused mappedEvents and anomalyEvidenceSelection.events work and reuse existing model aggregation where behavior is equivalent. Preserve role order, thread ceilings, precise and incomplete-cost behavior; claim no unmeasured speedup.
* Repair the actual sorting oracle or remove its unsupported nil-turn claim, preserve meaningful tool-name ordering, remove overwritten parity setup, and consolidate completion-date duplicates only with explicit assertion recipients.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
