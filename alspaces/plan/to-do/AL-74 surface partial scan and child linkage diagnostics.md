h3. Title: AL-74: Surface partial scan and child-linkage diagnostics

*TARGET*: AgenticLens (CLI / Live Adapter / HTTP DTO / Diagnostics / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-08, ADP-06
*Related*: AL-62 and AL-63: reuse ownership and capability contracts; no dependency on a broad schema rewrite.

*CONTEXT*

CLI scan hides malformed siblings and rejected targets; live child lookup failures look like a successful absence of children.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to surface partial scan and child-linkage diagnostics, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected CLI / Live Adapter / HTTP DTO / Diagnostics / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Present per-ticket rejection and fatal scan diagnostics while retaining valid siblings under the existing partial-scan policy. Distinguish rejected targets from nonexistent tickets.
* Keep supported live messages/parts visible when optional child lookup fails, with a typed completeness diagnostic reaching HTTP/CLI consumers.
* Test corrupt/unsupported/unreadable scan siblings and failed, successful-empty and successful-linked child lookup. Do not turn optional enrichment failure into an empty or wholly failed trace.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
