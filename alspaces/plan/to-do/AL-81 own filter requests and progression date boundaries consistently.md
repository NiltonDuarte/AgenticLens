h3. Title: AL-81: Own filter requests and progression date boundaries consistently

*TARGET*: AgenticLens (Frontend Fleet Comparisons Progression / API Boundary / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: FE-02, FE-03, FE-D2
*Related*: None

*CONTEXT*

Late filter responses overwrite current selections. Progression presets switch from latest-ledger time to server-current time when a type/prefix is selected. Duplicated state ownership lets the views diverge.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to own filter requests and progression date boundaries consistently, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Frontend Fleet Comparisons Progression / API Boundary / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Fence data, diagnostics, loading and errors by the active request, including stale rejections after newer successes. Preserve existing useLiveTrace and TicketDeepDive ownership semantics.
* Choose one documented progression time-window meaning and use it with and without filters; test a clock far beyond the newest ledger while toggling a type/prefix that includes all shown tickets.
* Extract only shared request/filter/date invariants that remove concrete duplication, and keep pure calculations out of views. Do not add a generic caching framework.
* Use deferred A/B response and rejection permutations and a real server-date contract fixture, not only immediately resolved mocked requests. FE-D2 benchmark arithmetic belongs to AL-86.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
