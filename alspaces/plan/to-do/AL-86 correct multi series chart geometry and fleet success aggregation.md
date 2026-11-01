h3. Title: AL-86: Correct multi-series chart geometry and fleet success aggregation

*TARGET*: AgenticLens (Frontend Pure Aggregation / Comparisons / Trend Chart / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-10, FE-11
*Related*: None

*CONTEXT*

Positive bar series share identical geometry and obscure each other. Overall fleet success averages ticket percentages, reporting 50 percent for 99 successes in 100 calls across unequal tickets.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to correct multi-series chart geometry and fleet success aggregation, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Frontend Pure Aggregation / Comparisons / Trend Chart / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Define grouped or stacked positive series with a matching scale; verify geometry for multiple positives, mixed signs, unknown gaps and the shipped token-distribution configuration.
* Compute overall success from selected total calls/errors in a pure function, preserving search/type/prefix/skill scope and explicit zero-call behavior. Verify unequal denominators.
* Address FE-D2 benchmark aggregation ownership as part of the fix, preserving configured scope and unknown-cost semantics. Do not introduce a charting framework rewrite.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
