h3. Title: AL-80: Align live and offline source decoding fidelity

*TARGET*: AgenticLens (OpenCode V1 Decoding / Live Adapter / Persistence / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: ADP-12
*Related*: AL-69: execution classification contract. AL-67 remains a broader architecture review, not a prerequisite.

*CONTEXT*

Live raw tool output trims meaningful whitespace that offline v1 decoding preserves; duplicated interpretation owns inconsistent field/status semantics.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to align live and offline source decoding fidelity, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected OpenCode V1 Decoding / Live Adapter / Persistence / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Share only the actual version-specific decoded representation needed by live/offline consumers; keep I/O/cache ownership in adapters and explicit projections in consumers.
* Preserve transient raw strings verbatim and verify parity for whitespace, structured values, malformed/unknown states and normal records.
* Preserve ledger privacy: no raw-payload expansion in serialization. Coordinate status behavior with AL-69 without inventing a universal source abstraction.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
