# AL-58 Iteration 03 Delta Plan: Remove Redundant Schema Queries and Speed Up Detection Tests

## Goal and boundary

Address the two low-priority findings from the AL-58 architecture and test-performance review. Remove v1 column checks that duplicate the schema-family contract, and reduce avoidable SQLite file I/O in synthetic schema rejection tests. Preserve the frozen contract and the shared/v1/v2 ownership established by iterations 01 and 02. This iteration changes only the OpenCode adapter implementation and its schema-family detection tests; it does not change extraction results, accepted schemas, diagnostics, fixtures, docs, CLI behavior, or runtime evidence claims.

## Grounding

- `backend/internal/adapters/opencode/schema_family_detection.go` requires `time_updated` in both the v1 `message` and `part` tables before a database is classified as v1.
- `backend/internal/adapters/opencode/schema_v1_session_extraction.go` nevertheless calls `schemaV1TableHasColumn` for those same columns and falls back to a constant timestamp expression. The strict detector runs before this reader, so supported v1 layouts always have the columns.
- `backend/internal/adapters/opencode/schema_family_detection_test.go` creates a file-backed SQLite database for each required-column omission case. The review profile measured `TestSchemaFamilyRejectsEveryRequiredTableColumnOmission` at 0.910 seconds and the package at 1.023 seconds. Its subtests use independent database handles and local schemas.
- The contract requires complete required-column validation and rejection of partial schemas; it does not permit treating a missing required column as an optional legacy layout. Read-only behavior remains unchanged.

## Planned changes

### V1 extraction

1. Remove `schemaV1TableHasColumn` and its two `PRAGMA table_info` queries from `schema_v1_session_extraction.go`.
2. Build the v1 message and part queries directly from `m.time_updated` and `p.time_updated`, which are guaranteed by schema-family detection.
3. Preserve the existing schema-family validation order, query behavior for supported databases, and contextual error behavior. Do not relax required columns or add a compatibility fallback.

### Schema detection tests

1. Replace file-backed temporary databases in synthetic schema rejection cases in `schema_family_detection_test.go` with an isolated in-memory SQLite database per subtest.
2. Mark independent subtests parallel where safe. Keep all schema setup local to each subtest and retain every existing table/column omission case and expected diagnostic assertion.
3. Retain disk-backed coverage where it is specifically needed to prove the production read-only file-opening boundary; do not change `TestOpenDBRemainsReadOnly` or unrelated fixtures.

## Verification

- Run focused, uncached adapter tests: `go test -count=1 ./internal/adapters/opencode`.
- Run the adapter race suite: `go test -race ./internal/adapters/opencode`.
- Run `make test` to confirm backend and frontend suites remain green.
- Profile with `go test -json ./internal/adapters/opencode` and summarize test/package elapsed times. Compare against the review baseline (0.910 seconds for the omission matrix; 1.023 seconds for the package) and confirm the matrix no longer creates on-disk databases. Report the observed timing; do not add a brittle timing assertion.
- Run `git diff --check`.

## Acceptance criteria

1. `schemaV1TableHasColumn` and both redundant metadata queries are removed; supported v1 extraction directly selects the validated `time_updated` columns.
2. Complete required-column validation and rejection of partial/mixed schemas remain unchanged; all current rejection assertions and omissions are retained.
3. Synthetic schema rejection tests use independent in-memory databases and may execute in parallel without sharing mutable state or relying on timing.
4. The read-only file-opening test and unrelated schema/runtime behavior remain unchanged.
5. Focused tests, race tests, `make test`, profiling, and diff checks pass; profiling results document the measured improvement without enforcing an absolute duration.
6. Changes remain within the v1 extraction reader and schema-family detection tests, plus this iteration's review artifact; no contract, evidence, compatibility claim, domain, frontend, or CLI changes are introduced.

## Handoff

This is a zero-behavior-change cleanup and test-performance iteration. Implementation begins only after the user approves this plan.
