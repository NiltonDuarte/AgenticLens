h3. Title: AL-83: Fix trace keyboard activation and disclosure state

*TARGET*: AgenticLens (Trace Waterfall / Live Session Tree / Trace Inspector / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-06, FE-07, FE-16, FE-D3
*Related*: None

*CONTEXT*

Parent key handlers intercept nested buttons; selected payloads cannot be hidden; Hide Reasoning changes only the label for short content.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to fix trace keyboard activation and disclosure state, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Trace Waterfall / Live Session Tree / Trace Inspector / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Give nested copy/payload/collapse controls native keyboard behavior without triggering parent selection; test Enter and Space and preserve child-session badge isolation.
* Separate selection-driven initial expansion from explicit disclosure state; selected payloads must hide and retain intended state through enrichment and controlled selection.
* Make reasoning disclosure match its label for short and long content; assert visible content, not merely label changes.
* Address FE-D3 composition-test overlap only with explicit assertion recipients preserving filters, denominator invariance, empty-to-recovered transitions and callback isolation. Other FE-D3 oracles remain owned by AL-84 and AL-87.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
