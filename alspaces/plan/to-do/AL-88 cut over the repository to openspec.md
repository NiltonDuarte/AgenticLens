# AL-88: Cut over the repository to OpenSpec

## Goal

AgenticLens itself uses OpenSpec for future work and its active backlog, while all historical telemetry remains accessible without a second new per-ticket workspace.

## Context

Platform and guidance support alone leave the live repository using alspaces/plan conventions. The final cutover must reconcile shared configuration, existing work and contributor entrypoints together. This is the adoption outcome of [PO-24](../../PO-24/SPEC.md), not permission to rewrite old session facts.

## Acceptance criteria

- Initialize and version the approved OpenSpec structure/schema and project guidance using the verified supported version. Select the new active/archive workspace roots while retaining an explicit legacy root for existing alspaces records. Bound each root appropriately; do not let nested active/archive roots or legacy plan folders duplicate tickets.
- Move remaining active roadmap specifications into ticket-prefixed OpenSpec changes. Preserve IDs, acceptance, readiness, dependencies and active-development ownership. Maintain a single editable backlog and a navigable dependency/readiness view; do not infer approval from OpenSpec artifact completion. Coordinate active ticket owners before relocating their writable artifacts.
- Update live contributor/setup entrypoints and relevant links to the new structure. Existing public/historical links retain a navigable route to their evidence. Preserve completed specifications, historical audits and ledgers; no bulk copy into duplicate ledgers, silent reticketing or unsupported schema conversion.
- Future PO/GEN/development work creates one OpenSpec change home with the established ticket convention. No new alspaces ticket workspace or independent plan index is required after cutover. Legacy-only adopting projects retain their supported workflow.
- Demonstrate creating and completing a representative change, source-backed publication, native dashboard/CLI detail, archival and late refresh using the repository's selected roots. Verify copied historical facts remain readable with original source stores unavailable.
- Provide a reviewed migration inventory and recoverable cutover. Failure leaves the previous supported configuration and facts usable; partial completion is visible and resumable. Confirm no root-relative path or guidance still recreates the retired workspace convention.

## Implementation notes

Do not adopt AL2-2's unimplemented compressed archive format as a prerequisite: OpenSpec archive folders are ordinary directories. The migration inventory covers active work separately from immutable historical evidence. Existing in-progress owners and unrelated ticket IDs must not be renumbered or overwritten during cutover.
