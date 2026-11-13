# AL-84: Resolve workspaces across configured roots

## Goal

Operators can list and inspect the same ticket from explicit active, archive and historical roots without depending on its current folder name.

## Context

AgenticLens recursively scans one configured directory, while detail/artifact readers assume a direct child named after the ticket. OpenSpec moves `al-123-add-login` to `archive/2026-11-13-al-123-add-login`. A stable `AL-123` must continue to address that change. This is the read-side foundation of [PO-24](../../PO-24/SPEC.md).

## Acceptance criteria

- Support multiple explicit shared workspace roots with bounded immediate-child discovery for the new mode, retaining existing single-root recursive configuration and explicit-override behavior. Do not infer roots when policy is absent or accept contradictory configuration silently.
- List, ticket detail, artifacts, event/trace lookup, comparison and live/read-side workspace selection use the same resolved directory. A ledger-backed ticket moved between configured roots retains identity, totals and accessible artifacts; configured standard OpenSpec files are read in place.
- Existing ledger identity is authoritative for its supported recorded ticket. A ledger-free ticket-prefixed change can be located using the explicit creation/naming policy. Detect conflicting folder/ledger identity and multiple distinct changes for one ticket; report ambiguity rather than choosing a root, merging history or double-counting it.
- Overlapping roots and physical aliases cannot enumerate the same directory twice. Explicit archive containers are not tickets; unrelated nested spec folders are not recursively discovered in immediate-child mode.
- Invalid shared policy fails clearly. An unavailable configured root produces actionable partial-result diagnostics while valid sibling roots remain readable. Expose enough location/error context in the API/dashboard to understand the result; unavailable telemetry is not measured zero.
- Copy supported ledgers to another machine fixture without original harness sources or capture state and retain normal reads. Reject unsafe traversal/escaping paths without an authenticity requirement on valid history.

## Implementation notes

Dynamic schema/spec-glob discovery is not required: existing configured artifact paths suffice initially. AL2-2 separately owns compressed archive creation; this ticket handles ordinary directories, not package ingestion or its precedence policy. Shared identity/location decisions must be usable by publication without duplicating a second resolver. Apply the canonical engineering and portable-history rules linked in PO-24.
