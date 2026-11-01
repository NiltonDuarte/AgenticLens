h3. Title: AL-77: Repair adapter test oracles and remove test-only database accessor

*TARGET*: AgenticLens (OpenCode Schema Detection Tests / Live SQLite / Tests)

*Status*: Ready
*Work type*: test
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: ADP-07, ADP-09, ADP-13
*Related*: None

*CONTEXT*

Required-column omission tests reject an already mixed schema, no-write tests inspect absent or unrelated directories, and an exported production DB accessor exists only for test introspection.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to repair adapter test oracles and remove test-only database accessor, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected OpenCode Schema Detection Tests / Live SQLite / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Build each required-column omission from one valid schema family, assert the missing-structure cause, and keep mixed-family rejection separate. Disabling required-column checks in a disposable mutation must fail the repaired test.
* Observe the exact configured populated workspace for no-write checks, fail traversal errors, and compare bytes. Preserve trace/hydration assertions when consolidating into existing effective tests; a write mutation must be detected.
* Remove the exported test-only DB accessor while retaining read-only enforcement proof through a package-internal or real boundary test. Compile all consumers and avoid replacing it with another unnecessary production hook.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
