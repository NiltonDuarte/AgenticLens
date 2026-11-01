h3. Title: AL-73: Make installation and release verification reliable

*TARGET*: AgenticLens (Installer / Downloader / Packaged Setup / Release Packaging)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: D1, D2, D5, D6
*Related*: Coordinate with AL-68 hook installation so canonical packaging and updates retain hook assets regardless of landing order. AL-75 owns compiler and dependency reproducibility.

*CONTEXT*

Only verified artifacts should replace installed binaries, failures must retain truthful update state, and release manifests must cover every required artifact.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Installer / Downloader / Packaged Setup / Release Packaging, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Verify installer artifacts and preserve failed update outcomes

Failed downloads can report successful update and advance version state; downloaders do not verify published checksums and execute arbitrary returned bytes. Duplicate shipped copies require parallel fixes.

* Validate the selected release artifact against published SHA256SUMS before execution/replacement; preserve the existing binary on empty, wrong or mismatched artifacts and verification failure. Pin wrapper/artifact provenance to the selected release.
* Return truthful total/partial failure and advance state only for successful replacement; verify total failure, partial failure and atomic successful replacement.
* Use one canonical packaged implementation or deterministic equivalence checks while preserving both legitimate installation destinations and differing defaults. Address D6 as part of shipping D1/D2 consistently.
* Use isolated local HTTP/hash stubs, never execute unverified fetched content in tests. Preserve release-only installation and report publisher-trust limits without claiming checksum validation authenticates the publisher.

h4. Reject incomplete release checksum manifests

A hashing failure hidden by a pipeline can publish a partial SHA256SUMS that release-verify still accepts.

* Propagate hashing failures and validate exact required artifact coverage before committing/publishing a manifest. Detect missing, duplicate, unexpected or unreadable entries according to the release inventory.
* Use a fault-injected hashing command that emits one valid entry then fails; packaging and independent verification must reject partial coverage.
* Preserve successful release packaging and host verification. No hosted release publication or tag mutation is needed to demonstrate this boundary.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
