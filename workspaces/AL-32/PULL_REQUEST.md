# feat(domain,adapters,cli): telemetry event pruning, database message stripping, and schema v3 migration (AL-32)

## Overview
Upgrades the AgenticLens fact ledger schema to version 3, introducing pure event persistence pruning and stripping error message strings to keep repository fact ledgers compact (<5 KB) and leak-free.

Previously, persisting entire raw turn and tool histories resulted in ledgers ballooning into megabytes with thousands of routine turns and potential secret leakage embedded in tool error messages. This implementation strictly separates the in-memory analysis fact plane (full fidelity) from the persisted ledger fact plane (pruned fidelity):
1. In-memory metrics derivation computes all aggregations and anomalies across the full raw event history.
2. Events are pruned via `domain.PruneEventsForPersistence` prior to persistence, omitting routine turns and retaining only bare tool failures (`status: error`), compactions, and anomaly-associated events with all `error_message` fields cleared to `nil`.
3. Schema version is promoted to 3, supported by a new migration step `StepV2ToV3Sanitized` in the migration engine and invariant validation.
4. Historical workspace ledgers (36 workspaces) have been migrated to schema version 3, dropping file sizes from hundreds of kilobytes to compact, leak-free JSON payloads.

- Contract: [workspaces/AL-32/CONTRACT.md](workspaces/AL-32/CONTRACT.md)
- Telemetry: [workspaces/AL-32/SUMMARY.md](workspaces/AL-32/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- Promoted `CurrentSchemaVersion` from `2` to `3` and added `MigrationStepSanitizedEvents = "schema_v3_sanitized"`.
- Implemented `PruneEventsForPersistence(ledger AgenticFactLedger) []EventFact` in `events.go`:
  - Retains bare failed tool calls (`Status == "error"`) with `ToolName`, `Status`, `Turn`, `Timestamp`, `DurationMs`, `ContextTokens`, and `ErrorMessage: nil`.
  - Retains compaction events with `ErrorMessage: nil`.
  - Retains anomaly-associated events with `ErrorMessage: nil`.
  - Omits routine successful tool calls and completed turns.
  - Returns stably sorted events via `SortEvents`.
- Updated `detectToolHighFailureRates` in `anomalies.go` to evaluate `ledger.Tools` first, preventing pruned event slices from skewing volume or failure rates.
- Created `StepV2ToV3Sanitized` migration step in `domain/migration/step_v2_v3.go`.
- Registered `StepV2ToV3Sanitized` in migration engine registry.
- Updated `migration/validator.go` to assert target version 3 invariants: identity preservation, bit-identical derived totals and workflow efficiency, `len(after.Events) <= len(before.Events)`, and zero non-nil error messages.

### Adapters Layer (`backend/internal/adapters/`)
- Stripped regex sanitization (`sanitizeTelemetryText`, `emailPattern`) from OpenCode SQLite extractor in `adapters/opencode/extractor.go`, recording bare `ErrorMessage: nil` on tool failure events.
- Updated `adapters/backfill/engine.go` to apply `domain.PruneEventsForPersistence` before atomic file writes and set schema version to 3.
- Updated `adapters/backfill/migration_runner.go` to default target version to 3.

### CLI Application (`backend/cmd/agentic-lens/`)
- Updated `runGenerate` to maintain the strict in-memory calculation order: metric derivation and markdown generation execute on the full unpruned fact ledger, while `persistedFactLedger` applies `domain.PruneEventsForPersistence` and sets `SchemaVersion = 3` before `adapters.SaveFactLedger`.
- Defaulted `runMigrate` and `runBackfill` target schema version to `domain.CurrentSchemaVersion` (3).

### Verification & Compliance
- Added domain unit tests in `events_test.go` and `migration/engine_test.go` testing v2 -> v3 migration, chained v1 -> v3 migration, idempotency, and calculation parity.
- Added migration runner integration tests in `migration_runner_test.go` verifying batch v3 upgrade, zero `"error_message"` keys, and file size limits.
- Added CLI integration tests in `main_test.go` (`TestCLIGenerate_V3PrunedEventsAndZeroErrorMessage`) verifying generated JSON structure, pruned events, zero error messages, and summary markdown preservation.
- Migrated all 36 historical workspace ledgers to schema version 3.
- Verified zero `any` or empty `interface{}` across Go backend.
- Verified zero emojis across all code, tests, and artifacts.
