# AL-57 iteration 03 migration implementation

Baseline: ee7558749c7101ddc4134cbdd22961bcba0a5a09. Writable implementation ownership: backend/internal/domain/migration only. Parent retains integration, telemetry, commits, and PR delivery. Initial checkout already contained parent-owned SUMMARY.md and iteration 03 planning changes; these were preserved.

Replaced decodeCostPresence with concrete schema-v1-to-v4 and schema-v5 ledger inputs, known session/invocation/task fields, and direct domain assembly. Shared schema fields have version-neutral names; version dispatch selects legacy compatibility or required v5 evidence validation. There is no recursive JSON walker, intermediate reserialization, arbitrary object classification, reflection, or byte-kind matching.

Named memberPresence probes preserve root/session mixed-shape rejection even for explicit null collections, and reject removed canonical invocation id members including null. RawMessage is confined to the declared nullable cost_usd member, where it distinguishes required-v5 omission from allowed null before nullable numeric decoding. Domain.ValidateCostEvidence validates amount/evidence combinations before observed amount or partial known subtotal enters the domain accumulator.

Legacy numeric zero/nonzero remain observed; missing/null remain unavailable. Legacy evidence with an empty state retains prior migration inference. v5 empty/missing/null evidence is invalid. Flat owner metadata, exact row occurrences, canonical session metadata, task metadata, and optional collection nil/empty distinctions are assembled directly. Existing source normalization, event sanitation, migration state, dropped-tool diagnostics, and step invariants remain in their owning code.

Migration intentionally retains its existing acceptance of unrelated/retired unknown fields. Runtime adapter schema strictness is separate. Unknown metadata containing invocation-like fields is not interpreted as a cost record. Malformed/trailing JSON remains rejected by the actual schema decoder/version probe.

Extended the existing TestCostEvidenceMigrationPreservesObservedHistoryAndMissingPresence with task amount presence across legacy schema families, v5 canonical/flat/task valid evidence and contradiction cases, and mixed-null collection guards. Added trailing JSON to TestDetectVersion. Existing ownership, occurrence, legacy cost, event sanitation, and invariant assertions were retained.

Verification against the current migration diff plus the shared domain validator, with embedded assets already present:

- backend: go test ./internal/domain/migration ./internal/domain: passed.
- backend: go test -race ./internal/domain/migration ./internal/domain: passed.
- backend: go vet ./internal/domain/migration ./internal/domain: passed.
- backend: go test -race ./internal/adapters/backfill: passed (includes migration runner and recovery integration).
- backend: go vet ./internal/adapters/backfill: passed.
- git diff --check -- backend/internal/domain/migration: passed.

These are focused migration checks. Integrated adapter/HTTP/CLI checks and the final make gate remain parent/tester-owned after adapter implementation is complete. Migration test-file ownership passes to the tester at this handoff; migration production ownership remains with this implementation agent for review corrections.
