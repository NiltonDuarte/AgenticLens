h3. Title: AL-89: Verify installer artifacts and preserve failed update outcomes

*TARGET*: AgenticLens (Installer Wrapper / Downloader / Packaged Setup / Tests)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit, follow-up allocation in iteration 09.
*Findings*: D1, D2, D6
*Related*: AL-68: coordinate installation entrypoints and preserve lifecycle-hook assets when either feature lands first.

*CONTEXT*

Failed downloads can report successful update and advance version state; downloaders do not verify published checksums and execute arbitrary returned bytes. Duplicate shipped copies require parallel fixes.

The linked audit contains affected paths, triggers, evidence and limits at its frozen baseline. Revalidate those paths before implementation; ticket creation does not claim a new reproduction or a completed fix.

*GOAL*

As an AgenticLens maintainer, I want to verify installer artifacts and preserve failed update outcomes, with observable behavior and regression evidence at the owning boundary.

*VERTICAL SLICE SCOPE*

* Deliver the correction through the affected Installer Wrapper / Downloader / Packaged Setup / Tests consumers and their existing user-facing or verification boundary, including meaningful regression coverage. Avoid a schema-only, helper-only or documentation-only fix for a behavioral defect.
* Keep pure transformations in the functional core and I/O in the shell. Limit shared abstractions to demonstrated duplication. Preserve persisted facts, source identity and baseline behavior when optional enrichment is unavailable.

*ACCEPTANCE CRITERIA*

* Validate the selected release artifact against published SHA256SUMS before execution/replacement; preserve the existing binary on empty, wrong or mismatched artifacts and verification failure. Pin wrapper/artifact provenance to the selected release.
* Return truthful total/partial failure and advance state only for successful replacement; verify total failure, partial failure and atomic successful replacement.
* Use one canonical packaged implementation or deterministic equivalence checks while preserving both legitimate installation destinations and differing defaults. Address D6 as part of shipping D1/D2 consistently.
* Use isolated local HTTP/hash stubs, never execute unverified fetched content in tests. Preserve release-only installation and report publisher-trust limits without claiming checksum validation authenticates the publisher.
* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with the approved concrete contract. This planning session does not modify application/test/runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md; strict typing, zero any/interface{}, zero emojis and no domain mocks. Pass affected tests, independent focused review and required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Record reviewed revision, behavioral evidence, reused checks and remaining environment limits. Update the finding disposition only after the correction is verified; publish implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
