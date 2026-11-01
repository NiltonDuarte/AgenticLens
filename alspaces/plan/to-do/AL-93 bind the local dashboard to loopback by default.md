h3. Title: AL-93: Bind the local dashboard to loopback by default

*TARGET*: AgenticLens (HTTP Listener / CLI / Embedded Dashboard / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: ADP-05
*Related*: None

*CONTEXT*

The dashboard advertises localhost but binds all interfaces by default. The user explicitly defines AgenticLens as local-only and deliberate remote exposure as operator-owned.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to bind the local dashboard to loopback by default, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected HTTP Listener / CLI / Embedded Dashboard / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Bind loopback by default and keep the advertised URL consistent with the listener.
* Prove standalone embedded UI and API access still work locally and assert the actual listener address.
* Do not add authentication, a new remote-deployment mode or mandatory remote-host configuration. Preserve ADP-05 at P2 default hardening under the user's accepted scope.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
