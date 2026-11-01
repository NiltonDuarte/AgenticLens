h3. Title: AL-90: Decouple compatibility checks from ticket artifacts and fixed ports

*TARGET*: AgenticLens (OpenCode Compatibility Script / Stable Test Fixtures / Make / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: D4
*Related*: None

*CONTEXT*

Maintained compatibility targets depend on AL-58 historical evidence and fixed ports; a foreign listener or exhausted readiness loop can falsely satisfy setup.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to decouple compatibility checks from ticket artifacts and fixed ports, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected OpenCode Compatibility Script / Stable Test Fixtures / Make / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Move maintained provider/config fixtures to a stable owned test location without altering historical AL-58 evidence. Relocating archival ticket workspaces must not break the checks.
* Allocate/configure isolated ports, verify launched child and expected protocol identity, fail readiness exhaustion before launching the client, and clean up processes.
* Exercise missing provider, occupied port and stalled startup with local fixtures; run supported real OpenCode compatibility environments when available and label unavailable environments explicitly.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
