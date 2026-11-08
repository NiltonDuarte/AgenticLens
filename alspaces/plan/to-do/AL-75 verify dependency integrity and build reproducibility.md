# AL-75: Verify dependency integrity and build reproducibility

## Goal

Maintainers can reproduce supported builds and decide dependency changes from current integrity, reachability and bundle evidence rather than historical warnings.

## Context

Investigate bounded integrity/advisory/bundle signals and align build inputs with declared toolchains. Historical observations are not proof of a current exploit or measured performance defect.

## Acceptance criteria

### Investigate dependency integrity and align reproducible build inputs

Bounded audit observations found lockfile entries without resolved/integrity metadata, npm advisory output, a large bundle warning, CI Go setup inconsistent with go.mod, and a Docker npm-ci fallback that can conceal lockfile drift. These are signals, not proven exploitable vulnerabilities or measured performance defects.

* Inspect current registry/audit evidence and dependency reachability before proposing updates; distinguish production exposure, development-only risk, tooling metadata and unavailable evidence. Preserve the historical audit counts as observations, not current vulnerability claims.
* Regenerate or explain incomplete lockfile metadata using the supported package manager and prove clean reproducible install. Align CI compiler setup/cache labels with the module toolchain and fail Docker builds on lockfile mismatch rather than silently installing a changed graph.
* Measure bundle composition/load impact before proposing optimization; record no-action conclusions when justified. Do not blanket-upgrade dependencies or promise a performance benefit without evidence.
* Coordinate package/lockfile edits with AL-74 and verify build, tests and embedded assets for accepted changes.

## Implementation notes

Related work: Coordinate package/lockfile cleanup with AL-74 and release packaging with AL-73; do not duplicate installer integrity verification.

Finding references: Bounded dependency/toolchain/reproducibility observations; not a confirmed exploit or performance defect.

* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
* [REV-5 REVIEW.md](../../REV-5/iterations/03/REVIEW.md)

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain evidence-led dependency and reproducibility investigation. REV-6 changed build/test ordering, not proof of lockfile integrity or dependency reachability. Historical audit counts are not current vulnerability evidence.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
