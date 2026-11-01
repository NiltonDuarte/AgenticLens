h3. Title: AL-72: Apply anomaly thresholds consistently through persistence and views

*TARGET*: AgenticLens (Domain / CLI / Persistence / Fleet and Live Projections / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: CC-03, CC-04
*Related*: None

*CONTEXT*

Custom anomaly thresholds disappear during persistence and fleet recomputation; ErrorRateThreshold is serialized but ignored in favor of fixed levels.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to apply anomaly thresholds consistently through persistence and views, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Domain / CLI / Persistence / Fleet and Live Projections / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Define the configured error-rate threshold and warning/critical semantics in the contract; apply the effective policy to event detection and aggregate fallback, and display truthful thresholds.
* Prove generate/save/reload/detail/fleet and live parity for settings above and below defaults, including exact boundaries and event pruning. Do not silently replace an already configured projection with default detection.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
