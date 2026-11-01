h3. Title: AL-76: Isolate backend tests from checkout and unrelated listeners

*TARGET*: AgenticLens (CLI HTTP Smoke Tests / Configuration Tests / Git Tests)

*Status*: Ready
*Work type*: test
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-12, ADP-08
*Related*: None

*CONTEXT*

Tests depend on live repository configuration/branch inventory. A fixed-port server smoke test can pass against an unrelated HTTP service, and readiness requests can exceed their stated timeout.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to isolate backend tests from checkout and unrelated listeners, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected CLI HTTP Smoke Tests / Configuration Tests / Git Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Use self-contained configuration and temporary real Git repositories with explicit branch inventories; no dependency on this checkout policy or assumed absent branches.
* Use isolated listeners and verify the launched process and expected fixture identity. Detect early child exit; give readiness requests and cleanup bounded lifetimes.
* Exercise an occupied-port/foreign-service scenario and an accepting nonresponding listener. Prove parser and branch-precedence regressions remain detectable.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
