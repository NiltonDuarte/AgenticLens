# AL-73: Verify release artifacts before installation

## Goal

Installers replace binaries only with verified release artifacts and report failures accurately; release verification rejects incomplete checksum inventories.

## Context

Only verified artifacts should replace installed binaries, failures must retain truthful update state, and release manifests must cover every required artifact.

## Acceptance criteria

### Verify installer artifacts and preserve failed update outcomes

Failed downloads can report successful update and advance version state; downloaders do not verify published checksums and execute arbitrary returned bytes. Duplicate shipped copies require parallel fixes.

* Validate the selected release artifact against published SHA256SUMS before execution/replacement; preserve the existing binary on empty, wrong or mismatched artifacts and verification failure. Pin wrapper/artifact provenance to the selected release.
* Return truthful total/partial failure and advance state only for successful replacement; verify total failure, partial failure and atomic successful replacement.
* Use one canonical packaged implementation or deterministic equivalence checks while preserving both legitimate installation destinations and differing defaults. Address D6 as part of shipping D1/D2 consistently.
* Use isolated local HTTP/hash stubs, never execute unverified fetched content in tests. Preserve release-only installation and report publisher-trust limits without claiming checksum validation authenticates the publisher.

### Reject incomplete release checksum manifests

A hashing failure hidden by a pipeline can publish a partial SHA256SUMS that release-verify still accepts.

* Propagate hashing failures and validate exact required artifact coverage before committing/publishing a manifest. Detect missing, duplicate, unexpected or unreadable entries according to the release inventory.
* Use a fault-injected hashing command that emits one valid entry then fails; packaging and independent verification must reject partial coverage.
* Preserve successful release packaging and host verification. No hosted release publication or tag mutation is needed to demonstrate this boundary.

## Implementation notes

Related work: Coordinate with AL-68 hook installation so canonical packaging and updates retain hook assets regardless of landing order. AL-75 owns compiler and dependency reproducibility.

Finding references: D1, D2, D5, D6

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain installer/checksum failure acceptance. FIX-16 and GEN-19 repaired release-candidate workflow behavior; that delivery does not establish installer artifact validation or complete checksum-manifest rejection.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
