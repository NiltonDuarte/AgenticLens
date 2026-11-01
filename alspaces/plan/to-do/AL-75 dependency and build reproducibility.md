h3. Title: AL-75: Verify dependency integrity and reproducible builds

*TARGET*: AgenticLens (Dependency Maintenance / Lockfile / Docker / CI Toolchain)

*Status*: Ready
*Work type*: chore
*Priority*: P3
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: Bounded dependency/toolchain/reproducibility observations; not a confirmed exploit or performance defect.
*Related*: Coordinate package/lockfile cleanup with AL-74 and release packaging with AL-73; do not duplicate installer integrity verification.

*CONTEXT*

Investigate bounded integrity/advisory/bundle signals and align build inputs with declared toolchains. Historical observations are not proof of a current exploit or measured performance defect.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Dependency Maintenance / Lockfile / Docker / CI Toolchain, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Investigate dependency integrity and align reproducible build inputs

Bounded audit observations found lockfile entries without resolved/integrity metadata, npm advisory output, a large bundle warning, CI Go setup inconsistent with go.mod, and a Docker npm-ci fallback that can conceal lockfile drift. These are signals, not proven exploitable vulnerabilities or measured performance defects.

* Inspect current registry/audit evidence and dependency reachability before proposing updates; distinguish production exposure, development-only risk, tooling metadata and unavailable evidence. Preserve the historical audit counts as observations, not current vulnerability claims.
* Regenerate or explain incomplete lockfile metadata using the supported package manager and prove clean reproducible install. Align CI compiler setup/cache labels with the module toolchain and fail Docker builds on lockfile mismatch rather than silently installing a changed graph.
* Measure bundle composition/load impact before proposing optimization; record no-action conclusions when justified. Do not blanket-upgrade dependencies or promise a performance benefit without evidence.
* Coordinate package/lockfile edits with AL-74 and verify build, tests and embedded assets for accepted changes.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
* [REV-5 REVIEW.md](../../REV-5/iterations/03/REVIEW.md)
