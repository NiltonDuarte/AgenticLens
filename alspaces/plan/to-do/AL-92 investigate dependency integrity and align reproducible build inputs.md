h3. Title: AL-92: Investigate dependency integrity and align reproducible build inputs

*TARGET*: AgenticLens (Dependency Maintenance / Lockfile / Docker / CI Toolchain)

*Status*: Ready
*Work type*: chore
*Priority*: P3
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: Bounded dependency/toolchain/reproducibility observations; not a confirmed exploit or performance defect.
*Related*: None

*CONTEXT*

Bounded audit observations found lockfile entries without resolved/integrity metadata, npm advisory output, a large bundle warning, CI Go setup inconsistent with go.mod, and a Docker npm-ci fallback that can conceal lockfile drift. These are signals, not proven exploitable vulnerabilities or measured performance defects.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to investigate dependency integrity and align reproducible build inputs, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Dependency Maintenance / Lockfile / Docker / CI Toolchain consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Inspect current registry/audit evidence and dependency reachability before proposing updates; distinguish production exposure, development-only risk, tooling metadata and unavailable evidence. Preserve the historical audit counts as observations, not current vulnerability claims.
* Regenerate or explain incomplete lockfile metadata using the supported package manager and prove clean reproducible install. Align CI compiler setup/cache labels with the module toolchain and fail Docker builds on lockfile mismatch rather than silently installing a changed graph.
* Measure bundle composition/load impact before proposing optimization; record no-action conclusions when justified. Do not blanket-upgrade dependencies or promise a performance benefit without evidence.
* Coordinate package/lockfile edits with AL-88 and verify build, tests and embedded assets for accepted changes.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 REVIEW.md](../../REV-5/iterations/03/REVIEW.md)
