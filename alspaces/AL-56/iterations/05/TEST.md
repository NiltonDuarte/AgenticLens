# AL-56 Iteration 05 Test Report

## Automated checks

- `make` passed: Go vet, frontend typecheck and lint, all Go race tests, all 719 frontend tests, frontend production build, and multi-platform CLI builds.
- `git diff --check` passed.
- Frontend dependency installation reported six npm audit advisories: three moderate, one high, and two critical.

## Migration and ledger audit

- The full test run includes regression coverage for v1-v3 migration, interim-v4 normalization, mixed canonical/legacy shape rejection, atomic failure, and repeated migration.
- CLI and migration-runner tests verify that removed unscoped tool totals are reported with their count and source-log recovery guidance.
- Audited all 84 tracked `agentic_metrics.json` ledgers. All parse as schema v4, contain only root-level `sessions` for telemetry facts, and contain unique nonblank session IDs (545 sessions total).
- All 545 prior invocations are preserved. The only invocation-field changes are the expected v4 `user_iterations` values; every other compared invocation field is unchanged.
- Session-attributed tool calls and errors are preserved. The migration removed 843 unscoped tool aggregate rows across 66 ledgers and recorded the corresponding drop count and recovery guidance in each affected ledger.
- The existing v3 event-persistence allowlist pruned 12 ineligible routine events from `testdata/workspaces/AL-2`; retained event metadata was preserved.
- No persisted session event contains nonempty prompt, reasoning, tool-input, tool-output, or error-message payloads.

## Live Codex smoke check

Not run in the tester phase. The repository workflow requires telemetry ingestion at the final iteration step; a live Codex generation would ingest the active session before implementation, review, and delivery work is complete. The Codex adapter and session-refresh behavior are covered by the fixture and regression suites included in `make`.

## Reviewer-fix retest

- After reviewer fixes, `make` passed again: Go vet, frontend typecheck/lint, all Go race tests, all 719 frontend tests, Vite production build, and multi-platform builds.
- Focused regression commands passed:
  - `go test ./internal/adapters -run '^TestSchemaMigrationDTOBoundaries$'`
  - `go test ./internal/adapters/backfill -run '^TestMigrationRunner_(NormalizesInterimV4AndReportsDroppedUnscopedTools|RejectsMixedCanonicalAndLegacyCollectionsAtomically|SanitizesCanonicalV4EventPayloads)$'`
  - `go test ./internal/domain/migration -run '^TestMigrationEngineNormalizesInterimV4AndDropsUnscopedTools$'`
- Verified the v4 migration fast path sanitizes canonical event payloads, the current DTO rejects `error_message`, mixed session/root fact shapes fail atomically, and duplicate tool names in one session fail validation.
- `git diff --check` passed after the retest.
- Dependency installation again reported six npm audit advisories: three moderate, one high, and two critical.

## Final migration-runner verification

- Current schema-v4 fixtures in `backend/internal/adapters/repository_test.go` and `backend/internal/adapters/http/server_test.go` use the required `sessions` collection and nested invocation/tool/event arrays. Strict validation remains enabled.
- Final `make` passed after fixture, migration-status, and canonical-validation fixes: Go vet, frontend typecheck/lint, all Go race tests, all 719 frontend tests, frontend production build, and multi-platform CLI builds.
- Focused boundary regressions passed:
  - `go test ./internal/adapters/backfill -run '^TestMigrationRunner_(ValidatesCanonicalV4BeforeUpToDateFastPath|ValidCanonicalV4IsUpToDate|SanitizesCanonicalV4EventPayloads)$'`
  - `go test ./internal/adapters -run '^TestSchemaMigrationDTOBoundaries$'`
  - `go test ./internal/domain/migration -run '^TestMigrationEngineNormalizesInterimV4AndDropsUnscopedTools$'`
- The migration-runner fast-path tests verify duplicate session IDs/tool rows, invalid tool counts, dangling and whitespace-padded session/invocation parents, padded IDs, missing required collections, and valid canonical-v4 up-to-date handling. Invalid ledgers are checked byte-for-byte for no modification; the valid ledger is checked for no modification. Event sanitization and strict `error_message` rejection continue to pass.
- Migration metadata tests verify removed unscoped rows leave status `completed` when no steps are pending and retain `partial` only when pending steps exist.
- `git diff --check` passed after the final changes.

## Final retest after canonical-v4 validation updates

- `git diff --check && make` passed. The full Go race suite passed, including repository scanning and HTTP fixtures under strict schema validation; all 719 frontend tests passed; frontend typecheck/lint/build and multi-platform CLI builds passed. Dependency installation reported six npm audit advisories (three moderate, one high, two critical).
- Canonical-v4 fast-path regressions passed in `go test -race ./...`: required session and nested collections, padded session/parent IDs, duplicate session IDs, duplicate tool rows, invalid counts, and dangling parents are rejected without modifying the original bytes. The valid canonical-v4 fixture remains up-to-date and unchanged.
- Event sanitization and rejection of `error_message` remain covered by passing migration-runner and DTO tests in the full suite.
- Migration metadata uses `completed` when `steps_pending` is empty and `partial` when pending steps remain; the v3-to-v4 step preserves this distinction. Existing migration tests cover pending-step preservation and completed status.
- Final `git diff --check` passed.
