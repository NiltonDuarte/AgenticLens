h3. Title: AL-73: Align ticket identity parsing and stable sorting across layers

*TARGET*: AgenticLens (Domain / CLI and API / Frontend Formatters and Filters / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-06, FE-01, CC-07
*Related*: None

*CONTEXT*

AL2 IDs are parsed as AL and sorted lexically in Go and TypeScript. Independently, descending equal-value comparisons violate strict ordering and repeatedly reorder tied rows.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to align ticket identity parsing and stable sorting across layers, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Domain / CLI and API / Frontend Formatters and Filters / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Agree a configured-prefix grammar across normalization, prefix extraction, filtering and natural ordering. Preserve documented legacy delimiterless formats without hardcoding AL or AL2.
* Test AL versus AL2, AL2-2 before AL2-10, custom digit/delimiter prefixes, case and leading zeroes through both language boundaries and consumer views.
* Ensure all affected descending comparisons treat equality correctly and follow a documented stable tie policy; repeated sorting must be idempotent for updated time, duration, invocation and identity ties. Preserve the valid cost/completion comparators.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
