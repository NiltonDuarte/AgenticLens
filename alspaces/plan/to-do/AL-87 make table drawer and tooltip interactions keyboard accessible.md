h3. Title: AL-87: Make table, drawer and tooltip interactions keyboard accessible

*TARGET*: AgenticLens (Frontend Tables / Anomaly Drawer / Tooltip CSS / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-13, FE-14, FE-15
*Related*: None

*CONTEXT*

Table sorting/navigation is mouse-only, the modal drawer does not own focus, and hover CSS overrides Escape tooltip dismissal.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to make table, drawer and tooltip interactions keyboard accessible, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Frontend Tables / Anomaly Drawer / Tooltip CSS / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Use keyboard-operable sorting/navigation with appropriate aria-sort and preserve tooltip propagation; cover TicketTable, comparison and landing headers.
* Implement drawer initial focus, containment/background behavior, Escape and focus return using a bounded native/existing mechanism.
* Give tooltip visibility one owner, preserve keyboard focus on Escape and keep it dismissed while the pointer remains hovered. Verify actual computed visibility with stylesheet-aware or browser evidence; class-only assertions are insufficient.
* Test keyboard flows and real visibility rather than role/label-only evidence, satisfying the tooltip portion of FE-D3.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
